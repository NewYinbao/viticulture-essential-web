package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func rhineFixture(t *testing.T, id string, prepare func(*Room, *Player)) (*Room, *Player) {
	t.Helper()
	r, p := visitorFixture("summer-01")
	if strings.Contains(id, "winter") {
		r.Phase = "winter"
	}
	p.Income = 3
	p.Wines = nil
	p.Grapes = nil
	for fi := range p.Fields {
		for vi := range p.Fields[fi].Vines {
			p.Fields[fi].Vines[vi].ID = "field-" + p.Fields[fi].Vines[vi].ID
		}
	}
	p.Hand = []Card{{ID: "v0", Type: "vine", Red: 1}, {ID: "v1", Type: "vine", White: 1}, {ID: "o0", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
	for _, fact := range rhineCardFacts {
		if fact.Card.ID == id {
			p.Hand = append(p.Hand, fact.Card)
		}
	}
	if prepare != nil {
		prepare(r, p)
	}
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Space: r.Phase + "_visitor"}
	if err := r.queueVisitor(p, id); err != nil {
		t.Fatalf("queue %s: %v", id, err)
	}
	return r, p
}
func rhineChoose(t *testing.T, r *Room, option string, a Action) {
	t.Helper()
	if len(r.Choices) == 0 {
		t.Fatal("missing pending choice")
	}
	c := r.Choices[0]
	a.Type = "choose"
	a.ChoiceID = c.ID
	a.Option = option
	if err := r.Apply(c.PlayerID, a); err != nil {
		t.Fatalf("%s %s: %v", c.Visitor.CardID, option, err)
	}
}

func TestRhinePrintedIndependentBranches(t *testing.T) {
	tests := []struct {
		id, option                             string
		a                                      Action
		coins, vp, income, hand, grapes, wines int
	}{
		{"summer-agent", "discard", Action{CardIDs: []string{"v0", "v1"}}, 5, 0, 0, -2, 0, 0},
		{"summer-botanist", "grape", Action{Color: "red"}, 0, 0, 0, 0, 1, 0},
		{"summer-cicerone", "coins", Action{}, 4, 0, 0, 0, 0, 0},
		{"summer-docent", "coins", Action{}, 3, 0, 0, 0, 0, 0},
		{"summer-embezzler", "vp_coins", Action{}, 6, -2, 0, 0, 0, 0},
		{"summer-fortuneTeller", "draw", Action{}, 0, 0, 0, 3, 0, 0},
		{"summer-freelancer", "draw", Action{Colors: []string{"winter"}}, 0, 0, 0, 1, 0, 0},
		{"summer-friendlyHelper", "wine", Action{Color: "white"}, 0, 0, 0, 0, 0, 1},
		{"summer-grapeBuyer", "large", Action{}, -5, 0, 0, 0, 2, 0},
		{"summer-philanthropist", "draw", Action{}, 0, 0, 0, 2, 0, 0},
		{"summer-premiumWineDealer", "sparkling", Action{}, -9, 0, 0, 0, 0, 1},
		{"summer-sculptor", "coins", Action{}, 3, 0, 0, 0, 0, 0},
		{"summer-sommelier", "build", Action{}, -2, 0, 0, 0, 0, 0},
		{"summer-subsidizer", "build", Action{Building: "cottage"}, -2, 0, 0, 0, 0, 0},
		{"summer-supporter", "vine_coins", Action{}, 2, 0, 0, 1, 0, 0},
		{"summer-traveller", "coins", Action{}, 3, 0, 0, 0, 0, 0},
		{"summer-wineTrader", "red", Action{}, -6, 0, 0, 0, 0, 1},
		{"summer-writer", "grape", Action{Color: "white"}, 0, 0, -2, 0, 1, 0},
		{"winter-advertiser", "draw", Action{}, 0, 0, -2, 3, 0, 0},
		{"winter-bargainer", "draw", Action{}, 0, 0, 0, 3, 0, 0},
		{"winter-brideToBe", "coins", Action{}, 3, 0, 0, 0, 0, 0},
		{"winter-bureaucrat", "resolve", Action{}, -1, 0, 0, 3, 0, 0},
		{"winter-cellarman", "grapes", Action{}, -4, 0, 0, 0, 2, 0},
		{"winter-duchess", "income", Action{}, -1, 0, 1, 0, 0, 0},
		{"winter-eliteOenologist", "upgrade", Action{}, 0, 0, 0, 0, 0, 0},
		{"winter-enthusiast", "draw", Action{}, 0, 0, 0, 2, 0, 0},
		{"winter-estateAgent", "draw", Action{Colors: []string{"summer", "summer", "summer"}}, 0, 0, 0, 3, 0, 0},
		{"winter-grapeWhisperer", "harvest", Action{Fields: []int{0, 1}}, 2, 0, 0, 0, 2, 0},
		{"winter-harvestExpert", "build", Action{}, -1, 0, 0, 0, 0, 0},
		{"winter-lecturer", "train", Action{}, -3, 0, 0, 0, 0, 0},
		{"winter-lovebirds", "draw", Action{}, 1, 0, 0, 1, 0, 0},
		{"winter-researcher", "draw", Action{}, 0, 0, 0, 2, 0, 0},
		{"winter-rhineSailor", "exchange", Action{CardIDs: []string{"v0", "v1", "o0"}}, 1, 0, 0, 0, 0, 0},
		{"winter-schoolTeacher", "train", Action{}, -4, 0, 0, 0, 0, 0},
		{"winter-shipper", "coins", Action{}, 3, 0, 0, 0, 0, 0},
		{"winter-skeptic", "upgrade", Action{}, -1, 0, 0, 0, 0, 0},
		{"winter-uncertifiedOenologist", "upgrade", Action{}, 0, -1, 0, 0, 0, 0},
		{"winter-wineStoreOwner", "buy", Action{Slot: 3, Color: "red"}, -3, 0, 0, 0, 0, 1},
		{"winter-winterAgent", "draw", Action{}, 0, 0, 0, 3, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			r, p := rhineFixture(t, "rhine-"+test.id, nil)
			before := *probePlayer(p)
			rhineChoose(t, r, test.option, test.a)
			actual := []int{p.Coins - before.Coins, p.VP - before.VP, p.Income - before.Income, len(p.Hand) - len(before.Hand), len(p.Grapes) - len(before.Grapes), len(p.Wines) - len(before.Wines)}
			expected := []int{test.coins, test.vp, test.income, test.hand, test.grapes, test.wines}
			for i, n := range expected {
				if actual[i] != n {
					t.Fatalf("result metric %d = %d; want %d (all %v)", i, actual[i], n, actual)
				}
			}
			if len(r.Choices) != 0 {
				t.Fatal("unexpected unfinished choice")
			}
		})
	}
}

func TestRhineCatalogReplacementAndBoardRequirements(t *testing.T) {
	counts := map[string]int{}
	for _, card := range RhineCatalog(ExpansionConfig{Board: "tuscany", Visitors: "rhine"}) {
		counts[card.Type]++
	}
	if counts["summer"] != 40 || counts["winter"] != 40 {
		t.Fatal(counts)
	}
	removed := map[string]bool{"rhine-summer-oldGeneral": true, "rhine-summer-structureReorganizer": true, "rhine-winter-influencer": true, "rhine-winter-lobbyist": true}
	base := RhineCatalog(ExpansionConfig{Board: "ee", Visitors: "rhine"})
	if len(base) != 76 {
		t.Fatalf("base catalog %d", len(base))
	}
	for _, card := range base {
		if removed[card.ID] {
			t.Fatal("Tuscany dependency leaked", card.ID)
		}
	}
	if err := (ExpansionConfig{Board: "ee", Visitors: "rhine"}).Playable(); err != nil {
		t.Fatal(err)
	}
}

func TestRhineDrawnOnlyDiscardPersistsAndCannotReplay(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-plantDealer", nil)
	rhineChoose(t, r, "draw", Action{})
	if len(r.Choices) != 1 || len(p.Hand) != 7 {
		t.Fatal("draw not committed as one continuation")
	}
	snapshot, _ := json.Marshal(r)
	var restored Room
	if err := json.Unmarshal(snapshot, &restored); err != nil {
		t.Fatal(err)
	}
	r = &restored
	p = r.Player(p.ID)
	c := r.Choices[0]
	bad := Action{Type: "choose", ChoiceID: c.ID, Option: "continue", CardIDs: []string{"v0", "v1"}}
	if r.Apply(p.ID, bad) == nil {
		t.Fatal("discarded old hand instead of newly drawn cards")
	}
	after, _ := json.Marshal(r)
	if string(after) != string(snapshot) {
		t.Fatal("invalid selection changed persisted game")
	}
	ids := strings.Split(c.Visitor.Rhine.Values["drawn"], ",")
	rhineChoose(t, r, "continue", Action{CardIDs: ids[:2]})
	if len(p.Hand) != 5 {
		t.Fatal("draw/discard count", len(p.Hand))
	}
	if r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "continue", CardIDs: ids[:2]}) == nil {
		t.Fatal("choice replay accepted")
	}
}

func TestRhineDuchessRequiresTwoDisjointOrders(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-duchess", func(r *Room, p *Player) { p.Wines = []Wine{{ID: "w1", Type: "red", Value: 3}} })
	if reason := r.rhineOptionReason(p, r.Choices[0].Visitor, "fill"); reason == "" {
		t.Fatal("one order advertised as sufficient")
	}
	before, _ := json.Marshal(r)
	c := r.Choices[0]
	if r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "fill", CardID: "o0", WineIDs: []string{"w1"}}) == nil {
		t.Fatal("one order accepted")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("income or wine spent despite impossible second order")
	}
	p.Hand = append(p.Hand, Card{ID: "o1", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}, Points: 3, Income: 1})
	p.Wines = append(p.Wines, Wine{ID: "w2", Type: "red", Value: 2})
	rhineChoose(t, r, "fill", Action{CardID: "o0", WineIDs: []string{"w2"}})
	if p.Income != 2 {
		t.Fatal("first order: income cost/reward", p.Income)
	}
	rhineChoose(t, r, "continue", Action{CardID: "o1", WineIDs: []string{"w1"}})
	if p.VP != 15 || p.Income != 3 {
		t.Fatal("second order result", p.VP, p.Income)
	}
}

func TestRhineBankerEachOpponentOwnsResponse(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-banker", nil)
	rhineChoose(t, r, "resolve", Action{})
	if p.Coins != 55 || len(r.Choices) != 2 {
		t.Fatal("banker setup")
	}
	q := r.Player(r.Choices[0].PlayerID)
	if q.ID == p.ID {
		t.Fatal("actor response included")
	}
	c := r.Choices[0]
	if r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "sell"}) == nil {
		t.Fatal("actor controlled opponent")
	}
	rhineChoose(t, r, "sell", Action{})
	if q.VP != 9 || q.Coins != 53 {
		t.Fatal("opponent exchange")
	}
	rhineChoose(t, r, "skip", Action{})
}

func TestRhineAnyTwoSparklingAndRecipeUniqueness(t *testing.T) {
	for _, colors := range [][]string{{"red", "red"}, {"white", "white"}, {"red", "white"}} {
		r, p := rhineFixture(t, "rhine-winter-brideToBe", func(r *Room, p *Player) {
			p.Buildings = append(p.Buildings, "medium_cellar", "large_cellar")
			p.Grapes = []Grape{{ID: "a", Color: colors[0], Value: 4}, {ID: "b", Color: colors[1], Value: 3}}
		})
		rhineChoose(t, r, "make", Action{Grapes: []int{0, 1}})
		if len(p.Wines) != 1 || p.Wines[0].Type != "sparkling" || p.Wines[0].Value != 7 || len(p.Grapes) != 0 {
			t.Fatal("any two grapes recipe", colors, p.Wines)
		}
	}
	p := &Player{Buildings: []string{"large_cellar"}, Grapes: []Grape{{ID: "a", Color: "red", Value: 5}}}
	if rhineTwoSparkling(p, []int{0, 0}) == nil {
		t.Fatal("same grape used twice")
	}
}

func TestRhineMiddlemanAndPremiumBottleMatching(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-middleman", func(r *Room, p *Player) {
		p.Hand[2].Requirements = []Wine{{Type: "red", Value: 5}, {Type: "white", Value: 2}}
		p.Wines = []Wine{{ID: "b", Type: "blush", Value: 5}, {ID: "s", Type: "sparkling", Value: 7}}
	})
	rhineChoose(t, r, "fill", Action{CardID: "o0", WineIDs: []string{"b", "s"}})
	if len(p.Wines) != 0 || p.VP != 12 {
		t.Fatal("value-only fill")
	}
	discard := r.Discards["order"]
	if discard[len(discard)-1].Requirements[0].Type != "red" {
		t.Fatal("original order altered")
	}
	for _, good := range []bool{false, true} {
		r, p := rhineFixture(t, "rhine-winter-premiumBuyer", func(r *Room, p *Player) {
			p.Hand[2].Requirements = []Wine{{Type: "red", Value: 4}, {Type: "red", Value: 1}}
			p.Wines = []Wine{{ID: "a", Type: "red", Value: 9}, {ID: "b", Type: "red", Value: 2}}
			if good {
				p.Wines[1].Value = 3
			}
		})
		rhineChoose(t, r, "fill", Action{CardID: "o0", WineIDs: []string{"a", "b"}})
		expected := 12
		if good {
			expected = 14
		}
		if p.VP != expected {
			t.Fatalf("individual +2 match: %d want %d", p.VP, expected)
		}
	}
}

func TestRhineMiddlemanDrawFundsMandatoryDiscard(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-middleman", func(r *Room, p *Player) { p.Coins = 0; p.Hand = []Card{p.Hand[len(p.Hand)-1]} })
	rhineChoose(t, r, "draw", Action{})
	if len(p.Hand) != 3 || len(r.Choices) != 1 {
		t.Fatal("expected newly drawn vines and payment choice")
	}
	id := p.Hand[0].ID
	rhineChoose(t, r, "discard", Action{CardIDs: []string{id}})
	if len(p.Hand) != 2 || p.Coins != 0 {
		t.Fatal("discard after draw")
	}
}

func TestRhineWineEngineerEachRepetitionCostsTwo(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-wineEngineer", func(r *Room, p *Player) { p.Coins = 6; p.Grapes = []Grape{{ID: "g", Color: "red", Value: 1}} })
	rhineChoose(t, r, "age", Action{Grapes: []int{0}})
	if p.Coins != 4 || p.Grapes[0].Value != 2 {
		t.Fatal("first repetition")
	}
	rhineChoose(t, r, "again", Action{Grapes: []int{0}})
	if p.Coins != 2 || p.Grapes[0].Value != 3 {
		t.Fatal("second repetition")
	}
	rhineChoose(t, r, "again", Action{Grapes: []int{0}})
	if p.Coins != 0 || p.Grapes[0].Value != 4 || len(r.Choices) != 0 {
		t.Fatal("third repetition cap")
	}
}

func TestRhineSonInLawPermanentFieldAndWinterTiming(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-sonInLaw", nil)
	rhineChoose(t, r, "field", Action{Field: 0})
	if p.RhineSonField == nil || *p.RhineSonField != 0 {
		t.Fatal("missing attachment")
	}
	for _, card := range r.Discards["summer"] {
		if card.ID == "rhine-summer-sonInLaw" {
			t.Fatal("permanent card was discarded")
		}
	}
	r.Config.Visitors = "rhine"
	r.Phase = "winter"
	r.TurnID = p.ID
	p.Fields[0].Vines = []Card{{ID: "big", Type: "vine", Red: 4, White: 3}}
	workers := p.Workers
	r.rhinePrepareWinterTurn()
	if len(r.Choices) != 1 {
		t.Fatal("missing first winter offer")
	}
	snapshot, _ := json.Marshal(r)
	var restored Room
	_ = json.Unmarshal(snapshot, &restored)
	r = &restored
	p = r.Player(p.ID)
	rhineChoose(t, r, "harvest", Action{Grapes: []int{2, 1}})
	if len(p.Grapes) != 2 || p.Grapes[0].Value+p.Grapes[1].Value != 3 || !p.Fields[0].Harvested || p.Workers != workers || r.TurnID != p.ID {
		t.Fatal("harvest changed primary action or exceeded total3")
	}
	r.rhinePrepareWinterTurn()
	if len(r.Choices) != 0 {
		t.Fatal("offered twice in same winter")
	}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "gain_coin"}); err != nil {
		t.Fatal(err)
	}
	if p.RhineWinterActionYear != 1 || p.Workers != workers-1 {
		t.Fatal("first ordinary winter action marker")
	}
	r.Year = 2
	r.TurnID = p.ID
	p.Fields[0].Harvested = false
	r.rhinePrepareWinterTurn()
	if len(r.Choices) != 1 {
		t.Fatal("did not offer next winter")
	}
	rhineChoose(t, r, "skip", Action{})
}

func TestRhineAdministratorMovesExactWorkerWithoutSpendingAnother(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-administrator", nil)
	seat := Seat{PlayerID: p.ID, Slot: 2, WorkerType: "soldato"}
	r.Context.Space = "summer_visitor"
	r.Context.TriggerSeat = &seat
	for i := range r.Spaces {
		if r.Spaces[i].ID == "summer_visitor" {
			r.Spaces[i].Occupied = []Seat{seat}
		}
	}
	workers := p.Workers
	rhineChoose(t, r, "plan", Action{Space: "draw_order", Slot: 1})
	if p.Workers != workers || len(r.Planned) != 1 || r.Planned[0].WorkerType != "soldato" {
		t.Fatal("worker identity lost or another worker spent")
	}
	for _, space := range r.Spaces {
		if space.ID == "summer_visitor" && len(space.Occupied) != 0 {
			t.Fatal("trigger worker was duplicated")
		}
	}
}

func TestRhineOldGeneralAndInfluencerPublicStars(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-oldGeneral", func(r *Room, p *Player) {
		r.Config.Board = "tuscany"
		r.Spaces = NewTuscanySpaces(3)
		r.Players[1].Influence = map[string]int{"pisa": 6}
	})
	rhineChoose(t, r, "resolve", Action{Influence: []InfluenceMove{{To: "pisa"}}})
	if p.Influence["pisa"] != 1 || len(r.Choices) != 1 {
		t.Fatal("general initial effect")
	}
	q := r.Player(r.Choices[0].PlayerID)
	stars := []RhineStarMove{{PlayerID: q.ID, From: "pisa"}, {PlayerID: q.ID, From: "pisa"}, {PlayerID: q.ID, From: "pisa"}}
	rhineChoose(t, r, "continue", Action{RhineStars: stars})
	if q.Influence["pisa"] != 3 {
		t.Fatal("opponent did not return3")
	}
	r, p = rhineFixture(t, "rhine-winter-influencer", func(r *Room, p *Player) {
		r.Config.Board = "tuscany"
		r.Spaces = NewTuscanySpaces(3)
		r.Players[1].Influence = map[string]int{"pisa": 1}
	})
	rhineChoose(t, r, "resolve", Action{Influence: []InfluenceMove{{To: "siena"}}})
	q = r.Players[1]
	rhineChoose(t, r, "continue", Action{RhineStars: []RhineStarMove{{PlayerID: q.ID, From: "pisa", To: "lucca"}}})
	if q.Influence["pisa"] != 0 || q.Influence["lucca"] != 1 {
		t.Fatal("any player's stars not moved")
	}
}

func TestRhineVineTraderPricesAndCheapBuyerExactIncrease(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-vineTrader", func(r *Room, p *Player) { p.Coins = 0 })
	rhineChoose(t, r, "trade", Action{CardIDs: []string{"v0", "v1"}})
	if p.Coins != 6 {
		t.Fatal("vine sale price")
	}
	rhineChoose(t, r, "continue", Action{Colors: []string{"vine", "winter"}})
	if p.Coins != 0 || len(p.Hand) != 3 {
		t.Fatal("mixed purchase prices")
	}
	r, p = rhineFixture(t, "rhine-winter-cheapBuyer", func(r *Room, p *Player) {
		p.Hand[2].Requirements = []Wine{{Type: "red", Value: 3}}
		p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
	})
	rhineChoose(t, r, "fill", Action{WineIDs: []string{"wine"}})
	if p.Wines[0].Value != 3 {
		t.Fatal("combined increase not2")
	}
	rhineChoose(t, r, "continue", Action{CardID: "o0", WineIDs: []string{"wine"}})
	p = &Player{Wines: []Wine{{ID: "ceiling", Type: "red", Value: 3}}}
	if rhineAgeTwoTotal(p, []string{"ceiling"}) == nil {
		t.Fatal("zero actual increase counted as2")
	}
}
