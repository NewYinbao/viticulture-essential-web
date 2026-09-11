package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func moorFixture(id string) (*Room, *Player) {
	r, p := visitorFixture("summer-01")
	r.Config.Visitors = "ee_moor"
	r.Phase = moorCard(id).Type
	r.Context = &ActionContext{ActorID: p.ID, SourceCardID: id, ReturnTurnID: p.ID, Space: r.Phase + "_visitor", TriggerSeat: &Seat{PlayerID: p.ID, Large: true}, LastWorker: true}
	r.Choices = nil
	p.Hand = append(p.Hand, moorCard(id))
	p.Buildings = []string{"trellis", "irrigation"}
	p.Wines = []Wine{{ID: "r2", Type: "red", Value: 2}, {ID: "w3", Type: "white", Value: 3}, {ID: "r4", Type: "red", Value: 4}}
	p.Fields[2].Vines = nil
	return r, p
}
func moorChoose(t *testing.T, r *Room, a Action) {
	t.Helper()
	c := r.Choices[0]
	a.Type = "choose"
	a.ChoiceID = c.ID
	if err := r.Apply(c.PlayerID, a); err != nil {
		t.Fatal(err)
	}
}
func TestMoorAllFortyInitialBranches(t *testing.T) {
	if len(MoorCatalog()) != 40 {
		t.Fatal("Moor catalogue must have 40 unique cards")
	}
	seen := map[string]bool{}
	for _, card := range MoorCatalog() {
		if seen[card.ID] || !card.Implemented || card.RuleSource == "" {
			t.Fatal(card)
		}
		seen[card.ID] = true
		for _, op := range visitorDefs[card.ID].Options {
			t.Run(card.ID+"/"+op, func(t *testing.T) {
				r, p := moorFixture(card.ID)
				a := Action{Option: op, Color: "white", CardID: "order", WineIDs: []string{"r2"}, Grapes: []int{0}, Field: 0, Fields: []int{0}, Building: "yoke", Slot: 2}
				switch card.ID {
				case "moor-summer-01":
					if op == "draw" {
						p.Buildings = append(p.Buildings, "yoke")
					}
				case "moor-summer-02":
					p.Hand = append(p.Hand, moorCard("moor-winter-05"))
					a.CardID = "moor-winter-05"
				case "moor-summer-03":
					a.Grapes = []int{0, 1, 2}
				case "moor-summer-07":
					a.Grapes = []int{0, 1}
				case "moor-summer-08":
					a.TargetIDs = []string{r.Players[1].ID}
				case "moor-summer-11":
					if op == "build" {
						p.Buildings = []string{"trellis"}
					} else {
						p.Hand = append(p.Hand, Card{ID: "irrigation-vine", Type: "vine", Irrigation: true, Red: 1})
						a.CardID = "irrigation-vine"
						a.Field = 2
					}
				case "moor-summer-12":
					a.Buildings = []string{"medium_cellar", "large_cellar", "yoke"}
					if op == "vp" {
						p.Buildings = []string{"trellis", "irrigation", "yoke", "cottage", "medium_cellar", "large_cellar"}
					}
				case "moor-summer-13":
					if op == "grape" {
						p.Buildings = append(p.Buildings, "windmill")
					}
				case "moor-summer-14":
					if op == "vp" {
						p.Grapes = append(p.Grapes, Grape{ID: "a", Color: "white", Value: 8}, Grape{ID: "b", Color: "red", Value: 8})
					}
				case "moor-summer-17":
					if op == "wine" {
						p.Buildings = append(p.Buildings, "tasting_room")
					}
				case "moor-summer-18":
					a.CardIDs = []string{"s0"}
				case "moor-winter-01":
					r.Spaces = append(r.Spaces, Space{ID: "old-any", Season: "any", Occupied: []Seat{{PlayerID: p.ID, Slot: 1, Season: "summer"}, {PlayerID: p.ID, Slot: 2, Season: "summer"}}})
					a.TargetIDs = []string{"old-any", "old-any"}
					a.Fields = []int{1, 2}
				case "moor-winter-03":
					a.Slot = 1
				case "moor-winter-07":
				case "moor-winter-09":
					p.Buildings = append(p.Buildings, "cottage")
				case "moor-winter-12":
					p.Hand = append(p.Hand, moorCard("moor-summer-07"))
					a.CardID = "moor-summer-07"
				case "moor-winter-13":
					if op == "plant" {
						a.CardID = "plant"
						a.Field = 2
					} else {
						a.WineIDs = []string{"r2", "w3", "r4"}
					}
				case "moor-winter-15":
					a.CardIDs = []string{"s0", "s1", "s2"}
					a.Colors = []string{"vine", "vine", "winter"}
				}
				if err := r.queueVisitor(p, card.ID); err != nil {
					t.Fatal(err)
				}
				before, _ := json.Marshal(p)
				moorChoose(t, r, a)
				after, _ := json.Marshal(p)
				if bytes.Equal(before, after) && len(r.Choices) < 1 && card.ID != "moor-summer-08" {
					t.Fatal("branch produced no resource change or continuation")
				}
				if card.ID == "moor-summer-08" && r.Players[1].Fields[0].Harvested {
					t.Fatal("Grape Thief consumed opponent harvest")
				}
			})
		}
	}
}
func TestMoorPrivatePoolPersistsAndDoesNotLeak(t *testing.T) {
	for _, id := range []string{"moor-summer-15", "moor-winter-17"} {
		t.Run(id, func(t *testing.T) {
			r, p := moorFixture(id)
			r.Context.Held = nil
			r.Choices = nil
			if err := r.queueVisitor(p, id); err != nil {
				t.Fatal(err)
			}
			op := "draw"
			if id == "moor-winter-17" {
				op = "order"
			}
			moorChoose(t, r, Action{Option: op})
			if len(r.Context.Held) != 3 {
				t.Fatal("private draw count")
			}
			if len(r.Revealed) != 0 {
				t.Fatal("private pool made public")
			}
			other, _ := json.Marshal(r.View(r.Players[1].ID))
			for _, c := range r.Context.Held {
				if bytes.Contains(other, []byte(c.ID)) {
					t.Fatal("private card leaked")
				}
			}
			saved, _ := json.Marshal(r)
			var restored Room
			if err := json.Unmarshal(saved, &restored); err != nil {
				t.Fatal(err)
			}
			r = &restored
			p = r.Player(p.ID)
			before, _ := json.Marshal(r)
			c := r.Choices[0]
			err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "keep", CardIDs: []string{"order"}})
			if err == nil {
				t.Fatal("unrelated card accepted")
			}
			after, _ := json.Marshal(r)
			if !bytes.Equal(before, after) {
				t.Fatal("invalid private choice changed pool")
			}
			n := 1
			if id == "moor-summer-15" {
				n = 2
			}
			ids := []string{}
			for _, c := range r.Context.Held[:n] {
				ids = append(ids, c.ID)
			}
			moorChoose(t, r, Action{Option: "keep", CardIDs: ids})
			for _, id := range ids {
				if cardIndex(p, id, "order") < 0 {
					t.Fatal("selected card missing")
				}
			}
		})
	}
}
func TestMoorNestedVisitorsPreserveParentAndSecond(t *testing.T) {
	r, p := moorFixture("moor-summer-02")
	p.Hand = append(p.Hand, moorCard("moor-winter-12"), moorCard("moor-summer-07"))
	r.Context.Remaining = 1
	r.Context.CoinAfter = 2
	r.Choices = nil
	if err := r.queueVisitor(p, "moor-summer-02"); err != nil {
		t.Fatal(err)
	}
	coins := p.Coins
	moorChoose(t, r, Action{Option: "play", CardID: "moor-winter-12"})
	if r.Context.Parent == nil || r.Context.SourceCardID != "moor-winter-12" {
		t.Fatal("child context not established")
	}
	moorChoose(t, r, Action{Option: "play", CardID: "moor-summer-07"})
	if r.Context.Parent.Parent == nil {
		t.Fatal("grandchild context lost")
	}
	data, _ := json.Marshal(r)
	var rr Room
	json.Unmarshal(data, &rr)
	r = &rr
	p = r.Player(p.ID)
	moorChoose(t, r, Action{Option: "coins"})
	if r.Context == nil || r.Context.Parent != nil || r.Choices[0].Visitor.Stage != "second" {
		t.Fatal("parent second visitor not resumed")
	}
	if p.Coins != coins+p.Wake {
		t.Fatal("parent coin bonus paid early")
	}
	moorChoose(t, r, Action{Option: "skip"})
	if r.Context != nil || p.Coins != coins+p.Wake+2 {
		t.Fatal("outer settlement incorrect")
	}
}
func TestMoorConcreteAmountsAndSpecialTiming(t *testing.T) {
	tests := []struct {
		id, op    string
		coins, vp int
	}{
		{"moor-summer-06", "coins", 7, 0}, {"moor-summer-06", "vp", 0, 2}, {"moor-summer-07", "coins", 1, 0},
		{"moor-winter-02", "wine", -7, 0}, {"moor-winter-04", "vp", -9, 3}, {"moor-winter-06", "coins", 3, 0}, {"moor-winter-09", "coins", 3, 0}, {"moor-winter-11", "coins", 4, 0}, {"moor-winter-11", "vp", 0, 2}, {"moor-winter-14", "wine", -9, 0}, {"moor-winter-18", "wine", -9, 0}, {"moor-winter-20", "wine", -3, 0}, {"moor-summer-20", "wine", -3, 0},
	}
	for _, tc := range tests {
		t.Run(tc.id+tc.op, func(t *testing.T) {
			r, p := moorFixture(tc.id)
			p.Wines = nil
			r.Context.Held = []Card{moorCard(tc.id)}
			oldCoins, oldVP := p.Coins, p.VP
			if err := r.moorEffect(p, &VisitorStep{CardID: tc.id, Stage: "effect", ActorID: p.ID}, Action{Option: tc.op, Color: "red", Grapes: []int{0}}); err != nil {
				t.Fatal(err)
			}
			if p.Coins != oldCoins+tc.coins || p.VP != oldVP+tc.vp {
				t.Fatalf("unexpected coin/vp delta %d %d", p.Coins-oldCoins, p.VP-oldVP)
			}
		})
	}
	r, p := moorFixture("moor-summer-09")
	before := p.VP
	moorChooseDirect(t, r, p, "moor-summer-09", Action{Option: "draw"})
	if p.VP != before+1 {
		t.Fatal("grande guest bonus")
	}
	r, p = moorFixture("moor-summer-16")
	p.TotalWorkers = 3
	workers := p.Workers
	coins := p.Coins
	moorChooseDirect(t, r, p, "moor-summer-16", Action{Option: "train"})
	if p.Coins != coins-3 || p.Workers != workers+1 {
		t.Fatal("Recruiter cost/readiness")
	}
}
func moorChooseDirect(t *testing.T, r *Room, p *Player, id string, a Action) {
	t.Helper()
	if err := r.moorEffect(p, &VisitorStep{CardID: id, Stage: "effect", ActorID: p.ID}, a); err != nil {
		t.Fatal(err)
	}
}
func TestMoorFuselFruitAndContracts(t *testing.T) {
	r, p := moorFixture("moor-summer-05")
	r.Choices = nil
	before := p.VP
	moorChooseDirect(t, r, p, "moor-summer-05", Action{Option: "offer"})
	for len(r.Choices) > 0 {
		moorChoose(t, r, Action{Option: "grape", Color: "white"})
	}
	if p.VP != before+2 {
		t.Fatal("Fusel must award once")
	}
	r, p = moorFixture("moor-summer-04")
	p.Fields[0].FruitDealer = true
	r.Choices = nil
	if err := r.performPlacement(p, Action{Space: "harvest", Fields: []int{0}}, false); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 {
		t.Fatal("Fruit Dealer must trigger once")
	}
	before = p.VP
	moorChoose(t, r, Action{Option: "vp"})
	if p.VP != before+1 {
		t.Fatal("fruit reward")
	}
	r, p = moorFixture("moor-summer-06")
	p.MoorContracts = []Card{moorCard("moor-summer-06"), moorCard("moor-summer-10")}
	p.Income = 3
	p.Grapes = nil
	before = p.VP
	r.finishGame()
	if p.VP != before-8 {
		t.Fatal("end game contract penalty")
	}
	r.finishGame()
	if p.VP != before-8 {
		t.Fatal("contract penalty applied twice")
	}
}
func TestMoorDualSeasonAndInvalidAtomicity(t *testing.T) {
	r, p := moorFixture("moor-winter-10")
	r.Phase = "summer"
	r.Context = nil
	r.Choices = nil
	p.Coins = 0
	if err := r.Apply(p.ID, Action{Type: "place", Space: "summer_visitor", CardID: "moor-winter-10", Slot: 2}); err != nil {
		t.Fatal(err)
	}
	moorChoose(t, r, Action{Option: "resolve", Grapes: []int{0}})
	if p.Coins != 4 {
		t.Fatal("summer Grape Ice branch")
	}
	if len(r.Discards["winter"]) == 0 {
		t.Fatal("blue card must discard to winter")
	}
	r, p = moorFixture("moor-summer-12")
	r.Choices = nil
	r.queueVisitor(p, "moor-summer-12")
	before, _ := json.Marshal(r)
	c := r.Choices[0]
	err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "build", Buildings: []string{"medium_cellar", "large_cellar", "medium_cellar"}})
	if err == nil {
		t.Fatal("duplicate building accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("multi-build failed to rollback")
	}
}
func TestMoorCatalogueNoBaseCollisionAndFields(t *testing.T) {
	base := map[string]bool{}
	for _, c := range Catalog() {
		base[c.ID] = true
	}
	for _, c := range MoorCatalog() {
		if base[c.ID] || !strings.HasPrefix(c.ID, "moor-"+c.Type+"-") {
			t.Fatal(c.ID)
		}
		r, p := moorFixture(c.ID)
		r.visitorPrompt(p, c.ID, "effect", visitorDefs[c.ID].Options)
		choice := r.Choices[0]
		for _, op := range choice.Options {
			if _, ok := choice.OptionFields[op]; !ok || choice.Labels[op] == "" {
				t.Fatal(fmt.Sprint(c.ID, op))
			}
		}
	}
	if _, err := CatalogForConfig(ExpansionConfig{Visitors: "ee_moor"}); err != nil {
		t.Fatal(err)
	}
}

func TestMoorSpecialTrainingAndMercadoOrdering(t *testing.T) {
	r, p := moorFixture("moor-summer-16")
	r.Config.SpecialWorkers = true
	r.SpecialWorkerPool = []string{"chef", "soldato"}
	p.TotalWorkers = 3
	p.SpecialWorkers = nil
	r.Players[1].Buildings = append(r.Players[1].Buildings, "academy")
	workers, coins := p.Workers, p.Coins
	other := r.Players[1].Coins
	moorChooseDirect(t, r, p, "moor-summer-16", Action{Option: "train", SpecialWorker: "chef"})
	if p.Coins != coins-5 || r.Players[1].Coins != other+1 || p.Workers != workers || !p.specialWorkerReady("chef", r.Year) {
		t.Fatal("special worker price/readiness or Academy toll")
	}
	r, p = moorFixture("moor-summer-02")
	r.Config.Structures = true
	p.Buildings = append(p.Buildings, "mercado")
	r.Decks["order"] = []Card{{ID: "mercado-test-order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 1}}
	p.Hand = append(p.Hand, moorCard("moor-winter-05"))
	r.Choices = nil
	r.queueVisitor(p, "moor-summer-02")
	moorChoose(t, r, Action{Option: "play", CardID: "moor-winter-05"})
	if r.Choices[0].Kind != "structure_mercado" {
		t.Fatal("draw trigger must precede nested visitor effect")
	}
	moorChoose(t, r, Action{Option: "skip"})
	if r.Choices[0].Visitor.CardID != "moor-winter-05" {
		t.Fatal("nested visitor continuation lost")
	}
	moorChoose(t, r, Action{Option: "vp"})
	if r.Context != nil {
		t.Fatal("nested action did not finish")
	}
}
func TestNestedActionBonusCompletesWithoutExtraChoice(t *testing.T) {
	r, p := moorFixture("moor-summer-07")
	parent := r.Context
	before := len(p.Hand)
	if err := r.startNestedAction(p, Action{Space: "draw_vine"}, true, nil); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != before+2 {
		t.Fatal("nested action lost printed bonus")
	}
	if r.Context.Parent != parent {
		t.Fatal("parent context lost")
	}
	r.finishNestedParentEffect(p, &VisitorStep{CardID: "moor-summer-07", Stage: "effect", ActorID: p.ID}, parent)
	r.Choices = nil
	r.finishVisitorContext()
	if r.Context != nil {
		t.Fatal("nested plain action did not settle")
	}
}
func TestMoorAgingChainsAndPrivateOrderRestrictions(t *testing.T) {
	r, p := moorFixture("moor-winter-16")
	p.Grapes = []Grape{{ID: "a", Color: "red", Value: 2}, {ID: "b", Color: "red", Value: 3}}
	if !moorCanAgeGrapes(p, 2) {
		t.Fatal("descending aging chain incorrectly unavailable")
	}
	if err := moorAgeGrapes(p, []int{0, 1}, 2, 1); err != nil {
		t.Fatal(err)
	}
	p.Wines = []Wine{{ID: "w", Type: "red", Value: 1}}
	r.Context.Held = []Card{{ID: "private", Type: "order", Requirements: []Wine{{Type: "white", Value: 9}}}}
	if r.moorOptionReason(p, &VisitorStep{CardID: "moor-summer-15", Stage: "private"}, "fill") == "" {
		t.Fatal("old hand order must not enable Principal's private fill")
	}
}
