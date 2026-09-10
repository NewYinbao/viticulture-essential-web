package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

//go:embed web/*
var assets embed.FS

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
func (a *App) identity(r *http.Request) (*Room, string) {
	s, ok := a.Store.Sessions[token(r)]
	if !ok {
		return nil, ""
	}
	return a.Store.Rooms[s.Code], s.PlayerID
}
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]any{"ok": true, "version": "ee-rules-fix"})
	})
	mux.HandleFunc("POST /api/create", a.enter)
	mux.HandleFunc("POST /api/join", a.enter)
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		defer a.mu.Unlock()
		room, id := a.identity(r)
		if room == nil {
			fail(w, 401, "会话无效，请重新加入")
			return
		}
		reply(w, 200, room.view(id))
	})
	mux.HandleFunc("POST /api/action", a.action)
	mux.HandleFunc("GET /api/events", a.events)
	sub, _ := fs.Sub(assets, "web")
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
		Name string `json:"name"`
		Code string `json:"code"`
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
	a.mu.Lock()
	defer a.mu.Unlock()
	var room *Room
	created := r.URL.Path == "/api/create"
	if created {
		if len(a.Store.Rooms) >= 100 {
			fail(w, 429, "房间数量上限100，请管理员归档存档后重启")
			return
		}
		code := ""
		for {
			code = strings.ToUpper(uid()[:6])
			if a.Store.Rooms[code] == nil {
				break
			}
		}
		room = &Room{Code: code, Phase: "lobby", Players: []*Player{}, Spaces: []Space{}, WakeSlots: []WakeSlot{}, Log: []string{}}
	} else {
		room = a.Store.Rooms[strings.ToUpper(strings.TrimSpace(in.Code))]
		if room == nil {
			fail(w, 404, "房间不存在")
			return
		}
		if room.Phase != "lobby" {
			fail(w, 400, "对局已开始；原玩家请用原浏览器恢复会话")
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
	p := makePlayer(in.Name)
	room.Players = append(room.Players, p)
	if created {
		room.HostID = p.ID
	}
	room.Revision++
	room.log(in.Name + " 加入酒庄")
	a.Store.Rooms[room.Code] = room
	t := uid()
	a.Store.Sessions[t] = Session{Code: room.Code, PlayerID: p.ID}
	if e := a.save(); e != nil {
		delete(a.Store.Sessions, t)
		if created {
			delete(a.Store.Rooms, room.Code)
		} else {
			old := new(Room)
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
	var act Action
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
	if act.Revision != room.Revision {
		fail(w, 409, "局面已经变化，请刷新操作")
		return
	}
	b, _ := json.Marshal(room)
	copy := new(Room)
	_ = json.Unmarshal(b, copy)
	if e := copy.apply(id, act); e != nil {
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
	reply(w, 200, copy.view(id))
}
func (a *App) events(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	room, id := a.identity(r)
	if room == nil {
		a.mu.Unlock()
		fail(w, 401, "会话无效")
		return
	}
	if len(a.subs) >= 128 {
		a.mu.Unlock()
		fail(w, 429, "连接数上限")
		return
	}
	code := room.Code
	ch := make(chan struct{}, 1)
	a.subs[ch] = code
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.subs, ch); a.mu.Unlock() }()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(w)
	push := func() error {
		a.mu.Lock()
		b, e := json.Marshal(a.Store.Rooms[code].view(id))
		a.mu.Unlock()
		if e != nil {
			return e
		}
		_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if _, e = fmt.Fprintf(w, "data: %s\n\n", b); e != nil {
			return e
		}
		return rc.Flush()
	}
	if push() != nil {
		return
	}
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			if push() != nil {
				return
			}
		case <-ticker.C:
			_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if _, e := io.WriteString(w, ": heartbeat\n\n"); e != nil {
				return
			}
			if rc.Flush() != nil {
				return
			}
		}
	}
}
func main() {
	addr := flag.String("addr", "127.0.0.1:3010", "监听地址，局域网使用0.0.0.0:3010")
	data := flag.String("data", "ee-data", "存档目录")
	flag.Parse()
	a, e := loadApp(*data)
	if e != nil {
		log.Fatal(e)
	}
	srv := &http.Server{Addr: *addr, Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	fmt.Println("葡萄酒庄园 · Go局域网开发预览 / 非完整原版")
	fmt.Println("本机访问 http://localhost:3010")
	_, port, _ := net.SplitHostPort(*addr)
	ips, _ := net.InterfaceAddrs()
	for _, ip := range ips {
		if n, ok := ip.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			fmt.Printf("网卡候选地址 http://%s:%s（需监听0.0.0.0且防火墙允许）\n", n.IP, port)
		}
	}
	abs, _ := filepath.Abs(*data)
	fmt.Println("存档：", abs)
	fmt.Println("仅限可信局域网，不要映射到公网。关闭此窗口停止服务器。")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}
