package game

import (
	"encoding/json"
	"os"
	"testing"
)

// Opt-in fixture export contains synthetic players only, never a saved room.
func TestRhineWriteBrowserFixtures(t *testing.T) {
	target := os.Getenv("RHINE_BROWSER_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	var views []map[string]any
	for id := range rhineDefinitions {
		r, p := rhineFixture(t, id, func(r *Room, p *Player) {
			r.Config = ExpansionConfig{Board: "tuscany", Visitors: "rhine"}
			r.Spaces = NewTuscanySpaces(3)
			p.Influence = map[string]int{"pisa": 1}
			p.Buildings = append(p.Buildings, "medium_cellar", "large_cellar")
			p.Wines = []Wine{{ID: "w1", Type: "red", Value: 3}, {ID: "w2", Type: "white", Value: 2}}
			p.Grapes = []Grape{{ID: "g1", Color: "red", Value: 3}, {ID: "g2", Color: "white", Value: 2}}
		})
		for _, o := range r.Choices[0].Options {
			c := r.Choices[0]
			c.Options = []string{o}
			c.ID = id + "_" + o
			r.Choices[0] = c
			r.Code = id + o
			views = append(views, r.View(p.ID))
		}
	}
	data, err := json.Marshal(views)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0600); err != nil {
		t.Fatal(err)
	}
}
