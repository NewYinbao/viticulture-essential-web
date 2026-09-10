package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func fixReload(t *testing.T, r *Room) *Room { return persistedRoom(t, r) }
func fixChoose(r *Room, a Action) error {
	c := r.Choices[0]
	a.Type = "choose"
	a.ChoiceID = c.ID
	return r.Apply(c.PlayerID, a)
}
func fixReject(t *testing.T, r *Room, run func() error) {
	t.Helper()
	before, _ := json.Marshal(r)
	if run() == nil {
		t.Fatal("invalid action accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("rollback changed state")
	}
}
func TestRulesFixGrayWorkerSaveAndYear(t *testing.T) {
	r := eeBegin(t, 3)
	p := r.Players[0]
	r.claimGrayWorker(p)
	before := p.Workers
	r = fixReload(t, r)
	p = r.Players[0]
	q := r.Players[1]
	n := q.Workers
	r.claimGrayWorker(q)
	r.claimGrayWorker(p)
	if q.Workers != n || p.Workers != before || r.GrayWorkerOwner != p.ID {
		t.Fatal("gray worker duplicated after reload")
	}
	r.finishYear()
	if r.GrayWorkerOwner != "" || p.Workers != p.TotalWorkers-1 {
		t.Fatal("gray worker not returned")
	}
	r = fixReload(t, r)
	q = r.Players[1]
	n = q.Workers
	r.claimGrayWorker(q)
	if q.Workers != n+1 || r.GrayWorkerOwner != q.ID {
		t.Fatal("next year reward missing")
	}
}
func TestRulesFixProducerManagerPersistRollback(t *testing.T) {
	r, p := visitorFixture("winter-32")
	p.Hand = append(p.Hand, Card{ID: "summer-32", Type: "summer"})
	for i := range r.Spaces {
		if r.Spaces[i].ID == "winter_visitor" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 2}}
		}
	}
	if e := r.Apply(p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 1, CardID: "winter-32"}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "action", Space: "summer_visitor", CardID: "summer-32"}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	fixReject(t, r, func() error {
		return fixChoose(r, Action{Option: "retrieve", TargetIDs: []string{"winter_visitor", "winter_visitor"}, Fields: []int{2, 1}})
	})
	r = fixReload(t, r)
	p = r.Players[0]
	n := p.Workers
	if e := fixChoose(r, Action{Option: "retrieve", TargetIDs: []string{"winter_visitor"}, Fields: []int{2}}); e != nil {
		t.Fatal(e)
	}
	if p.Workers != n+1 {
		t.Fatal("old seat not retrieved")
	}
	if len(r.Choices) != 1 || r.Choices[0].Visitor.Stage != "second" {
		t.Fatal("outer Manager bonus lost")
	}
}
func TestRulesFixYokeMotivatorReload(t *testing.T) {
	r, p := visitorFixture("winter-25")
	p.Buildings = append(p.Buildings, "yoke")
	p.YokeUsed = true
	p.LargeWorker = false
	for i := range r.Spaces {
		if r.Spaces[i].ID == "yoke" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 1, Large: true}}
		}
	}
	playVisitorFixture(t, r, p, "winter-25")
	chooseVisitor(t, r, "resolve", nil)
	r = fixReload(t, r)
	p = r.Players[0]
	if e := fixChoose(r, Action{Option: "retrieve"}); e != nil {
		t.Fatal(e)
	}
	if p.YokeUsed || !p.LargeWorker {
		t.Fatal("Motivator did not release yoke")
	}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "yoke" {
			if _, _, e := placement(&r.Spaces[i], p, Action{}); e != nil {
				t.Fatal(e)
			}
		}
	}
}
func TestRulesFixAgingBoundaries(t *testing.T) {
	for _, v := range []int{3, 4, 5, 6, 7, 8, 9} {
		p := NewPlayer("p")
		p.Wines = []Wine{{ID: "w", Type: "red", Value: v}}
		ageWines(p)
		want := v
		if v == 4 || v == 5 || v == 7 || v == 8 {
			want++
		}
		if p.Wines[0].Value != want {
			t.Fatalf("%d -> %d want %d", v, p.Wines[0].Value, want)
		}
	}
	r, p := visitorFixture("winter-14")
	p.Coins = 0
	p.Buildings = nil
	p.Wines = []Wine{{ID: "w", Type: "red", Value: 4}}
	p.Hand = []Card{{ID: "winter-14", Type: "winter"}, {ID: "o", Type: "order", Requirements: []Wine{{Type: "red", Value: 5}}}}
	playVisitorFixture(t, r, p, "winter-14")
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "fill", WineIDs: []string{"w"}}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "fill", CardID: "o", WineIDs: []string{"w"}}); e != nil {
		t.Fatal(e)
	}
}
func TestRulesFixMinimumAndHiddenDraw(t *testing.T) {
	for _, id := range []string{"summer-09", "winter-12", "winter-24", "winter-22", "winter-26", "winter-33", "winter-29"} {
		r, p := visitorFixture(id)
		p.Hand = []Card{{ID: id, Type: r.Phase}}
		p.Grapes = nil
		p.Wines = nil
		for i := range p.Fields {
			p.Fields[i].Vines = nil
		}
		fixReject(t, r, func() error {
			return r.Apply(p.ID, Action{Type: "place", Space: r.Phase + "_visitor", CardID: id, Slot: 2})
		})
	}
	r, p := visitorFixture("winter-02")
	playVisitorFixture(t, r, p, "winter-02")
	p.Grapes = nil
	fixReject(t, r, func() error { return fixChoose(r, Action{Option: "make"}) })
	r, p = visitorFixture("summer-11")
	p.Hand = []Card{{ID: "summer-11", Type: "summer"}}
	p.Buildings = nil
	r.Decks["vine"] = []Card{{ID: "unplantable", Type: "vine", Red: 4, Trellis: true}}
	playVisitorFixture(t, r, p, "summer-11")
	if e := fixChoose(r, Action{Option: "draw"}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "plant"}); e != nil {
		t.Fatal("optional post-draw softlock", e)
	}
}
func TestRulesFixDeclineBonus(t *testing.T) {
	for _, space := range []string{"draw_vine", "tour", "build", "sell_grapes"} {
		r := eeBegin(t, 3)
		p := r.Player(r.TurnID)
		p.Coins = 10
		p.Buildings = nil
		before := len(p.Hand)
		a := Action{Type: "place", Space: space, Slot: 1, DeclineBonus: true, Building: "trellis", Mode: "sell_field", Field: 0}
		if e := r.Apply(p.ID, a); e != nil {
			t.Fatal(e)
		}
		switch space {
		case "draw_vine":
			if len(p.Hand) != before+1 {
				t.Fatal("forced draw bonus")
			}
		case "tour":
			if p.Coins != 12 {
				t.Fatal("forced coin bonus")
			}
		case "build":
			if p.Coins != 8 {
				t.Fatal("forced discount")
			}
		case "sell_grapes":
			if p.VP != 0 {
				t.Fatal("forced VP bonus")
			}
		}
	}
	r, p := visitorFixture("winter-05")
	if e := r.Apply(p.ID, Action{Type: "place", Space: "winter_visitor", CardID: "winter-05", Slot: 1, DeclineBonus: true}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "draw"}); e != nil {
		t.Fatal(e)
	}
	if len(r.Choices) != 0 {
		t.Fatal("declined second visitor still queued")
	}
	r, p = visitorFixture("winter-05")
	r.Planned = []PlannedPlacement{{PlayerID: p.ID, Space: "draw_order", Slot: 1}}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "draw_order" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 1}}
		}
	}
	r.startNextPlanner()
	r = fixReload(t, r)
	p = r.Players[0]
	n := len(p.Hand)
	if e := fixChoose(r, Action{Option: "execute", DeclineBonus: true}); e != nil {
		t.Fatal(e)
	}
	if len(p.Hand) != n+1 {
		t.Fatal("Planner forced bonus")
	}
}

func TestRulesFixProducerYokeSaveRollback(t *testing.T) {
	r, p := visitorFixture("summer-32")
	p.Buildings = append(p.Buildings, "yoke")
	p.YokeUsed = true
	p.Fields[0].Harvested = true
	for i := range r.Spaces {
		if r.Spaces[i].ID == "yoke" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 1}}
		}
	}
	playVisitorFixture(t, r, p, "summer-32")
	r = fixReload(t, r)
	fixReject(t, r, func() error {
		return fixChoose(r, Action{Option: "retrieve", TargetIDs: []string{"yoke", "yoke"}, Fields: []int{1, 1}})
	})
	r = fixReload(t, r)
	if e := fixChoose(r, Action{Option: "retrieve", TargetIDs: []string{"yoke"}, Fields: []int{1}}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	p = r.Players[0]
	if p.YokeUsed || !p.Fields[0].Harvested {
		t.Fatal("yoke release changed annual field restriction")
	}
	r.TurnID = p.ID
	fixReject(t, r, func() error { return r.Apply(p.ID, Action{Type: "place", Space: "yoke", Mode: "harvest", Field: 0}) })
	if e := r.Apply(p.ID, Action{Type: "place", Space: "yoke", Mode: "harvest", Field: 1}); e != nil {
		t.Fatal(e)
	}
}
func TestRulesFixGrayWakeOrganizerActual(t *testing.T) {
	r, p := visitorFixture("summer-33")
	r.Phase = "wake"
	for _, q := range r.Players {
		q.Wake = 0
	}
	r.WakeSlots = nil
	for i := 1; i <= 7; i++ {
		r.WakeSlots = append(r.WakeSlots, WakeSlot{Slot: i})
	}
	before := p.Workers
	if e := r.Apply(p.ID, Action{Type: "wake", Slot: 7}); e != nil {
		t.Fatal(e)
	}
	if p.Workers != before+1 || r.GrayWorkerOwner != p.ID {
		t.Fatal("spring did not claim unique gray")
	}
	r = fixReload(t, r)
	p = r.Players[0]
	r.Phase = "summer"
	r.TurnID = p.ID
	r.Players[1].Wake = 1
	r.Players[2].Wake = 2
	playVisitorFixture(t, r, p, "summer-33")
	if e := fixChoose(r, Action{Option: "move", Slot: 4}); e != nil {
		t.Fatal(e)
	}
	r = fixReload(t, r)
	q := r.Players[1]
	q.Hand = append(q.Hand, Card{ID: "summer-33", Type: "summer"})
	r.Discards["summer"] = nil
	r.TurnID = q.ID
	before = q.Workers
	if e := r.Apply(q.ID, Action{Type: "place", Space: "summer_visitor", Slot: 1, Large: true, CardID: "summer-33"}); e != nil {
		t.Fatal(e)
	}
	if e := fixChoose(r, Action{Option: "move", Slot: 7}); e != nil {
		t.Fatal(e)
	}
	if q.Workers != before {
		t.Fatal("Organizer minted another gray after save")
	}
}
