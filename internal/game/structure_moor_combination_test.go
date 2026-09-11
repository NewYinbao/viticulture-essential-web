package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTrussFermentationPreservesMandatoryAging(t *testing.T) {
	for _, surplus := range []bool{false, true} {
		t.Run(map[bool]string{false: "cannot_consume_only_grape", true: "may_use_surplus"}[surplus], func(t *testing.T) {
			r, p := moorFixture("moor-winter-16")
			r.Config.Structures = true
			p.VP = 0
			p.Buildings = append(p.Buildings, "fermentation_tank")
			p.Grapes = nil
			p.Wines = nil
			for i := range p.Fields {
				p.Fields[i].Vines = nil
				p.Fields[i].Harvested = false
			}
			vine := Card{ID: "truss-vine", Type: "vine", Red: 2}
			if surplus {
				vine.White = 1
			}
			p.Fields[0].Vines = []Card{vine}
			r.Context.Remaining = 1
			if err := r.queueVisitor(p, "moor-winter-16"); err != nil {
				t.Fatal(err)
			}
			moorChoose(t, r, Action{Option: "harvest", Fields: []int{0}})
			if len(r.Choices) != 2 || r.Choices[0].Kind != "structure_fermentation" || r.Choices[1].Visitor.Stage != "age" {
				t.Fatalf("wrong ordering: %+v", r.Choices)
			}
			// The complete queue and forced X survive a process-style JSON restore.
			data, _ := json.Marshal(r)
			var restored Room
			if err := json.Unmarshal(data, &restored); err != nil {
				t.Fatal(err)
			}
			r = &restored
			p = r.Player(p.ID)
			choiceID := r.Choices[0].ID
			a := Action{Type: "choose", ChoiceID: choiceID, Option: "make", Recipes: [][]int{{0}}}
			if surplus {
				if err := r.Apply(p.ID, a); err != nil {
					t.Fatal(err)
				}
				if len(p.Wines) != 1 || len(p.Grapes) != 1 {
					t.Fatal("surplus fermentation failed")
				}
			} else {
				before, _ := json.Marshal(r)
				if err := r.Apply(p.ID, a); err == nil {
					t.Fatal("fermentation consumed the mandatory aging grape")
				}
				after, _ := json.Marshal(r)
				if !bytes.Equal(before, after) {
					t.Fatal("rejected fermentation did not roll back whole room")
				}
				moorChoose(t, r, Action{Option: "skip"})
			}
			if r.Choices[0].Visitor.MoorCount != 1 {
				t.Fatal("mandatory X lost")
			}
			moorChoose(t, r, Action{Option: "age", Grapes: []int{0}})
			if r.Choices[0].Visitor.Stage != "second" {
				t.Fatal("outer bonus visitor lost")
			}
			moorChoose(t, r, Action{Option: "skip"})
			if r.Context != nil || len(r.Choices) != 0 {
				t.Fatal("Truss did not finish")
			}
		})
	}
}
func TestCouncilmanMercadoFillRestoreAndNestedVisitor(t *testing.T) {
	r, p := moorFixture("moor-summer-02")
	r.Config.Structures = true
	p.VP = 0
	p.Coins = 0
	p.Buildings = append(p.Buildings, "mercado", "inn")
	p.Wines = []Wine{{ID: "red", Type: "red", Value: 1}, {ID: "white", Type: "white", Value: 1}}
	order := Card{ID: "council-private-order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}
	r.Decks["order"] = []Card{order}
	p.Hand = append(p.Hand, moorCard("moor-winter-05"))
	r.Context.Remaining = 1
	if err := r.queueVisitor(p, "moor-summer-02"); err != nil {
		t.Fatal(err)
	}
	moorChoose(t, r, Action{Option: "play", CardID: "moor-winter-05"})
	if r.Context.Parent == nil || r.Choices[0].Kind != "structure_mercado" || r.Choices[1].Visitor.CardID != "moor-winter-05" {
		t.Fatal("nested stack ordering lost")
	}
	public, _ := json.Marshal(r.View(r.Players[1].ID))
	if bytes.Contains(public, []byte(order.ID)) {
		t.Fatal("opponent saw just-drawn private order")
	}
	data, _ := json.Marshal(r)
	var restored Room
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	r = &restored
	p = r.Player(p.ID)
	a := Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "use", CardID: order.ID, WineIDs: []string{"red"}}
	if err := r.Apply(r.Players[1].ID, a); err == nil {
		t.Fatal("opponent resolved private Mercado")
	}
	if err := r.Apply(p.ID, a); err != nil {
		t.Fatal(err)
	}
	if p.VP != 2 || len(p.Wines) != 1 || r.Choices[0].Visitor.CardID != "moor-winter-05" {
		t.Fatal("Mercado did not finish before child effect")
	}
	if err := r.Apply(p.ID, a); err == nil {
		t.Fatal("replayed Mercado accepted")
	}
	moorChoose(t, r, Action{Option: "vp"})
	if p.VP != 3 || p.Coins != 2 {
		t.Fatalf("child should score only remaining white type, Inn once per visitor: VP=%d coins=%d", p.VP, p.Coins)
	}
	if r.Choices[0].Visitor.Stage != "second" || r.Context.Parent != nil {
		t.Fatal("parent second visitor not restored")
	}
	moorChoose(t, r, Action{Option: "skip"})
	if r.Context != nil || len(r.Choices) != 0 {
		t.Fatal("Councilman did not finish")
	}
}
