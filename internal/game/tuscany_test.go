package game

import (
	"encoding/json"
	"testing"
)

// These unit fixtures exercise implementation under development. They do not
// bypass the production start gate or declare the whole module available.
func tuscanyRoom(n int) *Room {
	r := &Room{Config: ExpansionConfig{Board: "tuscany", Visitors: "ee"}, Year: 1, Phase: "spring", Decks: map[string][]Card{}, Discards: map[string][]Card{}}
	for i := 0; i < n; i++ {
		p := NewPlayer("test")
		p.Wake = i + 2
		p.Season = "spring"
		r.Players = append(r.Players, p)
	}
	r.HostID = r.Players[0].ID
	r.TurnID = r.HostID
	r.Spaces = r.spacesForRules()
	for i := 1; i <= 7; i++ {
		r.WakeSlots = append(r.WakeSlots, WakeSlot{Slot: i})
	}
	for _, p := range r.Players {
		r.WakeSlots[p.Wake-1].PlayerID = p.ID
	}
	return r
}

func mustApplyT(t *testing.T, r *Room, id string, a Action) {
	t.Helper()
	if err := r.Apply(id, a); err != nil {
		t.Fatalf("%+v: %v", a, err)
	}
}
func choiceT(t *testing.T, r *Room, a Action) {
	t.Helper()
	if len(r.Choices) == 0 {
		t.Fatal("no choice")
	}
	c := r.Choices[0]
	a.Type = "choose"
	a.ChoiceID = c.ID
	if a.Option == "" && len(c.Options) > 0 {
		a.Option = c.Options[0]
	}
	mustApplyT(t, r, c.PlayerID, a)
}
func restoreT(t *testing.T, r *Room) *Room {
	t.Helper()
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	var q Room
	if e = json.Unmarshal(b, &q); e != nil {
		t.Fatal(e)
	}
	return &q
}
func rejectUnchangedT(t *testing.T, r *Room, id string, a Action) {
	t.Helper()
	b, _ := json.Marshal(r)
	if r.Apply(id, a) == nil {
		t.Fatal("invalid action accepted", a)
	}
	c, _ := json.Marshal(r)
	if string(b) != string(c) {
		t.Fatal("rejected action changed state")
	}
}

func TestTuscanyPrintedSlots(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5, 6} {
		r := tuscanyRoom(n)
		cap := 1
		if n >= 3 {
			cap = 2
		}
		if n >= 5 {
			cap = 3
		}
		counts := map[string]int{}
		for i := range r.Spaces {
			s := &r.Spaces[i]
			if s.Season == "any" {
				continue
			}
			counts[s.Season]++
			if s.Capacity != cap {
				t.Fatal(n, s)
			}
			for slot := 1; slot <= cap; slot++ {
				_, bonus, e := placement(s, r.Players[0], Action{Slot: slot})
				if e != nil || bonus != (s.BonusSlots[slot] != "") {
					t.Fatal(n, s.ID, slot, bonus, e)
				}
			}
			s.Occupied = nil
			for slot := 1; slot <= cap; slot++ {
				s.Occupied = append(s.Occupied, Seat{Slot: slot, PlayerID: r.Players[1].ID})
			}
			if _, _, e := placement(s, r.Players[0], Action{}); e == nil {
				t.Fatal("ordinary overflow")
			}
			if slot, bonus, e := placement(s, r.Players[0], Action{Large: true}); e != nil || slot != -1 || bonus {
				t.Fatal("grande overflow", s.ID)
			}
		}
		for _, season := range tuscanySeasons {
			if counts[season] != 4 {
				t.Fatal(counts)
			}
		}
	}
}

func TestTuscanyWakeAndIndividualPass(t *testing.T) {
	r := tuscanyRoom(2)
	r.tuscanySpring()
	p, q := r.Players[0], r.Players[1]
	rejectUnchangedT(t, r, p.ID, Action{Type: "wake", Slot: 1})
	mustApplyT(t, r, p.ID, Action{Type: "wake", Slot: 2})
	mustApplyT(t, r, q.ID, Action{Type: "wake", Slot: 5})
	if p.Coins != 0 || len(q.Hand) != 0 || r.Phase != "spring" {
		t.Fatal("wake gave rewards")
	}
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	if p.Coins != 1 || p.Season != "summer" || r.Phase != "spring" || r.TurnID != q.ID {
		t.Fatal("individual pass ordering")
	}
	r.Decks["winter"] = []Card{{ID: "private-winter", Type: "winter"}}
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	r = restoreT(t, r)
	if r.Phase != "spring" || r.Choices[0].Kind != "tuscany_draw" {
		t.Fatal("choice must finish before advancing")
	}
	choiceT(t, r, Action{Option: "winter"})
	if r.Phase != "summer" || r.TurnID != p.ID || len(r.Player(q.ID).Hand) != 1 {
		t.Fatal("summer entry")
	}
	if r.Player(p.ID).Workers != 2 {
		t.Fatal("season wrongly refreshed workers")
	}
}

func TestTuscanyWinterReturnUpkeepAndWakeReservation(t *testing.T) {
	r := tuscanyRoom(2)
	p, q := r.Players[0], r.Players[1]
	r.Phase = "winter"
	p.Season = "winter"
	q.Season = "winter"
	p.Coins = 1
	p.Income = 2
	q.Coins = 3
	p.Workers = 0
	p.LargeWorker = false
	p.Grapes = []Grape{{ID: "g", Color: "red", Value: 2}}
	p.Wines = []Wine{{ID: "w", Type: "red", Value: 3}}
	for _, s := range []string{"spring", "summer", "fall", "winter"} {
		for i := range r.Spaces {
			if r.Spaces[i].Season == s {
				r.Spaces[i].Occupied = append(r.Spaces[i].Occupied, Seat{Slot: 1, PlayerID: p.ID})
			}
		}
	}
	for i := 0; i < 9; i++ {
		p.Hand = append(p.Hand, Card{ID: NewID(), Type: "vine"})
	}
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	if p.Workers != 2 || !p.LargeWorker || p.Grapes[0].Value != 3 || p.Wines[0].Value != 3 || p.Coins != 1 {
		t.Fatal("retrieve/age must precede discard and income")
	}
	for _, s := range r.Spaces {
		if len(s.Occupied) != 0 {
			t.Fatal("worker not returned", s)
		}
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	q = r.Player(q.ID)
	choiceT(t, r, Action{CardIDs: []string{p.Hand[0].ID, p.Hand[1].ID}})
	choiceT(t, r, Action{})
	if p.Coins != 3 || r.Choices[0].Kind != "tuscany_next_wake" {
		t.Fatal("income order")
	}
	choiceT(t, r, Action{Option: "6"})
	if r.TurnID != q.ID || r.Phase != "winter" || p.Wake != 2 || p.NextWake != 6 {
		t.Fatal("premature year transition")
	}
	mustApplyT(t, r, q.ID, Action{Type: "place", Space: "train", Slot: 1}) // funds needed below are supplied before this action in fixture
	if q.Coins != 0 || q.TotalWorkers != 4 || q.Workers != 1 {
		t.Fatal("released winter slot or discount")
	}
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	choiceT(t, r, Action{})
	c := r.Choices[0]
	rejectUnchangedT(t, r, q.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "6"})
	choiceT(t, r, Action{Option: "4"})
	if r.Year != 2 || r.Phase != "spring" || r.TurnID != q.ID || p.Wake != 6 || q.Wake != 4 {
		t.Fatal("next year wake order")
	}
}

func visitorCardT(t *testing.T, id string) Card {
	t.Helper()
	for _, c := range Catalog() {
		if c.ID == id {
			return c
		}
	}
	t.Fatal(id)
	return Card{}
}

func TestTuscanyPlannerFallAndWinterReservations(t *testing.T) {
	r := tuscanyRoom(2)
	p, q := r.Players[0], r.Players[1]
	r.Phase = "summer"
	p.Season = "summer"
	q.Season = "summer"
	p.Hand = []Card{visitorCardT(t, "summer-29")}
	p.Fields[0].Vines = []Card{{ID: "v", Type: "vine", Red: 2}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-29"})
	choiceT(t, r, Action{Option: "plan", Space: "harvest", Slot: 1})
	if p.Fields[0].Harvested || len(p.Grapes) > 0 || len(r.Planned) != 1 {
		t.Fatal("plan executed early")
	}
	r.Planned = append(r.Planned, PlannedPlacement{PlayerID: q.ID, Space: "train", Slot: 1})
	for i := range r.Spaces {
		if r.Spaces[i].ID == "train" {
			r.Spaces[i].Occupied = append(r.Spaces[i].Occupied, Seat{PlayerID: q.ID, Slot: 1})
		}
	}
	q.Coins = 3
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	if r.Phase != "fall" || r.Choices[0].Kind != "planner" {
		t.Fatal("missing fall planner")
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	q = r.Player(q.ID)
	choiceT(t, r, Action{Option: "execute", Fields: []int{0}})
	if !p.Fields[0].Harvested || len(r.Planned) != 1 || r.Planned[0].Space != "train" {
		t.Fatal("future reservation consumed")
	}
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	if r.Phase != "winter" || r.Choices[0].Kind != "planner" {
		t.Fatal("missing winter planner")
	}
	choiceT(t, r, Action{Option: "execute"})
	if q.Coins != 0 || q.TotalWorkers != 4 || len(r.Planned) != 0 {
		t.Fatal("planner physical bonus")
	}
}

func TestTuscanyOrganizerChoicePreservesOuterVisitor(t *testing.T) {
	r := tuscanyRoom(3)
	p := r.Players[0]
	r.Phase = "summer"
	p.Hand = []Card{visitorCardT(t, "summer-33"), visitorCardT(t, "summer-01")}
	r.Decks["summer"] = []Card{{ID: "drawn-private", Type: "summer"}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", Slot: 2, CardID: "summer-33"})
	choiceT(t, r, Action{Option: "move", Slot: 5, Color: "summer"})
	if r.Choices[0].Kind != "tuscany_draw" || p.Passed {
		t.Fatal("Organizer passed before rewards/outer visitor")
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	choiceT(t, r, Action{Option: "summer"})
	if r.Choices[0].Visitor == nil || r.Choices[0].Visitor.Stage != "second" {
		t.Fatal("lost outer visitor")
	}
	choiceT(t, r, Action{Option: "skip"})
	if !p.Passed || p.Season != "fall" || p.Wake != 5 || r.Phase != "summer" {
		t.Fatal("Organizer pass timing")
	}
}

func TestTuscanyManagerPreviousActionNoWorkerOrBonus(t *testing.T) {
	r := tuscanyRoom(2)
	p := r.Players[0]
	r.Phase = "winter"
	p.Hand = []Card{visitorCardT(t, "winter-32")}
	r.Decks["vine"] = []Card{{ID: "one", Type: "vine"}, {ID: "two", Type: "vine"}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "winter_visitor", CardID: "winter-32", DeclineBonus: true})
	choiceT(t, r, Action{Option: "action", Space: "draw_vine"})
	if len(p.Hand) != 1 || p.Workers != 1 {
		t.Fatal("Manager got bonus or spent extra worker")
	}
	for _, sp := range r.Spaces {
		if sp.ID == "draw_vine" && len(sp.Occupied) > 0 {
			t.Fatal("Manager placed worker")
		}
	}
}

func TestTuscanyTradeAndWineBoundaries(t *testing.T) {
	r := tuscanyRoom(2)
	p := r.Players[0]
	r.Phase = "summer"
	p.Coins = 3
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "trade", Trades: []Trade{{Give: "coins", Receive: "vp"}}})
	if p.Coins != 0 || p.VP != 1 {
		t.Fatal("trade amounts")
	}
	r.TurnID = p.ID
	p.Workers = 2
	p.VP = -5
	r.Spaces = NewTuscanySpaces(2)
	rejectUnchangedT(t, r, p.ID, Action{Type: "place", Space: "trade", Trades: []Trade{{Give: "vp", Receive: "coins"}}})
	p.Grapes = []Grape{{ID: "old", Color: "red", Value: 9}}
	r.TurnID = p.ID
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "trade", Trades: []Trade{{Give: "grape", GrapeID: "old", Receive: "grape", Color: "white"}}})
	if len(p.Grapes) != 1 || p.Grapes[0].Color != "white" || p.Grapes[0].Value != 1 {
		t.Fatal("trade grape value")
	}
	for _, typ := range []string{"red", "white", "blush", "sparkling"} {
		r := tuscanyRoom(2)
		p := r.Players[0]
		r.Phase = "winter"
		p.Wines = []Wine{{ID: "sell", Type: typ, Value: 9}}
		mustApplyT(t, r, p.ID, Action{Type: "place", Space: "sell_wine", WineIDs: []string{"sell"}})
		if p.VP != map[string]int{"red": 1, "white": 1, "blush": 2, "sparkling": 4}[typ] || len(p.Wines) != 0 {
			t.Fatal("wine scoring", typ)
		}
	}
}

func TestTuscanyInfluenceFiniteAndScoring(t *testing.T) {
	r := tuscanyRoom(2)
	p, q := r.Players[0], r.Players[1]
	for i := 0; i < 6; i++ {
		if e := r.influence(p, InfluenceMove{To: "pisa"}); e != nil {
			t.Fatal(e)
		}
	}
	if p.Coins != 12 {
		t.Fatal(p.Coins)
	}
	if e := r.influence(p, InfluenceMove{From: "pisa", To: "siena"}); e != nil {
		t.Fatal(e)
	}
	if p.Coins != 12 || p.Influence["siena"] != 1 {
		t.Fatal("moving must not reward")
	}
	if e := r.influence(p, InfluenceMove{To: "pisa"}); e == nil {
		t.Fatal("seventh star")
	}
	q.Influence = map[string]int{"pisa": 5, "firenze": 1}
	r.scoreInfluence()
	if p.VP != 2 || q.VP != 2 {
		t.Fatal("tie or unique majority scoring", p.VP, q.VP)
	}
}

func TestTuscanyFinalThresholdAndNoDoubleUpkeep(t *testing.T) {
	for _, vp := range []int{20, 24, 25, 40} {
		r := tuscanyRoom(2)
		r.Phase = "winter"
		p := r.Players[0]
		p.VP = vp
		for _, q := range r.Players {
			q.Passed = true
			q.NextWake = q.Wake
			q.Coins = 8
			q.Income = 3
			q.Grapes = []Grape{{Color: "red", Value: 2}}
		}
		p.Influence = map[string]int{"siena": 1}
		r.tuscanyFinishYear()
		if (r.Phase == "finished") != (vp >= 25) {
			t.Fatal(vp, r.Phase)
		}
		if p.Coins != 8 || p.Grapes[0].Value != 2 {
			t.Fatal("upkeep applied twice")
		}
		if vp < 25 && p.VP != vp {
			t.Fatal("influence scored before final year")
		}
		if vp >= 25 && p.VP != vp+2 {
			t.Fatal("influence missing")
		}
	}
}

func TestTuscanySequentialBonusesAndRestoration(t *testing.T) {
	t.Run("second trade can spend cards from first trade", func(t *testing.T) {
		r := tuscanyRoom(3)
		p := r.Players[0]
		r.Phase = "summer"
		p.Coins = 3
		r.Decks["winter"] = []Card{{ID: "trade-a", Type: "winter"}, {ID: "trade-b", Type: "winter"}}
		mustApplyT(t, r, p.ID, Action{Type: "place", Space: "trade", Slot: 2, Trades: []Trade{{Give: "coins", Receive: "cards", Colors: []string{"winter", "winter"}}}})
		if len(p.Hand) != 2 || r.Choices[0].Kind != "tuscany_trade" || r.TurnID != p.ID {
			t.Fatal("second trade did not wait for first trade")
		}
		r = restoreT(t, r)
		p = r.Player(p.ID)
		c := r.Choices[0]
		rejectUnchangedT(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "trade", Trades: []Trade{{Give: "cards", CardIDs: []string{p.Hand[0].ID, p.Hand[0].ID}, Receive: "vp"}}})
		choiceT(t, r, Action{Option: "trade", Trades: []Trade{{Give: "cards", CardIDs: []string{p.Hand[0].ID, p.Hand[1].ID}, Receive: "vp"}}})
		if p.VP != 1 || p.Coins != 0 || len(p.Hand) != 0 || len(r.Discards["winter"]) != 2 || p.Workers != 1 || r.TurnID == p.ID || r.Context != nil {
			t.Fatal("trade continuation, payment or cleanup")
		}
	})
	t.Run("sixth star then bonus movement", func(t *testing.T) {
		r := tuscanyRoom(3)
		p := r.Players[0]
		p.Influence = map[string]int{"pisa": 5}
		mustApplyT(t, r, p.ID, Action{Type: "place", Space: "influence", Slot: 2, Influence: []InfluenceMove{{To: "siena"}}})
		if p.Coins != 1 || r.Choices[0].Kind != "tuscany_influence" {
			t.Fatal("sixth star must reward before bonus")
		}
		r = restoreT(t, r)
		p = r.Player(p.ID)
		c := r.Choices[0]
		rejectUnchangedT(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "place", Influence: []InfluenceMove{{To: "pisa"}}})
		choiceT(t, r, Action{Option: "place", Influence: []InfluenceMove{{From: "siena", To: "pisa"}}})
		if p.Coins != 1 || p.Influence["pisa"] != 6 || p.Influence["siena"] != 0 || r.TurnID == p.ID {
			t.Fatal("bonus created seventh star or movement reward")
		}
	})
	t.Run("coin before or after visitor and after save", func(t *testing.T) {
		for _, first := range []bool{false, true} {
			r := tuscanyRoom(2)
			p := r.Players[0]
			r.Phase = "winter"
			p.Coins = 1
			p.Hand = []Card{visitorCardT(t, "winter-34")}
			mustApplyT(t, r, p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 1, BonusFirst: first, CardID: "winter-34"})
			wantCoins := 1
			if first {
				wantCoins++
			}
			if p.Coins != wantCoins {
				t.Fatal("bonus paid at wrong time", first, p.Coins)
			}
			r = restoreT(t, r)
			p = r.Player(p.ID)
			choiceT(t, r, Action{Option: "buy"})
			if p.Coins != 1 || p.Income != 1 || r.Context != nil || r.TurnID == p.ID {
				t.Fatal("deferred coin lost or paid twice", first)
			}
		}
	})
}
