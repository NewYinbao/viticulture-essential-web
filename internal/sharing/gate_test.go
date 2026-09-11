package sharing

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewKey(t *testing.T) {
	a, e := NewKey()
	if e != nil {
		t.Fatal(e)
	}
	b, e := NewKey()
	if e != nil {
		t.Fatal(e)
	}
	if raw, e := hex.DecodeString(a); e != nil || len(raw) != 24 || a == b {
		t.Fatalf("invalid/randomness failure: %q %q", a, b)
	}
}
func TestGateRoutesAndLogin(t *testing.T) {
	key := strings.Repeat("a", 48)
	calls := 0
	h := Gate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }), key)
	for _, path := range []string{"/api/state", "/api/events", "/api/transport", "/health", "/static/js/app.js"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path+"?key="+key, nil))
		if w.Code != 401 {
			t.Errorf("unguarded %s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "/_access") || strings.Contains(w.Body.String(), key) {
		t.Fatal("invalid login page")
	}
	for _, value := range []string{"wrong", strings.Repeat("x", 2048)} {
		r := httptest.NewRequest("POST", "/_access", strings.NewReader(url.Values{"key": {value}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 || len(w.Result().Cookies()) != 0 {
			t.Fatal("invalid key accepted")
		}
	}
	r := httptest.NewRequest("POST", "/_access", strings.NewReader(url.Values{"key": {key}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-Forwarded-Proto", "https")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 303 || w.Header().Get("Location") != "/" {
		t.Fatal("login redirect")
	}
	cs := w.Result().Cookies()
	if len(cs) != 1 {
		t.Fatal("cookie missing")
	}
	c := cs[0]
	if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
		t.Fatalf("cookie security: %+v", c)
	}
	for _, tc := range []struct {
		path string
		code int
		body string
	}{{"/api/transport", 200, `{"poll":true}`}, {"/api/events", 501, "Use /api/state polling\n"}, {"/api/state", 204, ""}} {
		r = httptest.NewRequest("GET", tc.path, nil)
		r.AddCookie(c)
		w = httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.code || w.Body.String() != tc.body {
			t.Errorf("%s: %d %q", tc.path, w.Code, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Error("missing privacy headers")
		}
	}
	if calls != 1 {
		t.Fatalf("downstream calls %d", calls)
	}
}
func TestGateRejectsShortConfiguredKey(t *testing.T) {
	h := Gate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("short key allowed") }), "short")
	r := httptest.NewRequest("GET", "/api/state", nil)
	r.AddCookie(&http.Cookie{Name: "viticulture_gate", Value: "short"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}

func TestInvitationSurvivesGate(t *testing.T) {
	key := "0123456789012345678901234567890123456789012345678901"
	h := Gate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), key)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/_access?room=ABC123", strings.NewReader("key="+key))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(w, req)
	if w.Code != 303 || w.Header().Get("Location") != "/?room=ABC123" {
		t.Fatalf("invitation lost: %d %s", w.Code, w.Header().Get("Location"))
	}
}
