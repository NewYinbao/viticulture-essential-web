package game

import (
	"encoding/json"
	"testing"
)

func readyWorkerForTest(r *Room, p *Player, id string) {
	p.SpecialWorkers = append(p.SpecialWorkers, id)
	p.SpecialWorkerReady = map[string]int{id: r.Year}
	p.SpecialWorkerUsed = map[string]bool{}
}

func TestEETravelerUsesHiddenSummerSlotFromWinter(t *testing.T) {
	r := specialWorkerRoom("traveler")
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	readyWorkerForTest(r, p, "traveler")
	r.Phase, r.TurnID = "winter", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 3, WorkerType: "traveler"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 1 || len(r.space("draw_vine").Occupied) != 1 || r.space("draw_vine").Occupied[0].Slot != 3 {
		t.Fatalf("EE Traveler did not use hidden prior-season slot: hand=%+v seats=%+v", p.Hand, r.space("draw_vine").Occupied)
	}
}

func TestEEMessengerChoosesLiveResourcesAfterPersistentReservation(t *testing.T) {
	r := specialWorkerRoom("messenger")
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	readyWorkerForTest(r, p, "messenger")
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "make_wine", Slot: 1, WorkerType: "messenger", Recipes: [][]int{{99}}}); err != nil {
		t.Fatalf("reservation incorrectly validated future resources: %v", err)
	}
	if len(r.MessengerPlans) != 1 || r.MessengerPlans[0].Action != nil {
		t.Fatalf("reservation persisted private future inputs: %+v", r.MessengerPlans)
	}
	p.Grapes = []Grape{{ID: "live-red", Color: "red", Value: 2}}
	r.Phase, r.TurnID = "winter", p.ID
	r.startNextPlanner()
	if len(r.Choices) != 1 || r.Choices[0].Kind != "messenger" || r.Choices[0].ActionSpace != "make_wine" {
		t.Fatalf("missing Messenger resource continuation: %+v", r.Choices)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var restored Room
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatal(err)
	}
	r, p = &restored, restored.Player(p.ID)
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", Space: "make_wine", Recipes: [][]int{{0}}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Wines) != 1 || len(p.Grapes) != 0 || len(r.MessengerPlans) != 0 {
		t.Fatalf("Messenger did not use restored live resources: wines=%+v grapes=%+v plans=%+v", p.Wines, p.Grapes, r.MessengerPlans)
	}
}

func TestMessengerImpossibleActionEndsTurnWithoutReward(t *testing.T) {
	r := specialWorkerRoom("messenger")
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	readyWorkerForTest(r, p, "messenger")
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "fill_order", Slot: 1, WorkerType: "messenger"}); err != nil {
		t.Fatal(err)
	}
	r.Phase, r.TurnID = "winter", p.ID
	beforeVP := p.VP
	if !r.activateMessenger(p) {
		t.Fatal("Messenger reservation was not consumed")
	}
	if p.VP != beforeVP || len(r.MessengerPlans) != 0 || len(r.Choices) != 0 || r.TurnID == p.ID {
		t.Fatalf("failed Messenger gained reward or did not end turn: vp=%d plans=%+v choices=%+v turn=%s", p.VP, r.MessengerPlans, r.Choices, r.TurnID)
	}
}

func TestPlannerFeasibilityCoversTuscanyActionSeasons(t *testing.T) {
	for _, tc := range []struct {
		space string
		setup func(*Player)
	}{
		{"plant", func(p *Player) { p.Hand = []Card{{ID: "v", Type: "vine", Red: 1}} }},
		{"build", func(p *Player) { p.Coins = 10 }},
		{"trade", func(p *Player) { p.Coins = 3 }},
		{"influence", func(p *Player) {}},
		{"build_tour", func(p *Player) {}},
	} {
		t.Run(tc.space, func(t *testing.T) {
			r := specialWorkerRoom("messenger")
			p := r.Players[0]
			p.Hand, p.Grapes, p.Wines = nil, nil, nil
			p.Coins, p.VP = 0, 0
			tc.setup(p)
			if !r.plannerCanExecute(PlannedPlacement{PlayerID: p.ID, Space: tc.space, Slot: 1}) {
				t.Fatalf("Tuscany %s executable action was rejected", tc.space)
			}
		})
	}
}

func TestPlannerOnlySpecialWorkerGrandeRetrainAndFailedReservationRetention(t *testing.T) {
	r := specialWorkerRoom("traveler")
	p := r.Players[0]
	p.Workers, p.LargeWorker = 0, false
	readyWorkerForTest(r, p, "traveler")
	r.Phase = "summer"
	if err := r.specialPlayable(p, "summer-29"); err != nil {
		t.Fatalf("Planner rejected the only ready special worker: %v", err)
	}
	p.TotalWorkers, p.GrandeRemoved, p.Coins = 5, true, 4
	if !r.plannerCanExecute(PlannedPlacement{PlayerID: p.ID, Space: "train", Slot: 1}) {
		t.Fatal("Planner rejected legal Grande retraining")
	}

	p.SpecialWorkerUsed["traveler"] = true
	seat := Seat{Season: "winter", PlayerID: p.ID, Slot: 1, WorkerType: "traveler"}
	r.Phase = "winter"
	r.space("fill_order").Occupied = []Seat{seat}
	r.Planned = []PlannedPlacement{{PlayerID: p.ID, Space: "fill_order", Slot: 1, WorkerType: "traveler"}}
	r.startNextPlanner()
	if len(r.Planned) != 0 || len(r.space("fill_order").Occupied) != 1 || !p.SpecialWorkerUsed["traveler"] {
		t.Fatalf("failed special-worker reservation was refunded: planned=%+v seats=%+v used=%v", r.Planned, r.space("fill_order").Occupied, p.SpecialWorkerUsed)
	}
}

func TestRhineSonHarvestNestsAroundMessengerWithoutExtraTurn(t *testing.T) {
	r := specialWorkerRoom("messenger")
	r.Config.Board, r.Config.Visitors = "ee", "rhine"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	readyWorkerForTest(r, p, "messenger")
	field := 0
	p.RhineSonField = &field
	p.Fields[field].Vines = []Card{{ID: "son-vine", Type: "vine", Red: 1}}
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "make_wine", Slot: 1, WorkerType: "messenger"}); err != nil {
		t.Fatal(err)
	}
	r.Phase, r.TurnID = "winter", p.ID
	if !r.activateMessenger(p) {
		t.Fatal("Messenger was not activated")
	}
	r.rhinePrepareWinterTurn()
	if r.Context == nil || r.Context.Step != "rhine_first_winter" || r.Context.Parent == nil || r.Context.Parent.Step != "messenger" {
		t.Fatalf("Son-in-law did not persistently nest Messenger: %+v", r.Context)
	}
	b, _ := json.Marshal(r)
	var restored Room
	_ = json.Unmarshal(b, &restored)
	r, p = &restored, restored.Player(p.ID)
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "harvest", Grapes: []int{1, 0}}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "messenger" || r.TurnID != p.ID {
		t.Fatalf("Messenger continuation did not resume without spending a turn: choices=%+v turn=%s", r.Choices, r.TurnID)
	}
	c = r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", Space: "make_wine", Recipes: [][]int{{0}}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Wines) != 1 || r.TurnID == p.ID {
		t.Fatalf("Messenger after Son-in-law did not consume exactly the reserved turn: wines=%+v turn=%s", p.Wines, r.TurnID)
	}
}
