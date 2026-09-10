package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"vineyard/internal/game"
)

func TestSaveRestore(t *testing.T) {
	dir := t.TempDir()
	a, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	r := testRoom()
	begin(t, r)
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["secret"] = Session{Code: r.Code, PlayerID: r.HostID}
	if e = a.save(); e != nil {
		t.Fatal(e)
	}
	b, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	if b.Store.Rooms[r.Code].TurnID != r.TurnID || b.Store.Sessions["secret"].PlayerID != r.HostID {
		t.Fatal("restore")
	}
}

func request(h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHTTPAndReplay(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	h := a.routes()
	w := request(h, "POST", "/api/create", "", map[string]string{"name": "alice"})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var ca map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &ca)
	w = request(h, "POST", "/api/join", "", map[string]string{"name": "bob", "code": ca["code"]})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	r := a.Store.Rooms[ca["code"]]
	act := game.Action{Type: "start", Revision: r.Revision}
	w = request(h, "POST", "/api/action", ca["token"], act)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request(h, "POST", "/api/action", ca["token"], act)
	if w.Code != 409 {
		t.Fatal("replay not rejected", w.Code)
	}
	w = request(h, "GET", "/api/state?token=wrong", "", nil)
	if w.Code != 401 {
		t.Fatal("invalid session")
	}
	w = request(h, "POST", "/api/action", ca["token"], map[string]string{"type": "place", "extra": "x"})
	if w.Code != 400 {
		t.Fatal("unknown fields")
	}
	req := httptest.NewRequest("POST", "/api/create", strings.NewReader(`{"name":"evil"}`))
	req.Header.Set("Origin", "https://evil.example")
	ww := httptest.NewRecorder()
	h.ServeHTTP(ww, req)
	if ww.Code != 403 {
		t.Fatal("cross origin")
	}
}

func TestTransactionRollback(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	r := testRoom()
	begin(t, r)
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["x"] = Session{Code: r.Code, PlayerID: r.TurnID}
	before, _ := json.Marshal(r)
	w := request(a.routes(), "POST", "/api/action", "x", game.Action{Type: "place", Space: "build", Building: "unknown", Revision: r.Revision})
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	after, _ := json.Marshal(a.Store.Rooms[r.Code])
	if !bytes.Equal(before, after) {
		t.Fatal("invalid action modified state")
	}
}

func TestSSEPrivateState(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	r := testRoom()
	begin(t, r)
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["a"] = Session{Code: r.Code, PlayerID: r.Players[0].ID}
	server := httptest.NewServer(a.routes())
	defer server.Close()
	client := http.Client{Timeout: 2 * time.Second}
	res, e := client.Get(server.URL + "/api/events?token=a")
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	if !scanner.Scan() {
		t.Fatal("no SSE data")
	}
	line := scanner.Text()
	if !strings.HasPrefix(line, "data: ") {
		t.Fatal(line)
	}
	for _, c := range r.Players[1].Hand {
		if strings.Contains(line, c.ID) {
			t.Fatal("SSE hidden hand leak")
		}
	}
}

func TestEEVisitorHTTPAtomicReject(t *testing.T) {
	r := eeBegin(t, 2)
	a, _ := loadApp(t.TempDir())
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["ee"] = Session{Code: r.Code, PlayerID: r.TurnID}
	before, _ := json.Marshal(r)
	body, _ := json.Marshal(game.Action{Type: "place", Space: "summer_visitor", Revision: r.Revision})
	req := httptest.NewRequest("POST", "/api/action", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer ee")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, req)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "非自己的当季访客") {
		t.Fatal(w.Code, w.Body.String())
	}
	after, _ := json.Marshal(a.Store.Rooms[r.Code])
	if !bytes.Equal(before, after) {
		t.Fatal("rollback")
	}
}

func TestEEIgnoresPreviewSave(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte("not json"), 0600)
	if _, e := loadApp(dir); e != nil {
		t.Fatal(e)
	}
}

func TestEEWinterVisitorHTTPAtomicReject(t *testing.T) {
	r := eeBegin(t, 2)
	for i := 0; i < 2; i++ {
		if e := r.Apply(r.TurnID, game.Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	resolveTestChoices(t, r)
	a, _ := loadApp(t.TempDir())
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["winter"] = Session{Code: r.Code, PlayerID: r.TurnID}
	before, _ := json.Marshal(r)
	w := request(a.routes(), "POST", "/api/action", "winter", game.Action{Type: "place", Space: "winter_visitor", Revision: r.Revision})
	after, _ := json.Marshal(a.Store.Rooms[r.Code])
	if w.Code != 400 || !strings.Contains(w.Body.String(), "非自己的当季访客") || !bytes.Equal(before, after) {
		t.Fatal("winter visitor rejection must be explicit and atomic", w.Body.String())
	}
}

func TestVisitorPermissionRollbackReplayAndPersistence(t *testing.T) {
	r, p := visitorFixture("winter-38")
	dir := t.TempDir()
	app, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	app.Store.Rooms[r.Code] = r
	for i, q := range r.Players {
		app.Store.Sessions[fmt.Sprint(i)] = Session{Code: r.Code, PlayerID: q.ID}
	}
	send := func(token string, a game.Action) int {
		return request(app.routes(), "POST", "/api/action", token, a).Code
	}
	if code := send("0", game.Action{Type: "place", Space: "winter_visitor", CardID: "winter-38", Slot: 2}); code != 200 {
		t.Fatal(code)
	}
	r = app.Store.Rooms[r.Code]
	c := r.Choices[0]
	act := game.Action{Type: "choose", ChoiceID: c.ID, Option: "resolve", Revision: r.Revision}
	before, _ := json.Marshal(r)
	if send("1", act) != 400 {
		t.Fatal("wrong responder accepted")
	}
	after, _ := json.Marshal(app.Store.Rooms[r.Code])
	if !bytes.Equal(before, after) {
		t.Fatal("authority rollback")
	}
	if send("0", act) != 200 {
		t.Fatal("actor train")
	}
	if send("0", act) != 409 {
		t.Fatal("stale replay")
	}
	r = app.Store.Rooms[r.Code]
	c = r.Choices[0]
	act = game.Action{Type: "choose", ChoiceID: c.ID, Option: "train", Revision: r.Revision}
	if send("0", act) != 200 {
		t.Fatal("actor reply")
	}
	app, e = loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	r = app.Store.Rooms[r.Code]
	if r.Choices[0].PlayerID != r.Players[1].ID || r.Context.ActorID != p.ID {
		t.Fatal("lost multiplayer continuation")
	}
	c = r.Choices[0]
	act = game.Action{Type: "choose", ChoiceID: c.ID, Option: "train", Revision: r.Revision}
	// Current revision cannot replay an old choice either.
	bad := act
	bad.ChoiceID = "replayed"
	if send("1", bad) != 400 {
		t.Fatal("choice replay")
	}
	// Fail persistence after the candidate was valid: no authoritative mutation.
	app.dir = filepath.Join(dir, "not-a-directory")
	if e = os.WriteFile(app.dir, []byte("x"), 0600); e != nil {
		t.Fatal(e)
	}
	before, _ = json.Marshal(r)
	if send("1", act) != 500 {
		t.Fatal("expected save error")
	}
	after, _ = json.Marshal(app.Store.Rooms[r.Code])
	if !bytes.Equal(before, after) {
		t.Fatal("save rollback")
	}
	app.dir = dir
	if send("1", act) != 200 {
		t.Fatal("opponent train")
	}
	r = app.Store.Rooms[r.Code]
	if r.Players[0].VP != 11 || r.Players[1].Trained != 1 || r.Players[1].Workers != 2 {
		t.Fatal("training/reward timing")
	}
	c = r.Choices[0]
	if send("2", game.Action{Type: "choose", ChoiceID: c.ID, Option: "skip", Revision: r.Revision}) != 200 {
		t.Fatal("skip")
	}
	app, e = loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	r = app.Store.Rooms[r.Code]
	if r.Context != nil || len(r.Choices) != 0 || r.TurnID != r.Players[1].ID {
		t.Fatal("completion persistence")
	}
}

func TestImporterRealHTTPEmptyEscrowRestart(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(fmt.Sprint(restart), func(t *testing.T) {
			r, _ := visitorFixture("summer-23")
			dir := t.TempDir()
			app, e := loadApp(dir)
			if e != nil {
				t.Fatal(e)
			}
			app.Store.Rooms[r.Code] = r
			for _, p := range r.Players {
				app.Store.Sessions[p.ID] = Session{Code: r.Code, PlayerID: p.ID}
			}
			srv := httptest.NewServer(app.routes())
			defer func() { srv.Close() }()
			send := func(id string, a game.Action) {
				t.Helper()
				a.Revision = app.Store.Rooms[r.Code].Revision
				b, _ := json.Marshal(a)
				req, _ := http.NewRequest("POST", srv.URL+"/api/action", bytes.NewReader(b))
				req.Header.Set("Authorization", "Bearer "+id)
				res, e := srv.Client().Do(req)
				if e != nil {
					t.Fatal(e)
				}
				defer res.Body.Close()
				body, _ := io.ReadAll(res.Body)
				if res.StatusCode != 200 {
					t.Fatalf("%d %s", res.StatusCode, body)
				}
			}
			send(r.Players[0].ID, game.Action{Type: "place", Space: "summer_visitor", CardID: "summer-23", Slot: 2})
			choose := func(option string, ids []string) {
				q := app.Store.Rooms[r.Code].Choices[0]
				send(q.PlayerID, game.Action{Type: "choose", ChoiceID: q.ID, Option: option, CardIDs: ids})
			}
			choose("resolve", nil)
			if restart {
				srv.Close()
				app, e = loadApp(dir)
				if e != nil {
					t.Fatal(e)
				}
				srv = httptest.NewServer(app.routes())
			}
			choose("give", []string{"s0"})
			choose("give", []string{"s1", "s2"})
			choose("resolve", nil)
			loaded, e := loadApp(dir)
			if e != nil {
				t.Fatal(e)
			}
			got := loaded.Store.Rooms[r.Code]
			if got.Context != nil || len(got.Choices) != 0 || len(got.Players[0].Hand) != 10 {
				t.Fatalf("settlement failed: context=%v hand=%d", got.Context, len(got.Players[0].Hand))
			}
		})
	}
}

func TestVisitorFeasibilityRealHTTPRollbackAndRestart(t *testing.T) {
	r, p := visitorFixture("summer-05")
	dir := t.TempDir()
	app, e := loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	app.Store.Rooms[r.Code] = r
	app.Store.Sessions["actor"] = Session{Code: r.Code, PlayerID: p.ID}
	srv := httptest.NewServer(app.routes())
	defer func() { srv.Close() }()
	send := func(a game.Action, want int) {
		t.Helper()
		a.Revision = app.Store.Rooms[r.Code].Revision
		b, _ := json.Marshal(a)
		req, _ := http.NewRequest("POST", srv.URL+"/api/action", bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer actor")
		res, e := srv.Client().Do(req)
		if e != nil {
			t.Fatal(e)
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != want {
			t.Fatalf("%d %s", res.StatusCode, body)
		}
	}
	// Has plant capability but cannot pay for a building.
	p.Coins = 0
	send(game.Action{Type: "place", Space: "summer_visitor", CardID: "summer-05", Slot: 2}, 200)
	r = app.Store.Rooms[r.Code]
	before, _ := json.Marshal(r)
	c := r.Choices[0]
	send(game.Action{Type: "choose", ChoiceID: c.ID, Option: "vp_build"}, 400)
	after, _ := json.Marshal(app.Store.Rooms[r.Code])
	if !bytes.Equal(before, after) {
		t.Fatal("HTTP committed free VP")
	}
	srv.Close()
	app, e = loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	srv = httptest.NewServer(app.routes())
	r = app.Store.Rooms[r.Code]
	if r.Choices[0].ID != c.ID {
		t.Fatal("rejected step altered persisted choice")
	}
	send(game.Action{Type: "choose", ChoiceID: c.ID, Option: "vp_plant"}, 200)
	srv.Close()
	app, e = loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	srv = httptest.NewServer(app.routes())
	r = app.Store.Rooms[r.Code]
	c = r.Choices[0]
	send(game.Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}, 400)
	send(game.Action{Type: "choose", ChoiceID: c.ID, Option: "plant", CardIDs: []string{"plant"}, Fields: []int{2}}, 200)
	if app.Store.Rooms[r.Code].Context != nil {
		t.Fatal("not completed")
	}
}

func testRoom() *game.Room {
	a, b := game.NewPlayer("A"), game.NewPlayer("B")
	return &game.Room{Code: "TEST01", Phase: "lobby", HostID: a.ID, Players: []*game.Player{a, b}, Spaces: []game.Space{}, WakeSlots: []game.WakeSlot{}, Log: []string{}}
}

func begin(t *testing.T, r *game.Room) {
	t.Helper()
	if e := r.Apply(r.HostID, game.Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	resolveTestChoices(t, r)
	r.SpringLeader = 0
	r.TurnID = r.Players[0].ID
	for i := 1; r.Phase == "wake"; i++ {
		if e := r.Apply(r.TurnID, game.Action{Type: "wake", Slot: i, Color: "summer"}); e != nil {
			t.Fatal(e)
		}
	}
}

func resolveTestChoices(t *testing.T, r *game.Room) {
	t.Helper()
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		a := game.Action{Type: "choose", ChoiceID: c.ID}
		switch c.Kind {
		case "papa":
			a.Option = "coins"
		case "fall":
			a.Option = "winter"
		case "discard":
			for _, card := range r.Player(c.PlayerID).Hand[:c.Count] {
				a.CardIDs = append(a.CardIDs, card.ID)
			}
		default:
			t.Fatalf("unexpected choice %s", c.Kind)
		}
		if e := r.Apply(c.PlayerID, a); e != nil {
			t.Fatal(e)
		}
	}
}

func eeRoom(n int) *game.Room {
	r := &game.Room{Phase: "lobby", Code: "EE"}
	for i := 0; i < n; i++ {
		r.Players = append(r.Players, game.NewPlayer("p"))
	}
	r.HostID = r.Players[0].ID
	return r
}

func eeBegin(t *testing.T, n int) *game.Room {
	t.Helper()
	r := eeRoom(n)
	if e := r.Apply(r.HostID, game.Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		if e := r.Apply(c.PlayerID, game.Action{Type: "choose", ChoiceID: c.ID, Option: "coins"}); e != nil {
			t.Fatal(e)
		}
	}
	for slot := 1; r.Phase == "wake"; slot++ {
		if e := r.Apply(r.TurnID, game.Action{Type: "wake", Slot: slot, Color: "summer"}); e != nil {
			t.Fatal(e)
		}
	}
	return r
}

func visitorFixture(id string) (*game.Room, *game.Player) {
	r := eeRoom(3)
	r.InitDecks()
	r.Phase = strings.Split(id, "-")[0]
	r.Year = 1
	r.Spaces = game.NewSpaces(3)
	for i, p := range r.Players {
		p.Wake = i + 1
		p.Coins = 50
		p.VP = 10
		p.Buildings = []string{"trellis", "irrigation"}
		p.Hand = []game.Card{{ID: "plant", Name: "C", Type: "vine", White: 4}, {ID: "order", Type: "order", Requirements: []game.Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
		for k := 0; k < 5; k++ {
			p.Hand = append(p.Hand, game.Card{ID: fmt.Sprint("s", k), Type: "summer"})
		}
		p.Grapes = []game.Grape{{ID: "g0", Color: "red", Value: 4}, {ID: "g1", Color: "white", Value: 3}, {ID: "g2", Color: "red", Value: 2}}
		p.Wines = []game.Wine{{ID: "w9", Type: "red", Value: 9}, {ID: "w7", Type: "white", Value: 7}, {ID: "w1", Type: "red", Value: 1}}
		for f := range p.Fields {
			p.Fields[f].Vines = []game.Card{{ID: fmt.Sprint("v", f), Name: "A", Type: "vine", Red: 1}, {ID: fmt.Sprint("u", f), Name: "B", Type: "vine", White: 1}}
		}
	}
	p := r.Players[0]
	r.TurnID = p.ID
	for _, c := range game.Catalog() {
		if c.ID == id {
			p.Hand = append(p.Hand, c)
		}
	}
	r.Discards["winter"] = []game.Card{{ID: "old-winter", Type: "winter"}}
	r.Discards["summer"] = []game.Card{{ID: "old-summer", Type: "summer"}}
	return r, p
}
