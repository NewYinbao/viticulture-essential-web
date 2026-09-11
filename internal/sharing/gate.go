package sharing

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"net/url"
)

func NewKey() (string, error) {
	var b [24]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}

// Gate protects every route, including health and assets; no credentials in URLs.
func Gate(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		valid := func(s string) bool { return len(key) >= 32 && subtle.ConstantTimeCompare([]byte(s), []byte(key)) == 1 }
		if r.URL.Path == "/_access" && r.Method == "POST" {
			r.Body = http.MaxBytesReader(w, r.Body, 1024)
			if err := r.ParseForm(); err != nil || !valid(r.PostForm.Get("key")) {
				http.Error(w, "Invalid access key", 403)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "viticulture_gate", Value: key, Path: "/", HttpOnly: true, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https", SameSite: http.SameSiteLaxMode})
			target := "/"
			if room := r.URL.Query().Get("room"); room != "" {
				target += "?room=" + url.QueryEscape(room)
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
			return
		}
		cookie, err := r.Cookie("viticulture_gate")
		if err != nil || !valid(cookie.Value) {
			if r.Method == "GET" && r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, `<!doctype html><html lang="zh"><meta name="viewport" content="width=device-width"><title>酒庄访问口令</title><h1>酒庄 · 私人邀请</h1><p>请向房主索取本次启动的访问口令，不要向陌生人转发。</p><form method="post" action="%s"><label>访问口令 <input name="key" type="password" required autocomplete="off"></label><button>进入</button></form></html>`, html.EscapeString("/_access?room="+url.QueryEscape(r.URL.Query().Get("room"))))
			} else {
				http.Error(w, "Access key required; open /", 401)
			}
			return
		}
		if r.URL.Path == "/api/transport" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"poll":true}`)
			return
		}
		if r.URL.Path == "/api/events" {
			http.Error(w, "Use /api/state polling", 501)
			return
		}
		next.ServeHTTP(w, r)
	})
}
