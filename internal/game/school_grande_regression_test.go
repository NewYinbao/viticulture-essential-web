package game

import "testing"

func TestSchoolRetrainsOnlyRemainingGrandeWithoutSpecialModule(t *testing.T) {
	r, p := visitorFixture("summer-01")
	r.Config.Structures = true
	r.Config.SpecialWorkers = false
	r.Context, r.Choices = nil, nil
	p.StructureSlots = []string{"school"}
	p.Buildings = append(p.Buildings, "school")
	p.GrandeRemoved, p.LargeWorker = true, false
	p.TotalWorkers, p.Workers, p.Coins = 5, 5, 0
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "structure_action:" + p.ID + ":school"}); err != nil {
		t.Fatal(err)
	}
	if p.GrandeRemoved || !p.LargeWorker || p.TotalWorkers != 6 || p.Workers != 4 || p.Coins != 1 {
		t.Fatalf("School should train the remaining grande immediately after placing one worker: %+v", p)
	}
}
