package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"vineyard/internal/game"
	"vineyard/web"
)

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, e string) {
	reply(w, status, map[string]string{"error": e})
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return fmt.Errorf("trailing input")
	}
	return nil
}
func token(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if r.Method == "GET" {
		return r.URL.Query().Get("token")
	}
	return ""
}
func (a *App) identity(r *http.Request) (*game.Room, string) {
	s, ok := a.Store.Sessions[token(r)]
	if !ok {
		return nil, ""
	}
	room := a.Store.Rooms[s.Code]
	if room == nil || room.Player(s.PlayerID) == nil {
		return nil, ""
	}
	return room, s.PlayerID
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/transport", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]bool{"poll": false}) })
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]any{"ok": true, "version": "ee-rules-fix"})
	})
	mux.HandleFunc("POST /api/create", a.enter)
	mux.HandleFunc("POST /api/join", a.enter)
	mux.HandleFunc("POST /api/password", a.password)
	mux.HandleFunc("POST /api/logout", a.logout)
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		room, id := a.identity(r)
		if room == nil {
			fail(w, 401, "会话无效，请重新加入")
			return
		}
		if !a.requirePassword(w, room, id) {
			return
		}
		reply(w, 200, a.playerView(room, id))
	})
	mux.HandleFunc("POST /api/action", a.action)
	mux.HandleFunc("GET /api/events", a.events)
	sub := web.Assets()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'")
		if origin := r.Header.Get("Origin"); origin != "" {
			u, e := url.Parse(origin)
			if e != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				fail(w, 403, "不允许跨站请求")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func (a *App) enter(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name     string `json:"name"`
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if e := decode(w, r, &in); e != nil {
		fail(w, 400, "请求格式错误")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len([]rune(in.Name)) < 1 || len([]rune(in.Name)) > 20 {
		fail(w, 400, "昵称需要1–20个字")
		return
	}
	for _, c := range in.Name {
		if c < 32 {
			fail(w, 400, "昵称包含控制字符")
			return
		}
	}
	if !validPassword(in.Password) {
		fail(w, 400, "密码需要8–128个字符")
		return
	}
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	if !a.beginAuth(w, r, in.Code+":"+in.Name) {
		return
	}
	defer a.endAuth()
	// Password checks run outside the game mutex, so login cannot freeze a table.
	if r.URL.Path == "/api/join" {
		a.mu.Lock()
		room := a.Store.Rooms[in.Code]
		id := ""
		if room != nil {
			for _, p := range room.Players {
				if p.Name == in.Name {
					id = p.ID
					break
				}
			}
		}
		record, protected := a.Store.Passwords[id]
		a.mu.Unlock()
		if room == nil {
			fail(w, 404, "房间不存在")
			return
		}
		if id != "" {
			if !protected {
				fail(w, 403, "旧座位尚未设置密码，请在原浏览器中设置；不能凭昵称认领")
				return
			}
			if !record.matches(in.Password) {
				fail(w, 401, "昵称或密码不正确")
				return
			}
			a.mu.Lock()
			defer a.mu.Unlock()
			room = a.Store.Rooms[in.Code]
			if room == nil || room.Player(id) == nil || a.Store.Passwords[id] != record {
				fail(w, 409, "座位已变化，请重新加入")
				return
			}
			t, sessions := a.replaceSession(room, id)
			if err := a.save(); err != nil {
				a.restoreSessions(t, sessions)
				fail(w, 500, "存档失败，登录未生效")
				return
			}
			a.notify(room.Code)
			reply(w, 200, map[string]string{"token": t, "code": room.Code})
			return
		}
	}
	password := makePassword(in.Password)
	a.mu.Lock()
	defer a.mu.Unlock()
	var room *game.Room
	created := r.URL.Path == "/api/create"
	if created {
		if len(a.Store.Rooms) >= 100 {
			fail(w, 429, "房间数量上限100，请管理员归档存档后重启")
			return
		}
		code := ""
		for {
			code = strings.ToUpper(game.NewID()[:6])
			if a.Store.Rooms[code] == nil {
				break
			}
		}
		room = &game.Room{Code: code, Phase: "lobby", Players: []*game.Player{}, Spaces: []game.Space{}, WakeSlots: []game.WakeSlot{}, Log: []string{}}
	} else {
		room = a.Store.Rooms[in.Code]
		if room == nil {
			fail(w, 404, "房间不存在")
			return
		}
		if room.Phase != "lobby" {
			fail(w, 400, "对局已开始；原玩家请输入原昵称和密码")
			return
		}
		if len(room.Players) >= 6 {
			fail(w, 400, "房间已满")
			return
		}
		for _, p := range room.Players {
			if p.Name == in.Name {
				fail(w, 400, "该昵称已被使用")
				return
			}
		}
	}
	before, _ := json.Marshal(room)
	p := game.NewPlayer(in.Name)
	room.Players = append(room.Players, p)
	if created {
		room.HostID = p.ID
	}
	room.Revision++
	room.AddLog(in.Name + " 加入酒庄")
	a.Store.Rooms[room.Code] = room
	t := game.NewID()
	a.Store.Sessions[t] = Session{Code: room.Code, PlayerID: p.ID}
	a.Store.Passwords[p.ID] = password
	if e := a.save(); e != nil {
		delete(a.Store.Sessions, t)
		delete(a.Store.Passwords, p.ID)
		if created {
			delete(a.Store.Rooms, room.Code)
		} else {
			old := new(game.Room)
			_ = json.Unmarshal(before, old)
			a.Store.Rooms[room.Code] = old
		}
		log.Print("save: ", e)
		fail(w, 500, "存档失败，加入未生效")
		return
	}
	a.notify(room.Code)
	reply(w, 200, map[string]string{"token": t, "code": room.Code})
}
func (a *App) action(w http.ResponseWriter, r *http.Request) {
	var act game.Action
	if e := decode(w, r, &act); e != nil {
		fail(w, 400, "请求格式错误")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	room, id := a.identity(r)
	if room == nil {
		fail(w, 401, "会话无效")
		return
	}
	if !a.requirePassword(w, room, id) {
		return
	}
	if act.Revision != room.Revision {
		fail(w, 409, "局面已经变化，请刷新操作")
		return
	}
	b, _ := json.Marshal(room)
	copy := new(game.Room)
	_ = json.Unmarshal(b, copy)
	if e := copy.Apply(id, act); e != nil {
		fail(w, 400, e.Error())
		return
	}
	copy.Revision++
	a.Store.Rooms[room.Code] = copy
	if e := a.save(); e != nil {
		a.Store.Rooms[room.Code] = room
		log.Print("save: ", e)
		fail(w, 500, "存档失败，本次操作未生效")
		return
	}
	a.notify(room.Code)
	reply(w, 200, a.playerView(copy, id))
}

// Handler returns the application HTTP handler.
func (a *App) Handler() http.Handler { return a.routes() }
