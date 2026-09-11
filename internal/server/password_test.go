package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testPassword = "cellar-password-2026"

func authEnter(t *testing.T, a *App, path, name, code, password string, status int) map[string]string {
	t.Helper()
	w := request(a.routes(), "POST", path, "", map[string]string{"name": name, "code": code, "password": password})
	if w.Code != status {
		t.Fatalf("%s: got %d want %d: %s", path, w.Code, status, w.Body.String())
	}
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if status != 200 && result["token"] != "" {
		t.Fatal("rejection issued token")
	}
	return result
}

func TestPasswordSeatLoginAndPrivateState(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	for _, password := range []string{"", "short", strings.Repeat("a", 129)} {
		authEnter(t, a, "/api/create", "Alice", "", password, 400)
	}
	host := authEnter(t, a, "/api/create", "Alice", "", testPassword, 200)
	guest := authEnter(t, a, "/api/join", "Bob", host["code"], testPassword, 200)
	r := a.Store.Rooms[host["code"]]
	begin(t, r)
	before, _ := json.Marshal(r)
	for _, phase := range []string{"lobby", "summer", "finished"} {
		r.Phase = phase
		snapshot, _ := json.Marshal(a.Store)
		authEnter(t, a, "/api/join", "Alice", host["code"], "wrong-password", 401)
		after, _ := json.Marshal(a.Store)
		if !bytes.Equal(snapshot, after) {
			t.Fatal("failed login mutated state")
		}
	}
	_ = json.Unmarshal(before, r)
	for len(r.Players) < 6 {
		r.Players = append(r.Players, testRoom().Players[0])
	}
	// A full, started room still lets its authenticated owner recover the seat.
	restored := authEnter(t, a, "/api/join", "Alice", host["code"], testPassword, 200)
	if len(r.Players) != 6 || r.Players[0].ID != a.Store.Sessions[restored["token"]].PlayerID {
		t.Fatal("login created a seat")
	}
	if request(a.routes(), "GET", "/api/state", host["token"], nil).Code != 401 {
		t.Fatal("old login still valid")
	}
	for _, identity := range []map[string]string{restored, guest} {
		w := request(a.routes(), "GET", "/api/state", identity["token"], nil)
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		for _, secret := range []string{testPassword, `"Passwords"`, `"Salt"`, `"Hash"`} {
			if strings.Contains(w.Body.String(), secret) {
				t.Fatal("credentials leaked in View")
			}
		}
		own := a.Store.Sessions[identity["token"]].PlayerID
		for _, p := range r.Players {
			if p.ID != own {
				for _, c := range p.Hand {
					if strings.Contains(w.Body.String(), c.ID) {
						t.Fatal("opponent hand leaked")
					}
				}
			}
		}
	}
	if a.Store.Passwords[r.Players[0].ID].Salt == a.Store.Passwords[r.Players[1].ID].Salt {
		t.Fatal("salt reused")
	}
}

func TestLegacyPasswordEnrollmentAndDurableRotation(t *testing.T) {
	dir := t.TempDir()
	a, _ := loadApp(dir)
	r := testRoom()
	begin(t, r)
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["original-browser"] = Session{Code: r.Code, PlayerID: r.HostID}
	a.Store.Sessions["other-browser"] = Session{Code: r.Code, PlayerID: r.HostID}
	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	a, _ = loadApp(dir)
	authEnter(t, a, "/api/join", r.Players[0].Name, r.Code, testPassword, 403)
	for _, token := range []string{"original-browser", "other-browser"} {
		for _, path := range []string{"/api/state", "/api/events"} {
			w := request(a.routes(), "GET", path, token, nil)
			if w.Code != 403 || strings.Contains(w.Body.String(), `"hand"`) || strings.Contains(w.Body.String(), a.enrollmentKeys[r.HostID]) {
				t.Fatal("legacy view leaked", w.Body.String())
			}
		}
		w := request(a.routes(), "POST", "/api/password", token, map[string]string{"password": testPassword})
		if w.Code != 403 {
			t.Fatal("bearer alone claimed legacy seat")
		}
	}
	change := func(token, old, next string, status int) string {
		t.Helper()
		w := request(a.routes(), "POST", "/api/password", token, map[string]string{"password": next, "currentPassword": old, "enrollmentCode": a.enrollmentKeys[r.HostID]})
		if w.Code != status {
			t.Fatalf("password %d: %s", w.Code, w.Body.String())
		}
		var result map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &result)
		return result["token"]
	}
	change("not-owner", "", testPassword, 401)
	next := change("original-browser", "", testPassword, 200)
	if next == "" {
		t.Fatal("no replacement session")
	}
	for _, old := range []string{"original-browser", "other-browser"} {
		if request(a.routes(), "GET", "/api/state", old, nil).Code != 401 {
			t.Fatal("old session survived enrollment")
		}
	}
	change(next, "wrong-password", "new-password-2026", 401)
	rotated := change(next, testPassword, "new-password-2026", 200)
	a, _ = loadApp(dir)
	if request(a.routes(), "GET", "/api/state", next, nil).Code != 401 {
		t.Fatal("old token restored")
	}
	if request(a.routes(), "GET", "/api/state", rotated, nil).Code != 200 {
		t.Fatal("new token lost")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "ee-state-v1.json"))
	if bytes.Contains(data, []byte(testPassword)) || bytes.Contains(data, []byte("new-password-2026")) {
		t.Fatal("plaintext password saved")
	}
	authEnter(t, a, "/api/join", r.Players[0].Name, r.Code, testPassword, 401)
	login := authEnter(t, a, "/api/join", r.Players[0].Name, r.Code, "new-password-2026", 200)
	if request(a.routes(), "POST", "/api/logout", login["token"], map[string]string{}).Code != 200 {
		t.Fatal("logout")
	}
	a, _ = loadApp(dir)
	if request(a.routes(), "GET", "/api/state", login["token"], nil).Code != 401 {
		t.Fatal("logout not durable")
	}
}

func TestPasswordSaveFailureRollback(t *testing.T) {
	for _, operation := range []string{"create", "join", "login", "password", "logout"} {
		t.Run(operation, func(t *testing.T) {
			a, _ := loadApp(t.TempDir())
			host := authEnter(t, a, "/api/create", "Host", "", testPassword, 200)
			before, _ := json.Marshal(a.Store)
			a.dir = filepath.Join(a.dir, "file")
			if err := os.WriteFile(a.dir, []byte("x"), 0600); err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "create":
				authEnter(t, a, "/api/create", "Next", "", testPassword, 500)
			case "join":
				authEnter(t, a, "/api/join", "Next", host["code"], testPassword, 500)
			case "login":
				authEnter(t, a, "/api/join", "Host", host["code"], testPassword, 500)
			case "password", "logout":
				w := request(a.routes(), "POST", "/api/"+operation, host["token"], map[string]string{"password": "new-password-2026", "currentPassword": testPassword})
				if w.Code != 500 {
					t.Fatal(w.Code, w.Body.String())
				}
			}
			after, _ := json.Marshal(a.Store)
			if !bytes.Equal(before, after) {
				t.Fatal("failed persistence modified identity or game")
			}
		})
	}
}

func TestPasswordRevokesAlreadyOpenSSE(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	host := authEnter(t, a, "/api/create", "Host", "", testPassword, 200)
	srv := httptest.NewServer(a.routes())
	defer srv.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(srv.URL + "/api/events?token=" + host["token"])
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "data: ") {
		t.Fatal("missing initial private view")
	}
	authEnter(t, a, "/api/join", "Host", host["code"], testPassword, 200)
	var rest []string
	for scanner.Scan() {
		rest = append(rest, scanner.Text())
	}
	if scanner.Err() != nil {
		t.Fatal(scanner.Err())
	}
	if !strings.Contains(strings.Join(rest, "\n"), "event: session-ended") {
		t.Fatal("old SSE not explicitly closed", rest)
	}
	if strings.Contains(strings.Join(rest, "\n"), `"hand"`) {
		t.Fatal("revoked SSE received more state")
	}
	res, err := client.Get(srv.URL + "/api/events?token=" + host["token"])
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode != 401 {
		t.Fatal("revoked SSE reconnected")
	}
}

func TestPasswordAttemptLimitsAndMalformedCredential(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	host := authEnter(t, a, "/api/create", "Host", "", testPassword, 200)
	for i := 0; i < 8; i++ {
		authEnter(t, a, "/api/join", "Host", host["code"], "wrong-password", 401)
	}
	authEnter(t, a, "/api/join", "Host", host["code"], "wrong-password", 429)
	if (PasswordRecord{Salt: "bad", Hash: "bad"}).matches(testPassword) {
		t.Fatal("corrupt record accepted")
	}
}
