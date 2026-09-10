package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			send := func(id string, a Action) {
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
			send(r.Players[0].ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-23", Slot: 2})
			choose := func(option string, ids []string) {
				q := app.Store.Rooms[r.Code].Choices[0]
				send(q.PlayerID, Action{Type: "choose", ChoiceID: q.ID, Option: option, CardIDs: ids})
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
	send := func(a Action, want int) {
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
	send(Action{Type: "place", Space: "summer_visitor", CardID: "summer-05", Slot: 2}, 200)
	r = app.Store.Rooms[r.Code]
	before, _ := json.Marshal(r)
	c := r.Choices[0]
	send(Action{Type: "choose", ChoiceID: c.ID, Option: "vp_build"}, 400)
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
	send(Action{Type: "choose", ChoiceID: c.ID, Option: "vp_plant"}, 200)
	srv.Close()
	app, e = loadApp(dir)
	if e != nil {
		t.Fatal(e)
	}
	srv = httptest.NewServer(app.routes())
	r = app.Store.Rooms[r.Code]
	c = r.Choices[0]
	send(Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}, 400)
	send(Action{Type: "choose", ChoiceID: c.ID, Option: "plant", CardIDs: []string{"plant"}, Fields: []int{2}}, 200)
	if app.Store.Rooms[r.Code].Context != nil {
		t.Fatal("not completed")
	}
}
