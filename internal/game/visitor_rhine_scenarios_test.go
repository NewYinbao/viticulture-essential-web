package game

import "testing"

func TestRhineRemainingPrintedBranches(t *testing.T) {
	tests := []struct {
		id, option string
		a          Action
		prepare    func(*Room, *Player)
		verify     func(*testing.T, *Room, *Player)
	}{
		{"summer-ampelograph", "plant", Action{CardID: "v0", Field: 1}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 52 || len(p.Fields[1].Vines) != 3 {
				t.Fatal("plant with spare capacity", p.Coins, p.Fields[1])
			}
		}},
		{"summer-contractor", "vp_build", Action{Building: "cottage"}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 46 || p.VP != 11 || !has(p, "cottage") {
				t.Fatal("two distinct contractor benefits")
			}
		}},
		{"summer-dismantler", "destroy", Action{Building: "trellis", Color: "red"}, nil, func(t *testing.T, r *Room, p *Player) {
			if has(p, "trellis") || len(p.Wines) != 1 || p.Wines[0].Value != 2 {
				t.Fatal("printed building cost becomes wine")
			}
		}},
		{"summer-earlyBuyer", "fill", Action{CardID: "o0", WineIDs: []string{"wine"}}, func(r *Room, p *Player) { p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}} }, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 52 || p.VP != 12 || len(p.Wines) != 0 {
				t.Fatal("fill and early buyer reward")
			}
		}},
		{"summer-grapeMerchant", "sell", Action{Grapes: []int{0}}, func(r *Room, p *Player) { p.Grapes = []Grape{{ID: "g", Color: "red", Value: 4}} }, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 56 || len(p.Grapes) != 0 {
				t.Fatal("three times regular grape sale price")
			}
		}},
		{"summer-miller", "build", Action{}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 48 || !has(p, "windmill") {
				t.Fatal("fixed two coin price")
			}
		}},
		{"summer-owner", "build_draw", Action{Building: "windmill"}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 45 || p.VP != 11 || len(p.Hand) != 4 {
				t.Fatal("newly built windmill qualifies after actions")
			}
		}},
		{"summer-peasant", "coins", Action{}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 52 || len(p.Hand) != 4 {
				t.Fatal("order and coins")
			}
		}},
		{"summer-reorganizer", "coins", Action{CardIDs: []string{"v0", "o0"}}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 55 || len(p.Hand) != 1 {
				t.Fatal("discard exactly two for coins")
			}
		}},
		{"summer-wineLover", "income", Action{}, func(r *Room, p *Player) { p.Buildings = append(p.Buildings, "tasting_room") }, func(t *testing.T, r *Room, p *Player) {
			if p.Income != 4 {
				t.Fatal("tasting room income")
			}
		}},
		{"winter-cellarmaster", "discard", Action{WineIDs: []string{"wine"}}, func(r *Room, p *Player) { p.Wines = []Wine{{ID: "wine", Type: "red", Value: 9}} }, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 54 || p.VP != 12 || len(p.Wines) != 0 {
				t.Fatal("highest cellar wine reward")
			}
		}},
		{"winter-chemist", "harvest", Action{Fields: []int{0}}, func(r *Room, p *Player) { p.Fields[0].Vines = []Card{{Red: 3, White: 2}} }, func(t *testing.T, r *Room, p *Player) {
			if p.VP != 11 || !p.Fields[0].Harvested || len(p.Grapes) != 2 || p.Grapes[0].Value+p.Grapes[1].Value != 3 {
				t.Fatal("harvest both colors minus one")
			}
		}},
		{"winter-craftsman", "upgrade_vp", Action{}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 46 || p.VP != 11 || !has(p, "medium_cellar") {
				t.Fatal("upgrade and point")
			}
		}},
		{"winter-endorser", "grape_orders", Action{Grapes: []int{0}}, func(r *Room, p *Player) { p.Grapes = []Grape{{ID: "g", Color: "red", Value: 2}} }, func(t *testing.T, r *Room, p *Player) {
			if len(p.Hand) != 6 || len(p.Grapes) != 0 {
				t.Fatal("grape exchanged for three orders")
			}
		}},
		{"winter-grapeVendor", "draw", Action{Grapes: []int{0}}, func(r *Room, p *Player) {
			p.Grapes = []Grape{{ID: "g", Color: "red", Value: 5}}
			p.Buildings = append(p.Buildings, "tasting_room")
		}, func(t *testing.T, r *Room, p *Player) {
			if len(p.Hand) != 6 || p.Coins != 54 || len(p.Grapes) != 0 {
				t.Fatal("round up half value and tasting room reward")
			}
		}},
		{"winter-hiredHand", "coins_harvest", Action{Fields: []int{0}}, nil, func(t *testing.T, r *Room, p *Player) {
			if p.Coins != 52 || !p.Fields[0].Harvested {
				t.Fatal("two of four benefits")
			}
		}},
		{"winter-specialHarvester", "separate", Action{Fields: []int{0}}, func(r *Room, p *Player) { p.Fields[0].Vines = []Card{{Red: 3}, {Red: 2, White: 1}} }, func(t *testing.T, r *Room, p *Player) {
			if len(p.Grapes) != 3 || !p.Fields[0].Harvested {
				t.Fatal("separate vines remain separate grapes", p.Grapes)
			}
		}},
		{"winter-zymologist", "make", Action{Recipes: [][]int{{0}}}, func(r *Room, p *Player) { p.Grapes = []Grape{{ID: "g", Color: "white", Value: 5}} }, func(t *testing.T, r *Room, p *Player) {
			if len(p.Wines) != 1 || p.Wines[0].Value != 5 || has(p, "medium_cellar") {
				t.Fatal("wine above cellar without constructing cellar")
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			r, p := rhineFixture(t, "rhine-"+tc.id, tc.prepare)
			rhineChoose(t, r, tc.option, tc.a)
			tc.verify(t, r, p)
			if len(r.Choices) > 0 {
				t.Fatal("unexpected continuation")
			}
		})
	}
}

func TestRhineBrickMasonTwoBuildingsOneCost(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-brickMason", nil)
	rhineChoose(t, r, "build", Action{Building: "cottage"})
	if p.Coins != 41 || !has(p, "cottage") || len(r.Choices) != 1 {
		t.Fatal("first build and single upfront price")
	}
	rhineChoose(t, r, "continue", Action{Building: "windmill"})
	if p.Coins != 41 || !has(p, "windmill") {
		t.Fatal("second free build charged again")
	}
}

func TestRhineGreenskeeperAndPlantReorganizer(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-greenskeeper", nil)
	rhineChoose(t, r, "extra", Action{CardID: "v0", Field: 0})
	rhineChoose(t, r, "continue", Action{CardID: "v1", Field: 0})
	if p.Coins != 52 || len(p.Fields[0].Vines) != 4 {
		t.Fatal("threshold grants additional vine")
	}
	r, p = rhineFixture(t, "rhine-summer-plantReorganizer", nil)
	original := p.Fields[0].Vines[0].ID
	rhineChoose(t, r, "replant", Action{})
	rhineChoose(t, r, "continue", Action{Colors: []string{"order"}})
	rhineChoose(t, r, "continue", Action{Fields: []int{0}, CardIDs: []string{original}})
	if cardIndex(p, original, "vine") < 0 {
		t.Fatal("uproot must return vine to hand")
	}
	rhineChoose(t, r, "continue", Action{CardID: original, Field: 1})
	if len(p.Fields[0].Vines) != 1 || len(p.Fields[1].Vines) != 3 {
		t.Fatal("vine could not be replanted")
	}
}

func TestRhineAccountantAndTheologianContinuedChoices(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-accountant", nil)
	rhineChoose(t, r, "resolve", Action{})
	if len(p.Hand) != 6 || p.Coins != 51 || len(r.Choices) != 2 {
		t.Fatal("accountant actor and opponents")
	}
	q := r.Player(r.Choices[0].PlayerID)
	hand := len(q.Hand)
	rhineChoose(t, r, "draw", Action{})
	rhineChoose(t, r, "skip", Action{})
	if len(q.Hand) != hand+1 {
		t.Fatal("opponent optional card")
	}
	r, p = rhineFixture(t, "rhine-winter-theologian", nil)
	rhineChoose(t, r, "harvest", Action{Fields: []int{0}})
	rhineChoose(t, r, "continue", Action{Color: "white"})
	if len(p.Grapes) != 2 || !p.Fields[0].Harvested {
		t.Fatal("extra one grape separate from ordinary yield")
	}
}

func TestRhineStructureReorganizerAndLobbyist(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-structureReorganizer", func(r *Room, p *Player) { r.Config.Board = "tuscany"; r.Spaces = NewTuscanySpaces(3) })
	rhineChoose(t, r, "draw", Action{Colors: []string{"order"}})
	rhineChoose(t, r, "destroy", Action{Building: "trellis"})
	rhineChoose(t, r, "build", Action{Building: "windmill"})
	if has(p, "trellis") || !has(p, "windmill") || p.Coins != 50 || len(p.Hand) != 4 {
		t.Fatal("any card and optional rebuild without structures module")
	}
	r, p = rhineFixture(t, "rhine-winter-lobbyist", func(r *Room, p *Player) {
		r.Config.Board = "tuscany"
		r.Spaces = NewTuscanySpaces(3)
		r.Players[1].VP = 20
		r.Players[1].Influence = map[string]int{"pisa": 2}
	})
	q := r.Players[1]
	rhineChoose(t, r, "resolve", Action{RhineStars: []RhineStarMove{{PlayerID: q.ID, From: "pisa"}, {PlayerID: q.ID, From: "pisa"}}})
	if p.Coins != 48 || p.VP != 7 || q.Influence["pisa"] != 0 {
		t.Fatal("cost per returned star and twenty point penalty")
	}
}

func TestRhineVirtuosoBonusAndNestedVisitor(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-virtuoso", nil)
	workers := p.Workers
	rhineChoose(t, r, "action", Action{Space: "tour", Slot: 1})
	if p.Coins != 53 || p.Workers != workers || r.Context != nil || len(r.Choices) != 0 {
		t.Fatal("free action with occupied-slot bonus must finish once", p.Coins, r.Context)
	}
	r, p = rhineFixture(t, "rhine-winter-virtuoso", func(r *Room, p *Player) {
		for _, f := range rhineCardFacts {
			if f.Card.ID == "rhine-summer-docent" {
				p.Hand = append(p.Hand, f.Card)
			}
		}
	})
	workers = p.Workers
	rhineChoose(t, r, "action", Action{Space: "summer_visitor", CardID: "rhine-summer-docent"})
	if r.Context == nil || r.Context.Parent == nil {
		t.Fatal("lost virtual action parent")
	}
	rhineChoose(t, r, "coins", Action{})
	if p.Coins != 53 || p.Workers != workers || len(r.Choices) != 0 || r.Context != nil {
		t.Fatal("nested visitor did not restore parent")
	}
}

func TestRhineExceptionalMakeHonorsCharmatAndFermentationContinuation(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-zymologist", func(r *Room, p *Player) {
		r.Config.Structures = true
		p.Buildings = append(p.Buildings, "charmat")
		p.Grapes = []Grape{{ID: "r", Color: "red", Value: 4}, {ID: "w", Color: "white", Value: 3}}
	})
	rhineChoose(t, r, "make", Action{Recipes: [][]int{{0, 1}}, RecipeTypes: []string{"sparkling"}})
	if len(p.Wines) != 1 || p.Wines[0].Type != "sparkling" || p.Wines[0].Value != 7 || has(p, "large_cellar") {
		t.Fatal("Charmat permitted exceptional sparkling without gaining cellar")
	}
	r, p = rhineFixture(t, "rhine-winter-specialHarvester", func(r *Room, p *Player) {
		r.Config.Structures = true
		p.Buildings = append(p.Buildings, "fermentation_tank")
	})
	rhineChoose(t, r, "separate", Action{Fields: []int{0}})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_fermentation" || r.Context == nil {
		t.Fatal("manual harvest lost fermentation choice")
	}
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}); err != nil {
		t.Fatal(err)
	}
	if r.Context != nil || len(r.Choices) != 0 {
		t.Fatal("fermentation skip did not complete original visitor")
	}
}

func TestRhineInfluencerBanquetHallBenefitsMoverForOwnAndOpponentStars(t *testing.T) {
	for _, hall := range []bool{false, true} {
		r, p := rhineFixture(t, "rhine-winter-influencer", func(r *Room, p *Player) {
			r.Config.Board = "tuscany"
			r.Config.Structures = true
			r.Spaces = NewTuscanySpaces(3)
			p.Influence = map[string]int{"pisa": 1}
			r.Players[1].Influence = map[string]int{"pisa": 1}
			if hall {
				p.Buildings = append(p.Buildings, "banquet_hall")
			}
		})
		rhineChoose(t, r, "resolve", Action{Influence: []InfluenceMove{{To: "pisa"}}})
		q := r.Players[1]
		coins, opponentCoins := p.Coins, q.Coins
		rhineChoose(t, r, "continue", Action{RhineStars: []RhineStarMove{{PlayerID: p.ID, From: "pisa", To: "siena"}, {PlayerID: q.ID, From: "pisa", To: "siena"}}})
		expected := coins
		if hall {
			expected += 2
		}
		if p.Coins != expected || q.Coins != opponentCoins || p.Influence["siena"] != 1 || q.Influence["siena"] != 1 {
			t.Fatal("Banquet Hall must reward the mover for each moved star", hall, p.Coins, q.Coins)
		}
	}
}

func TestRhineInfluencerBanquetDrawQueuesMercado(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-influencer", func(r *Room, p *Player) {
		r.Config.Board = "tuscany"
		r.Config.Structures = true
		r.Spaces = NewTuscanySpaces(3)
		p.Buildings = append(p.Buildings, "banquet_hall", "mercado")
		p.Influence = map[string]int{"pisa": 1}
		p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
		r.Decks["order"] = []Card{{ID: "market-order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
	})
	rhineChoose(t, r, "resolve", Action{Influence: []InfluenceMove{{To: "pisa"}}})
	rhineChoose(t, r, "continue", Action{RhineStars: []RhineStarMove{{PlayerID: p.ID, From: "pisa", To: "arezzo"}}})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_mercado" || r.Choices[0].CardIDs[0] != "market-order" {
		t.Fatal("Banquet Hall order draw must preserve its Mercado response")
	}
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}); err != nil {
		t.Fatal(err)
	}
	if r.Context != nil || len(r.Choices) != 0 {
		t.Fatal("Mercado did not restore completed influencer")
	}
}

func TestRhineSupervisorAwardsOnlyNewSparklingWine(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-supervisor", func(r *Room, p *Player) {
		p.Buildings = append(p.Buildings, "medium_cellar", "large_cellar")
		p.Wines = []Wine{{ID: "old", Type: "sparkling", Value: 8}}
		p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 4}, {ID: "r2", Color: "red", Value: 2}, {ID: "w1", Color: "white", Value: 1}, {ID: "w3", Color: "white", Value: 3}}
	})
	rhineChoose(t, r, "make", Action{Recipes: [][]int{{0, 1, 2}, {3}}})
	if p.VP != 11 || len(p.Wines) != 3 || len(p.Grapes) != 0 {
		t.Fatal("Supervisor counted old sparkling or ordinary white wine", p.VP, p.Wines)
	}
}

func TestRhineAdministratorReservationExecutesAtWinterStart(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-administrator", nil)
	seat := Seat{PlayerID: p.ID, Slot: 2, WorkerType: "soldato"}
	r.Context.TriggerSeat = &seat
	r.Context.TriggerSpace = "summer_visitor"
	for i := range r.Spaces {
		if r.Spaces[i].ID == "summer_visitor" {
			r.Spaces[i].Occupied = []Seat{seat}
		}
	}
	workers := p.Workers
	rhineChoose(t, r, "plan", Action{Space: "draw_order", Slot: 1})
	if len(r.Choices) != 0 || len(r.Planned) != 1 {
		t.Fatal("reservation resolved before future season")
	}
	r.Phase = "winter"
	r.startNextPlanner()
	if len(r.Choices) != 1 || r.Choices[0].Kind != "planner" || r.Context.WorkerType != "soldato" {
		t.Fatal("winter reservation lost worker kind")
	}
	hand := len(p.Hand)
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != hand+2 || p.Workers != workers || len(r.Planned) != 0 || len(r.Choices) != 0 {
		t.Fatal("future action/bonus not exactly once")
	}
}

func TestRhineLaborerPaysOnceThenMakesFreshHarvest(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-laborer", nil)
	rhineChoose(t, r, "harvest_make", Action{Fields: []int{0}})
	if p.VP != 9 || len(p.Grapes) != 2 || len(r.Choices) != 1 {
		t.Fatal("laborer cost and harvest")
	}
	rhineChoose(t, r, "continue", Action{Recipes: [][]int{{0}, {1}}})
	if p.VP != 9 || len(p.Grapes) != 0 || len(p.Wines) != 2 {
		t.Fatal("laborer must make harvested grapes without another VP cost")
	}
}

func TestRhineWineEngineerKeepsOriginalGrapesAndBranch(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-wineEngineer", func(r *Room, p *Player) {
		p.Grapes = []Grape{{ID: "r", Color: "red", Value: 1}, {ID: "w", Color: "white", Value: 1}, {ID: "other", Color: "red", Value: 6}}
	})
	rhineChoose(t, r, "age", Action{Grapes: []int{0, 1}})
	c := r.Choices[0]
	coins := p.Coins
	for _, a := range []Action{{Type: "choose", ChoiceID: c.ID, Option: "wine", Color: "red"}, {Type: "choose", ChoiceID: c.ID, Option: "again", Grapes: []int{2}}} {
		if r.Apply(p.ID, a) == nil {
			t.Fatal("engineer changed branch/original grapes")
		}
		if p.Coins != coins {
			t.Fatal("rejected selection charged coins")
		}
	}
	rhineChoose(t, r, "again", Action{Grapes: []int{1}})
	rhineChoose(t, r, "done", Action{})
	if p.Grapes[0].Value != 2 || p.Grapes[1].Value != 3 || p.Grapes[2].Value != 6 {
		t.Fatal("fixed grape subset")
	}
}

func TestRhineDestroyCellarRetainsBuiltLargeCellarAndExistingWine(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-brickMason", func(r *Room, p *Player) {
		p.Buildings = append(p.Buildings, "medium_cellar", "large_cellar")
		p.Wines = []Wine{{ID: "old", Type: "red", Value: 8}}
	})
	rhineChoose(t, r, "destroy", Action{Building: "medium_cellar"})
	if has(p, "medium_cellar") || !has(p, "large_cellar") || len(p.Wines) != 1 || p.Wines[0].Value != 8 {
		t.Fatal("demolition incorrectly removed other building or wine")
	}
}

func TestRhineSonInLawPrecedesFirstReservedWinterAction(t *testing.T) {
	r, p := rhineFixture(t, "rhine-summer-sonInLaw", nil)
	rhineChoose(t, r, "field", Action{Field: 0})
	r.Config.Visitors = "rhine"
	r.Phase = "winter"
	p.Workers = 0
	p.LargeWorker = false
	r.Planned = []PlannedPlacement{{PlayerID: p.ID, Space: "make_wine", Slot: 2}}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "make_wine" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: p.ID, Slot: 2}}
		}
	}
	r.startNextPlanner()
	r.rhinePrepareWinterTurn()
	if r.Context == nil || r.Context.Step != "rhine_first_winter" || r.Context.Parent == nil {
		t.Fatal("Son-in-law skipped before first reserved action")
	}
	rhineChoose(t, r, "harvest", Action{Grapes: []int{1, 1}})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "planner" || len(p.Grapes) != 2 {
		t.Fatal("reserved action was lost after harvest")
	}
	c := r.Choices[0]
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "execute", Recipes: [][]int{{0}}}); err != nil {
		t.Fatal(err)
	}
	if p.RhineWinterActionYear != r.Year || len(r.Choices) != 0 || len(p.Wines) != 1 {
		t.Fatal("reserved first action not counted")
	}
}

func TestRhineChemistHarvestMachinePreservesYieldRuleAndOnePoint(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-chemist", func(r *Room, p *Player) {
		r.Config.Structures = true
		p.Buildings = append(p.Buildings, "harvest_machine", "fermentation_tank")
		p.Fields[0].Vines = []Card{{Red: 3, White: 2}}
		p.Fields[1].Vines = []Card{{Red: 2, White: 3}}
		p.Fields[2].Sold = true
	})
	rhineChoose(t, r, "harvest", Action{HarvestAll: true, Fields: []int{0, 1}})
	if p.VP != 11 || len(p.Grapes) != 4 || !p.Fields[0].Harvested || !p.Fields[1].Harvested || p.Fields[2].Harvested || len(r.Choices) != 1 || r.Choices[0].Kind != "structure_fermentation" {
		t.Fatal("Harvest Machine changed special yields or repeated one-card reward")
	}
}

func TestRhineSpecialHarvesterHarvestMachineSeparatesEveryVine(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-specialHarvester", func(r *Room, p *Player) {
		r.Config.Structures = true
		p.Buildings = append(p.Buildings, "harvest_machine")
		p.Fields[0].Vines = []Card{{Red: 3}, {Red: 1}}
		p.Fields[1].Vines = []Card{{White: 3}, {White: 1}}
		p.Fields[2].Sold = true
	})
	rhineChoose(t, r, "separate", Action{HarvestAll: true, Fields: []int{0, 1}})
	if len(p.Grapes) != 4 || !p.Fields[0].Harvested || !p.Fields[1].Harvested || p.VP != 10 {
		t.Fatal("each vine should remain a separate harvested yield")
	}
}
