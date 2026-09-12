package game

import "testing"

func TestCardLibraryTotalsAndIsolation(t *testing.T) {
	groups := CardLibrary()
	expected := []int{190, 40, 80, 36, 11}
	ids := map[string]bool{}
	total := 0
	dependent := 0
	for i, g := range groups {
		if len(g.Cards) != expected[i] {
			t.Fatalf("%s: %d", g.ID, len(g.Cards))
		}
		for _, c := range g.Cards {
			if ids[c.ID] {
				t.Fatalf("duplicate %s", c.ID)
			}
			ids[c.ID] = true
			total++
			if c.RequiresTuscany {
				dependent++
			}
		}
	}
	if total != 357 || dependent != 4 {
		t.Fatalf("total %d; Tuscany-only %d", total, dependent)
	}
	groups[0].Cards[0].Name = "changed"
	if CardLibrary()[0].Cards[0].Name == "changed" {
		t.Fatal("mutable shared catalogue")
	}
}
