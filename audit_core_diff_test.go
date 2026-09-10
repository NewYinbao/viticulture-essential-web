//go:build audit_diff

package main

import "testing"

// These regression specifications intentionally fail on confirmed current differences.
func TestAuditDiffPlantMustNotOfferUproot(t *testing.T) {
	p, q := makePlayer("a"), makePlayer("b")
	p.Wake = 1
	q.Wake = 2
	p.Fields[0].Vines = []Card{{ID: "v", Type: "vine", Red: 1}}
	r := &Room{Players: []*Player{p, q}, Phase: "summer", Year: 1, TurnID: p.ID, Spaces: spaces(2)}
	if err := r.apply(p.ID, Action{Type: "place", Space: "plant", Mode: "uproot", Field: 0, CardID: "v"}); err == nil {
		t.Fatal("plant accepted uproot without a yoke")
	}
}
func TestAuditDiffYokeReopensAfterProducer(t *testing.T) {
	p := makePlayer("a")
	p.Coins = 2
	p.Buildings = []string{"yoke"}
	p.YokeUsed = true
	r := &Room{Players: []*Player{p}, Spaces: spaces(2), Context: &ActionContext{Space: "summer_visitor"}}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "yoke" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 1}}
		}
	}
	if err := r.visitorSequence(p, &VisitorStep{CardID: "summer-32"}, Action{TargetIDs: []string{"yoke"}, Fields: []int{1}}); err != nil {
		t.Fatal(err)
	}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "yoke" {
			if _, _, err := placement(&r.Spaces[i], p, Action{}); err != nil {
				t.Fatalf("empty yoke remains locked: %v", err)
			}
		}
	}
}
func TestAuditDiffProducerMayRetrieveEarlierWorkerOnSameAction(t *testing.T) {
	p := makePlayer("a")
	p.Coins = 2
	r := &Room{Players: []*Player{p}, Spaces: spaces(3), Context: &ActionContext{Space: "summer_visitor"}}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "summer_visitor" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 2}}
		}
	}
	if err := r.visitorSequence(p, &VisitorStep{CardID: "summer-32"}, Action{TargetIDs: []string{"summer_visitor"}, Fields: []int{2}}); err != nil {
		t.Fatalf("earlier worker forbidden: %v", err)
	}
}
func TestAuditDiffSwapNeedsCapacityNotBuildings(t *testing.T) {
	p := makePlayer("a")
	// Reachable via Horticulturist or Sharecropper planting without structures.
	p.Fields[0].Vines = []Card{{ID: "v1", Type: "vine", Red: 2, Trellis: true}}
	p.Fields[1].Vines = []Card{{ID: "v2", Type: "vine", White: 1}}
	r := &Room{Players: []*Player{p}}
	if err := r.visitorSequence(p, &VisitorStep{CardID: "summer-11"}, Action{Option: "swap", Fields: []int{0, 1}, CardIDs: []string{"v1", "v2"}}); err != nil {
		t.Fatalf("legal swap rejected: %v", err)
	}
}
func TestAuditDiffBaseWorkerAndEndgameControls(t *testing.T) {
	p, q := makePlayer("a"), makePlayer("b")
	p.VP = 20
	p.TotalWorkers = 6
	p.Workers = 0
	p.LargeWorker = false
	p.Trained = 3
	r := &Room{Players: []*Player{p, q}, Year: 1, Spaces: spaces(2)}
	r.finishYear()
	if r.Phase != "finished" || p.Workers != 5 || !p.LargeWorker || p.Trained != 0 {
		t.Fatal("endgame or worker recovery regression")
	}
	for _, n := range []int{2, 3, 4, 5, 6} {
		s := spaces(n)[0]
		want := 1
		if n >= 3 {
			want = 2
		}
		if n >= 5 {
			want = 3
		}
		if s.Capacity != want {
			t.Fatal("capacity")
		}
	}
}
