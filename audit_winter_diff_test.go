//go:build audit_diff

package main

import "testing"

// These assertions encode the rules/reference expectation; known differences fail.
// No production, existing tests, HTTP server or save files are written.
func auditWinterChoice(t *testing.T, r *Room, p *Player, a Action) error {
	t.Helper()
	a.Type = "choose"
	a.ChoiceID = r.Choices[0].ID
	return r.apply(p.ID, a)
}
func auditWinterStart(t *testing.T, r *Room, p *Player, id string) {
	t.Helper()
	if e := r.apply(p.ID, Action{Type: "place", Space: "winter_visitor", CardID: id}); e != nil {
		t.Fatal(e)
	}
}
func TestAuditWinterDiffHarvestExpert(t *testing.T) {
	r, p := visitorFixture("winter-18")
	auditWinterStart(t, r, p, "winter-18")
	before := len(p.Hand)
	if e := auditWinterChoice(t, r, p, Action{Option: "harvest", Field: 0}); e != nil {
		t.Fatal(e)
	}
	if !p.Fields[0].Harvested {
		t.Fatal("harvest did not occur")
	}
	if len(p.Hand) != before+1 {
		t.Errorf("Harvest Expert: expected harvest AND draw 1 vine, hand delta=%d", len(p.Hand)-before)
	}
}
func TestAuditWinterDiffAgeBeyondCellar(t *testing.T) {
	for _, id := range []string{"winter-04", "winter-15"} {
		t.Run(id, func(t *testing.T) {
			r, p := visitorFixture(id)
			p.Buildings = nil
			p.Wines = []Wine{{ID: "zym-red4", Type: "red", Value: 4}}
			auditWinterStart(t, r, p, id)
			if e := auditWinterChoice(t, r, p, Action{Option: "age"}); e != nil {
				t.Fatal(e)
			}
			if p.Wines[0].Value != 6 {
				t.Errorf("Zymologist wine already above small cellar: expected 4 -> 6, actual %d", p.Wines[0].Value)
			}
		})
	}
}
func TestAuditWinterDiffMasterVintnerAgeBeyondCellar(t *testing.T) {
	r, p := visitorFixture("winter-14")
	p.Buildings = nil
	p.Wines = []Wine{{ID: "zym-red4", Type: "red", Value: 4}}
	p.Hand = append(p.Hand, Card{ID: "red5-order", Type: "order", Requirements: []Wine{{Type: "red", Value: 5}}, Points: 2})
	auditWinterStart(t, r, p, "winter-14")
	if e := auditWinterChoice(t, r, p, Action{Option: "fill", WineIDs: []string{"zym-red4"}}); e != nil {
		t.Fatalf("expected legal 4 -> 5 before fill, actual error: %v", e)
	}
}
func TestAuditWinterDiffHarvesterMinimumFAQ(t *testing.T) {
	r, p := visitorFixture("winter-12")
	for i := range p.Fields {
		p.Fields[i].Vines = nil
	}
	if e := r.apply(p.ID, Action{Type: "place", Space: "winter_visitor", CardID: "winter-12"}); e != nil {
		return
	} // Correct minimum preflight may reject before the effect.
	before := p.VP
	if e := auditWinterChoice(t, r, p, Action{Option: "vp"}); e == nil {
		t.Errorf("FAQ requires up-to effect to some extent; expected reject zero harvest, actual VP delta=%d", p.VP-before)
	}
}
func TestAuditWinterDiffManagerProducerOtherWorker(t *testing.T) {
	r, p := visitorFixture("winter-32")
	p.Hand = append(p.Hand, Card{ID: "summer-32", Type: "summer"})
	// Earlier worker on same winter-visitor action; current Manager worker is slot 1.
	r.Spaces = spaces(3)
	for i := range r.Spaces {
		if r.Spaces[i].ID == "winter_visitor" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 0}}
		}
	}
	if e := r.apply(p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 1, CardID: "winter-32"}); e != nil {
		t.Fatal(e)
	}
	if e := auditWinterChoice(t, r, p, Action{Option: "action", Space: "summer_visitor", CardID: "summer-32"}); e != nil {
		t.Fatal(e)
	}
	if e := auditWinterChoice(t, r, p, Action{Option: "retrieve", TargetIDs: []string{"winter_visitor"}, Fields: []int{0}}); e != nil {
		t.Fatalf("Java r153+ allows earlier worker, excludes only triggering worker: %v", e)
	}
}
func TestAuditWinterDiffZymologistControls(t *testing.T) {
	for _, tc := range []struct {
		name   string
		grapes []Grape
		recipe []int
		typ    string
		value  int
	}{
		{"red", []Grape{{ID: "r", Color: "red", Value: 4}}, []int{0}, "red", 4},
		{"blush", []Grape{{ID: "r", Color: "red", Value: 2}, {ID: "w", Color: "white", Value: 2}}, []int{0, 1}, "blush", 4},
		{"sparkling", []Grape{{ID: "r", Color: "red", Value: 1}, {ID: "r2", Color: "red", Value: 2}, {ID: "w", Color: "white", Value: 4}}, []int{0, 1, 2}, "sparkling", 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, p := visitorFixture("winter-33")
			p.Buildings = nil
			p.Wines = nil
			p.Grapes = tc.grapes
			auditWinterStart(t, r, p, "winter-33")
			if e := auditWinterChoice(t, r, p, Action{Option: "make", Recipes: [][]int{tc.recipe}}); e != nil {
				t.Fatal(e)
			}
			if len(p.Buildings) != 0 || len(p.Wines) != 1 || p.Wines[0].Type != tc.typ || p.Wines[0].Value != tc.value {
				t.Fatalf("wrong output: buildings=%v wines=%v", p.Buildings, p.Wines)
			}
		})
	}
}
func TestAuditWinterDiffManagerBonusPreservedControl(t *testing.T) {
	r, p := visitorFixture("winter-32")
	p.Hand = append(p.Hand, Card{ID: "summer-01", Type: "summer"})
	if e := r.apply(p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 1, CardID: "winter-32"}); e != nil {
		t.Fatal(e)
	}
	if e := auditWinterChoice(t, r, p, Action{Option: "action", Space: "summer_visitor", CardID: "summer-01"}); e != nil {
		t.Fatal(e)
	}
	if e := auditWinterChoice(t, r, p, Action{Option: "vp"}); e != nil {
		t.Fatal(e)
	}
	if len(r.Choices) != 1 || r.Choices[0].Visitor.Stage != "second" || r.Context.Space != "winter_visitor" {
		t.Fatalf("outer winter bonus lost: choices=%+v context=%+v", r.Choices, r.Context)
	}
}
