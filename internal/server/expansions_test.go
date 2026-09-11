package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"vineyard/internal/game"
)

func TestExpansionHTTPConfigurationPersistenceAndPrivacy(t *testing.T) {
	dir := t.TempDir()
	a, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	h := a.routes()
	enter := func(path, name, code string) map[string]string {
		w := request(h, "POST", path, "", map[string]string{"password": "test-password-123", "name": name, "code": code})
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		var v map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &v)
		return v
	}
	host := enter("/api/create", "Host", "")
	guest := enter("/api/join", "Guest", host["code"])
	room := func() *game.Room { return a.Store.Rooms[host["code"]] }
	checkRejected := func(token string, act game.Action, status int) {
		before, _ := json.Marshal(a.Store)
		w := request(h, "POST", "/api/action", token, act)
		if w.Code != status {
			t.Fatal(w.Code, w.Body.String())
		}
		after, _ := json.Marshal(a.Store)
		if !bytes.Equal(before, after) {
			t.Fatal("rejected configuration mutated state")
		}
	}
	cfg := &game.ExpansionConfig{Board: "ee", Visitors: "ee"}
	checkRejected(guest["token"], game.Action{Type: "configure", Revision: room().Revision, Config: cfg}, 400)
	checkRejected(host["token"], game.Action{Type: "configure", Revision: 0, Config: cfg}, 409)
	for _, bad := range []game.ExpansionConfig{{Visitors: "moor+rhine"}, {Visitors: "ee+rhine"}, {Board: "bordeaux"}} {
		checkRejected(host["token"], game.Action{Type: "configure", Revision: room().Revision, Config: &bad}, 400)
	}
	rev := room().Revision
	w := request(h, "POST", "/api/action", host["token"], game.Action{Type: "configure", Revision: rev, Config: cfg})
	if w.Code != 200 || room().Revision != rev+1 {
		t.Fatal(w.Body.String())
	}
	restored, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	if restored.Store.Rooms[host["code"]].Config != *cfg {
		t.Fatal("config not persisted")
	}
	cfg.Structures = true
	rev = room().Revision
	w = request(h, "POST", "/api/action", host["token"], game.Action{Type: "configure", Revision: rev, Config: cfg})
	if w.Code != 200 || room().Config != *cfg {
		t.Fatal("structure configuration should be accepted", w.Body.String())
	}
	w = request(h, "POST", "/api/action", host["token"], game.Action{Type: "start", Revision: room().Revision})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	checkRejected(host["token"], game.Action{Type: "configure", Revision: room().Revision, Config: cfg}, 400)
	for _, identity := range []map[string]string{host, guest} {
		w = request(h, "GET", "/api/state", identity["token"], nil)
		var v struct {
			Config  game.ExpansionConfig `json:"config"`
			Players []map[string]any     `json:"players"`
			Hand    []game.Card          `json:"hand"`
		}
		if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil {
			t.Fatal(e)
		}
		if v.Config != *cfg {
			t.Fatal("poll config", v.Config)
		}
		for _, p := range v.Players {
			if _, ok := p["hand"]; ok {
				t.Fatal("private hand in public players")
			}
		}
		if strings.Contains(w.Body.String(), `"decks":`) || strings.Contains(w.Body.String(), `"discards":`) {
			t.Fatal("deck identities leaked")
		}
	}
	server := httptest.NewServer(h)
	defer server.Close()
	for _, identity := range []map[string]string{host, guest} {
		client := http.Client{Timeout: 3 * time.Second}
		res, e := client.Get(server.URL + "/api/events?token=" + identity["token"])
		if e != nil {
			t.Fatal(e)
		}
		scanner := bufio.NewScanner(res.Body)
		if !scanner.Scan() {
			res.Body.Close()
			t.Fatal("no event")
		}
		line := scanner.Text()
		res.Body.Close()
		if !strings.Contains(line, `"config":{"board":"ee"`) {
			t.Fatal("SSE config missing")
		}
		playerID := a.Store.Sessions[identity["token"]].PlayerID
		for _, p := range room().Players {
			if p.ID != playerID {
				for _, c := range p.Hand {
					if strings.Contains(line, c.ID) {
						t.Fatal("SSE private hand leaked")
					}
				}
			}
		}
	}
}

func TestLegacySaveDefaultsToEEConfiguration(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	r := testRoom()
	begin(t, r)
	data, _ := json.Marshal(r)
	var legacy map[string]any
	_ = json.Unmarshal(data, &legacy)
	delete(legacy, "config")
	data, _ = json.Marshal(legacy)
	var loaded game.Room
	if e := json.Unmarshal(data, &loaded); e != nil {
		t.Fatal(e)
	}
	a.Store.Rooms[loaded.Code] = &loaded
	if got := loaded.View(loaded.HostID)["config"].(game.ExpansionConfig); got.Board != "ee" || got.Visitors != "ee" || got.Structures || got.SpecialWorkers {
		t.Fatal(got)
	}
}
