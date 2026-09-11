package game

import "testing"

func TestFarmerTrainingDiscountOffPrintedBonus(t *testing.T) {
	for _, coins := range []int{3, 4} {
		r := specialWorkerRoom("farmer", "chef")
		p := r.Players[0]
		readySpecial(r, p, "farmer")
		r.Phase, p.Season, p.Coins = "winter", "winter", coins
		p.Workers, p.LargeWorker = 0, false
		r.space("train").Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1}}
		reasons, _, _ := r.availability(p.ID)
		if reasons["train"] != "" {
			t.Fatalf("Farmer discount hidden: %s", reasons["train"])
		}
		if err := r.Apply(p.ID, Action{Type: "place", Space: "train", Slot: 2, WorkerType: "farmer", SpecialWorker: "regular"}); err != nil {
			t.Fatal(err)
		}
		if r.Choices[0].Kind != "special_farmer" {
			t.Fatal("missing Farmer reward choice")
		}
		if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "discount"}); err != nil {
			t.Fatal(err)
		}
		if p.Coins != coins-3 {
			t.Fatalf("Farmer paid %d, expected 3", coins-p.Coins)
		}
	}
}

func TestFarmerTrainingClassChoiceRetainsRewardChoice(t *testing.T) {
	r := specialWorkerRoom("farmer", "chef")
	p := r.Players[0]
	readySpecial(r, p, "farmer")
	r.Phase, p.Season, p.Coins = "winter", "winter", 5
	r.space("train").Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1}}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "train", Slot: 2, WorkerType: "farmer"}); err != nil {
		t.Fatal(err)
	}
	if r.Choices[0].Kind != "special_train" {
		t.Fatal("missing training class selection")
	}
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "chef"}); err != nil {
		t.Fatal(err)
	}
	if r.Choices[0].Kind != "special_farmer" {
		t.Fatal("training class selection skipped Farmer reward")
	}
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "discount"}); err != nil {
		t.Fatal(err)
	}
	if p.Coins != 1 || !p.hasSpecialWorker("chef") || p.SpecialWorkerReady["chef"] != r.Year+1 {
		t.Fatal("special training lost discount or identity")
	}
}

func TestTrainPreviewIncludesSoldatoFee(t *testing.T) {
	for _, coins := range []int{4, 5} {
		r := specialWorkerRoom("chef", "soldato")
		p := r.Players[0]
		r.Phase, p.Season, p.Coins = "winter", "winter", coins
		r.space("train").Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1, WorkerType: "soldato"}}
		for _, option := range r.workerPlacements(p, r.space("train")) {
			if option.Toll != 1 {
				t.Fatalf("missing entry fee: %+v", option)
			}
		}
		reasons, _, _ := r.availability(p.ID)
		if (reasons["train"] == "") != (coins == 5) {
			t.Fatalf("coins=%d wrong training reason: %s", coins, reasons["train"])
		}
		if err := r.Apply(p.ID, Action{Type: "place", Space: "train", Slot: 2, SpecialWorker: "regular"}); (err == nil) != (coins == 5) {
			t.Fatalf("coins=%d engine mismatch: %v", coins, err)
		}
	}
}
