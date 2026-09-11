package game

import "testing"

func TestAutumnBuildStructureMatAndField(t *testing.T) {
	for _, mode := range []string{"mat", "field", "build"} {
		t.Run(mode, func(t *testing.T) {
			r := structureRoom()
			p := r.Players[0]
			r.Phase = "fall"
			p.Season = "fall"
			structureHand(p, "storehouse")
			beforeCoins := p.Coins
			a := Action{Type: "place", Space: "build_tour", Mode: mode, Building: "structure-storehouse", Field: 0, StructureSlot: 0, DeclineBonus: true}
			mustApplyT(t, r, p.ID, a)
			if !p.hasStructure("storehouse") || p.VP != 1 || p.Coins != beforeCoins-2 {
				t.Fatalf("wrong autumn build: %+v", p)
			}
			if mode == "field" {
				if p.Fields[0].Structure != "storehouse" {
					t.Fatal("field destination lost")
				}
			} else if p.StructureSlots[0] != "storehouse" {
				t.Fatal("mat destination lost")
			}
		})
	}
}
