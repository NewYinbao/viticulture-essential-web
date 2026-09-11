package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func plannerSpecialFixture(t *testing.T, worker string) (*Room, *Player) {
	t.Helper()
	r := specialWorkerRoom(worker, "oracle")
	p := r.Players[0]
	readySpecial(r, p, worker)
	r.Phase, r.TurnID = "summer", p.ID
	for _, q := range r.Players {
		q.Season = "summer"
	}
	trigger := Seat{Season: "summer", PlayerID: p.ID, Slot: 1}
	r.space("summer_visitor").Occupied = []Seat{trigger}
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Space: "summer_visitor", TriggerSpace: "summer_visitor", TriggerSeat: &trigger, Data: map[string]string{}}
	r.Choices = []Choice{{ID: NewID(), PlayerID: p.ID, Kind: "visitor", Options: []string{"plan"}, Count: 1, Visitor: &VisitorStep{CardID: "summer-29", Stage: "effect", ActorID: p.ID}}}
	return r, p
}

func reserveWithPlanner(t *testing.T, r *Room, p *Player, worker, space string, slot int) {
	t.Helper()
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "plan", Space: space, Slot: slot, WorkerType: worker}); err != nil {
		t.Fatal(err)
	}
}

func advanceToPlannerByApply(t *testing.T, r *Room) {
	t.Helper()
	for i := 0; i < 30; i++ {
		if len(r.Choices) > 0 && r.Choices[0].Kind == "planner" {
			return
		}
		if len(r.Choices) > 0 {
			c := r.Choices[0]
			if err := r.Apply(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: c.Options[0]}); err != nil {
				t.Fatalf("advance choice %s: %v", c.Kind, err)
			}
			continue
		}
		if err := r.Apply(r.TurnID, Action{Type: "pass"}); err != nil {
			t.Fatalf("advance pass in %s: %v", r.Phase, err)
		}
	}
	t.Fatalf("Planner choice not reached: phase=%s choices=%+v", r.Phase, r.Choices)
}

func TestPlannerFarmerChoicePersistsUntilFutureAction(t *testing.T) {
	r, p := plannerSpecialFixture(t, "farmer")
	r.Decks["order"] = []Card{{ID: "o1", Type: "order"}, {ID: "o2", Type: "order"}, {ID: "o3", Type: "order"}}
	reserveWithPlanner(t, r, p, "farmer", "draw_order", 2)
	if r.Context == nil || r.Context.Step != "planner_placement" || r.Context.Parent == nil || len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("Farmer placement choice did not nest around Planner: context=%+v choices=%+v", r.Context, r.Choices)
	}
	r = roundTripVisitor(t, r)
	p = r.Player(p.ID)
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "draw_order"}); err != nil {
		t.Fatal(err)
	}
	if r.Context != nil || len(r.Planned) != 1 || r.Planned[0].SpecialWorker == nil || r.Planned[0].SpecialWorker.FarmerBonus != "draw_order" || len(p.Hand) != 0 {
		t.Fatalf("Farmer choice was not saved without early reward: context=%+v plan=%+v hand=%d", r.Context, r.Planned, len(p.Hand))
	}
	r = roundTripVisitor(t, r)
	p = r.Player(p.ID)
	advanceToPlannerByApply(t, r)
	r = roundTripVisitor(t, r)
	p = r.Player(p.ID)
	c = r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 2 || len(r.Choices) != 0 || len(r.Planned) != 0 {
		t.Fatalf("Farmer future action or bonus repeated/lost: hand=%d choices=%+v plans=%+v", len(p.Hand), r.Choices, r.Planned)
	}
}

func TestPlannerPlacementChoiceRestoresOuterVisitorContinuation(t *testing.T) {
	r, p := plannerSpecialFixture(t, "farmer")
	r.Context.Remaining = 1
	reserveWithPlanner(t, r, p, "farmer", "draw_order", 2)
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "draw_order"}); err != nil {
		t.Fatal(err)
	}
	if r.Context == nil || r.Context.Step == "planner_placement" || len(r.Choices) != 1 || r.Choices[0].Visitor == nil || r.Choices[0].Visitor.Stage != "second" {
		t.Fatalf("outer two-visitor continuation was not restored: context=%+v choices=%+v", r.Context, r.Choices)
	}
	r = roundTripVisitor(t, r)
	p = r.Player(p.ID)
	c = r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}); err != nil {
		t.Fatal(err)
	}
	if r.Context != nil || len(r.Choices) != 0 || len(r.Planned) != 1 || r.Planned[0].SpecialWorker.FarmerBonus != "draw_order" {
		t.Fatalf("outer visitor completion lost Planner state: context=%+v choices=%+v plans=%+v", r.Context, r.Choices, r.Planned)
	}
}

func TestPlannerProfessoreAndInnkeeperResolveAtReservation(t *testing.T) {
	t.Run("professore", func(t *testing.T) {
		r, p := plannerSpecialFixture(t, "professore")
		p.Workers--
		r.space("plant").Occupied = []Seat{{Season: "summer", PlayerID: p.ID, Slot: 1}}
		reserveWithPlanner(t, r, p, "professore", "draw_order", 2)
		if len(r.Choices) != 1 || r.Choices[0].Kind != "special_professore" {
			t.Fatalf("missing immediate Professore choice: %+v", r.Choices)
		}
		r = roundTripVisitor(t, r)
		p = r.Player(p.ID)
		before := p.Workers
		c := r.Choices[0]
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "plant:1"}); err != nil {
			t.Fatal(err)
		}
		if p.Workers != before+1 || len(r.space("plant").Occupied) != 0 || r.Context != nil || len(r.Planned) != 1 {
			t.Fatalf("Professore did not finish inside outer Planner: workers=%d seat=%+v context=%+v", p.Workers, r.space("plant").Occupied, r.Context)
		}
		advanceToPlannerByApply(t, r)
		c = r.Choices[0]
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
			t.Fatal(err)
		}
		if len(r.Choices) != 0 {
			t.Fatalf("Professore triggered again in the future: %+v", r.Choices)
		}
	})

	t.Run("innkeeper", func(t *testing.T) {
		r, p := plannerSpecialFixture(t, "innkeeper")
		q := r.Players[1]
		q.Hand = []Card{{ID: "guest", Type: "summer"}, {ID: "guest2", Type: "summer"}}
		r.space("draw_order").Occupied = []Seat{{Season: "summer", PlayerID: q.ID, Slot: 1}}
		reserveWithPlanner(t, r, p, "innkeeper", "draw_order", 2)
		if len(r.Choices) != 1 || r.Choices[0].Kind != "special_innkeeper" {
			t.Fatalf("missing immediate Innkeeper choice: %+v", r.Choices)
		}
		c := r.Choices[0]
		beforeCoins := p.Coins
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: q.ID + "|summer"}); err != nil {
			t.Fatal(err)
		}
		if p.Coins != beforeCoins-1 || len(p.Hand) != 1 || len(q.Hand) != 1 || r.Context != nil {
			t.Fatalf("Innkeeper did not resolve at reservation: coins=%d hands=%d/%d context=%+v", p.Coins, len(p.Hand), len(q.Hand), r.Context)
		}
		advanceToPlannerByApply(t, r)
		c = r.Choices[0]
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
			t.Fatal(err)
		}
		if len(r.Choices) != 0 || len(p.Hand) != 2 || len(q.Hand) != 1 {
			t.Fatalf("Innkeeper repeated instead of only taking the order action: choices=%+v hands=%d/%d", r.Choices, len(p.Hand), len(q.Hand))
		}
	})
}

func TestPlannerLocksPostActionSpecialWorkerQualification(t *testing.T) {
	t.Run("merchant eligible at placement", func(t *testing.T) {
		r, p := plannerSpecialFixture(t, "merchant")
		for _, q := range r.Players[1:] {
			q.Season = "fall"
		}
		reserveWithPlanner(t, r, p, "merchant", "draw_order", 2)
		for _, q := range r.Players[1:] {
			q.Season = "summer"
		}
		advanceToPlannerByApply(t, r)
		for _, q := range r.Players[1:] {
			q.Season = "fall"
		}
		c := r.Choices[0]
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
			t.Fatal(err)
		}
		if len(r.Choices) != 1 || r.Choices[0].Kind != "special_merchant" {
			t.Fatalf("Merchant placement qualification was recomputed: %+v", r.Choices)
		}
	})

	for _, tc := range []struct {
		worker string
		bonus  bool
		kind   string
	}{
		{worker: "mafioso", bonus: false, kind: "special_mafioso"},
		{worker: "politico", bonus: true, kind: "special_politico"},
	} {
		t.Run(tc.worker, func(t *testing.T) {
			r, p := plannerSpecialFixture(t, tc.worker)
			slot := 2
			if tc.bonus {
				slot = 1
			}
			reserveWithPlanner(t, r, p, tc.worker, "draw_order", slot)
			if tc.bonus {
				delete(r.space("draw_order").BonusSlots, slot)
			} else {
				r.space("draw_order").BonusSlots[slot] = "draw_order"
			}
			advanceToPlannerByApply(t, r)
			c := r.Choices[0]
			if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
				t.Fatal(err)
			}
			if len(r.Choices) != 1 || r.Choices[0].Kind != tc.kind {
				t.Fatalf("%s qualification was recomputed: %+v", tc.worker, r.Choices)
			}
			if tc.worker == "politico" && r.Choices[0].SpecialBonus != "draw_order" {
				t.Fatalf("Politico lost the placement-time bonus: %+v", r.Choices[0])
			}
		})
	}
}

func TestPlannerOracleWaitsForActualDrawAndFailedExecutionRollsBack(t *testing.T) {
	r, p := plannerSpecialFixture(t, "oracle")
	r.Decks["order"] = []Card{{ID: "o1", Type: "order"}, {ID: "o2", Type: "order"}, {ID: "o3", Type: "order"}}
	reserveWithPlanner(t, r, p, "oracle", "draw_order", 2)
	if r.Context != nil || len(p.Hand) != 0 || len(r.Choices) != 0 {
		t.Fatalf("Oracle triggered before the future draw: context=%+v hand=%d choices=%+v", r.Context, len(p.Hand), r.Choices)
	}
	advanceToPlannerByApply(t, r)
	c := r.Choices[0]
	before, _ := json.Marshal(r)
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", Space: "forged"}); err == nil {
		t.Fatal("forged Planner action was accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("failed Planner execution did not roll back plan/context/choice")
	}
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", BonusOverride: "draw_order"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 2 || len(r.Choices) != 1 || r.Choices[0].Kind != "special_oracle" {
		t.Fatalf("Oracle did not trigger on actual draw: hand=%d choices=%+v", len(p.Hand), r.Choices)
	}
}
