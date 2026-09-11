package game

import "testing"

func TestCardWorkshopDiscountAppliesToFixedBuildings(t *testing.T) {
	for _, visitor := range []bool{false, true} {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"workshop"}
		p.Coins = 1
		var err error
		if visitor {
			err = r.visitorBuild(p, "trellis", 0)
		} else {
			err = r.perform(p, Action{Space: "build", Building: "trellis"})
		}
		if err != nil {
			t.Fatalf("visitor=%v: %v", visitor, err)
		}
		if p.Coins != 0 || !has(p, "trellis") {
			t.Fatalf("wrong fixed build cost: %+v", p)
		}
	}
}
func TestCardMixerRejectsTwoOfSameWine(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"mixer", "medium_cellar", "large_cellar"}
	p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 3}, {ID: "w1", Color: "white", Value: 3}, {ID: "r2", Color: "red", Value: 2}, {ID: "w2", Color: "white", Value: 3}}
	if err := r.performStructureAction(p, mustStructure("mixer"), Action{Recipes: [][]int{{0, 1}, {2, 3}}}); err == nil {
		t.Fatal("accepted 2 blush wines; card permits only 1 of each type")
	}
	if len(p.Grapes) != 4 || len(p.Wines) != 0 {
		t.Fatal("invalid Mixer spent resources")
	}
}
func TestCardFermentationOnYokeAndVisitorHarvest(t *testing.T) {
	for _, source := range []string{"yoke", "single", "visitor"} {
		t.Run(source, func(t *testing.T) {
			r := structureRoom()
			p := r.Players[0]
			p.Buildings = []string{"fermentation_tank", "yoke"}
			p.Fields[0].Vines = []Card{{ID: "v1", Type: "vine", Red: 1}}
			r.Context = &ActionContext{ActorID: p.ID, Space: "winter_visitor", Remaining: 1}
			var err error
			switch source {
			case "yoke":
				err = r.performEEPlacement(p, Action{Space: "yoke", Mode: "harvest", Field: 0}, false)
			case "single":
				err = r.performEEPlacement(p, Action{Space: "harvest", Field: 0}, false)
			case "visitor":
				err = r.visitorHarvest(p, Action{Fields: []int{0}}, 1)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_fermentation" {
				t.Fatalf("missing fermentation after %s: %+v", source, r.Choices)
			}
		})
	}
}
func TestCardAcademyFeeAppliesToVisitorTraining(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	q := r.Players[1]
	q.Buildings = []string{"academy"}
	p.Coins = 2
	other := q.Coins
	if err := r.visitorTrain(p, 1, false); err != nil {
		t.Fatal(err)
	}
	if p.Coins != 0 || q.Coins != other+1 {
		t.Fatalf("Academy fee missing: %d / %d", p.Coins, q.Coins)
	}
}

func TestCardCharmatIsAvailableAtEveryWineSource(t *testing.T) {
	for _, source := range []string{"public", "wine_press", "mixer", "fermentation", "visitor"} {
		t.Run(source, func(t *testing.T) {
			r := structureRoom()
			p := r.Players[0]
			p.Buildings = []string{"charmat", "medium_cellar", "large_cellar", source}
			p.Grapes = []Grape{{ID: "r", Color: "red", Value: 4}, {ID: "w", Color: "white", Value: 3}}
			a := Action{Space: "make_wine", Recipes: [][]int{{0, 1}}, RecipeTypes: []string{"sparkling"}}
			var err error
			switch source {
			case "public":
				err = r.performEEPlacement(p, a, false)
			case "wine_press", "mixer":
				err = r.performStructureAction(p, mustStructure(source), a)
			case "fermentation":
				r.Context = &ActionContext{ActorID: p.ID, Space: "harvest"}
				r.enqueue(p.ID, "structure_fermentation", []string{"make", "skip"}, 1)
				a.Type = "choose"
				a.ChoiceID = r.Choices[0].ID
				a.Option = "make"
				err = r.resolveChoice(p.ID, a)
			case "visitor":
				err = r.visitorMake(p, a, 1)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Wines) != 1 || p.Wines[0].Type != "sparkling" || p.Wines[0].Value != 7 {
				t.Fatalf("Charmat lost at %s: %+v", source, p.Wines)
			}
		})
	}
}
func TestCardCharmatMixedBatchAndRequiresOwnership(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"charmat", "mixer", "medium_cellar", "large_cellar"}
	p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 4}, {ID: "w1", Color: "white", Value: 3}, {ID: "r2", Color: "red", Value: 3}, {ID: "w2", Color: "white", Value: 2}}
	a := Action{Recipes: [][]int{{0, 1}, {2, 3}}, RecipeTypes: []string{"sparkling", ""}}
	if err := r.performStructureAction(p, mustStructure("mixer"), a); err != nil {
		t.Fatal(err)
	}
	if len(p.Wines) != 2 || p.Wines[0].Type != "sparkling" || p.Wines[1].Type != "blush" || p.VP != 1 {
		t.Fatalf("mixed batch: %+v", p)
	}
	p.Buildings = []string{"medium_cellar", "large_cellar"}
	p.Wines = nil
	p.Grapes = []Grape{{ID: "r", Color: "red", Value: 4}, {ID: "w", Color: "white", Value: 3}}
	if err := r.visitorMake(p, Action{Recipes: [][]int{{0, 1}}, RecipeTypes: []string{"sparkling"}}, 1); err == nil {
		t.Fatal("Charmat used without ownership")
	}
	if len(p.Grapes) != 2 || len(p.Wines) != 0 {
		t.Fatal("failed Charmat spent resources")
	}
}
func TestCardFermentationDoesNotOverwriteOuterAction(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"fermentation_tank", "harvest_machine"}
	for i := 0; i < 3; i++ {
		p.Fields[i].Vines = []Card{{ID: NewID(), Type: "vine", Red: 1}}
	}
	outer := Action{Space: "summer_visitor", CardID: "summer-22"}
	r.Context = &ActionContext{ActorID: p.ID, PendingAction: &outer, Remaining: 1}
	if err := r.visitorHarvest(p, Action{HarvestAll: true}, 1); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Context.PendingAction.CardID != "summer-22" {
		t.Fatal("nested harvest overwrote continuation or offered reward per field")
	}
	for _, f := range p.Fields {
		if !f.Harvested {
			t.Fatal("Harvest Machine missed field")
		}
	}
	a := Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "make", Recipes: [][]int{{0}}}
	if err := r.resolveChoice(p.ID, a); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Visitor == nil || r.Choices[0].Visitor.Stage != "second" {
		t.Fatal("Fermentation lost outer visitor continuation")
	}
}
func TestCardVisitorSpecialTrainingCostAndImmediateAvailability(t *testing.T) {
	for _, now := range []bool{false, true} {
		r := specialWorkerRoom("oracle", "mafioso")
		p := r.Players[0]
		q := r.Players[1]
		q.Buildings = []string{"academy"}
		q.Coins = 0
		p.Coins = 4
		before := p.Workers
		if err := r.visitorTrain(p, 2, now, "oracle"); err != nil {
			t.Fatal(err)
		}
		ready := r.Year + 1
		if now {
			ready = r.Year
		}
		if p.Coins != 0 || q.Coins != 1 || p.Workers != before || p.SpecialWorkerReady["oracle"] != ready {
			t.Fatalf("special training cost/availability mismatch: %+v", p)
		}
	}
}

func TestCardOptionalStructureCannotStrandMandatoryVisitor(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"tap_room"}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
	p.Hand = []Card{{ID: "order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2}}
	r.Context = &ActionContext{ActorID: p.ID, Space: "winter_visitor"}
	r.enqueue(p.ID, "structure_visitor_bonus", []string{"tap_room", "skip"}, 1)
	r.visitorPrompt(p, "winter-14", "fill", []string{"fill"})
	a := Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "tap_room", WineIDs: []string{"wine"}}
	if err := r.Apply(p.ID, a); err == nil {
		t.Fatal("Tap Room stranded mandatory fill")
	}
	if len(p.Wines) != 1 || p.VP != 0 || len(r.Choices) != 2 {
		t.Fatal("rejected optional reward was not atomic")
	}
	a.Option = "skip"
	if err := r.Apply(p.ID, a); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Visitor.Stage != "fill" {
		t.Fatal("mandatory continuation lost")
	}
}
func TestPrintedStructureCardCostsAndCategories(t *testing.T) {
	// Manually transcribed from the 36 numbered printed images, not the catalogue.
	costs := []int{2, 3, 2, 2, 5, 4, 7, 4, 3, 8, 3, 3, 2, 5, 5, 5, 4, 3, 3, 5, 3, 3, 2, 4, 3, 4, 5, 5, 2, 3, 4, 3, 2, 9, 3, 2}
	actions := map[int]bool{1: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 10: true, 11: true, 12: true, 22: true, 32: true}
	residual := map[int]bool{13: true, 33: true, 34: true, 35: true, 36: true}
	for i, d := range structureDefs {
		category := "enhancement"
		if actions[i+1] {
			category = "action"
		}
		if residual[i+1] {
			category = "residual"
		}
		if d.Cost != costs[i] || d.RequiresWorker != actions[i+1] || d.Category != category {
			t.Fatalf("printed image %d mismatch: %+v", i+1, d)
		}
	}
}
