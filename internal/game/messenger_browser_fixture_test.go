package game

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMessengerBrowserFixtureExport(t *testing.T) {
	target := os.Getenv("MESSENGER_BROWSER_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	r := specialWorkerRoom("messenger", "traveler")
	r.Code, r.Revision = "MESSENGERUI", 1
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	p.Name = "信使玩家"
	readyWorkerForTest(r, p, "messenger")
	p.Grapes = []Grape{{ID: "future-grape", Color: "red", Value: 1}}
	r.Phase, r.TurnID = "summer", p.ID
	for i, q := range r.Players {
		q.Wake = i + 1
		q.Season = "summer"
		q.Passed = i > 0
	}
	rooms := map[string]*Room{r.Code: r}
	sessions := map[string]any{}
	tokens := []string{}
	for i, q := range r.Players {
		token := "messenger-ui-" + string(rune('a'+i))
		tokens = append(tokens, token)
		sessions[token] = map[string]string{"Code": r.Code, "PlayerID": q.ID}
	}
	b, err := json.MarshalIndent(map[string]any{
		"Rooms": rooms, "Sessions": sessions,
		"case": map[string]any{"code": r.Code, "tokens": tokens, "playerId": p.ID},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, b, 0600); err != nil {
		t.Fatal(err)
	}
}
