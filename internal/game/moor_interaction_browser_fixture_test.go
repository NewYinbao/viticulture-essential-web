package game

import (
	"encoding/json"
	"os"
	"testing"
)

// Seeded starting states for actual HTTP/UI integration tests, not natural games.
func TestMoorInteractionBrowserFixtureExport(t *testing.T) {
	target := os.Getenv("MOOR_INTERACTION_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	rooms, sessions := map[string]*Room{}, map[string]any{}
	cases := []map[string]string{}
	for _, fixture := range []struct{ code, card string }{{"COUNCIL", "moor-summer-02"}, {"TRUSS", "moor-winter-16"}, {"FRUIT", "moor-summer-04"}} {
		r, p := moorFixture(fixture.card)
		r.Code, r.Revision = fixture.code, 1
		r.Config.Structures = true
		r.Context, r.Choices = nil, nil
		p.Name, p.Coins, p.VP = "Moor交互夹具", 10, 0
		p.Workers, p.LargeWorker, p.TotalWorkers = 2, true, 3
		p.Hand, p.Grapes, p.Wines, p.Buildings = []Card{moorCard(fixture.card)}, nil, nil, nil
		p.StructureSlots = []string{"", ""}
		for i := range p.Fields {
			p.Fields[i].Vines = nil
			p.Fields[i].Harvested = false
		}
		for _, other := range r.Players {
			other.PapaResolved = true
			if other.ID != p.ID {
				other.Passed = true
			}
		}
		switch fixture.code {
		case "COUNCIL":
			p.Coins = 0
			p.Buildings, p.StructureSlots = []string{"mercado", "inn"}, []string{"mercado", "inn"}
			p.Hand = append(p.Hand, moorCard("moor-winter-05"), Card{ID: "old-private-order", Type: "order", Requirements: []Wine{{Type: "sparkling", Value: 7}}, Points: 4})
			p.Wines = []Wine{{ID: "red", Type: "red", Value: 1}, {ID: "white", Type: "white", Value: 1}}
			r.Decks["order"] = []Card{{ID: "new-private-order", Name: "仅限本次新订单", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}}
		case "TRUSS":
			p.Buildings, p.StructureSlots = []string{"fermentation_tank", "medium_cellar"}, []string{"fermentation_tank", ""}
			p.Fields[0].Vines = []Card{{ID: "truss-vine-a", Type: "vine", Red: 2, White: 2}}
			p.Fields[1].Vines = []Card{{ID: "truss-vine-b", Type: "vine", Red: 1}}
		case "FRUIT":
			p.Coins = 4
			p.Buildings = []string{"yoke"}
			p.Fields[0].Vines = []Card{{ID: "fruit-vine", Type: "vine", Red: 2, White: 1}}
		}
		rooms[r.Code] = r
		for _, player := range r.Players {
			token := r.Code + "-" + player.ID
			sessions[token] = map[string]string{"Code": r.Code, "PlayerID": player.ID}
		}
		cases = append(cases, map[string]string{"code": r.Code, "card": fixture.card, "token": r.Code + "-" + p.ID, "opponentToken": r.Code + "-" + r.Players[1].ID, "playerId": p.ID})
	}
	data, err := json.MarshalIndent(map[string]any{"Rooms": rooms, "Sessions": sessions, "cases": cases}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(target, data, 0600); err != nil {
		t.Fatal(err)
	}
}
