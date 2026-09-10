package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func rejectSpecial(t *testing.T, r *Room, id string, a Action) {
	t.Helper()
	before, _ := json.Marshal(r)
	if e := r.Apply(id, a); e == nil {
		t.Fatal("accepted invalid action")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("failed action mutated state")
	}
}
func specialWake(r *Room) {
	r.WakeSlots = nil
	for i := 1; i <= 7; i++ {
		w := WakeSlot{Slot: i}
		for _, p := range r.Players {
			if p.Wake == i {
				w.PlayerID = p.ID
			}
		}
		r.WakeSlots = append(r.WakeSlots, w)
	}
}
func TestSpecialQueenBranchesRightSeatAndPersistence(t *testing.T) {
	for _, opt := range []string{"vp", "cards", "coins"} {
		t.Run(opt, func(t *testing.T) {
			r, p := visitorFixture("winter-11")
			p.Wake = 2
			r.Players[1].Wake = 3
			r.Players[2].Wake = 1
			q := r.Players[2]
			q.Passed = true
			q.VP = -4
			coins := p.Coins
			hand := len(p.Hand) - 1
			qh := len(q.Hand)
			playVisitorFixture(t, r, p, "winter-11")
			chooseVisitor(t, r, "resolve", nil)
			if r.Choices[0].PlayerID != q.ID {
				t.Fatal("right neighbor used wake order or seat +1")
			}
			c := r.Choices[0]
			rejectSpecial(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: opt})
			r = roundTripVisitor(t, r)
			p = r.Players[0]
			q = r.Players[2]
			chooseVisitor(t, r, opt, func(a *Action) { a.CardIDs = []string{"plant", "s0"} })
			if !q.Passed || r.Context != nil || len(r.Choices) != 0 || r.TurnID != r.Players[1].ID {
				t.Fatal("Queen continuation")
			}
			switch opt {
			case "vp":
				if q.VP != -5 {
					t.Fatal(q.VP)
				}
			case "cards":
				if len(p.Hand) != hand+2 || len(q.Hand) != qh-2 {
					t.Fatal("transfer")
				}
			case "coins":
				if p.Coins != coins+3 || q.Coins != 47 {
					t.Fatal("payment")
				}
			}
		})
	}
}
func TestSpecialQueenImpossibleAndForgedPayments(t *testing.T) {
	r, p := visitorFixture("winter-11")
	q := r.Players[2]
	q.VP = -5
	q.Coins = 2
	q.Hand = q.Hand[:1]
	rejectSpecial(t, r, p.ID, Action{Type: "place", Space: "winter_visitor", CardID: "winter-11"})
	q.Coins = 3
	playVisitorFixture(t, r, p, "winter-11")
	chooseVisitor(t, r, "resolve", nil)
	c := r.Choices[0]
	for _, opt := range []string{"vp", "cards", "skip"} {
		rejectSpecial(t, r, q.ID, Action{Type: "choose", ChoiceID: c.ID, Option: opt})
	}
	chooseVisitor(t, r, "coins", nil)
	if q.VP != -5 || q.Coins != 0 {
		t.Fatal("floor")
	}
	r, p = visitorFixture("winter-11")
	playVisitorFixture(t, r, p, "winter-11")
	chooseVisitor(t, r, "resolve", nil)
	c = r.Choices[0]
	for _, ids := range [][]string{{"s0", "s0"}, {"s0", "forged"}, {"s0"}} {
		rejectSpecial(t, r, c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: "cards", CardIDs: ids})
	}
}
func TestSpecialOrganizerAllRowsRewards(t *testing.T) {
	for row := 1; row <= 7; row++ {
		t.Run(fmt.Sprint(row), func(t *testing.T) {
			r, p := visitorFixture("summer-33")
			p.Wake = row%7 + 1
			r.Players[1].Wake = (row+1)%7 + 1
			r.Players[2].Wake = (row+2)%7 + 1
			specialWake(r)
			old := p.Wake
			coins, vp, hand := p.Coins, p.VP, len(p.Hand)-1
			workers := p.Workers - 1
			playVisitorFixture(t, r, p, "summer-33")
			chooseVisitor(t, r, "move", func(a *Action) { a.Slot = row; a.Color = "winter" })
			if p.Wake != row || !p.Passed || r.WakeSlots[old-1].PlayerID != "" || r.WakeSlots[row-1].PlayerID != p.ID {
				t.Fatal("wake or pass")
			}
			if row == 4 {
				coins++
			}
			if row == 6 {
				vp++
			}
			if row == 7 {
				workers++
			}
			if row == 2 || row == 3 || row == 5 {
				hand++
			}
			if p.Coins != coins || p.VP != vp || p.Workers != workers || len(p.Hand) != hand {
				t.Fatal("reward")
			}
		})
	}
}
func TestSpecialOrganizerDoubleVisitorAndManager(t *testing.T) {
	for _, manager := range []bool{false, true} {
		t.Run(fmt.Sprint(manager), func(t *testing.T) {
			id := "summer-33"
			space := "summer_visitor"
			if manager {
				id = "winter-32"
				space = "winter_visitor"
			}
			r, p := visitorFixture(id)
			specialWake(r)
			p.Hand = append(p.Hand, Card{ID: "winter-21", Type: "winter"}, Card{ID: "summer-14", Type: "summer"})
			if manager {
				p.Hand = append(p.Hand, Card{ID: "summer-33", Type: "summer"})
			}
			if e := r.Apply(p.ID, Action{Type: "place", Space: space, Slot: 1, CardID: id}); e != nil {
				t.Fatal(e)
			}
			if manager {
				chooseVisitor(t, r, "action", func(a *Action) { a.Space = "summer_visitor"; a.CardID = "summer-33" })
			}
			chooseVisitor(t, r, "move", func(a *Action) { a.Slot = 5; a.Color = "summer" })
			if p.Passed || r.Context == nil || !r.Context.PassAfter || r.Choices[0].Visitor.Stage != "second" {
				t.Fatal("lost outer continuation")
			}
			r = roundTripVisitor(t, r)
			p = r.Players[0]
			next := "summer-14"
			if manager {
				next = "winter-21"
			}
			chooseVisitor(t, r, "play", func(a *Action) { a.CardID = next })
			opt := "coins"
			if manager {
				opt = "resolve"
			}
			chooseVisitor(t, r, opt, nil)
			phase := "summer"
			if manager {
				phase = "winter"
			}
			if !p.Passed || r.Phase != phase || r.Context != nil || r.TurnID != r.Players[1].ID {
				t.Fatal("pass broke season or turn")
			}
		})
	}
}
func TestSpecialOrganizerRejectsOccupiedAndMissingColor(t *testing.T) {
	r, p := visitorFixture("summer-33")
	specialWake(r)
	playVisitorFixture(t, r, p, "summer-33")
	c := r.Choices[0]
	for _, slot := range []int{0, 1, 2, 8, 5} {
		rejectSpecial(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "move", Slot: slot})
	}
}
func enterSpecialWinter(t *testing.T, r *Room) {
	t.Helper()
	for r.Phase == "summer" {
		if e := r.Apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	for r.Phase == "fall" {
		c := r.Choices[0]
		if e := r.Apply(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: c.Options[0]}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestSpecialPlannerDeferredBonusAndPersistence(t *testing.T) {
	r, p := visitorFixture("summer-29")
	before := len(p.Hand) - 1
	playVisitorFixture(t, r, p, "summer-29")
	chooseVisitor(t, r, "plan", func(a *Action) { a.Space = "draw_order"; a.Slot = 1 })
	if p.Workers != 0 || len(p.Hand) != before || len(r.Planned) != 1 {
		t.Fatal("planner executed early or worker count")
	}
	r = roundTripVisitor(t, r)
	p = r.Players[0]
	enterSpecialWinter(t, r)
	if r.Choices[0].Kind != "planner" || r.Context.Step != "planner" {
		t.Fatal("missing winter start choice")
	}
	before = len(p.Hand)
	c := r.Choices[0]
	rejectSpecial(t, r, r.Players[1].ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"})
	rejectSpecial(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", Space: "train"})
	r = roundTripVisitor(t, r)
	p = r.Players[0]
	c = r.Choices[0]
	a := Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}
	if e := r.Apply(p.ID, a); e != nil {
		t.Fatal(e)
	}
	if p.Workers != 0 || len(p.Hand) != before+2 || len(r.Planned) != 0 || r.Context != nil || r.TurnID != p.ID {
		t.Fatal("deferred bonus or normal first turn")
	}
	rejectSpecial(t, r, p.ID, a)
}
func TestSpecialPlannerMultiplayerVisitorContinuation(t *testing.T) {
	r, p := visitorFixture("summer-29")
	p.Hand = append(p.Hand, Card{ID: "winter-38", Type: "winter"})
	playVisitorFixture(t, r, p, "summer-29")
	chooseVisitor(t, r, "plan", func(a *Action) { a.Space = "winter_visitor"; a.Slot = 2 })
	// A second preplacement is a persisted fixture of a replayed Planner.
	q := r.Players[1]
	q.Workers--
	r.Planned = append(r.Planned, PlannedPlacement{PlayerID: q.ID, Space: "draw_order", Slot: 1})
	for i := range r.Spaces {
		if r.Spaces[i].ID == "draw_order" {
			r.Spaces[i].Occupied = append(r.Spaces[i].Occupied, Seat{PlayerID: q.ID, Slot: 1})
		}
	}
	enterSpecialWinter(t, r)
	c := r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", CardID: "winter-38"}); e != nil {
		t.Fatal(e)
	}
	chooseVisitor(t, r, "resolve", nil)
	chooseVisitor(t, r, "skip", nil)
	chooseVisitor(t, r, "train", nil)
	r = roundTripVisitor(t, r)
	chooseVisitor(t, r, "skip", nil)
	if r.Choices[0].Kind != "planner" || r.Choices[0].PlayerID != q.ID {
		t.Fatal("lost next reservation")
	}
	c = r.Choices[0]
	if e := r.Apply(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); e != nil {
		t.Fatal(e)
	}
	if r.Context != nil || len(r.Planned) != 0 || r.TurnID != r.Players[0].ID || r.Players[0].Workers != 0 {
		t.Fatal("normal winter resumed incorrectly")
	}
}
func TestSpecialPlannerUnavailableWorkerSpaceAndWinter(t *testing.T) {
	r, p := visitorFixture("summer-29")
	p.Workers = 1
	p.LargeWorker = false
	rejectSpecial(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-29"})
	r, p = visitorFixture("summer-29")
	playVisitorFixture(t, r, p, "summer-29")
	c := r.Choices[0]
	for _, sp := range []string{"tour", "gain_coin", "yoke", "forged"} {
		rejectSpecial(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "plan", Space: sp})
	}
	chooseVisitor(t, r, "plan", func(a *Action) { a.Space = "train"; a.Slot = 2 })
	p.Coins = 0
	enterSpecialWinter(t, r)
	if len(r.Choices) != 0 || r.Context != nil || len(r.Planned) != 0 || p.Workers != 0 {
		t.Fatal("impossible reservation should exhaust without refund")
	}
	r, p = visitorFixture("winter-32")
	p.Hand = append(p.Hand, Card{ID: "summer-29", Type: "summer"})
	playVisitorFixture(t, r, p, "winter-32")
	c = r.Choices[0]
	rejectSpecial(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "action", Space: "summer_visitor", CardID: "summer-29"})
}
func TestSpecialPlannerGrandeOverflowAndRetrieval(t *testing.T) {
	r, p := visitorFixture("summer-29")
	for i := range r.Spaces {
		if r.Spaces[i].ID == "draw_order" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1}, {PlayerID: r.Players[2].ID, Slot: 2}}
		}
	}
	playVisitorFixture(t, r, p, "summer-29")
	chooseVisitor(t, r, "plan", func(a *Action) { a.Space = "draw_order"; a.Large = true; a.Slot = -1 })
	if p.LargeWorker {
		t.Fatal("grande not consumed")
	}
	enterSpecialWinter(t, r)
	before := len(p.Hand)
	c := r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); e != nil {
		t.Fatal(e)
	}
	if len(p.Hand) != before+1 || p.LargeWorker {
		t.Fatal("overflow bonus or refund")
	}
	r, p = visitorFixture("summer-29")
	p.Hand = append(p.Hand, Card{ID: "summer-32", Type: "summer"})
	playVisitorFixture(t, r, p, "summer-29")
	chooseVisitor(t, r, "plan", func(a *Action) { a.Space = "draw_order"; a.Slot = 1 })
	r.TurnID = p.ID
	p.Workers = 1
	if e := r.Apply(p.ID, Action{Type: "place", Space: "summer_visitor", CardID: "summer-32", Slot: 1}); e != nil {
		t.Fatal(e)
	}
	chooseVisitor(t, r, "retrieve", func(a *Action) { a.TargetIDs = []string{"draw_order"}; a.Fields = []int{1} })
	if len(r.Planned) != 0 {
		t.Fatal("retrieved worker retained delayed action")
	}
	chooseVisitor(t, r, "skip", nil)
}
