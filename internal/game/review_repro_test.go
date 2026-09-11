package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReviewMessengerPrivateSelection(t *testing.T) {
	r := specialWorkerRoom("messenger", "oracle")
	p, q := r.Players[0], r.Players[1]
	readySpecial(r, p, "messenger")
	p.Hand = []Card{visitorCardT(t, "summer-01")}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", Slot: 1, WorkerType: "messenger", CardID: "summer-01"})
	b, _ := json.Marshal(r.View(q.ID))
	if strings.Contains(string(b), "\"cardId\":\"summer-01\"") {
		t.Fatalf("opponent sees unrevealed hand selection in messengerPlans: %s", mustReviewJSON(r.View(q.ID)["messengerPlans"]))
	}
}
func mustReviewJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
func TestReviewOracleCardConservation(t *testing.T) {
	r := specialWorkerRoom("oracle", "politico")
	p := r.Players[0]
	readySpecial(r, p, "oracle")
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "draw_vine", Slot: 2, WorkerType: "oracle"})
	id := r.Choices[0].Options[0]
	choiceT(t, r, Action{Option: id})
	total := len(p.Hand) + len(r.Decks["vine"]) + len(r.Discards["vine"])
	if total != 2 {
		t.Fatalf("Oracle discarded %s disappeared: hand=%d deck=%d discards=%d total=%d", id, len(p.Hand), len(r.Decks["vine"]), len(r.Discards["vine"]), total)
	}
}
func TestReviewSpecialWorkerAnnualAccounting(t *testing.T) {
	for _, board := range []string{"ee", "tuscany"} {
		t.Run(board, func(t *testing.T) {
			r := specialWorkerRoom("oracle", "farmer")
			p := r.Players[0]
			readySpecial(r, p, "oracle")
			r.Config.Board = board
			if board == "ee" {
				r.finishYear()
			} else {
				r.Phase = "winter"
				r.tuscanyPass(p)
			}
			if p.Workers != p.TotalWorkers-1-len(p.SpecialWorkers) {
				t.Fatalf("total=%d regular=%d large=%v special=%v; trained special also became regular", p.TotalWorkers, p.Workers, p.LargeWorker, p.SpecialWorkers)
			}
		})
	}
}
func TestReviewSchoolSpecialWorkerAccounting(t *testing.T) {
	r := specialWorkerRoom("oracle", "farmer")
	p := r.Players[0]
	before := p.Workers
	if e := r.trainWorkerNow(p, "oracle"); e != nil {
		t.Fatal(e)
	}
	if p.Workers != before {
		t.Fatalf("training special immediately added ordinary worker too: %d -> %d", before, p.Workers)
	}
}
func TestReviewEEFermentationMustFinishTurn(t *testing.T) {
	r := structureRoom()
	r.Config.Board = "ee"
	r.Spaces = r.spacesForRules()
	r.Phase = "winter"
	p, q := r.Players[0], r.Players[1]
	p.Buildings = []string{"fermentation_tank"}
	p.Fields[0].Vines = []Card{{ID: "vine-x", Type: "vine", Red: 1}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "harvest", Fields: []int{0}})
	if len(r.Choices) != 1 {
		t.Fatalf("missing fermentation choice: %+v", r.Choices)
	}
	r = restoreT(t, r)
	choiceT(t, r, Action{Option: "skip"})
	if r.TurnID != q.ID || r.Context != nil {
		t.Fatalf("resolved structure choice retained actor turn and context: turn=%s actor=%s next=%s context=%+v", r.TurnID, p.ID, q.ID, r.Context)
	}
}

func TestReviewMafiosoMustAcceptNewTargets(t *testing.T) {
	r := specialWorkerRoom("mafioso", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "mafioso")
	r.Phase = "summer"
	p.Season = "summer"
	p.Hand = []Card{{ID: "vine-old", Type: "vine", Red: 1}, {ID: "vine-new", Type: "vine", White: 1}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "plant", Slot: 2, WorkerType: "mafioso", CardID: "vine-old", Field: 0})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_mafioso" {
		t.Fatalf("missing Mafioso: %+v", r.Choices)
	}
	c := r.Choices[0]
	e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "repeat", CardID: "vine-new", Field: 1})
	if e != nil {
		t.Fatalf("Mafioso ignores second legal vine/field and replays consumed card: %v", e)
	}
}
func TestReviewMercadoCannotFillOldOrder(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	r.Phase = "fall"
	p.Season = "fall"
	p.Buildings = []string{"mercado"}
	p.Hand = []Card{{ID: "old-order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}}
	p.Wines = []Wine{{ID: "red-wine", Type: "red", Value: 1}}
	r.Decks = map[string][]Card{"order": {{ID: "new-order", Type: "order", Requirements: []Wine{{Type: "white", Value: 7}}, Points: 5}}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "draw_order", Slot: 1})
	if len(r.Choices) == 0 {
		return
	}
	c := r.Choices[0]
	e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "use", CardID: "old-order", WineIDs: []string{"red-wine"}})
	if e == nil {
		t.Fatal("Mercado permitted delivery of old hand order instead of just-drawn unfillable order")
	}
}
func TestReviewFermentationMustUseSubmittedRecipe(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	r.Phase = "fall"
	p.Season = "fall"
	p.Buildings = []string{"fermentation_tank"}
	p.Grapes = []Grape{{ID: "red", Color: "red", Value: 1}, {ID: "white", Color: "white", Value: 2}}
	p.Fields[0].Vines = []Card{{ID: "vine-x", Type: "vine", Red: 1}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "harvest", Slot: 1, Fields: []int{0}})
	choiceT(t, r, Action{Option: "make", Recipes: [][]int{{1}}})
	if len(p.Wines) != 1 || p.Wines[0].Type != "white" {
		t.Fatalf("Fermentation ignored chosen white recipe: %+v", p.Wines)
	}
}

func TestReviewOnlyOwnedVisitorStructureBonus(t *testing.T) {
	r := structureRoom()
	r.Phase = "summer"
	p := r.Players[0]
	p.Season = "summer"
	p.Buildings = []string{"tavern"}
	p.Hand = []Card{visitorCardT(t, "summer-01")}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", Slot: 1, CardID: "summer-01"})
	choiceT(t, r, Action{Option: "coins"})
	for _, c := range r.Choices {
		for _, o := range c.Options {
			if strings.Contains(o, "tap_room") {
				t.Fatalf("Tavern alone grants unowned Tap Room choice: %+v", r.Choices)
			}
		}
	}
}
func TestReviewStructureFieldsCannotPlantOrSell(t *testing.T) {
	for _, mode := range []string{"plant", "sell"} {
		t.Run(mode, func(t *testing.T) {
			r := structureRoom()
			r.Phase = "summer"
			p := r.Players[0]
			p.Season = "summer"
			p.Buildings = []string{"silo"}
			p.Fields[0].Structure = "silo"
			p.Hand = []Card{{ID: "vine-x", Type: "vine", Red: 1}}
			a := Action{Type: "place", Space: "plant", Slot: 1, CardID: "vine-x", Field: 0}
			if mode == "sell" {
				a = Action{Type: "place", Space: "flip_field", Slot: 1, Mode: "sell_field", Field: 0}
			}
			if err := r.Apply(p.ID, a); err == nil {
				t.Fatalf("%s accepted field occupied by structure: %+v", mode, p.Fields[0])
			}
		})
	}
}
func TestReviewGazeboManagerPreservesSecondVisitor(t *testing.T) {
	r := tuscanyRoom(3)
	r.Config.Structures = true
	r.Phase = "winter"
	p := r.Players[0]
	p.Season = "winter"
	p.Buildings = []string{"gazebo"}
	p.Hand = []Card{visitorCardT(t, "winter-32"), visitorCardT(t, "winter-34")}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 2, CardID: "winter-32"})
	choiceT(t, r, Action{Option: "action", Space: "tour"})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_influence" {
		t.Fatalf("missing gazebo choice: %+v", r.Choices)
	}
	r = restoreT(t, r)
	choiceT(t, r, Action{Option: "skip"})
	if len(r.Choices) != 1 || r.Choices[0].Visitor == nil || r.Choices[0].Visitor.Stage != "second" {
		t.Fatalf("Gazebo nested choice drops outer winter bonus: choices=%+v context=%+v turn=%s", r.Choices, r.Context, r.TurnID)
	}
}

func TestReviewMercadoBoundOrderSurvivesRestoreAndRejectsOtherSelection(t *testing.T) {
	r := structureRoom()
	p, q := r.Players[0], r.Players[1]
	r.Phase = "fall"
	p.Season = "fall"
	p.Buildings = []string{"mercado"}
	p.Hand = []Card{{ID: "old", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 2}}
	r.Decks = map[string][]Card{"order": {{ID: "new", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}, Points: 3}}}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "draw_order", Slot: 1})
	r = restoreT(t, r)
	p = r.Player(p.ID)
	c := r.Choices[0]
	if len(c.CardIDs) != 1 || c.CardIDs[0] != "new" {
		t.Fatalf("lost bound order: %+v", c)
	}
	a := Action{Type: "choose", ChoiceID: c.ID, Option: "use", CardID: "old", WineIDs: []string{"wine"}}
	rejectUnchangedT(t, r, p.ID, a)
	a.CardID = "new"
	rejectUnchangedT(t, r, q.ID, a)
	mustApplyT(t, r, p.ID, a)
	rejectUnchangedT(t, r, p.ID, a)
	if p.VP != 3 || len(p.Wines) != 0 || len(p.Hand) != 1 || p.Hand[0].ID != "old" {
		t.Fatalf("wrong delivery: %+v", p)
	}
}
func TestReviewOracleDiscardReturnsToFiniteDeck(t *testing.T) {
	r := specialWorkerRoom("oracle", "farmer")
	p := r.Players[0]
	readySpecial(r, p, "oracle")
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "draw_vine", Slot: 2, WorkerType: "oracle"})
	r = restoreT(t, r)
	p = r.Player(p.ID)
	c := r.Choices[0]
	discarded := c.Options[0]
	choiceT(t, r, Action{Option: discarded})
	if len(r.Discards["vine"]) != 1 || r.Discards["vine"][0].ID != discarded {
		t.Fatal("missing finite discard")
	}
	card, ok := r.take("vine")
	if !ok || card.ID != discarded {
		t.Fatal("discard never returned after shuffle")
	}
}
