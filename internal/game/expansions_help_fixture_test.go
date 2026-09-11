package game

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Optional browser fixture export exists only in the test binary. It never adds
// a production endpoint or makes unfinished modules available to configure/start.
func TestExpansionHelpBrowserFixtures(t *testing.T) {
	type session struct {
		Code     string
		PlayerID string `json:"playerId"`
	}
	bundle := struct {
		Rooms    map[string]*Room
		Sessions map[string]session
	}{map[string]*Room{}, map[string]session{}}
	newFixture := func(code string, cfg ExpansionConfig) *Room {
		r := tuscanyRoom(3)
		r.Code, r.Revision, r.Ruleset = code, 1, "ee-base-v1"
		// Initialize the real EE finite decks before applying a development config.
		r.Config = ExpansionConfig{}
		if err := r.InitDecks(); err != nil {
			t.Fatal(err)
		}
		r.Config = cfg.Normalized()
		if r.tuscany() {
			r.Ruleset = "tuscany-essential-v1"
		}
		r.tuscanySpring()
		r.Phase, r.TurnID = "summer", r.Players[0].ID
		for i, p := range r.Players {
			p.Name = []string{"山丘庄主", "河谷庄主", "林间庄主"}[i]
			p.Wake, p.Season, p.Passed = i+2, "summer", false
			p.Coins, p.Income, p.VP = 9, 2, 0
			p.Grapes = []Grape{{ID: code + fmt.Sprint(i) + "-red", Color: "red", Value: 2}}
			p.Hand = []Card{}
			for _, typ := range []string{"vine", "summer", "winter"} {
				r.draw(p, typ)
			}
			r.WakeSlots[i+1].PlayerID = p.ID
			bundle.Sessions[code+fmt.Sprint(i)] = session{code, p.ID}
		}
		r.Spaces = r.spacesForRules()
		bundle.Rooms[code] = r
		return r
	}
	index := 0
	for _, board := range []string{"ee", "tuscany"} {
		for _, visitors := range []string{"ee", "ee_moor", "rhine"} {
			for _, structures := range []bool{false, true} {
				for _, workers := range []bool{false, true} {
					r := newFixture(fmt.Sprintf("CFG%02d", index), ExpansionConfig{board, structures, workers, visitors})
					r.Phase = "lobby"
					index++
				}
			}
		}
	}
	tuscany := ExpansionConfig{Board: "tuscany", Visitors: "ee"}
	// Isolated development starting states, never a production configuration bypass.
	for _, season := range []string{"spring", "summer", "fall", "winter"} {
		r := newFixture("MAIN"+season, tuscany)
		r.Phase = season
		for _, p := range r.Players {
			p.Season = season
			p.Wines = []Wine{{ID: p.ID + "-wine", Type: "red", Value: 2}}
		}
	}
	setHand := func(r *Room, ids ...string) {
		for _, p := range r.Players {
			out := []Card{}
			for _, c := range p.Hand {
				keep := true
				for _, id := range ids {
					if c.ID == id {
						keep = false
					}
				}
				if keep {
					out = append(out, c)
				}
			}
			p.Hand = out
		}
		for typ, deck := range r.Decks {
			out := []Card{}
			for _, c := range deck {
				keep := true
				for _, id := range ids {
					if c.ID == id {
						keep = false
					}
				}
				if keep {
					out = append(out, c)
				}
			}
			r.Decks[typ] = out
		}
		for _, id := range ids {
			r.Players[0].Hand = append(r.Players[0].Hand, visitorCardT(t, id))
		}
	}
	r := newFixture("MAINmanager", tuscany)
	r.Phase = "winter"
	for _, p := range r.Players {
		p.Season = "winter"
	}
	setHand(r, "winter-32")
	mustApplyT(t, r, r.TurnID, Action{Type: "place", Space: "winter_visitor", Slot: 1, CardID: "winter-32", DeclineBonus: true})
	r = newFixture("MAINplanner", tuscany)
	setHand(r, "summer-29", "winter-34")
	r.Players[0].Coins, r.Players[0].Income = 0, 0
	mustApplyT(t, r, r.TurnID, Action{Type: "place", Space: "summer_visitor", Slot: 2, CardID: "summer-29", DeclineBonus: true})
	choiceT(t, r, Action{Option: "plan", Space: "winter_visitor", Slot: 1})
	for i := 0; i < 20 && r.Phase != "winter"; i++ {
		if len(r.Choices) > 0 {
			choiceT(t, r, Action{Option: r.Choices[0].Options[0]})
		} else {
			mustApplyT(t, r, r.TurnID, Action{Type: "pass"})
		}
	}
	if len(r.Choices) == 0 || r.Choices[0].Kind != "planner" {
		t.Fatal("planner UI fixture did not reach actual reservation")
	}
	r = newFixture("TSTAR", tuscany)
	r.Phase = "spring"
	mustApplyT(t, r, r.TurnID, Action{Type: "place", Space: "influence", Slot: 2, Influence: []InfluenceMove{{To: "siena"}}})
	if r.Choices[0].Kind != "tuscany_influence" {
		t.Fatal("expected actual influence bonus")
	}
	r = newFixture("TTRADE", tuscany)
	mustApplyT(t, r, r.TurnID, Action{Type: "place", Space: "trade", Slot: 2, Trades: []Trade{{Give: "coins", Receive: "vp"}}})
	if r.Choices[0].Kind != "tuscany_trade" {
		t.Fatal("expected actual second trade")
	}
	r = newFixture("TDRAW", tuscany)
	r.Phase, r.Players[0].Wake = "spring", 5
	mustApplyT(t, r, r.TurnID, Action{Type: "pass"})
	if r.Choices[0].Kind != "tuscany_draw" {
		t.Fatal("expected actual season reward")
	}
	r = newFixture("TUPKEEP", tuscany)
	r.Phase = "winter"
	mustApplyT(t, r, r.TurnID, Action{Type: "pass"})
	if r.Choices[0].Kind != "tuscany_upkeep" {
		t.Fatal("expected actual winter cleanup")
	}
	r = newFixture("EEQUEEN", ExpansionConfig{})
	r.Phase = "winter"
	// The Queen is a genuine EE card. Keep it out of all other hands/decks.
	for typ, deck := range r.Decks {
		out := []Card{}
		for _, c := range deck {
			if c.ID != "winter-11" {
				out = append(out, c)
			}
		}
		r.Decks[typ] = out
	}
	for _, p := range r.Players {
		out := []Card{}
		for _, c := range p.Hand {
			if c.ID != "winter-11" {
				out = append(out, c)
			}
		}
		p.Hand = out
	}
	r.Players[0].Hand = append(r.Players[0].Hand, Card{ID: "winter-11", Type: "winter", Implemented: true})
	mustApplyT(t, r, r.TurnID, Action{Type: "place", Space: "winter_visitor", Slot: 2, CardID: "winter-11"})
	if r.Choices[0].Visitor.Stage != "effect" {
		t.Fatal("expected actual visitor effect")
	}
	if filename := os.Getenv("VITICULTURE_HELP_FIXTURE"); filename != "" {
		data, err := json.Marshal(bundle)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
