package game

import "testing"

func structureRoom() *Room {
	r := tuscanyRoom(2)
	r.Config.Structures = true
	r.Spaces = r.spacesForRules()
	r.Phase = "spring"
	r.TurnID = r.Players[0].ID
	r.Players[0].Coins = 20
	return r
}

func structureHand(p *Player, id string) {
	d, _ := structureDef(id)
	p.Hand = append(p.Hand, Card{ID: structureCardID(d.ID), Type: "structure", StructureID: d.ID, StructureCost: d.Cost, Name: d.Name, Description: d.Description})
}

func TestAll36StructureCardsHaveAuditableRules(t *testing.T) {
	if len(structureDefs) != 36 {
		t.Fatalf("got %d structures", len(structureDefs))
	}
	seen := map[string]bool{}
	for _, d := range structureDefs {
		if seen[d.ID] || d.ID == "" || d.Name == "" || d.Cost <= 0 || d.Description == "" || d.RuleSource == "" {
			t.Fatalf("bad structure %+v", d)
		}
		seen[d.ID] = true
	}
	workers := map[string]bool{"cask": true, "wine_cave": true, "trading_post": true, "shop": true, "wine_press": true, "school": true, "wine_bar": true, "ristorante": true, "guest_house": true, "cafe": true, "label_factory": true, "mixer": true}
	for _, d := range structureDefs {
		if d.RequiresWorker != workers[d.ID] {
			t.Fatalf("worker icon mismatch for %s", d.ID)
		}
	}
}

func TestStructureBuildMatFieldStudioWorkshopAndDestroy(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	structureHand(p, "cask")
	if err := r.perform(p, Action{Space: "build", Building: "structure-cask", Mode: "mat", StructureSlot: 0}); err != nil {
		t.Fatal(err)
	}
	if !p.hasStructure("cask") || p.VP != 1 || p.Coins != 18 || p.StructureSlots[0] != "cask" {
		t.Fatalf("build mat: %+v", p)
	}
	structureHand(p, "studio")
	structureHand(p, "workshop")
	if err := r.perform(p, Action{Space: "build", Building: "structure-studio", Mode: "field", Field: 0}); err != nil {
		t.Fatal(err)
	}
	if p.Fields[0].Structure != "studio" || p.VP != 2 {
		t.Fatalf("build field: %+v", p)
	}
	if err := r.perform(p, Action{Space: "build", Building: "structure-workshop", Mode: "mat", StructureSlot: 1}); err != nil {
		t.Fatal(err)
	}
	if p.Coins != 10 {
		t.Fatalf("workshop cost/build: %d", p.Coins)
	}
	if err := r.destroyStructure(p, "studio"); err != nil {
		t.Fatal(err)
	}
	if p.Fields[0].Structure != "" || p.VP != 4 {
		t.Fatalf("destroy altered score/field: %+v", p)
	}
}

func TestStructureActionsAndWineEnhancements(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"medium_cellar", "large_cellar", "patio", "penthouse", "charmat"}
	p.Grapes = []Grape{{ID: "r", Color: "red", Value: 4}, {ID: "w", Color: "white", Value: 4}}
	if err := r.perform(p, Action{Space: "make_wine", Mode: "sparkling", Recipes: [][]int{{0, 1}}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Wines) != 1 || p.Wines[0].Type != "sparkling" || p.Wines[0].Value != 8 || p.Coins != 22 || p.VP != 1 {
		t.Fatalf("charmat/patio/penthouse: %+v", p)
	}
	p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 4}, {ID: "r2", Color: "red", Value: 4}, {ID: "w2", Color: "white", Value: 4}}
	if err := r.perform(p, Action{Space: "make_wine", Mode: "sparkling", Recipes: [][]int{{0, 1, 2}}}); err == nil {
		t.Fatal("Charmat must reject the normal 2 red + 1 white sparkling recipe")
	}
	if len(p.Grapes) != 3 || len(p.Wines) != 1 {
		t.Fatalf("rejected Charmat recipe mutated resources: %+v", p)
	}
	structureHand(p, "cask")
	if err := r.perform(p, Action{Space: "build", Building: "structure-cask", Mode: "mat", StructureSlot: 0}); err != nil {
		t.Fatal(err)
	}
	if err := r.performStructureAction(p, mustStructure("cask"), Action{WineIDs: []string{p.Wines[0].ID}}); err != nil {
		t.Fatal(err)
	}
	if p.Wines[0].Value != 9 {
		t.Fatalf("cask age: %+v", p.Wines)
	}
	q := r.Players[1]
	q.Coins = 50
	for _, id := range []string{"distiller", "storehouse", "silo", "dock", "statue"} {
		structureHand(q, id)
	}
	for i, id := range []string{"distiller", "storehouse", "silo"} {
		if err := r.buildStructure(q, Action{Building: structureCardID(id), Mode: "field", Field: i}, 0); err != nil {
			t.Fatal(id, err)
		}
	}
	for i, id := range []string{"dock", "statue"} {
		if err := r.buildStructure(q, Action{Building: structureCardID(id), Mode: "mat", StructureSlot: i}, 0); err != nil {
			t.Fatal(id, err)
		}
	}
	r.Decks = map[string][]Card{"vine": {{ID: "v", Type: "vine"}}, "order": {{ID: "o", Type: "order"}}}
	r.Discards = map[string][]Card{}
	r.finishYear()
	if q.VP < 2 || len(q.Hand) < 2 {
		t.Fatalf("residual structures not applied: vp=%d hand=%d", q.VP, len(q.Hand))
	}
}

func TestStructureAgingRejectsOccupiedSlotsAtomically(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"medium_cellar", "large_cellar"}
	p.Wines = []Wine{{ID: "red-2", Type: "red", Value: 2}, {ID: "red-3", Type: "red", Value: 3}}
	if err := r.performStructureAction(p, mustStructure("cask"), Action{WineIDs: []string{"red-2"}}); err == nil {
		t.Fatal("cask must reject an occupied destination slot")
	}
	if p.Wines[0].Value != 2 || p.Wines[1].Value != 3 {
		t.Fatalf("failed cask changed wine slots: %+v", p.Wines)
	}
	if err := r.performStructureAction(p, mustStructure("wine_cave"), Action{WineIDs: []string{"red-2", "red-2"}}); err == nil {
		t.Fatal("wine cave must reject duplicate bottle selection")
	}
	if p.Wines[0].Value != 2 || p.Wines[1].Value != 3 {
		t.Fatalf("failed wine cave changed wine slots: %+v", p.Wines)
	}
}

func mustStructure(id string) StructureDef {
	d, ok := structureDef(id)
	if !ok {
		panic(id)
	}
	return d
}

func TestStructureVisitorTriggersAndBarnChoice(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"inn", "tap_room", "tavern", "barn"}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 3}}
	p.Grapes = []Grape{{ID: "g1", Color: "red", Value: 1}, {ID: "g2", Color: "white", Value: 1}}
	r.Context = &ActionContext{ActorID: p.ID}
	if !r.queueStructureVisitorBonus(p) || r.Choices[len(r.Choices)-1].Kind != "structure_visitor_bonus" {
		t.Fatal("visitor trigger missing")
	}
	c := r.Choices[len(r.Choices)-1]
	if err := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "tap_room+tavern", WineIDs: []string{"wine"}, GrapeIDs: []string{"g1", "g2"}}); err != nil {
		t.Fatal(err)
	}
	if p.VP != 5 || len(p.Wines) != 0 || len(p.Grapes) != 0 {
		t.Fatalf("visitor bonus: %+v", p)
	}
}

func TestStructureVisitorBonusRequiresExactEligibleSelections(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"tavern"}
	p.Grapes = []Grape{{ID: "green", Color: "green", Value: 9}, {ID: "red", Color: "red", Value: 1}, {ID: "white", Color: "white", Value: 1}}
	r.Context = &ActionContext{ActorID: p.ID}
	if !r.queueStructureVisitorBonus(p) {
		t.Fatal("tavern bonus should be offered with two eligible grapes")
	}
	c := r.Choices[len(r.Choices)-1]
	before := append([]Grape(nil), p.Grapes...)
	if err := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "tavern", GrapeIDs: []string{"green", "red"}}); err == nil {
		t.Fatal("tavern accepted a green grape")
	}
	if len(p.Grapes) != len(before) || p.Grapes[0].ID != before[0].ID || p.VP != 0 {
		t.Fatalf("invalid tavern choice mutated state: %+v", p)
	}
	if err := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "tavern", GrapeIDs: []string{"red", "white"}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Grapes) != 1 || p.Grapes[0].ID != "green" || p.VP != 3 {
		t.Fatalf("valid tavern choice not applied: %+v", p)
	}
}

func TestStructurePrivateRewardsAndDestroyDiscard(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"medium_cellar", "large_cellar", "academy"}
	p.Wines = []Wine{{ID: "w", Type: "red", Value: 4}}
	p.Coins = 3
	r.Decks = map[string][]Card{"order": {{ID: "o", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}, Points: 2}}}
	r.Discards = map[string][]Card{}
	if err := r.performStructureAction(p, mustStructure("cask"), Action{WineIDs: []string{"w"}}); err != nil {
		t.Fatal(err)
	}
	if len(p.Wines) != 1 || p.Wines[0].Value != 6 || len(r.Decks["order"]) != 0 || len(p.Hand) != 1 {
		t.Fatalf("cask reward: %+v decks=%v", p, r.Decks)
	}
	if err := r.performStructureAction(p, mustStructure("wine_bar"), Action{WineIDs: []string{"w"}}); err != nil {
		t.Fatal(err)
	}
	if p.VP != 2 || len(p.Wines) != 0 {
		t.Fatalf("wine bar reward: %+v", p)
	}
	p.Grapes = []Grape{{ID: "g", Color: "red", Value: 1}}
	if err := r.performStructureAction(p, mustStructure("cafe"), Action{GrapeID: "g"}); err != nil {
		t.Fatal(err)
	}
	if p.Coins != 6 || p.VP != 3 {
		t.Fatalf("cafe reward: %+v", p)
	}
	q := r.Players[1]
	q.Buildings = []string{"academy"}
	p.Coins = 5
	if err := r.performStructureAction(p, mustStructure("school"), Action{}); err != nil {
		t.Fatal(err)
	}
	if p.TotalWorkers != 4 || p.Coins != 5 || q.Coins != 1 {
		t.Fatalf("school/academy reward: p=%+v q=%+v", p, q)
	}
	p.Coins = 10
	structureHand(p, "statue")
	if err := r.buildStructure(p, Action{Building: "structure-statue", Mode: "mat", StructureSlot: 0}, 0); err != nil {
		t.Fatal(err)
	}
	if err := r.destroyStructure(p, "statue"); err != nil {
		t.Fatal(err)
	}
	if len(r.Discards["structure"]) != 1 || r.Discards["structure"][0].ID != "structure-statue" || p.VP < 1 {
		t.Fatalf("destroy discard: %+v", r.Discards)
	}
}

func TestSchoolCanTrainSpecialWorkerThroughPersistentChoice(t *testing.T) {
	r := specialWorkerRoom("farmer")
	r.Config.Structures = true
	r.Spaces = r.spacesForRules()
	r.Phase = "spring"
	p := r.Players[0]
	p.Buildings = []string{"school"}
	p.Coins = 10
	r.Context = &ActionContext{ActorID: p.ID, PendingAction: &Action{Space: "structure_action", Building: "school"}}
	if err := r.performStructureAction(p, mustStructure("school"), Action{Building: "school"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_train" || p.Coins != 10 {
		t.Fatalf("school did not preserve special-worker choice atomically: choices=%+v player=%+v", r.Choices, p)
	}
	c := r.Choices[0]
	if err := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "farmer"}); err != nil {
		t.Fatal(err)
	}
	if !p.hasSpecialWorker("farmer") || p.SpecialWorkerReady["farmer"] != r.Year || p.TotalWorkers != 4 || p.Workers != 2 || p.Coins != 10 {
		t.Fatalf("school special-worker training wrong: %+v", p)
	}
}

func TestStructureMarketAndLabelFactoryChoices(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"mercado", "label_factory"}
	p.Coins = 3
	p.Wines = []Wine{{ID: "w", Type: "red", Value: 3}}
	p.Hand = []Card{{ID: "o", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}, Points: 2}}
	r.Decks = map[string][]Card{"order": {{ID: "o2", Type: "order"}}}
	r.Discards = map[string][]Card{}
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID}
	r.draw(p, "order")
	if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_mercado" {
		t.Fatalf("mercado choice: %+v", r.Choices)
	}
	r.Choices = nil
	if err := r.perform(p, Action{Space: "fill_order", CardID: "o", WineIDs: []string{"w"}}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 0 {
		t.Fatalf("label factory must not trigger on ordinary delivery: %+v", r.Choices)
	}
	p.Hand = []Card{{ID: "o3", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}, Points: 2}}
	p.Wines = []Wine{{ID: "w3", Type: "red", Value: 3}}
	p.Coins = 3
	if err := r.performStructureAction(p, mustStructure("label_factory"), Action{CardID: "o3", WineIDs: []string{"w3"}}); err != nil {
		t.Fatal(err)
	}
	if p.VP != 6 || p.Coins != 0 || len(p.Hand) != 0 {
		t.Fatalf("label factory result: %+v", p)
	}
}

func TestStructureInfluenceLuccaAndSide2Rewards(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	r.Decks["structure"] = []Card{{ID: "structure-test", Type: "structure"}}
	if err := r.influence(p, InfluenceMove{To: "lucca"}); err != nil {
		t.Fatal(err)
	}
	if p.Influence["lucca"] != 1 || len(p.Hand) != 1 || p.Hand[0].Type != "structure" {
		t.Fatalf("Lucca structure reward: influence=%v hand=%+v", p.Influence, p.Hand)
	}
	// Tuscany permits moving stars only after all six of the player's stars
	// have been placed. The Lucca reward itself must not be repeated by moving.
	p.Influence["pisa"] = 5
	coinsBeforeMove := p.Coins
	if err := r.influence(p, InfluenceMove{From: "lucca", To: "pisa"}); err != nil {
		t.Fatal(err)
	}
	if p.Coins != coinsBeforeMove {
		t.Fatal("moving a star must not repeat its region reward")
	}
}

func TestStructureActionAndPassiveBoundaries(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"mixer", "label_factory", "studio"}
	p.Grapes = []Grape{{ID: "r", Color: "red", Value: 3}, {ID: "w", Color: "white", Value: 3}}
	if err := r.performStructureAction(p, mustStructure("mixer"), Action{Recipes: [][]int{{0}}}); err == nil {
		t.Fatal("mixer must reject a single-color wine")
	}
	p.Coins = 10
	if err := r.perform(p, Action{Space: "build", Building: "cottage"}); err != nil {
		t.Fatal(err)
	}
	if p.VP != 1 {
		t.Fatalf("studio should score on a later base-structure build: %d", p.VP)
	}
}

func TestHarvestMachineAllFieldsAndFermentationChoice(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"harvest_machine"}
	p.Fields[0].Vines = []Card{{ID: "vine-1", Type: "vine", Red: 2}}
	p.Fields[1].Vines = []Card{{ID: "vine-2", Type: "vine", White: 1}}
	p.Fields[2].Sold = true
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID}
	if err := r.performEEPlacement(p, Action{Space: "harvest", Mode: "all"}, false); err != nil {
		t.Fatal(err)
	}
	if !p.Fields[0].Harvested || !p.Fields[1].Harvested || p.Fields[2].Harvested {
		t.Fatalf("harvest machine selected wrong fields: %+v", p.Fields)
	}
	if len(p.Grapes) != 2 {
		t.Fatalf("harvest machine grapes: %+v", p.Grapes)
	}
}

func TestStatueCannotTriggerFinalYear(t *testing.T) {
	r := tuscanyRoom(2)
	r.Config.Structures = true
	r.Phase = "winter"
	for _, p := range r.Players {
		p.NextWake = 1
		p.Wake = 1
	}
	p := r.Players[0]
	p.Buildings = []string{"statue"}
	p.VP = 24
	r.tuscanyFinishYear()
	if r.FinalYear || r.Phase != "spring" || r.Year != 2 || p.VP != 25 {
		t.Fatalf("statue incorrectly ended year: final=%v phase=%s year=%d vp=%d", r.FinalYear, r.Phase, r.Year, p.VP)
	}
}

func TestEEStatueCannotTriggerFinalYear(t *testing.T) {
	r := structureRoom()
	r.Phase = "winter"
	for _, p := range r.Players {
		p.NextWake = 1
		p.Wake = 1
	}
	p := r.Players[0]
	p.Buildings = []string{"statue"}
	p.VP = 19
	r.finishYear()
	if r.FinalYear || r.Phase != "wake" || p.VP != 20 {
		t.Fatalf("EE Statue incorrectly ended year: final=%v phase=%s vp=%d", r.FinalYear, r.Phase, p.VP)
	}
}

func TestStructureCompoundCostsAreAtomic(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"medium_cellar"}
	p.Wines = []Wine{{ID: "existing", Type: "red", Value: 2}}
	p.Grapes = []Grape{{ID: "g1", Color: "red", Value: 1}, {ID: "g2", Color: "red", Value: 2}}
	beforeGrapes, beforeWines := append([]Grape(nil), p.Grapes...), append([]Wine(nil), p.Wines...)
	if err := makeWines(p, [][]int{{0}, {1}}, 2); err == nil {
		t.Fatal("second bottle should fail because both target cellar slots are occupied")
	}
	if len(p.Grapes) != len(beforeGrapes) || len(p.Wines) != len(beforeWines) || p.Grapes[0].ID != beforeGrapes[0].ID || p.Wines[0].ID != beforeWines[0].ID {
		t.Fatalf("failed multi-bottle action mutated resources: grapes=%+v wines=%+v", p.Grapes, p.Wines)
	}
}

func TestStructureCompoundDiscardAndRestaurantAreAtomic(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Hand = []Card{{ID: "visitor-1", Type: "summer"}, {ID: "order-1", Type: "order"}}
	if err := r.visitorCards(p, []string{"visitor-1", "missing"}, 2, true); err == nil {
		t.Fatal("invalid second discard must fail")
	}
	if len(p.Hand) != 2 || len(r.Discards["summer"]) != 0 {
		t.Fatalf("invalid discard partially mutated state: hand=%+v discards=%+v", p.Hand, r.Discards)
	}
	p.Wines = []Wine{{ID: "wine-1", Type: "red", Value: 2}}
	p.Grapes = []Grape{{ID: "fruit-1", Color: "green", Value: 1}}
	if err := r.performStructureAction(p, mustStructure("ristorante"), Action{WineIDs: []string{"wine-1"}, GrapeID: "fruit-1"}); err == nil {
		t.Fatal("ristorante must reject a non-red/white grape")
	}
	if len(p.Wines) != 1 || len(p.Grapes) != 1 || p.Wines[0].ID != "wine-1" {
		t.Fatalf("failed restaurant action consumed a resource: %+v", p)
	}
}

func TestStructureDiscardKeepsCatalogueMetadataAndAutoMatSlot(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Coins = 20
	structureHand(p, "statue")
	if err := r.buildStructure(p, Action{Building: "structure-statue", Mode: "mat", StructureSlot: -1}, 0); err != nil {
		t.Fatal(err)
	}
	if len(p.StructureSlots) != 2 || p.StructureSlots[0] != "statue" {
		t.Fatalf("empty construction mat did not auto-select first slot: %+v", p.StructureSlots)
	}
	if err := r.destroyStructure(p, "statue"); err != nil {
		t.Fatal(err)
	}
	if len(r.Discards["structure"]) != 1 {
		t.Fatalf("structure was not discarded: %+v", r.Discards)
	}
	c := r.Discards["structure"][0]
	if c.StructureCost != 9 || c.Description == "" || c.RuleSource == "" || !c.Implemented {
		t.Fatalf("discarded structure lost catalogue metadata: %+v", c)
	}
}

func TestEEStructureDraftPassesFourRoundsAndResumesWake(t *testing.T) {
	r := &Room{Config: ExpansionConfig{Board: "ee", Structures: true, Visitors: "ee"}, Players: []*Player{NewPlayer("A"), NewPlayer("B")}, Phase: "setup"}
	r.HostID = r.Players[0].ID
	r.SpringLeader = 0
	r.Decks = map[string][]Card{"structure": {}}
	for i, c := range StructureCatalog() {
		c.ID = c.ID + "-" + string(rune('a'+i))
		r.Decks["structure"] = append(r.Decks["structure"], c)
	}
	if !r.beginStructureDraft() {
		t.Fatal("EE structure draft did not start")
	}
	if len(r.Choices) != 2 || r.StructureDraftRound != 1 {
		t.Fatalf("initial draft state: choices=%d round=%d", len(r.Choices), r.StructureDraftRound)
	}
	for r.Phase == "setup" {
		c := r.Choices[0]
		if err := r.resolveChoice(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: c.Options[0]}); err != nil {
			t.Fatal(err)
		}
	}
	if r.Phase != "wake" || r.StructureDraftRound != 0 || len(r.StructureDrafts) != 0 {
		t.Fatalf("draft did not resume game: phase=%s round=%d pools=%v", r.Phase, r.StructureDraftRound, r.StructureDrafts)
	}
	for _, p := range r.Players {
		if len(p.Hand) != 4 {
			t.Fatalf("player did not retain four drafted structures: %s has %d", p.Name, len(p.Hand))
		}
	}
}

func TestEEStructureDraftStartsAfterTheLastPapaChoice(t *testing.T) {
	r := &Room{Config: ExpansionConfig{Board: "ee", Structures: true, Visitors: "ee"}, Players: []*Player{NewPlayer("A"), NewPlayer("B")}, Phase: "setup"}
	for i, p := range r.Players {
		p.Papa = Card{ID: "papa-" + string(rune('1'+i))}
	}
	r.Decks = map[string][]Card{"structure": StructureCatalog()}
	r.Choices = []Choice{{ID: "p1", PlayerID: r.Players[0].ID, Kind: "papa", Options: []string{"gift", "coins"}}, {ID: "p2", PlayerID: r.Players[1].ID, Kind: "papa", Options: []string{"gift", "coins"}}}
	for _, c := range append([]Choice(nil), r.Choices...) {
		if err := r.resolveChoice(c.PlayerID, Action{Type: "choose", ChoiceID: c.ID, Option: "coins"}); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.Choices) != 2 || r.Choices[0].Kind != "structure_draft" {
		t.Fatalf("last Papa did not enter structure draft: %+v", r.Choices)
	}
}

func TestMercadoTriggersOnYearEndOrderDrawWithoutActionContext(t *testing.T) {
	r := structureRoom()
	p := r.Players[0]
	p.Buildings = []string{"mercado"}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 3}}
	p.Hand = []Card{{ID: "existing", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}}}
	r.Decks = map[string][]Card{"order": []Card{{ID: "drawn", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}}}}
	r.Discards = map[string][]Card{}
	r.Context = nil
	r.draw(p, "order")
	if len(r.Choices) != 1 || r.Choices[0].Kind != "structure_mercado" {
		t.Fatalf("Mercado did not queue after context-free draw: %+v", r.Choices)
	}
}

func TestMercadoYearEndChoiceResumesBeforeEEYearAdvance(t *testing.T) {
	r := &Room{Config: ExpansionConfig{Board: "ee", Structures: true, Visitors: "ee"}, Players: []*Player{NewPlayer("A"), NewPlayer("B")}, Phase: "year_end", Year: 1}
	p := r.Players[0]
	p.Buildings = []string{"dock", "mercado"}
	p.Hand = []Card{{ID: "existing", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}}}
	p.Wines = []Wine{{ID: "wine", Type: "red", Value: 3}}
	r.Decks = map[string][]Card{"order": []Card{{ID: "drawn", Type: "order", Requirements: []Wine{{Type: "red", Value: 2}}}}}
	r.Discards = map[string][]Card{}
	r.finishYear()
	if r.Year != 1 || r.StructureDraftRound != 0 || len(r.Choices) != 1 || r.Choices[0].Kind != "structure_mercado" {
		t.Fatalf("year advanced before Mercado resolution: year=%d choices=%+v", r.Year, r.Choices)
	}
	c := r.Choices[0]
	if err := r.resolveChoice(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "skip"}); err != nil {
		t.Fatal(err)
	}
	if r.Year != 2 || r.Phase != "wake" || len(r.Choices) != 0 {
		t.Fatalf("year-end did not resume after Mercado: year=%d phase=%s choices=%+v", r.Year, r.Phase, r.Choices)
	}
}
