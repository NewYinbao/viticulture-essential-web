package game

import "testing"

func TestTrussHarvestMachineAgesActualFieldCount(t *testing.T) {
	for _, supplied := range [][]int{nil, {0}} {
		r, p := moorFixture("moor-winter-16")
		r.Config.Structures = true
		p.Buildings = append(p.Buildings, "harvest_machine")
		p.Grapes = nil
		for i := range p.Fields {
			p.Fields[i].Sold = false
			p.Fields[i].Harvested = false
			p.Fields[i].Vines = []Card{{Red: i + 1}}
		}
		if err := r.queueVisitor(p, "moor-winter-16"); err != nil {
			t.Fatal(err)
		}
		moorChoose(t, r, Action{Option: "harvest", HarvestAll: true, Fields: supplied})
		if len(p.Grapes) != 3 || len(r.Choices) != 1 || r.Choices[0].Visitor.MoorCount != 3 {
			t.Fatalf("all three harvested fields must determine mandatory aging: fields=%v choices=%+v", supplied, r.Choices)
		}
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "age", Grapes: []int{0}}); err == nil {
			t.Fatal("client's shorter fields list reduced mandatory aging")
		}
		moorChoose(t, r, Action{Option: "age", Grapes: []int{0, 1, 2}})
	}
}
