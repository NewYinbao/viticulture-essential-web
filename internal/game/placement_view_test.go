package game

import (
	"encoding/json"
	"testing"
)

func placementHas(options []WorkerPlacement, worker string, slot int) bool {
	for _, o := range options {
		if o.WorkerType == worker && o.Slot == slot {
			return true
		}
	}
	return false
}

func TestPlacementViewOnlySpecialWorkerAndNoMutation(t *testing.T) {
	r := specialWorkerRoom("mafioso", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "mafioso")
	p.Workers, p.LargeWorker = 0, false
	s := r.space("tour")
	before, _ := json.Marshal(r)
	options := r.workerPlacements(p, s)
	if !placementHas(options, "mafioso", 1) {
		t.Fatalf("ready special worker hidden: %+v", options)
	}
	for _, o := range options {
		if o.WorkerType == "" {
			t.Fatal("invented ordinary worker")
		}
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("preview mutated room")
	}
	reasons, _, _ := r.availability(p.ID)
	if reasons[s.ID] != "" {
		t.Fatalf("special worker action blocked: %s", reasons[s.ID])
	}
	if err := r.Apply(p.ID, Action{Type: "place", Space: s.ID, Slot: 1, WorkerType: "mafioso"}); err != nil {
		t.Fatal(err)
	}
}

func TestPlacementViewChefFindsBumpableOccupiedSlot(t *testing.T) {
	r := specialWorkerRoom("chef", "oracle")
	p, q := r.Players[0], r.Players[1]
	readySpecial(r, p, "chef")
	p.Workers, p.LargeWorker = 0, false
	s := r.space("tour")
	s.Occupied = []Seat{{PlayerID: q.ID, Slot: 1, WorkerType: "chef"}, {PlayerID: q.ID, Slot: 2}}
	options := r.workerPlacements(p, s)
	if placementHas(options, "chef", 1) || !placementHas(options, "chef", 2) {
		t.Fatalf("incorrect Chef slots: %+v", options)
	}
	before := q.Workers
	if err := r.Apply(p.ID, Action{Type: "place", Space: s.ID, Slot: 2, WorkerType: "chef"}); err != nil {
		t.Fatal(err)
	}
	if q.Workers != before+1 {
		t.Fatal("bumped worker not returned")
	}
}

func TestTravelerExtraSlotsOnlyInPreviousSeason(t *testing.T) {
	for _, previous := range []bool{false, true} {
		r := specialWorkerRoom("traveler", "oracle")
		p := r.Players[0]
		readySpecial(r, p, "traveler")
		p.Workers, p.LargeWorker = 0, false
		s := r.space("tour")
		s.Capacity = 1
		if previous {
			r.Phase = "summer"
			p.Season = "summer"
		}
		options := r.workerPlacements(p, s)
		if placementHas(options, "traveler", 3) != previous {
			t.Fatalf("previous=%v slots=%+v", previous, options)
		}
		before, _ := json.Marshal(r)
		err := r.Apply(p.ID, Action{Type: "place", Space: s.ID, Slot: 3, WorkerType: "traveler"})
		if (err == nil) != previous {
			t.Fatalf("previous=%v err=%v", previous, err)
		}
		if !previous {
			after, _ := json.Marshal(r)
			if string(before) != string(after) {
				t.Fatal("invalid placement mutated room")
			}
		}
	}
}

func TestPlacementViewHidesDisabledModuleAndUnaffordableToll(t *testing.T) {
	r := specialWorkerRoom("oracle", "soldato")
	p := r.Players[0]
	readySpecial(r, p, "oracle")
	p.Workers, p.LargeWorker = 0, false
	s := r.space("tour")
	r.Config.SpecialWorkers = false
	if len(r.workerPlacements(p, s)) != 0 {
		t.Fatal("disabled special worker exposed")
	}
	r.Config.SpecialWorkers = true
	s.Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1, WorkerType: "soldato"}}
	p.Coins = 0
	if len(r.workerPlacements(p, s)) != 0 {
		t.Fatal("unaffordable toll advertised")
	}
}
