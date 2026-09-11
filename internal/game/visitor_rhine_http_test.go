package game

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRhineWriteHTTPFixtures(t *testing.T) {
	target := os.Getenv("RHINE_HTTP_FIXTURES")
	if target == "" {
		t.Skip("HTTP fixture export not requested")
	}
	rooms := map[string]*Room{}
	sessions := map[string]map[string]string{}
	cases := []map[string]string{}
	add := func(code, kind string, r *Room, p *Player) {
		r.Code = code
		r.HostID = p.ID
		r.TurnID = p.ID
		r.Config.Visitors = "rhine"
		for i, q := range r.Players {
			q.Name = []string{"甲", "乙", "丙"}[i]
		}
		token, other := "rhine-"+code+"-a", "rhine-"+code+"-b"
		rooms[code] = r
		sessions[token] = map[string]string{"Code": code, "playerId": p.ID}
		sessions[other] = map[string]string{"Code": code, "playerId": r.Players[1].ID}
		cases = append(cases, map[string]string{"code": code, "kind": kind, "token": token, "otherToken": other, "ownerId": p.ID})
	}
	r, p := rhineFixture(t, "rhine-summer-plantDealer", nil)
	add("RHINEPLANT", "plant", r, p)
	r, p = rhineFixture(t, "rhine-winter-virtuoso", func(r *Room, p *Player) {
		for _, c := range RhineCatalog(ExpansionConfig{Visitors: "rhine"}) {
			if c.ID == "rhine-summer-docent" {
				p.Hand = append(p.Hand, c)
			}
		}
	})
	add("RHINEVIRTUOSO", "virtuoso", r, p)
	r, p = rhineFixture(t, "rhine-summer-sonInLaw", nil)
	rhineChoose(t, r, "field", Action{Field: 0})
	r.Config.Visitors = "rhine"
	r.Phase = "winter"
	r.TurnID = p.ID
	r.rhinePrepareWinterTurn()
	add("RHINESON", "son", r, p)

	r, p = rhineFixture(t, "rhine-winter-trainer", func(r *Room, p *Player) { p.TotalWorkers = 2; p.GrandeRemoved = true; p.LargeWorker = false })
	add("RHINETRAINER", "trainer", r, p)
	r, p = rhineFixture(t, "rhine-winter-tutor", func(r *Room, p *Player) {
		p.LargeWorker = false
		r.space("winter_visitor").Occupied = []Seat{{PlayerID: p.ID, Slot: 1, Large: true}}
	})
	add("RHINETUTOR", "tutor", r, p)
	bundle := map[string]any{"Rooms": rooms, "Sessions": sessions, "cases": cases, "method": "Synthetic pending Rhino visitor fixtures; real authenticated HTTP and visible browser actions; no natural-game claim"}
	b, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(target, b, 0600); err != nil {
		t.Fatal(err)
	}
}
