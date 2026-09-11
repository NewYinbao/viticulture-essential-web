package game

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPlannerSpecialBrowserFixtureExport(t *testing.T) {
	target := os.Getenv("PLANNER_SPECIAL_BROWSER_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	r, p := plannerSpecialFixture(t, "farmer")
	r.Code, r.Revision = "PLANNERFARMER", 1
	p.Name = "Planner Farmer"
	r.Decks["order"] = []Card{{ID: "order-1", Type: "order"}, {ID: "order-2", Type: "order"}, {ID: "order-3", Type: "order"}}
	sessions := map[string]any{}
	tokens := []string{}
	for i, q := range r.Players {
		token := "planner-farmer-" + string(rune('a'+i))
		tokens = append(tokens, token)
		sessions[token] = map[string]string{"Code": r.Code, "PlayerID": q.ID}
	}
	b, err := json.MarshalIndent(map[string]any{
		"Rooms": map[string]*Room{r.Code: r}, "Sessions": sessions,
		"case": map[string]any{"code": r.Code, "tokens": tokens, "playerId": p.ID},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, b, 0600); err != nil {
		t.Fatal(err)
	}
}
