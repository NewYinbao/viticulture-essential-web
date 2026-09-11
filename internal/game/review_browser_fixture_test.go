package game

import (
	"encoding/json"
	"os"
	"testing"
)

func TestReviewBrowserFixtures(t *testing.T) {
	target := os.Getenv("REVIEW_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	rooms := map[string]*Room{}
	sessions := map[string]any{}
	add := func(code string, r *Room) {
		r.Code = code
		r.Revision = 1
		r.Ruleset = "ee-base-v1"
		if r.tuscany() {
			r.Ruleset = "tuscany-essential-v1"
		}
		for i, p := range r.Players {
			p.Name = "审查玩家"
			token := code
			if i > 0 {
				token = code + "-other"
			}
			sessions[token] = map[string]string{"Code": code, "playerId": p.ID}
		}
		r.View(r.Players[0].ID)
		rooms[code] = r
	}
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"guest_house"}
	p.StructureSlots = []string{"guest_house", ""}
	p.Hand = []Card{visitorCardT(t, "summer-01"), visitorCardT(t, "winter-34")}
	add("GUEST", r)
	r = structureRoom()
	p = r.Players[0]
	r.Config.Board = "ee"
	r.Phase = "winter"
	r.Spaces = r.spacesForRules()
	p.Season = "winter"
	p.Buildings = []string{"fermentation_tank"}
	p.StructureSlots = []string{"fermentation_tank", ""}
	p.Grapes = []Grape{{ID: "red", Color: "red", Value: 1}, {ID: "white", Color: "white", Value: 2}}
	p.Fields[0].Vines = []Card{{ID: "vine-x", Type: "vine", Red: 1}}
	add("FERMENT", r)
	r = specialWorkerRoom("mafioso", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "mafioso")
	r.Phase = "summer"
	p.Season = "summer"
	p.Hand = []Card{{ID: "vine-old", Type: "vine", Name: "红葡萄", Red: 1}, {ID: "vine-new", Type: "vine", Name: "白葡萄", White: 1}}
	add("MAFIOSO", r)
	r = structureRoom()
	p = r.Players[0]
	r.Phase = "fall"
	p.Season = "fall"
	p.Buildings = []string{"mercado"}
	p.StructureSlots = []string{"mercado", ""}
	p.Hand = []Card{{ID: "old-order", Type: "order", Name: "旧订单", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}}
	p.Wines = []Wine{{ID: "red-wine", Type: "red", Value: 1}, {ID: "white-wine", Type: "white", Value: 2}}
	r.Decks = map[string][]Card{"order": {{ID: "new-order", Type: "order", Name: "新订单", Requirements: []Wine{{Type: "white", Value: 2}}, Points: 3}}}
	add("MARKET", r)
	b, e := json.Marshal(map[string]any{"Rooms": rooms, "Sessions": sessions})
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(target, b, 0600); e != nil {
		t.Fatal(e)
	}
}
