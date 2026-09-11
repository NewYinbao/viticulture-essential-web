package game

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestTuscanyAuditPlannerCoinBeforeRestoration(t *testing.T) {
	r := tuscanyRoom(2)
	p, q := r.Players[0], r.Players[1]
	r.Phase = "summer"
	p.Season, q.Season = "summer", "summer"
	p.Hand = []Card{visitorCardT(t, "summer-29"), visitorCardT(t, "winter-34")}
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "summer_visitor", DeclineBonus: true, CardID: "summer-29"})
	choiceT(t, r, Action{Option: "plan", Space: "winter_visitor", Slot: 1})
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	mustApplyT(t, r, p.ID, Action{Type: "pass"})
	mustApplyT(t, r, q.ID, Action{Type: "pass"})
	if r.Phase != "winter" || len(r.Choices) != 1 || r.Choices[0].Kind != "planner" {
		t.Fatal("executable coin-first reservation was discarded", r.Phase, r.Planned, r.Choices)
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	c := r.Choices[0]
	for _, a := range []Action{
		{Type: "choose", ChoiceID: c.ID, Option: "execute", CardID: "winter-34"},
		{Type: "choose", ChoiceID: c.ID, Option: "execute", CardID: "winter-34", BonusFirst: true, DeclineBonus: true},
	} {
		rejectUnchangedT(t, r, p.ID, a)
	}
	choiceT(t, r, Action{Option: "execute", CardID: "winter-34", BonusFirst: true})
	if p.Coins != 1 || len(r.Planned) != 0 {
		t.Fatal("coin or reservation count", p.Coins, r.Planned)
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	choiceT(t, r, Action{Option: "buy"})
	if p.Coins != 0 || p.Income != 1 || p.Workers != 0 || !p.LargeWorker || r.TurnID != p.ID || r.Context != nil {
		t.Fatal("planner continuation lost, extra worker spent or bonus repeated", p, r.Context)
	}
	// A plain physical slot must not borrow a bonus from another slot.
	r = tuscanyRoom(5)
	r.Phase = "winter"
	p = r.Players[0]
	p.Hand = []Card{visitorCardT(t, "winter-34")}
	if r.plannerCanExecute(PlannedPlacement{PlayerID: p.ID, Space: "winter_visitor", Slot: 3}) {
		t.Fatal("plain slot received another slots lira")
	}
}

func TestTuscanyAuditOrganizerNoLegacyColorAndManagerResume(t *testing.T) {
	for _, winter := range []bool{false, true} {
		t.Run(fmt.Sprint(winter), func(t *testing.T) {
			r := tuscanyRoom(3)
			p := r.Players[0]
			r.Phase = "summer"
			p.Season = "summer"
			space, card := "summer_visitor", "summer-33"
			p.Hand = []Card{visitorCardT(t, "summer-33")}
			if winter {
				r.Phase = "winter"
				p.Season = "winter"
				space, card = "winter_visitor", "winter-32"
				p.Hand = append(p.Hand, visitorCardT(t, card))
			}
			r.Decks["summer"] = []Card{{ID: "summer-reward", Type: "summer"}}
			r.Decks["winter"] = []Card{{ID: "fall-reward", Type: "winter"}}
			mustApplyT(t, r, p.ID, Action{Type: "place", Space: space, Slot: 2, CardID: card})
			if winter {
				choiceT(t, r, Action{Option: "action", Space: "summer_visitor", CardID: "summer-33"})
			}
			r = restoreT(t, r)
			p = r.Player(p.ID)
			choiceT(t, r, Action{Option: "move", Slot: 5}) // No obsolete EE Color field.
			if p.Passed {
				t.Fatal("passed before reward and second outer visitor")
			}
			r = restoreT(t, r)
			p = r.Player(p.ID)
			if winter {
				c := r.Choices[0]
				rejectUnchangedT(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "place", Influence: []InfluenceMove{{To: "not-a-region"}}})
				choiceT(t, r, Action{Option: "place", Influence: []InfluenceMove{{To: "siena"}}})
			} else {
				choiceT(t, r, Action{Option: "summer"})
			}
			if len(r.Choices) != 1 || r.Choices[0].Visitor == nil || r.Choices[0].Visitor.Stage != "second" {
				t.Fatal("outer visitor continuation missing", r.Choices)
			}
			r = restoreT(t, r)
			p = r.Player(p.ID)
			choiceT(t, r, Action{Option: "skip"})
			if winter {
				if p.Coins != 1 || p.Influence["siena"] != 1 || p.Season != "year_end" || r.Choices[0].Kind != "tuscany_upkeep" {
					t.Fatal("winter Organizer did not pass exactly once", p)
				}
				r = restoreT(t, r)
				p = r.Player(p.ID)
				choiceT(t, r, Action{})
				choiceT(t, r, Action{Option: "6"})
				for _, sp := range r.Spaces {
					for _, seat := range sp.Occupied {
						if seat.PlayerID == p.ID {
							t.Fatal("worker not returned")
						}
					}
				}
			} else if p.Season != "fall" || len(p.Hand) != 2 {
				t.Fatal("summer/fall reward sequence", p)
			}
			if !p.Passed || r.TurnID == p.ID || r.Context != nil || r.Resume != "" {
				t.Fatal("pass resume stuck")
			}
		})
	}
}

// Start from a development-only initial wake fixture, then use only Apply and
// JSON restore until finished. No mid-game resource, score, turn or season edits.
func TestTuscanyAuditNaturalActionGame(t *testing.T) {
	for _, n := range []int{2, 3, 4, 5, 6} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			r := tuscanyRoom(n)
			r.tuscanySpring()
			// Finite EE decks; no expansion initialization or production gate bypass.
			for _, c := range Catalog() {
				if r.cardColor(c.Type) {
					r.Decks[c.Type] = append(r.Decks[c.Type], c)
				}
			}
			done := map[string]bool{}
			steps := 0
			for ; steps < 10000 && r.Phase != "finished"; steps++ {
				p := r.Player(r.TurnID)
				if p == nil {
					t.Fatal("missing actor", r.Phase)
				}
				if len(r.Choices) > 0 {
					c := r.Choices[0]
					a := Action{Type: "choose", ChoiceID: c.ID}
					switch c.Kind {
					case "tuscany_draw":
						a.Option = c.Options[0]
					case "tuscany_influence":
						a.Option = "skip"
					case "tuscany_trade":
						a.Option = "skip"
					case "tuscany_upkeep", "tuscany_next_wake":
						a.Option = c.Options[0]
					case "discard":
						for i := 0; i < c.Count; i++ {
							a.CardIDs = append(a.CardIDs, p.Hand[i].ID)
						}
					default:
						t.Fatal("unexpected choice", c.Kind)
					}
					mustApplyT(t, r, p.ID, a)
				} else if r.Phase == "wake" {
					slot := 2
					for r.WakeSlots[slot-1].PlayerID != "" {
						slot++
					}
					mustApplyT(t, r, p.ID, Action{Type: "wake", Slot: slot})
				} else {
					key := fmt.Sprintf("%d/%s/%s", r.Year, r.Phase, p.ID)
					if done[key] || r.Phase == "winter" {
						mustApplyT(t, r, p.ID, Action{Type: "pass"})
					} else {
						a := Action{Type: "place", Space: "gain_coin"}
						switch r.Phase {
						case "spring":
							a.Space = "tour"
						case "summer":
							a.Space = "trade"
							a.Trades = []Trade{{Give: "coins", Receive: "vp"}}
						case "fall":
							a.Space = "build_tour"
							a.Mode = "tour"
						}
						// Ordinary workers first; reserve the grande for contested summer trade.
						err := r.Apply(p.ID, a)
						if err != nil && p.LargeWorker && (r.Phase == "summer" || p.Workers == 0) {
							a.Large = true
							err = r.Apply(p.ID, a)
						}
						if err != nil {
							mustApplyT(t, r, p.ID, Action{Type: "place", Space: "gain_coin", Large: p.Workers == 0})
						}
						done[key] = true
					}
				}
				r = restoreT(t, r)
			}
			if r.Phase != "finished" || !r.FinalYear || len(r.WinnerIDs) == 0 {
				t.Fatal("natural game did not finish", r.Year, steps)
			}
			for _, p := range r.Players {
				if !p.Passed || p.Season != "ready" {
					t.Fatal("finished before all winter passes", p)
				}
			}
			b, _ := json.Marshal(r)
			rejectUnchangedT(t, r, r.Players[0].ID, Action{Type: "pass"})
			after, _ := json.Marshal(r)
			if string(b) != string(after) {
				t.Fatal("finished snapshot mutated")
			}
			t.Logf("%d players: finished year %d via %d Apply/restore steps", n, r.Year, steps)
		})
	}
}

func TestTuscanyAuditPersonalSeasonRewards(t *testing.T) {
	for _, entering := range []string{"summer", "fall", "winter"} {
		for row := 1; row <= 7; row++ {
			t.Run(fmt.Sprintf("%s/%d", entering, row), func(t *testing.T) {
				r := tuscanyRoom(2)
				p, q := r.Players[0], r.Players[1]
				r.Phase = tuscanySeasons[seasonIndex(entering)-1]
				p.Season, q.Season = r.Phase, r.Phase
				p.Wake = row
				q.Wake = 8 - row
				if q.Wake == row {
					q.Wake = 1
				}
				p.Grapes = []Grape{{ID: "g", Color: "red", Value: 2}}
				p.Wines = []Wine{{ID: "w", Type: "red", Value: 2}}
				if entering == "fall" {
					p.Buildings = append(p.Buildings, "cottage")
				}
				for _, typ := range []string{"vine", "order", "summer", "winter"} {
					r.Decks[typ] = []Card{{ID: typ + "1", Type: typ}, {ID: typ + "2", Type: typ}}
				}
				phase := r.Phase
				mustApplyT(t, r, p.ID, Action{Type: "pass"})
				for len(r.Choices) > 0 {
					r = restoreT(t, r)
					p = r.Player(p.ID)
					c := r.Choices[0]
					rejectUnchangedT(t, r, q.ID, Action{Type: "choose", ChoiceID: c.ID, Option: c.Options[0]})
					a := Action{Option: c.Options[0]}
					if c.Kind == "tuscany_influence" {
						a.Influence = []InfluenceMove{{To: "siena"}}
					}
					choiceT(t, r, a)
				}
				r = restoreT(t, r)
				p = r.Player(p.ID)
				if r.Phase != phase || r.TurnID != q.ID || !p.Passed || p.Season != entering {
					t.Fatal("next season opened before everyone passed", r.Phase, p)
				}
				wantCards, wantCoins, wantVP, wantWorkers, wantGrape := 0, 0, 0, 2, 2
				switch entering {
				case "summer":
					switch row {
					case 2:
						wantCoins = 1
					case 3, 4, 5:
						wantCards = 1
					case 6:
						wantVP = 1
					case 7:
						wantWorkers = 3
					}
				case "fall":
					wantCards = 1
					if row >= 3 {
						wantCards++
					}
				case "winter":
					switch row {
					case 5:
						wantCoins = 1
						if p.Influence["siena"] != 1 {
							t.Fatal("missing winter star")
						}
					case 6:
						wantGrape = 3
					case 7:
						if r.FirstPlayerID != p.ID {
							t.Fatal("missing next-year first marker")
						}
					}
				}
				if len(p.Hand) != wantCards || p.Coins != wantCoins || p.VP != wantVP || p.Workers != wantWorkers || p.Grapes[0].Value != wantGrape || p.Wines[0].Value != 2 {
					t.Fatalf("reward mismatch for %s/%d: %+v", entering, row, p)
				}
				rejectUnchangedT(t, r, p.ID, Action{Type: "pass"})
			})
		}
	}
}

func TestTuscanyAuditFinalWinterWaitAndMajorities(t *testing.T) {
	for _, finalVP := range []int{24, 25, 41} {
		t.Run(fmt.Sprint(finalVP), func(t *testing.T) {
			r := tuscanyRoom(3)
			r.Phase = "winter"
			p, q, z := r.Players[0], r.Players[1], r.Players[2]
			p.VP = finalVP
			q.VP = 23
			p.Influence = map[string]int{"pisa": 2, "siena": 1}
			q.Influence = map[string]int{"pisa": 2, "arezzo": 1}
			z.Influence = map[string]int{"pisa": 1}
			for _, a := range r.Players {
				a.Season = "winter"
				a.Income = 2
			}
			// The leading player passes first; influence cannot score until the others pass.
			for i := 0; i < 3; i++ {
				actor := r.Player(r.TurnID)
				mustApplyT(t, r, actor.ID, Action{Type: "pass"})
				r = restoreT(t, r)
				choiceT(t, r, Action{})
				choiceT(t, r, Action{})
				if i < 2 && (r.Phase != "winter" || r.Player(p.ID).VP != finalVP || r.Player(q.ID).VP != 23) {
					t.Fatal("scored before all winter passes")
				}
			}
			p, q, z = r.Player(p.ID), r.Player(q.ID), r.Player(z.ID)
			if finalVP < 25 {
				if r.Year != 2 || r.Phase != "spring" || p.VP != 24 || q.VP != 23 {
					t.Fatal("map majority prematurely triggered end")
				}
			} else if r.Phase != "finished" || p.VP != finalVP+2 || q.VP != 25 || z.VP != 0 {
				t.Fatal("25 threshold, no score cap, majority or tied region", p.VP, q.VP, z.VP, r.Phase)
			}
			for _, a := range r.Players {
				if a.Coins != 2 {
					t.Fatal("upkeep duplicated", a.Coins)
				}
			}
		})
	}
}

func TestTuscanyAuditTradeInvalidReceiptRollback(t *testing.T) {
	r := tuscanyRoom(3)
	r.Phase = "summer"
	p := r.Players[0]
	p.Coins = 3
	r.Decks["winter"] = []Card{{ID: "hidden-a", Type: "winter"}, {ID: "hidden-b", Type: "winter"}}
	rejectUnchangedT(t, r, p.ID, Action{Type: "place", Space: "trade", Slot: 2, Trades: []Trade{{Give: "coins", Receive: "cards", Colors: []string{"winter", "orange"}}}})
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: "trade", Slot: 2, Trades: []Trade{{Give: "coins", Receive: "cards", Colors: []string{"winter", "winter"}}}})
	r = restoreT(t, r)
	p = r.Player(p.ID)
	c := r.Choices[0]
	rejectUnchangedT(t, r, p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "trade", Trades: []Trade{{Give: "cards", CardIDs: []string{p.Hand[0].ID, p.Hand[1].ID}, Receive: "grape", Color: "pink"}}})
	choiceT(t, r, Action{Option: "trade", Trades: []Trade{{Give: "cards", CardIDs: []string{p.Hand[0].ID, p.Hand[1].ID}, Receive: "grape", Color: "white"}}})
	if p.Coins != 0 || len(p.Hand) != 0 || len(p.Grapes) != 1 || p.Grapes[0].Value != 1 || p.Grapes[0].Color != "white" || len(r.Discards["winter"]) != 2 {
		t.Fatal("trade resource conservation")
	}
}
