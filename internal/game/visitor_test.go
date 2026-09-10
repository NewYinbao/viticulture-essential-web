package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func visitorFixture(id string) (*Room, *Player) {
	r := eeRoom(3)
	r.InitDecks()
	r.Phase = strings.Split(id, "-")[0]
	r.Year = 1
	r.Spaces = NewSpaces(3)
	for i, p := range r.Players {
		p.Wake = i + 1
		p.Coins = 50
		p.VP = 10
		p.Buildings = []string{"trellis", "irrigation"}
		p.Hand = []Card{{ID: "plant", Name: "C", Type: "vine", White: 4}, {ID: "order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
		for k := 0; k < 5; k++ {
			p.Hand = append(p.Hand, Card{ID: fmt.Sprint("s", k), Type: "summer"})
		}
		p.Grapes = []Grape{{ID: "g0", Color: "red", Value: 4}, {ID: "g1", Color: "white", Value: 3}, {ID: "g2", Color: "red", Value: 2}}
		p.Wines = []Wine{{ID: "w9", Type: "red", Value: 9}, {ID: "w7", Type: "white", Value: 7}, {ID: "w1", Type: "red", Value: 1}}
		for f := range p.Fields {
			p.Fields[f].Vines = []Card{{ID: fmt.Sprint("v", f), Name: "A", Type: "vine", Red: 1}, {ID: fmt.Sprint("u", f), Name: "B", Type: "vine", White: 1}}
		}
	}
	p := r.Players[0]
	r.TurnID = p.ID
	for _, c := range Catalog() {
		if c.ID == id {
			p.Hand = append(p.Hand, c)
		}
	}
	r.Discards["winter"] = []Card{{ID: "old-winter", Type: "winter"}}
	r.Discards["summer"] = []Card{{ID: "old-summer", Type: "summer"}}
	return r, p
}
func visitorAction(r *Room, c Choice) Action {
	a := Action{Type: "choose", ChoiceID: c.ID, Option: c.Options[0], Building: "yoke", Buildings: []string{"yoke", "cottage"}, CardIDs: []string{"plant"}, Fields: []int{2}, Field: 0, Recipes: [][]int{{0}}, WineIDs: []string{"w9"}, Grapes: []int{0}, Color: "red", Colors: []string{"summer", "winter"}, TargetIDs: []string{r.Players[1].ID}, CardID: "order", Space: "tour"}
	return a
}
func prepareVisitorAction(r *Room, c Choice, option string) Action {
	a := visitorAction(r, c)
	a.Option = option
	id := c.Visitor.CardID
	p := r.Player(c.PlayerID)
	switch id {
	case "summer-29":
		a.Space = "draw_order"
	case "summer-33":
		a.Slot = 4
	case "summer-09", "summer-19", "summer-24":
		if option == "uproot" {
			a.CardIDs = []string{"v0"}
			a.Fields = []int{0}
			if id == "summer-19" {
				a.CardIDs = append(a.CardIDs, "v1")
				a.Fields = append(a.Fields, 1)
			}
		}
	case "summer-11":
		if option == "swap" {
			a.CardIDs = []string{"v0", "v1"}
			a.Fields = []int{0, 1}
		}
	case "summer-15":
		a.CardIDs = []string{"s0", "s1"}
		if option == "vp" {
			a.CardIDs = append(a.CardIDs, "s2", "s3")
		}
	case "summer-16":
		a.CardIDs = []string{"s0", "s1", "s2"}
	case "summer-20", "winter-09":
		a.CardIDs = []string{"s0", "s1"}
	case "summer-23":
		a.CardIDs = []string{"s0"}
	case "summer-32":
		a.TargetIDs = []string{"tour"}
		a.Fields = []int{2}
	case "winter-14":
		a.WineIDs = []string{"w1"}
	case "winter-35":
		a.CardIDs = []string{"s0"}
	case "winter-37":
		if c.Visitor.Stage == "take" {
			a.CardIDs = []string{r.Revealed[0].ID, r.Revealed[1].ID}
		}
	}
	if option == "fill" || strings.HasPrefix(option, "fill_") {
		a.WineIDs = []string{"w9"}
	}
	if id == "winter-14" {
		a.WineIDs = []string{"w1"}
	}
	if id == "winter-13" && option == "vp" {
		p.TotalWorkers = 6
	}
	if id == "winter-34" && option == "sell" {
		p.Income = 2
	}
	return a
}
func playVisitorFixture(t *testing.T, r *Room, p *Player, id string) {
	t.Helper()
	if e := r.Apply(p.ID, Action{Type: "place", Space: strings.Split(id, "-")[0] + "_visitor", CardID: id, Slot: 2}); e != nil {
		t.Fatal(e)
	}
}

// Every catalogued card has an explicit acceptance or rejection test. These are
// branch execution tests, not a claim that every numerical boundary is covered.
func TestVisitorEveryCardEveryInitialOption(t *testing.T) {
	unsupported := map[string]bool{}
	count := 0
	for _, card := range Catalog() {
		if card.Type != "summer" && card.Type != "winter" {
			continue
		}
		count++
		t.Run(card.ID, func(t *testing.T) {
			d, ok := visitorDefs[card.ID]
			if unsupported[card.ID] {
				r, p := visitorFixture(card.ID)
				before, _ := json.Marshal(r)
				e := r.Apply(p.ID, Action{Type: "place", Space: card.Type + "_visitor", CardID: card.ID})
				after, _ := json.Marshal(r)
				if e == nil || !strings.Contains(e.Error(), "NOT_IMPLEMENTED") || !bytes.Equal(before, after) || card.Implemented {
					t.Fatal("unsupported must reject atomically", e)
				}
				return
			}
			if !ok || !card.Implemented {
				t.Fatal("missing card implementation")
			}
			for _, option := range d.Options {
				t.Run(option, func(t *testing.T) {
					r, p := visitorFixture(card.ID)
					if card.ID == "summer-32" {
						r.Spaces[1].Occupied = []Seat{{PlayerID: p.ID, Slot: 2}}
					}
					if card.ID == "winter-25" {
						for i, q := range r.Players {
							q.LargeWorker = false
							r.Spaces[1].Occupied = append(r.Spaces[1].Occupied, Seat{PlayerID: q.ID, Slot: i + 1, Large: true})
						}
					}
					playVisitorFixture(t, r, p, card.ID)
					before, _ := json.Marshal(r)
					c0 := r.Choices[0]
					invalid := Action{Type: "choose", ChoiceID: c0.ID, Option: "forged"}
					if e := r.Apply(p.ID, invalid); e == nil {
						t.Fatal("forged option")
					}
					invalid.Option = c0.Options[0]
					if e := r.Apply(r.Players[1].ID, invalid); e == nil {
						t.Fatal("foreign actor")
					}
					after, _ := json.Marshal(r)
					if !bytes.Equal(before, after) {
						t.Fatal("per-card invalid command rollback")
					}
					for step := 0; len(r.Choices) > 0 && step < 20; step++ {
						c := r.Choices[0]
						opt := c.Options[0]
						if step == 0 {
							opt = option
						}
						a := prepareVisitorAction(r, c, opt)
						if e := r.Apply(c.PlayerID, a); e != nil {
							t.Fatalf("%s %s: %v", c.Visitor.Stage, opt, e)
						}
					}
					if len(r.Choices) != 0 || r.Context != nil || r.TurnID != r.Players[1].ID {
						t.Fatal("continuation did not restore actor turn order")
					}
				})
			}
		})
	}
	if count != 76 {
		t.Fatal(count)
	}
}

func TestVisitorInnkeeperDoesNotCoverOldTopOrRetrieveSelf(t *testing.T) {
	r, p := visitorFixture("winter-19")
	playVisitorFixture(t, r, p, "winter-19")
	if r.Discards["winter"][0].ID != "old-winter" || len(r.Discards["winter"]) != 1 {
		t.Fatal("current card covered old top")
	}
	c := r.Choices[0]
	a := prepareVisitorAction(r, c, "take")
	if e := r.Apply(p.ID, a); e != nil {
		t.Fatal(e)
	}
	if cardIndex(p, "winter-19", "winter") >= 0 || cardIndex(p, "old-winter", "winter") < 0 || len(r.Discards["winter"]) != 1 || r.Discards["winter"][0].ID != "winter-19" {
		t.Fatal("Innkeeper timing")
	}
}

func TestVisitorFailedMultiCostRollsBack(t *testing.T) {
	r, p := visitorFixture("summer-16")
	playVisitorFixture(t, r, p, "summer-16")
	c := r.Choices[0]
	a := prepareVisitorAction(r, c, "discard")
	a.CardIDs = []string{"s0", "s1", "forged"}
	before, _ := json.Marshal(r)
	if e := r.Apply(p.ID, a); e == nil {
		t.Fatal("forgery accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("wine/card/choice rollback")
	}
}

func roundTripVisitor(t *testing.T, r *Room) *Room { return persistedRoom(t, r) }
func chooseVisitor(t *testing.T, r *Room, option string, edit func(*Action)) {
	t.Helper()
	c := r.Choices[0]
	a := prepareVisitorAction(r, c, option)
	if edit != nil {
		edit(&a)
	}
	if e := r.Apply(c.PlayerID, a); e != nil {
		t.Fatal(e)
	}
}
func TestVisitorImporterEscrowSuccessAndRefund(t *testing.T) {
	for _, success := range []bool{false, true} {
		t.Run(fmt.Sprint(success), func(t *testing.T) {
			r, p := visitorFixture("summer-23")
			before := len(p.Hand) - 1
			playVisitorFixture(t, r, p, "summer-23")
			chooseVisitor(t, r, "resolve", nil)
			q := r.Players[1]
			n := len(q.Hand)
			chooseVisitor(t, r, "give", func(a *Action) { a.CardIDs = []string{"s0", "s1"} })
			if len(p.Hand) != before || len(q.Hand) != n-2 {
				t.Fatal("escrow committed prematurely")
			}
			// Escrow identities are absent from every view, including the actor's.
			for _, who := range r.Players {
				b, _ := json.Marshal(r.View(who.ID))
				if bytes.Contains(b, []byte("escrow")) {
					t.Fatal("private escrow exposed")
				}
			}
			r = roundTripVisitor(t, r)
			p = r.Players[0]
			q = r.Players[1]
			if success {
				chooseVisitor(t, r, "give", func(a *Action) { a.CardIDs = []string{"s0"} })
			} else {
				chooseVisitor(t, r, "skip", nil)
			}
			chooseVisitor(t, r, "resolve", nil)
			if len(p.Hand) != before+3 {
				t.Fatal("must gain exactly three")
			}
			if success {
				if len(q.Hand) != n-2 {
					t.Fatal("successful escrow refunded")
				}
			} else {
				if len(q.Hand) != n {
					t.Fatal("failed escrow not refunded")
				}
			}
			if r.Context != nil {
				t.Fatal("escrow continuation retained")
			}
		})
	}
}
func TestVisitorMentorRewardBeforeNextOpponentAndPassedState(t *testing.T) {
	r, p := visitorFixture("winter-17")
	r.Players[1].Passed = true
	playVisitorFixture(t, r, p, "winter-17")
	chooseVisitor(t, r, "resolve", nil)
	chooseVisitor(t, r, "skip", nil)
	before := len(p.Hand)
	chooseVisitor(t, r, "make", nil)
	c := r.Choices[0]
	if c.PlayerID != p.ID || c.Visitor.Stage != "reward" || strings.Join(c.Options, ",") != "vine,summer" {
		t.Fatal("wrong Mentor reward/order")
	}
	r = roundTripVisitor(t, r)
	p = r.Players[0]
	chooseVisitor(t, r, "vine", nil)
	if len(p.Hand) != before+1 || !r.Players[1].Passed || r.Choices[0].PlayerID != r.Players[2].ID {
		t.Fatal("reward/passed continuation")
	}
	chooseVisitor(t, r, "skip", nil)
	if r.TurnID != r.Players[2].ID {
		t.Fatal("passed player reactivated")
	}
}
func TestVisitorInnkeeperAndCaravanPendingPersistence(t *testing.T) {
	r, p := visitorFixture("winter-19")
	playVisitorFixture(t, r, p, "winter-19")
	r = roundTripVisitor(t, r)
	p = r.Players[0]
	if len(r.Context.Held) != 1 || r.Context.Held[0].ID != "winter-19" {
		t.Fatal("lost resolving zone")
	}
	chooseVisitor(t, r, "take", nil)
	if cardIndex(p, "winter-19", "winter") >= 0 {
		t.Fatal("self retrieved")
	}
	r, p = visitorFixture("winter-37")
	playVisitorFixture(t, r, p, "winter-37")
	chooseVisitor(t, r, "reveal", nil)
	if len(r.Revealed) != 4 {
		t.Fatal("Caravan four colors")
	}
	for _, q := range r.Players {
		v := r.View(q.ID)
		if len(v["revealed"].([]Card)) != 4 {
			t.Fatal("Caravan must be public")
		}
	}
	r = roundTripVisitor(t, r)
	chooseVisitor(t, r, "take", nil)
	if len(r.Revealed) != 0 {
		t.Fatal("revealed not cleared")
	}
}
func TestVisitorConfirmedAuditNumbers(t *testing.T) {
	cases := []struct {
		id, opt           string
		coins, vp, income int
	}{
		{"summer-02", "buy", 41, 13, 0}, {"summer-02", "sell", 56, 8, 0},
		{"summer-08", "buy", 44, 12, 0}, {"summer-08", "sell", 59, 7, 0},
		{"summer-26", "grape", 50, 10, 1}, {"summer-26", "wine", 50, 10, 2},
		{"winter-16", "grape", 50, 11, 0}, {"winter-16", "wine", 50, 11, 0},
		{"winter-18", "harvest", 50, 10, 0}, {"winter-18", "build", 49, 10, 0},
		{"winter-31", "train", 47, 10, 0}, {"winter-34", "buy", 49, 10, 1},
	}
	for _, tt := range cases {
		t.Run(tt.id+tt.opt, func(t *testing.T) {
			r, p := visitorFixture(tt.id)
			before := len(p.Hand) - 1
			playVisitorFixture(t, r, p, tt.id)
			chooseVisitor(t, r, tt.opt, nil)
			if p.Coins != tt.coins || p.VP != tt.vp || p.Income != tt.income {
				t.Fatalf("coins/vp/income: %d/%d/%d", p.Coins, p.VP, p.Income)
			}
			if tt.id == "winter-18" {
				want := before
				if tt.opt == "harvest" {
					want++
				}
				if tt.opt == "harvest" && !p.Fields[0].Harvested {
					t.Fatal("Harvest Expert failed to harvest")
				}
				if len(p.Hand) != want {
					t.Fatal("Harvest Expert must harvest AND draw; build must not draw")
				}
			}
			if tt.id == "winter-31" && p.Workers != 2 {
				t.Fatal("Governess worker must be usable immediately after placing one")
			}
		})
	}
}
func TestVisitorNoPaymentBelowBounds(t *testing.T) {
	for _, tc := range []struct {
		id, opt   string
		coins, vp int
	}{{"summer-02", "buy", 8, 0}, {"summer-02", "sell", 0, -4}, {"summer-08", "sell", 0, -3}, {"winter-07", "train", 0, -5}, {"winter-34", "buy", 0, 0}} {
		t.Run(tc.id+tc.opt, func(t *testing.T) {
			r, p := visitorFixture(tc.id)
			playVisitorFixture(t, r, p, tc.id)
			p.Coins = tc.coins
			p.VP = tc.vp
			before, _ := json.Marshal(r)
			a := prepareVisitorAction(r, r.Choices[0], tc.opt)
			if e := r.Apply(p.ID, a); e == nil {
				t.Fatal("unaffordable payment")
			}
			after, _ := json.Marshal(r)
			if !bytes.Equal(before, after) {
				t.Fatal("payment rollback")
			}
		})
	}
}
