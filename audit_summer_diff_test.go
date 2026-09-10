//go:build audit_diff

package main

import "testing"

// Expected-rule assertions: these intentionally fail on the audited implementation.
func auditSummerChoose(r *Room, a Action) error {
	c := r.Choices[0]
	a.Type = "choose"
	a.ChoiceID = c.ID
	return r.apply(c.PlayerID, a)
}
func TestAuditSummerDiffPlanterEmptyReward(t *testing.T) {
	r, p := visitorFixture("summer-09")
	p.Hand = []Card{{ID: "summer-09", Type: "summer"}}
	for i := range p.Fields {
		p.Fields[i].Vines = nil
	}
	before := p.Coins
	err := r.apply(p.ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-09", Slot: 2})
	if err != nil {
		return
	}
	err = auditSummerChoose(r, Action{Option: "plant"})
	if err == nil {
		t.Fatalf("empty Planter accepted: coins %d -> %d; no vines in hand or fields", before, p.Coins)
	}
}
func TestAuditSummerDiffEmptyMandatoryExtent(t *testing.T) {
	for _, tc := range []struct{ id, option string }{{"summer-07", "make"}, {"summer-28", "plant"}, {"summer-35", "plant"}, {"summer-32", "retrieve"}, {"summer-38", "pay"}} {
		t.Run(tc.id, func(t *testing.T) {
			r, p := visitorFixture(tc.id)
			if tc.id == "summer-32" {
				r.Spaces[1].Occupied = []Seat{{PlayerID: p.ID, Slot: 2}}
			}
			playVisitorFixture(t, r, p, tc.id)
			if err := auditSummerChoose(r, Action{Option: tc.option}); err == nil {
				t.Fatal("zero selected resources accepted; FAQ requires fulfilling up-to to some extent")
			}
		})
	}
}
func TestAuditSummerDiffLandscaperNoStructures(t *testing.T) {
	r, p := visitorFixture("summer-11")
	p.Buildings = nil
	p.Fields[0].Vines = []Card{{ID: "merlot", Type: "vine", Name: "Merlot", Red: 3, Irrigation: true}}
	p.Fields[1].Vines = []Card{{ID: "sangio", Type: "vine", Name: "Sangiovese", Red: 1}}
	// Merlot can legally have been planted with Horticulturist or Sharecropper.
	playVisitorFixture(t, r, p, "summer-11")
	if err := auditSummerChoose(r, Action{Option: "swap", Fields: []int{0, 1}, CardIDs: []string{"merlot", "sangio"}}); err != nil {
		t.Fatalf("legal capacity-preserving switch rejected: %v", err)
	}
}
func TestAuditSummerDiffOrganizerCannotMintGrayWorker(t *testing.T) {
	r, p := visitorFixture("summer-33")
	// This player took the sole gray worker at wake row 7, then uses Organizer to leave.
	p.Wake = 7
	p.Workers++
	r.GrayWorkerOwner = p.ID
	playVisitorFixture(t, r, p, "summer-33")
	if err := auditSummerChoose(r, Action{Option: "move", Slot: 4}); err != nil {
		t.Fatal(err)
	}
	// Importer/Peddler or a reshuffle allows another player to play the same card this year.
	q := r.Players[1]
	q.Hand = append(q.Hand, Card{ID: "summer-33", Type: "summer"})
	r.Discards["summer"] = nil
	r.TurnID = q.ID
	before := q.Workers
	if err := r.apply(q.ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-33", Large: true, Slot: 1}); err != nil {
		t.Fatal(err)
	}
	if err := auditSummerChoose(r, Action{Option: "move", Slot: 7}); err != nil {
		t.Fatal(err)
	}
	if q.Workers != before {
		t.Fatalf("row 7 is empty but gray worker is already owned: workers %d -> %d", before, q.Workers)
	}
}
