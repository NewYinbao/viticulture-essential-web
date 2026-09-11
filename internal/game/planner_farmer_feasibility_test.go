package game

import "testing"

func TestPlannerFeasibilityUsesLockedFarmerDiscount(t *testing.T) {
	r, p := plannerSpecialFixture(t, "farmer")
	reserveWithPlanner(t, r, p, "farmer", "train", 2)
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("missing Farmer placement choice: %+v", r.Choices)
	}
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "discount"}); err != nil {
		t.Fatal(err)
	}
	playerID := p.ID
	r = roundTripVisitor(t, r)
	p = r.Player(playerID)
	p.Coins = 3
	if len(r.Planned) != 1 || r.Planned[0].SpecialWorker == nil || r.Planned[0].SpecialWorker.FarmerBonus != "discount" {
		t.Fatalf("Farmer discount was not locked: %+v", r.Planned)
	}
	withoutLockedBonus := r.Planned[0]
	withoutLockedBonus.SpecialWorker = &PlannedWorkerState{PlacementHandled: true}
	if r.plannerCanExecute(withoutLockedBonus) {
		t.Fatal("Planner treated Farmer identity without a locked discount as affordable")
	}
	if !r.plannerCanExecute(r.Planned[0]) {
		t.Fatal("Planner discarded a train action affordable only with the locked Farmer discount")
	}
	r.Phase = "winter"
	for _, q := range r.Players {
		q.Season = "winter"
	}
	r.startNextPlanner()
	if len(r.Choices) != 1 || r.Choices[0].Kind != "planner" {
		t.Fatalf("executable Farmer reservation was removed: %+v", r.Choices)
	}
	if r.Choices[0].SpecialBonus != "discount" {
		t.Fatalf("actor's Planner choice omitted the locked Farmer discount: %+v", r.Choices[0])
	}
	other := r.Players[1]
	if pending, ok := r.View(other.ID)["pendingChoice"].(map[string]string); !ok || len(pending) != 2 || pending["kind"] != "planner" || pending["playerId"] != p.ID {
		t.Fatalf("locked Farmer discount leaked to another player: %#v", r.View(other.ID)["pendingChoice"])
	}
	c = r.Choices[0]
	before := p.TotalWorkers
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", SpecialWorker: "regular"}); err != nil {
		t.Fatal(err)
	}
	if p.Coins != 0 || p.TotalWorkers != before+1 || len(r.Choices) != 0 || len(r.Planned) != 0 {
		t.Fatalf("Farmer-discounted training did not execute once: coins=%d total=%d choices=%+v plans=%+v", p.Coins, p.TotalWorkers, r.Choices, r.Planned)
	}
}
