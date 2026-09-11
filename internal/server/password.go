package server

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
	"vineyard/internal/game"
)

const passwordIterations = 600000

// Credentials belong to a seat, never to the public Player or Room projection.
// PBKDF2-HMAC-SHA256 uses the OWASP work factor and a fresh 128-bit salt.
type PasswordRecord struct {
	Salt string
	Hash string
}

func validPassword(password string) bool {
	n := utf8.RuneCountInString(password)
	return utf8.ValidString(password) && n >= 8 && n <= 128
}

func makePassword(password string) PasswordRecord {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if err != nil {
		panic(err)
	} // Constant, valid KDF parameters.
	return PasswordRecord{hex.EncodeToString(salt), hex.EncodeToString(key)}
}

func (p PasswordRecord) matches(password string) bool {
	salt, e1 := hex.DecodeString(p.Salt)
	want, e2 := hex.DecodeString(p.Hash)
	if e1 != nil || e2 != nil || len(salt) != 16 || len(want) != 32 {
		return false
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	return err == nil && subtle.ConstantTimeCompare(key, want) == 1
}

type authAttempt struct {
	Count int
	Until time.Time
}

// Bound both expensive hash work and limiter memory. Forwarded headers are not
// trusted: callers cannot evade the limit by inventing a proxy IP address.
func (a *App) beginAuth(w http.ResponseWriter, r *http.Request, seat string) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	a.authMu.Lock()
	now := time.Now()
	for key, item := range a.authAttempts {
		if !now.Before(item.Until) {
			delete(a.authAttempts, key)
		}
	}
	allowed := len(a.authAttempts) < 4096
	for key, limit := range map[string]int{"ip:" + host: 60, "seat:" + seat: 8} {
		item := a.authAttempts[key]
		if item.Count >= limit {
			allowed = false
		}
		if allowed {
			if item.Count == 0 {
				item.Until = now.Add(time.Minute)
			}
			item.Count++
			a.authAttempts[key] = item
		}
	}
	a.authMu.Unlock()
	if allowed {
		select {
		case a.authSlots <- struct{}{}:
			return true
		default:
		}
	}
	w.Header().Set("Retry-After", "60")
	fail(w, 429, "尝试过于频繁，请稍后再试")
	return false
}

func (a *App) endAuth() { <-a.authSlots }

func (a *App) playerView(room *game.Room, id string) map[string]any {
	v := room.View(id)
	_, v["passwordSet"] = a.Store.Passwords[id]
	return v
}

// Legacy bearer tokens may already have been copied by the name-only version.
// They cannot establish ownership alone. Enrollment codes never enter HTTP,
// saves or public logs; only the local server console receives them at startup.
func (a *App) WriteEnrollmentCodes(w io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, room := range a.Store.Rooms {
		for _, p := range room.Players {
			if key := a.enrollmentKeys[p.ID]; key != "" {
				fmt.Fprintf(w, "旧座位验证码 %s / %s / %s（仅交给该玩家，重启后更换）\n", key, room.Code, p.Name)
			}
		}
	}
}

func (a *App) requirePassword(w http.ResponseWriter, room *game.Room, id string) bool {
	if _, ok := a.Store.Passwords[id]; ok {
		return true
	}
	reply(w, 403, map[string]any{
		"error":            "请先用服务器本机验证码为旧座位设置密码",
		"passwordRequired": true, "code": room.Code, "youId": id, "name": room.Player(id).Name,
	})
	return false
}

// Caller holds a.mu. Each seat has one active browser session. A successful
// login/password change invalidates old tokens, including already-open SSE.
func (a *App) replaceSession(room *game.Room, id string) (string, map[string]Session) {
	old := map[string]Session{}
	for t, s := range a.Store.Sessions {
		if s.Code == room.Code && s.PlayerID == id {
			old[t] = s
			delete(a.Store.Sessions, t)
		}
	}
	t := game.NewID()
	a.Store.Sessions[t] = Session{Code: room.Code, PlayerID: id}
	return t, old
}

func (a *App) restoreSessions(t string, old map[string]Session) {
	delete(a.Store.Sessions, t)
	for key, value := range old {
		a.Store.Sessions[key] = value
	}
}

func (a *App) password(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Password        string `json:"password"`
		CurrentPassword string `json:"currentPassword"`
		EnrollmentCode  string `json:"enrollmentCode"`
	}
	if decode(w, r, &in) != nil || !validPassword(in.Password) || len(in.CurrentPassword) > 512 {
		fail(w, 400, "密码需要8–128个字符")
		return
	}
	a.mu.Lock()
	room, id := a.identity(r)
	if room == nil {
		a.mu.Unlock()
		fail(w, 401, "会话无效，请重新加入")
		return
	}
	code := room.Code
	previous, protected := a.Store.Passwords[id]
	enrollment := a.enrollmentKeys[id]
	a.mu.Unlock()
	if !a.beginAuth(w, r, code+":"+id) {
		return
	}
	defer a.endAuth()
	if !protected && (enrollment == "" || subtle.ConstantTimeCompare([]byte(strings.TrimSpace(in.EnrollmentCode)), []byte(enrollment)) != 1) {
		fail(w, 403, "旧座位验证码不正确，请向服务器房主索取")
		return
	}
	if protected && !previous.matches(in.CurrentPassword) {
		fail(w, 401, "原密码不正确")
		return
	}
	next := makePassword(in.Password)
	a.mu.Lock()
	defer a.mu.Unlock()
	room, currentID := a.identity(r)
	if room == nil || currentID != id || a.Store.Passwords[id] != previous {
		fail(w, 401, "会话已变化，请重新加入")
		return
	}
	a.Store.Passwords[id] = next
	t, sessions := a.replaceSession(room, id)
	if err := a.save(); err != nil {
		if protected {
			a.Store.Passwords[id] = previous
		} else {
			delete(a.Store.Passwords, id)
		}
		a.restoreSessions(t, sessions)
		fail(w, 500, "存档失败，密码未修改")
		return
	}
	a.notify(code)
	delete(a.enrollmentKeys, id)
	reply(w, 200, map[string]string{"token": t, "code": code})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t := token(r)
	s, ok := a.Store.Sessions[t]
	if !ok {
		fail(w, 401, "会话无效")
		return
	}
	delete(a.Store.Sessions, t)
	if err := a.save(); err != nil {
		a.Store.Sessions[t] = s
		fail(w, 500, "存档失败，退出未生效")
		return
	}
	a.notify(s.Code)
	reply(w, 200, map[string]bool{"ok": true})
}
