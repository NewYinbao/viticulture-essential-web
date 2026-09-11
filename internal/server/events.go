package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *App) events(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	room, id := a.identity(r)
	if room == nil {
		a.mu.Unlock()
		fail(w, 401, "会话无效")
		return
	}
	if !a.requirePassword(w, room, id) {
		a.mu.Unlock()
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
		current, currentID := a.identity(r)
		if current == nil || currentID != id || current.Code != code {
			a.mu.Unlock()
			_ = rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_, _ = io.WriteString(w, "event: session-ended\ndata: {}\n\n")
			_ = rc.Flush()
			return fmt.Errorf("session revoked")
		}
		b, e := json.Marshal(a.playerView(current, id))
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
