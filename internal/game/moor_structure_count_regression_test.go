package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMoorBuilderCountsEachConstructedStructureOnce(t *testing.T) {
	r, p := moorFixture("moor-summer-12")
	r.Config.Structures = true
	p.Buildings = nil
	p.StructureSlots = []string{"", ""}
	p.Fields[0].Vines = nil
	p.Coins = 100
	p.VP = 0
	for i, id := range []string{"cask", "wine_cave", "trading_post"} {
		p.Hand = append(p.Hand, structureCardForID(id))
		a := Action{Building: id, Mode: "mat", StructureSlot: i}
		if i == 2 {
			a.Mode, a.Field = "field", 0
		}
		if err := r.buildStructure(p, a, 0); err != nil {
			t.Fatal(err)
		}
	}
	if count := moorBuildingCount(p); count != 3 {
		t.Fatalf("three physical structures counted as %d", count)
	}
	if err := r.queueVisitor(p, "moor-summer-12"); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(r)
	a := Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "vp"}
	if err := r.Apply(p.ID, a); err == nil {
		t.Fatal("Builder scored with only three physical structures")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid Builder score changed state")
	}
	for _, id := range []string{"trellis", "irrigation", "yoke"} {
		if err := r.visitorBuild(p, id, 0); err != nil {
			t.Fatal(err)
		}
	}
	if count := moorBuildingCount(p); count != 6 {
		t.Fatalf("six physical structures counted as %d", count)
	}
	points := p.VP
	if err := r.Apply(p.ID, a); err != nil {
		t.Fatal(err)
	}
	if p.VP != points+2 {
		t.Fatal("Builder did not award two points at six structures")
	}
}
