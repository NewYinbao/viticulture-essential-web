package game

import "testing"

func specialWorkerRoom(ids ...string) *Room {
	r := tuscanyRoom(3)
	r.Config.SpecialWorkers = true
	r.SpecialWorkerPool = append([]string{}, ids...)
	r.Decks = map[string][]Card{
		"vine":  {{ID: "vine-a", Type: "vine"}, {ID: "vine-b", Type: "vine"}},
		"order": {{ID: "order-a", Type: "order"}}, "summer": {}, "winter": {},
	}
	for _, p := range r.Players {
		p.Season = "spring"
		p.Coins = 10
	}
	return r
}

func trainSpecialDirect(t *testing.T, r *Room, p *Player, id string) {
	t.Helper()
	if err := r.trainWorker(p, id, 0); err != nil {
		t.Fatal(err)
	}
	if !p.hasSpecialWorker(id) || p.SpecialWorkerReady[id] != r.Year+1 || p.Coins != 5 {
		t.Fatalf("special training state: %+v", p)
	}
}

func TestSpecialWorkerCatalogueAndTrainingChoice(t *testing.T) {
	if got := len(SpecialWorkerCatalog()); got != 11 {
		t.Fatal(got)
	}
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	r.Phase = "winter"
	if err := r.Apply(p.ID, Action{Type: "place", Space: "train", Slot: 1}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_train" {
		t.Fatalf("missing training choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "farmer"})
	if !p.hasSpecialWorker("farmer") || p.Coins != 6 || p.TotalWorkers != 4 || p.Workers != 1 {
		t.Fatalf("training choice not applied: %+v", p)
	}
	if err := r.trainWorker(p, "farmer", 0); err == nil {
		t.Fatal("duplicate special worker accepted")
	}
}

func TestSpecialWorkerFarmerAndIdentityPersist(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	trainSpecialDirect(t, r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	r.Phase = "spring"
	r.TurnID = p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 1, WorkerType: "farmer"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("missing Farmer choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "draw_vine"})
	if len(p.Hand) != 2 || !p.SpecialWorkerUsed["farmer"] {
		t.Fatalf("Farmer bonus/identity not applied: hand=%d used=%v", len(p.Hand), p.SpecialWorkerUsed)
	}
	q := restoreT(t, r)
	if q.Spaces[0].Occupied[0].WorkerType != "farmer" {
		t.Fatalf("seat identity not persisted: %+v", q.Spaces[0].Occupied)
	}
}

func TestSpecialWorkerChefBumpAndSoldatoToll(t *testing.T) {
	r := specialWorkerRoom("chef", "soldato")
	p, q := r.Players[0], r.Players[1]
	trainSpecialDirect(t, r, p, "chef")
	p.SpecialWorkerReady["chef"] = r.Year
	r.space("draw_vine").Occupied = []Seat{{PlayerID: q.ID, Slot: 1}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 1, WorkerType: "chef"}); err != nil {
		t.Fatal(err)
	}
	if len(r.space("draw_vine").Occupied) != 1 || r.space("draw_vine").Occupied[0].PlayerID != p.ID || q.Workers != 3 {
		t.Fatal("Chef did not bump and return regular worker")
	}

	r = specialWorkerRoom("soldato", "oracle")
	p, q = r.Players[0], r.Players[1]
	trainSpecialDirect(t, r, p, "soldato")
	p.SpecialWorkerReady["soldato"] = r.Year
	r.space("draw_vine").Occupied = []Seat{{PlayerID: p.ID, Slot: 1, WorkerType: "soldato"}}
	p.SpecialWorkerUsed["soldato"] = true
	q.Coins = 2
	r.TurnID = q.ID
	if err := r.Apply(q.ID, Action{Type: "place", Space: "draw_vine", Slot: 2}); err != nil {
		t.Fatal(err)
	}
	if q.Coins != 1 || p.Coins != 6 {
		t.Fatalf("Soldato toll not paid: q=%d p=%d", q.Coins, p.Coins)
	}
}

func TestSpecialWorkerOracleAndPoliticoChoices(t *testing.T) {
	r := specialWorkerRoom("oracle", "politico")
	p := r.Players[0]
	trainSpecialDirect(t, r, p, "oracle")
	p.SpecialWorkerReady["oracle"] = r.Year
	r.Phase = "spring"
	r.TurnID = p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 2, WorkerType: "oracle"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_oracle" || len(r.Choices[0].Options) != 2 {
		t.Fatalf("missing Oracle choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: r.Choices[0].Options[0]})
	if len(p.Hand) != 1 {
		t.Fatalf("Oracle did not keep one of two cards: %d", len(p.Hand))
	}

	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	trainSpecialDirect(t, r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase = "spring"
	r.TurnID = p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 2, WorkerType: "politico"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_politico" {
		t.Fatalf("missing Politico choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "repeat_bonus"})
	if p.Coins != 8 {
		t.Fatalf("Politico surcharge/bonus not applied: %d", p.Coins)
	}
}

func TestSpecialWorkerMafiosoProfessoreInnkeeperMerchantTraveler(t *testing.T) {
	r := specialWorkerRoom("mafioso", "professore")
	p := r.Players[0]
	trainSpecialDirect(t, r, p, "mafioso")
	p.SpecialWorkerReady["mafioso"] = r.Year
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "mafioso"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_mafioso" {
		t.Fatalf("missing Mafioso choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "repeat"})
	if p.Coins != 9 {
		t.Fatalf("Mafioso did not repeat a non-bonus action: %d", p.Coins)
	}

	r = specialWorkerRoom("professore", "oracle")
	p = r.Players[0]
	trainSpecialDirect(t, r, p, "professore")
	p.SpecialWorkerReady["professore"] = r.Year
	p.Workers = 1
	r.space("draw_vine").Occupied = []Seat{{PlayerID: p.ID, Slot: 1}}
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "professore"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_professore" {
		t.Fatalf("missing Professore choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "draw_vine:1"})
	if p.Workers != 2 {
		t.Fatalf("Professore did not return a regular worker: %d", p.Workers)
	}

	r = specialWorkerRoom("innkeeper", "merchant")
	p, q := r.Players[0], r.Players[1]
	trainSpecialDirect(t, r, p, "innkeeper")
	p.SpecialWorkerReady["innkeeper"] = r.Year
	q.Hand = []Card{{ID: "gift-summer", Type: "summer"}}
	r.Phase, r.TurnID = "spring", p.ID
	r.space("tour").Occupied = []Seat{{PlayerID: q.ID, Slot: 1}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 2, WorkerType: "innkeeper"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_innkeeper" {
		t.Fatalf("missing Innkeeper choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: q.ID + "|summer"})
	if p.Coins != 7 || q.Coins != 11 || len(p.Hand) != 1 {
		t.Fatalf("Innkeeper transfer wrong: p=%d q=%d hand=%d", p.Coins, q.Coins, len(p.Hand))
	}

	r = specialWorkerRoom("merchant", "oracle")
	p, q = r.Players[0], r.Players[1]
	trainSpecialDirect(t, r, p, "merchant")
	p.SpecialWorkerReady["merchant"] = r.Year
	for _, other := range r.Players[1:] {
		other.Season = "summer"
	}
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "merchant"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_merchant" {
		t.Fatalf("missing Merchant choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "order"})
	if len(p.Hand) != 1 || p.Hand[0].Type != "order" {
		t.Fatalf("Merchant did not draw selected type: %+v", p.Hand)
	}

	r = specialWorkerRoom("traveler", "oracle")
	p = r.Players[0]
	trainSpecialDirect(t, r, p, "traveler")
	p.SpecialWorkerReady["traveler"] = r.Year
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 1, WorkerType: "traveler"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 2 {
		t.Fatalf("Traveler did not immediately execute previous-season action: %d", len(p.Hand))
	}
}

func TestSpecialWorkerMessengerPersistsAndExecutesAtFutureSeason(t *testing.T) {
	r := specialWorkerRoom("messenger", "oracle")
	p := r.Players[0]
	trainSpecialDirect(t, r, p, "messenger")
	p.SpecialWorkerReady["messenger"] = r.Year
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "plant", Slot: 1, WorkerType: "messenger"}); err != nil {
		t.Fatal(err)
	}
	if len(r.MessengerPlans) != 1 || r.MessengerPlans[0].WorkerType != "messenger" || len(r.Choices) != 0 {
		t.Fatalf("Messenger was not reserved: %+v choices=%+v", r.MessengerPlans, r.Choices)
	}
	p.Hand = []Card{{ID: "reserved-vine", Type: "vine", Red: 1}}
	r.Phase = "summer"
	if !r.activateMessenger(p) || len(r.Choices) != 1 || r.Choices[0].Kind != "messenger" {
		t.Fatalf("Messenger did not execute on future season: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "execute", Space: "plant", CardID: "reserved-vine", Field: 0})
}

func TestSpecialWorkerUsedResetsAtTuscanyYearEnd(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	p.SpecialWorkers = []string{"farmer"}
	p.SpecialWorkerReady = map[string]int{"farmer": 1}
	p.SpecialWorkerUsed = map[string]bool{"farmer": true}
	p.NextWake = p.Wake
	for _, q := range r.Players {
		q.Passed = true
		q.NextWake = q.Wake
	}
	r.Phase = "winter"
	r.tuscanyFinishYear()
	if p.SpecialWorkerUsed["farmer"] {
		t.Fatal("special worker remained used across year")
	}
}

func readySpecial(r *Room, p *Player, id string) {
	if err := r.trainWorker(p, id, 0); err != nil {
		panic(err)
	}
	p.SpecialWorkerReady[id] = r.Year
}

func TestSpecialWorkerHiddenBonusAndOverflowRules(t *testing.T) {
	r := specialWorkerRoom("farmer", "traveler")
	p, q := r.Players[0], r.Players[1]
	readySpecial(r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	// In a 2-player game Farmer still sees the second printed bonus.
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "farmer"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("hidden bonus missing: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "coin"})
	if p.Coins != 8 {
		t.Fatalf("Farmer did not use hidden bonus: %d", p.Coins)
	}

	r = specialWorkerRoom("soldato", "oracle")
	p, q = r.Players[0], r.Players[1]
	readySpecial(r, p, "soldato")
	p.SpecialWorkerReady["soldato"] = r.Year
	r.space("draw_vine").Occupied = []Seat{{PlayerID: p.ID, Slot: 1, WorkerType: "soldato"}}
	r.Phase, r.TurnID = "spring", q.ID
	q.Coins = 2
	if err := r.Apply(q.ID, Action{Type: "place", Space: "draw_vine", WorkerType: "", Slot: 0}); err != nil {
		t.Fatal(err)
	}
	if len(r.space("draw_vine").Occupied) != 2 || q.Coins != 1 {
		t.Fatalf("Soldato full-action overflow failed: %+v coins=%d", r.space("draw_vine").Occupied, q.Coins)
	}

	r = specialWorkerRoom("traveler", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "traveler")
	p.SpecialWorkerReady["traveler"] = r.Year
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 2, WorkerType: "traveler"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 1 || r.space("draw_vine").Occupied[0].Slot != 2 {
		t.Fatalf("Traveler hidden slot failed: hand=%d seats=%+v", len(p.Hand), r.space("draw_vine").Occupied)
	}
}

func TestSpecialWorkerFarmerNormalizesEEHiddenBonus(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(2)
	p := r.Players[0]
	readySpecial(r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "farmer"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" || r.Choices[0].Options[0] != "coin" {
		t.Fatalf("EE Farmer exposed non-canonical reward keys: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "coin"})
	if p.Coins != 8 {
		t.Fatalf("EE Farmer hidden coin reward wrong: %d", p.Coins)
	}
}

func TestSpecialWorkerFarmerChoosesTuscanyAlternateBonus(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	p.Fields[0].Vines = []Card{{ID: "harvest-vine", Type: "vine", Red: 1}}
	r.Phase, r.TurnID = "fall", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "harvest", Slot: 2, WorkerType: "farmer", Fields: []int{0}}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("Farmer alternate bonus choice missing: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "coin"})
	if p.Coins != 6 || len(p.Grapes) != 1 || !p.Fields[0].Harvested {
		t.Fatalf("Farmer alternate coin reward wrong: coins=%d grapes=%d harvested=%v", p.Coins, len(p.Grapes), p.Fields[0].Harvested)
	}
}

func TestSpecialWorkerPendingChoiceSurvivesRestore(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	p.Fields[0].Vines = []Card{{ID: "persistent-vine", Type: "vine", Red: 1}}
	r.Phase, r.TurnID = "fall", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "harvest", Slot: 2, WorkerType: "farmer", Fields: []int{0}}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Context == nil || r.Context.PendingAction == nil {
		t.Fatalf("missing serializable Farmer continuation: choices=%+v context=%+v", r.Choices, r.Context)
	}
	q := restoreT(t, r)
	if len(q.Choices) != 1 || q.Choices[0].Kind != "special_farmer" || q.Context == nil || q.Context.PendingAction == nil {
		t.Fatalf("Farmer continuation did not survive restore: choices=%+v context=%+v", q.Choices, q.Context)
	}
	choiceT(t, q, Action{Option: "coin"})
	if q.Context != nil || q.Players[0].Coins != 6 || !q.Players[0].Fields[0].Harvested {
		t.Fatalf("restored Farmer continuation did not finish exactly once: coins=%d context=%+v harvested=%v", q.Players[0].Coins, q.Context, q.Players[0].Fields[0].Harvested)
	}
}

func politicoRepeat(t *testing.T, r *Room, p *Player, payload Action) {
	t.Helper()
	choiceT(t, r, Action{Option: "repeat_bonus"})
	if len(r.Choices) != 1 || r.Choices[0].SpecialBonus == "" {
		t.Fatalf("complex Politico choice missing: %+v", r.Choices)
	}
	payload.Option = "confirm"
	choiceT(t, r, payload)
}

func TestSpecialWorkerPoliticoAllComplexBonuses(t *testing.T) {
	// Plant, including the resource-bearing second choice and payment rollback.
	r := specialWorkerRoom("politico", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	p.Hand = []Card{{ID: "v1", Type: "vine", Red: 1}, {ID: "v2", Type: "vine", White: 1}}
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "plant", Slot: 1, WorkerType: "politico", CardIDs: []string{"v1"}, Fields: []int{0}}); err != nil {
		t.Fatal(err)
	}
	before := p.Coins
	politicoRepeat(t, r, p, Action{CardIDs: []string{"v2"}, Fields: []int{1}})
	if len(p.Hand) != 0 || len(p.Fields[0].Vines) != 1 || len(p.Fields[1].Vines) != 1 || p.Coins != before-1 {
		t.Fatalf("Politico plant wrong: hand=%+v fields=%+v coins=%d", p.Hand, p.Fields, p.Coins)
	}

	// Discount is a direct second refund after the completed build.
	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "build", Slot: 2, WorkerType: "politico", Building: "trellis"}); err != nil {
		t.Fatal(err)
	}
	choiceT(t, r, Action{Option: "repeat_bonus"})
	if p.Coins != 4 || !has(p, "trellis") {
		t.Fatalf("Politico discount wrong: coins=%d buildings=%v", p.Coins, p.Buildings)
	}

	// Harvest and make wine both use the ordinary multi-resource validators.
	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "fall", p.ID
	p.Fields[0].Vines = []Card{{ID: "f0", Type: "vine", Red: 1}}
	p.Fields[1].Vines = []Card{{ID: "f1", Type: "vine", White: 1}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "harvest", Slot: 1, WorkerType: "politico", Fields: []int{0}}); err != nil {
		t.Fatal(err)
	}
	politicoRepeat(t, r, p, Action{Fields: []int{1}})
	if !p.Fields[0].Harvested || !p.Fields[1].Harvested {
		t.Fatal("Politico harvest did not take both fields")
	}

	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "fall", p.ID
	p.Grapes = []Grape{{ID: "g0", Color: "red", Value: 2}, {ID: "g1", Color: "white", Value: 2}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "make_wine", Slot: 1, WorkerType: "politico", Recipes: [][]int{{0}}}); err != nil {
		t.Fatal(err)
	}
	politicoRepeat(t, r, p, Action{Recipes: [][]int{{0}}})
	if len(p.Wines) != 2 {
		t.Fatalf("Politico wine wrong: %+v", p.Wines)
	}

	// Build/tour chooses the repeated mode, while visitor starts a full
	// nested visitor continuation without re-queuing the outer Politico.
	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "fall", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "build_tour", Slot: 1, WorkerType: "politico", Mode: "tour"}); err != nil {
		t.Fatal(err)
	}
	politicoRepeat(t, r, p, Action{Mode: "tour"})
	if p.Coins != 10 {
		t.Fatalf("Politico build/tour wrong: %d", p.Coins)
	}

	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "summer", p.ID
	p.Hand = []Card{{ID: "summer-01", Type: "summer"}, {ID: "summer-02", Type: "summer"}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "summer_visitor", Slot: 2, WorkerType: "politico", CardID: "summer-01"}); err != nil {
		t.Fatal(err)
	}
	choiceT(t, r, Action{Option: "vp"})
	choiceT(t, r, Action{Option: "skip"})
	politicoRepeat(t, r, p, Action{Color: "summer", CardID: "summer-02"})
	choiceT(t, r, Action{Option: "sell"})
	choiceT(t, r, Action{Option: "skip"})
	if p.Coins != 10 || p.VP != -2 {
		t.Fatalf("Politico visitor nested continuation wrong: coins=%d vp=%d choices=%+v", p.Coins, p.VP, r.Choices)
	}
}

func TestSpecialWorkerPoliticoInfluenceAndTradeBranches(t *testing.T) {
	r := specialWorkerRoom("politico", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "influence", Slot: 2, WorkerType: "politico", Influence: []InfluenceMove{{To: "firenze"}}}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "tuscany_influence" {
		t.Fatalf("base influence bonus missing: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "place", Influence: []InfluenceMove{{To: "firenze"}}})
	politicoRepeat(t, r, p, Action{Influence: []InfluenceMove{{To: "pisa"}}})
	if p.Influence["firenze"] != 2 || p.Influence["pisa"] != 1 || p.Coins != 6 {
		t.Fatalf("Politico influence wrong: influence=%v coins=%d", p.Influence, p.Coins)
	}

	r = specialWorkerRoom("politico", "oracle")
	p = r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	p.Grapes = []Grape{{ID: "g", Color: "red", Value: 1}}
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "trade", Slot: 2, WorkerType: "politico", Trades: []Trade{{Give: "grape", GrapeID: "g", Receive: "coins"}}}); err != nil {
		t.Fatal(err)
	}
	choiceT(t, r, Action{Option: "trade", Trades: []Trade{{Give: "coins", Receive: "grape", Color: "red"}}})
	politicoRepeat(t, r, p, Action{Trades: []Trade{{Give: "grape", GrapeID: p.Grapes[0].ID, Receive: "coins"}}})
	if p.Coins != 7 || len(p.Grapes) != 0 {
		t.Fatalf("Politico trade wrong: coins=%d grapes=%+v", p.Coins, p.Grapes)
	}
}

func TestSpecialWorkerPoliticoWorksWithEEBoard(t *testing.T) {
	r := specialWorkerRoom("politico", "oracle")
	r.Config.Board = "ee"
	r.Spaces = NewSpaces(3)
	p := r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "summer", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", Slot: 1, WorkerType: "politico"}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].SpecialBonus != "coin" {
		t.Fatalf("EE Politico bonus missing: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "repeat_bonus"})
	if p.Coins != 8 {
		t.Fatalf("EE Politico did not charge and repeat coin reward: %d", p.Coins)
	}
}

func TestSpecialWorkerDeclinedBonusDoesNotDiscountOrEnablePolitico(t *testing.T) {
	r := specialWorkerRoom("politico", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "politico")
	p.SpecialWorkerReady["politico"] = r.Year
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "build", Slot: 1, WorkerType: "politico", Building: "trellis", DeclineBonus: true}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 0 || p.Coins != 3 || !has(p, "trellis") {
		t.Fatalf("declined reward still affected build/Politico: choices=%+v coins=%d buildings=%v", r.Choices, p.Coins, p.Buildings)
	}

	r = specialWorkerRoom("farmer", "oracle")
	p = r.Players[0]
	p.Coins = 10
	r.Phase, r.TurnID = "winter", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "train", Slot: 1, DeclineBonus: true}); err != nil {
		t.Fatal(err)
	}
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_train" {
		t.Fatalf("missing training choice after declining discount: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "farmer"})
	if p.Coins != 5 || !p.hasSpecialWorker("farmer") {
		t.Fatalf("declined training reward still discounted: coins=%d workers=%v", p.Coins, p.SpecialWorkers)
	}
}

func TestSpecialWorkerOracleCountsTuscanyPrintedDraw(t *testing.T) {
	r := specialWorkerRoom("oracle", "farmer")
	p := r.Players[0]
	readySpecial(r, p, "oracle")
	p.SpecialWorkerReady["oracle"] = r.Year
	r.Decks["vine"] = []Card{{ID: "v1", Type: "vine"}, {ID: "v2", Type: "vine"}, {ID: "v3", Type: "vine"}}
	r.Phase, r.TurnID = "spring", p.ID
	if err := r.Apply(p.ID, Action{Type: "place", Space: "draw_vine", Slot: 1, WorkerType: "oracle"}); err != nil {
		t.Fatal(err)
	}
	if len(p.Hand) != 3 || len(r.Choices) != 1 || r.Choices[0].Kind != "special_oracle" || len(r.Choices[0].Options) != 3 {
		t.Fatalf("Oracle did not include Tuscany printed draw: hand=%d choices=%+v", len(p.Hand), r.Choices)
	}
	choiceT(t, r, Action{Option: r.Choices[0].Options[0]})
	if len(p.Hand) != 2 {
		t.Fatalf("Oracle discarded the wrong number after printed draw: hand=%d", len(p.Hand))
	}
}

func TestSpecialWorkerMerchantRecognizesWinterYearEnd(t *testing.T) {
	r := specialWorkerRoom("merchant", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "merchant")
	p.SpecialWorkerReady["merchant"] = r.Year
	r.Phase = "winter"
	for _, q := range r.Players[1:] {
		q.Season = "year_end"
	}
	if !r.merchantEligible(p) {
		t.Fatal("Merchant rejected opponents who already passed winter")
	}
	r.Players[1].Season = "winter"
	if r.merchantEligible(p) {
		t.Fatal("Merchant accepted an opponent still in winter")
	}
}

func TestSpecialWorkerPlannerContinuationSurvivesPreChoice(t *testing.T) {
	r := specialWorkerRoom("farmer", "oracle")
	p := r.Players[0]
	readySpecial(r, p, "farmer")
	p.SpecialWorkerReady["farmer"] = r.Year
	p.SpecialWorkerUsed["farmer"] = true
	r.Decks["order"] = []Card{{ID: "o1", Type: "order"}, {ID: "o2", Type: "order"}}
	r.Phase = "fall"
	r.space("draw_order").Occupied = []Seat{{PlayerID: p.ID, Slot: 1, WorkerType: "farmer"}}
	r.Planned = []PlannedPlacement{{PlayerID: p.ID, Space: "draw_order", Slot: 1, WorkerType: "farmer"}}
	r.startNextPlanner()
	if len(r.Choices) != 1 || r.Choices[0].Kind != "planner" {
		t.Fatalf("missing planner choice: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "execute"})
	if len(r.Choices) != 1 || r.Choices[0].Kind != "special_farmer" {
		t.Fatalf("missing Farmer choice in planner continuation: %+v", r.Choices)
	}
	choiceT(t, r, Action{Option: "draw_order"})
	if r.Context != nil || len(r.Planned) != 0 || len(r.Choices) != 0 || len(p.Hand) != 2 {
		t.Fatalf("planner did not resume after special choice: context=%+v planned=%+v choices=%+v hand=%d", r.Context, r.Planned, r.Choices, len(p.Hand))
	}
}
