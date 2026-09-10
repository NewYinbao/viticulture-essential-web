package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func eeRoom(n int) *Room {
	r := &Room{Phase: "lobby", Code: "EE"}
	for i := 0; i < n; i++ {
		r.Players = append(r.Players, makePlayer("p"))
	}
	r.HostID = r.Players[0].ID
	return r
}
func eeBegin(t *testing.T, n int) *Room {
	t.Helper()
	r := eeRoom(n)
	if e := r.apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		if e := r.apply(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: "coins"}); e != nil {
			t.Fatal(e)
		}
	}
	for slot := 1; r.Phase == "wake"; slot++ {
		if e := r.apply(r.TurnID, Action{Type: "wake", Slot: slot, Color: "summer"}); e != nil {
			t.Fatal(e)
		}
	}
	return r
}
func TestEEFixedCatalog(t *testing.T) {
	r := eeRoom(2)
	r.initDecks()
	for typ, n := range map[string]int{"vine": 42, "order": 36, "summer": 38, "winter": 38, "mama": 18, "papa": 18} {
		if len(r.Decks[typ]) != n {
			t.Fatalf("%s=%d", typ, len(r.Decks[typ]))
		}
	}
	seen := map[string]bool{}
	queen := false
	for _, c := range catalog() {
		if seen[c.ID] {
			t.Fatal("duplicate ID")
		}
		seen[c.ID] = true
		if c.Name == "Queen" {
			queen = true
		}
		if c.Type == "order" && (len(c.Requirements) == 0 || c.Points < 2 || c.Income < 1) {
			t.Fatal(c)
		}
		if c.ID == "vine-09" && (!c.Trellis || c.Irrigation) {
			t.Fatal("Pinot facilities")
		}
		if c.ID == "vine-25" && (c.Trellis || !c.Irrigation) {
			t.Fatal("Merlot facilities")
		}
	}
	if !queen {
		t.Fatal("missing Queen")
	}
}
func TestEEFiniteDeckAndPrivacy(t *testing.T) {
	r := eeRoom(2)
	r.initDecks()
	p := r.Players[0]
	for i := 0; i < 43; i++ {
		r.draw(p, "vine")
	}
	if len(p.Hand) != 42 {
		t.Fatal(len(p.Hand))
	}
	c := p.Hand[0]
	p.Hand = p.Hand[1:]
	r.discard(c)
	r.draw(p, "vine")
	if len(p.Hand) != 42 || len(r.Discards["vine"]) != 0 {
		t.Fatal("reshuffle")
	}
	r.draw(r.Players[1], "order")
	b, _ := json.Marshal(r.view(p.ID))
	for _, card := range r.Players[1].Hand {
		if bytes.Contains(b, []byte(card.ID)) {
			t.Fatal("private hand leaked")
		}
	}
	for _, d := range r.Decks {
		for _, card := range d {
			if bytes.Contains(b, []byte(card.ID)) {
				t.Fatal("deck sequence leaked")
			}
		}
	}
}
func TestEEParentsAllOptions(t *testing.T) {
	for i, g := range papaGifts {
		for _, option := range []string{"gift", "coins"} {
			r := eeRoom(2)
			p := r.Players[0]
			for _, c := range catalog() {
				if c.Type == "papa" && cardNumber(c) == i {
					p.Papa = c
				}
			}
			p.Coins = g.Cash
			r.enqueue(p.ID, "papa", []string{"gift", "coins"}, 1)
			c := r.Choices[0]
			if e := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: option}); e != nil {
				t.Fatal(e)
			}
			if !p.PapaResolved {
				t.Fatal("unresolved")
			}
			if option == "coins" && p.Coins != g.Cash+g.Alternative {
				t.Fatal("cash")
			}
			if option == "gift" {
				switch g.Gift {
				case "vp":
					if p.VP != 1 {
						t.Fatal("vp")
					}
				case "worker":
					if p.TotalWorkers != 4 || p.Workers != 3 {
						t.Fatal("worker")
					}
				default:
					if !has(p, g.Gift) {
						t.Fatal("building")
					}
				}
			}
		}
	}
}
func TestEEChoiceSaveResume(t *testing.T) {
	a, _ := loadApp(t.TempDir())
	r := eeRoom(6)
	if e := r.apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	if len(r.Choices) != 6 || r.Phase != "setup" {
		t.Fatal("setup")
	}
	seen := map[string]bool{}
	for _, p := range r.Players {
		if seen[p.Mama.ID] || seen[p.Papa.ID] {
			t.Fatal("parents not finite")
		}
		seen[p.Mama.ID] = true
		seen[p.Papa.ID] = true
		n := 3
		if cardNumber(p.Mama) >= 16 {
			n = 2
		}
		if len(p.Hand) != n {
			t.Fatal("mama cards")
		}
	}
	a.Store.Rooms[r.Code] = r
	if e := a.save(); e != nil {
		t.Fatal(e)
	}
	a2, e := loadApp(a.dir)
	if e != nil {
		t.Fatal(e)
	}
	q := a2.Store.Rooms[r.Code]
	if q.Choices[0].ID != r.Choices[0].ID || len(q.Decks["mama"]) != 12 {
		t.Fatal("restore")
	}
	c := q.Choices[0]
	if q.apply(q.Players[1].ID, Action{Type: "choose", ChoiceID: c.ID, Option: "gift"}) == nil {
		t.Fatal("choice ownership")
	}
}
func TestEEBonusGrandeAndInfiniteCoin(t *testing.T) {
	r := eeBegin(t, 3)
	p := r.player(r.TurnID)
	before := p.Coins
	if e := r.apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, Large: true}); e != nil {
		t.Fatal(e)
	}
	if p.Coins != before+3 {
		t.Fatal("bonus")
	}
	p = r.player(r.TurnID)
	if e := r.apply(p.ID, Action{Type: "place", Space: "tour", Slot: 2}); e != nil {
		t.Fatal(e)
	}
	p = r.player(r.TurnID)
	before = p.Coins
	if e := r.apply(p.ID, Action{Type: "place", Space: "tour", Large: true}); e != nil {
		t.Fatal(e)
	}
	if p.Coins != before+2 {
		t.Fatal("overflow bonus")
	}
	for i := 0; i < 3; i++ {
		if e := r.apply(r.TurnID, Action{Type: "place", Space: "gain_coin"}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestEEFallAndDiscardTie(t *testing.T) {
	r := eeBegin(t, 2)
	r.Players[0].Buildings = append(r.Players[0].Buildings, "cottage")
	for i := 0; i < 2; i++ {
		if e := r.apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Phase != "fall" || len(r.Choices) != 2 {
		t.Fatal("fall")
	}
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		if e := r.apply(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: func() string {
			if c.Count == 2 {
				return "winter_winter"
			}
			return "winter"
		}()}); e != nil {
			t.Fatal(e)
		}
	}
	for _, p := range r.Players {
		p.VP = 20
		p.Coins = 0
		p.Income = 0
		for len(p.Hand) < 9 {
			r.draw(p, "vine")
		}
	}
	for i := 0; i < 2; i++ {
		if e := r.apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Phase != "year_end" {
		t.Fatal(r.Phase)
	}
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		p := r.player(c.PlayerID)
		ids := []string{}
		for _, card := range p.Hand[:c.Count] {
			ids = append(ids, card.ID)
		}
		if e := r.apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, CardIDs: ids}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Phase != "finished" || len(r.WinnerIDs) != 2 {
		t.Fatal("shared win", r.WinnerIDs)
	}
}
func TestEEMultipleWineAndFieldYoke(t *testing.T) {
	r := eeRoom(2)
	p := r.Players[0]
	p.Coins = 0
	if e := r.performPlacement(p, Action{Space: "sell_grapes", Mode: "sell_field", Field: 0}, false); e != nil {
		t.Fatal(e)
	}
	if !p.Fields[0].Sold || p.Coins != 5 {
		t.Fatal("sale")
	}
	if e := r.performPlacement(p, Action{Space: "sell_grapes", Mode: "buy_field", Field: 0}, false); e != nil {
		t.Fatal(e)
	}
	p.Grapes = []Grape{{ID: "a", Color: "red", Value: 3}, {ID: "b", Color: "white", Value: 2}}
	if e := makeWines(p, [][]int{{0}, {1}}, 2); e != nil {
		t.Fatal(e)
	}
	if len(p.Wines) != 2 || len(p.Grapes) != 0 {
		t.Fatal("multi wine")
	}
	p.Fields[0].Vines = []Card{{ID: "v", Type: "vine", Red: 1}}
	if e := uproot(p, Action{Field: 0, CardID: "v"}); e != nil {
		t.Fatal(e)
	}
	if len(p.Hand) != 1 || len(p.Fields[0].Vines) != 0 {
		t.Fatal("uproot")
	}
}
func TestEEVisitorHTTPAtomicReject(t *testing.T) {
	r := eeBegin(t, 2)
	a, _ := loadApp(t.TempDir())
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["ee"] = Session{Code: r.Code, PlayerID: r.TurnID}
	before, _ := json.Marshal(r)
	body, _ := json.Marshal(Action{Type: "place", Space: "summer_visitor", Revision: r.Revision})
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

func TestEEOrderDescendingSameColor(t *testing.T) {
	for _, ids := range [][]string{{"high", "low"}, {"low", "high"}} {
		p := makePlayer("order")
		c := Card{ID: "order", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}, {Type: "red", Value: 5}}, Points: 3, Income: 1}
		p.Hand = []Card{c}
		p.Wines = []Wine{{ID: "high", Type: "red", Value: 5}, {ID: "low", Type: "red", Value: 2}}
		if e := fillOrder(p, Action{CardID: c.ID, WineIDs: ids}); e != nil {
			t.Fatal("valid same-color matching rejected", e)
		}
		if len(p.Wines) != 0 || len(p.Hand) != 0 || p.VP != 3 || p.Income != 1 {
			t.Fatal("order consumption/rewards")
		}
	}
}
func TestEENeutralWorkerReturnedAtYearEnd(t *testing.T) {
	r := eeBegin(t, 2)
	// Start a fresh wake selection, then take the temporary worker honestly.
	r.spring()
	if e := r.apply(r.TurnID, Action{Type: "wake", Slot: 7}); e != nil {
		t.Fatal(e)
	}
	p := r.player(r.WakeSlots[6].PlayerID)
	if p.Workers != 3 || p.TotalWorkers != 3 {
		t.Fatal("neutral worker must not increase permanent total")
	}
	if e := r.apply(r.TurnID, Action{Type: "wake", Slot: 1}); e != nil {
		t.Fatal(e)
	}
	for r.Year == 1 {
		resolveTestChoices(t, r)
		if r.Year != 1 {
			break
		}
		if e := r.apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if p.Workers != 2 || p.TotalWorkers != 3 || !p.LargeWorker {
		t.Fatal("neutral worker not returned")
	}
	for _, s := range r.Spaces {
		if len(s.Occupied) != 0 {
			t.Fatal("workers not cleared")
		}
	}
}
func TestEEParentChoiceView(t *testing.T) {
	r := eeRoom(2)
	if e := r.apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	c := r.Choices[0]
	p := r.player(c.PlayerID)
	g := papaGifts[cardNumber(p.Papa)]
	own := r.view(p.ID)
	options := own["parentOptions"].(map[string]any)
	if options["gift"] != g.Gift || options["coins"] != g.Alternative || options["baseCoins"] != g.Cash {
		t.Fatal("missing parent details")
	}
	if own["pendingChoice"].(Choice).ID != c.ID || !own["legal"].(map[string]bool)["canChoose"] {
		t.Fatal("owner cannot choose")
	}
	other := r.view(r.Players[1].ID)
	b, _ := json.Marshal(other["pendingChoice"])
	if bytes.Contains(b, []byte(c.ID)) || other["legal"].(map[string]bool)["canChoose"] {
		t.Fatal("choice authority leaked")
	}
}
func TestEEWinterVisitorHTTPAtomicReject(t *testing.T) {
	r := eeBegin(t, 2)
	for i := 0; i < 2; i++ {
		if e := r.apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	resolveTestChoices(t, r)
	a, _ := loadApp(t.TempDir())
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["winter"] = Session{Code: r.Code, PlayerID: r.TurnID}
	before, _ := json.Marshal(r)
	w := request(a.routes(), "POST", "/api/action", "winter", Action{Type: "place", Space: "winter_visitor", Revision: r.Revision})
	after, _ := json.Marshal(a.Store.Rooms[r.Code])
	if w.Code != 400 || !strings.Contains(w.Body.String(), "非自己的当季访客") || !bytes.Equal(before, after) {
		t.Fatal("winter visitor rejection must be explicit and atomic", w.Body.String())
	}
}

func TestEEInvalidWakeColorDoesNotClaimSlot(t *testing.T) {
	r := eeRoom(2)
	if e := r.apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	resolveTestChoices(t, r)
	before, _ := json.Marshal(r)
	if r.apply(r.TurnID, Action{Type: "wake", Slot: 5}) == nil {
		t.Fatal("missing color accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid wake claimed slot")
	}
}
func TestEEFinalYearReturnsPlacedWorkers(t *testing.T) {
	r := eeBegin(t, 2)
	if e := r.apply(r.TurnID, Action{Type: "place", Space: "tour"}); e != nil {
		t.Fatal(e)
	}
	r.Players[0].VP = 20
	r.finishYear()
	if r.Phase != "finished" {
		t.Fatal("not finished")
	}
	for _, s := range r.Spaces {
		if len(s.Occupied) > 0 {
			t.Fatal("final year left placed workers")
		}
	}
	for _, p := range r.Players {
		if p.Workers != p.TotalWorkers-1 || !p.LargeWorker {
			t.Fatal("final recall")
		}
	}
}
