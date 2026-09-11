package game

import (
	"encoding/json"
	"os"
	"testing"
)

// Isolated seeded interaction fixtures, not natural-play evidence.
func TestWorkerBrowserFixtureExport(t *testing.T) {
	target := os.Getenv("WORKER_BROWSER_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	rooms, sessions := map[string]*Room{}, map[string]any{}
	cases := []map[string]string{}
	for _, code := range []string{"CHEFBUMP", "SPECIALONLY", "DISCOUNTTRAIN", "TRAINSHORT", "TRAINEXACT", "FARMERTRAIN"} {
		r := specialWorkerRoom("chef", "soldato")
		r.Code, r.Revision = code, 1
		p, q := r.Players[0], r.Players[1]
		p.Name, q.Name = "交互夹具", "对手"
		p.PapaResolved = true
		p.Workers, p.LargeWorker = 2, true
		space := "tour"
		switch code {
		case "CHEFBUMP":
			readySpecial(r, p, "chef")
			p.Workers, p.LargeWorker, p.Coins = 0, false, 0
			r.space(space).Occupied = []Seat{{PlayerID: q.ID, Slot: 1, WorkerType: "soldato"}}
		case "SPECIALONLY":
			readySpecial(r, p, "chef")
			p.Workers, p.LargeWorker, p.Coins = 0, false, 5
		case "FARMERTRAIN":
			space, r.Phase = "train", "winter"
			for _, player := range r.Players {
				player.Season = "winter"
			}
			r.SpecialWorkerPool = []string{"farmer", "chef"}
			readySpecial(r, p, "farmer")
			p.Workers, p.LargeWorker, p.Coins = 0, false, 3
			r.space(space).Occupied = []Seat{{PlayerID: q.ID, Slot: 1}}
		case "DISCOUNTTRAIN", "TRAINSHORT", "TRAINEXACT":
			space, r.Phase = "train", "winter"
			for _, player := range r.Players {
				player.Season = "winter"
			}
			p.Coins = 4
			if code != "DISCOUNTTRAIN" {
				r.space(space).Occupied = []Seat{{PlayerID: q.ID, Slot: 1, WorkerType: "soldato"}}
			}
			if code == "TRAINEXACT" {
				p.Coins = 5
			}
		}
		rooms[code] = r
		token := code + "-" + p.ID
		sessions[token] = map[string]string{"Code": code, "PlayerID": p.ID}
		cases = append(cases, map[string]string{"code": code, "token": token, "playerId": p.ID, "space": space})
	}
	data, err := json.MarshalIndent(map[string]any{"Rooms": rooms, "Sessions": sessions, "cases": cases}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(target, data, 0600); err != nil {
		t.Fatal(err)
	}
}
