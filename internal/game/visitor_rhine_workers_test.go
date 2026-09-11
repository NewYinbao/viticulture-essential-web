package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRhineTrainerTrainingAndPermanentAvailableLoss(t *testing.T) {
	for _, kind := range []string{"regular", "grande", "soldato"} {
		t.Run(kind, func(t *testing.T) {
			r, p := rhineFixture(t, "rhine-winter-trainer", func(r *Room, p *Player) {
				r.Config.SpecialWorkers = true
				r.SpecialWorkerPool = []string{"soldato"}
				if kind == "soldato" {
					p.SpecialWorkers = []string{"soldato"}
					p.TotalWorkers++
					p.ensureSpecialMaps()
				}
				r.claimGrayWorker(p)
			})
			before := p.TotalWorkers
			workers := p.Workers
			rhineChoose(t, r, "lose", Action{WorkerID: "available:" + kind})
			if p.TotalWorkers != before-1 || p.VP != 12 || !p.GrayWorkerAvailable {
				t.Fatal("permanent cost or gray identity lost")
			}
			if kind == "regular" && p.Workers != workers-1 {
				t.Fatal("regular cost")
			}
			if kind == "grande" && (!p.GrandeRemoved || p.LargeWorker) {
				t.Fatal("grande cost")
			}
			if kind == "soldato" && p.hasSpecialWorker(kind) {
				t.Fatal("special cost")
			}
		})
	}
	r, p := rhineFixture(t, "rhine-winter-trainer", func(r *Room, p *Player) {
		p.TotalWorkers = 1
		p.Workers = 0
		p.LargeWorker = false
		p.GrandeRemoved = true
		r.claimGrayWorker(p)
	})
	c := r.Choices[0]
	before := p.TotalWorkers
	if r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "lose", WorkerID: "available:regular"}) == nil || p.TotalWorkers != before || !p.GrayWorkerAvailable {
		t.Fatal("gray was sacrificed as permanent")
	}
	rhineChoose(t, r, "train", Action{SpecialWorker: "grande"})
	if p.GrandeRemoved || !p.LargeWorker || p.Coins != 47 {
		t.Fatal("Trainer must train grande for3 immediately")
	}
}

func TestRhineTutorPlacedWorkerCancelsReservationsAndExchangeSurvivesReload(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-tutor", func(r *Room, p *Player) {
		p.TotalWorkers = 4
		p.SpecialWorkers = []string{"messenger"}
		p.ensureSpecialMaps()
		p.SpecialWorkerUsed["messenger"] = true
		seat := Seat{PlayerID: p.ID, Slot: 1, WorkerType: "messenger"}
		r.space("draw_order").Occupied = []Seat{seat}
		r.Planned = []PlannedPlacement{{PlayerID: p.ID, Space: "draw_order", Slot: 1, WorkerType: "messenger"}}
		r.MessengerPlans = append([]PlannedPlacement{}, r.Planned...)
	})
	rhineChoose(t, r, "exchange", Action{WorkerID: "placed:draw_order:0"})
	if p.TotalWorkers != 3 || p.hasSpecialWorker("messenger") || len(r.space("draw_order").Occupied) != 0 || len(r.Planned) != 0 || len(r.MessengerPlans) != 0 {
		t.Fatal("Tutor did not open slot/cancel exact reservation")
	}
	rhineChoose(t, r, "continue", Action{CardIDs: []string{"v0", "v1"}})
	b, _ := json.Marshal(r)
	var restored Room
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatal(err)
	}
	r = &restored
	p = r.Player(p.ID)
	if len(r.Choices) != 1 || r.Choices[0].Visitor.Rhine.Values["exchangeCount"] != "2" {
		t.Fatal("private exchange count not persisted")
	}
	before := len(p.Hand)
	rhineChoose(t, r, "continue", Action{Colors: []string{"summer", "order"}})
	if len(p.Hand) != before+2 || len(r.Choices) != 0 {
		t.Fatal("Tutor did not draw exact chosen types")
	}
}

func TestRhineTutorMayLoseItsTriggerGrandeAndRetrainWithoutResurrection(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-tutor", func(r *Room, p *Player) {
		p.LargeWorker = false
		r.space("winter_visitor").Occupied = []Seat{{PlayerID: p.ID, Slot: 1, Large: true}}
	})
	seat := r.space("winter_visitor").Occupied[0]
	r.Context.TriggerSeat = &seat
	r.Context.TriggerSpace = "winter_visitor"
	rhineChoose(t, r, "exchange", Action{WorkerID: "placed:winter_visitor:0"})
	rhineChoose(t, r, "continue", Action{})
	rhineChoose(t, r, "continue", Action{})
	if !p.GrandeRemoved || len(r.space("winter_visitor").Occupied) != 0 {
		t.Fatal("Tutor cannot remove triggering grande")
	}
	p.resetAnnualWorkers()
	if p.LargeWorker || p.Workers != 2 {
		t.Fatal("yearly reset resurrected lost grande")
	}
	coins := p.Coins
	if err := r.trainWorker(p, "grande", 0); err != nil {
		t.Fatal(err)
	}
	if p.GrandeRemoved || p.LargeWorker || p.Coins != coins-4 {
		t.Fatal("grande supply/cost/readiness")
	}
	p.resetAnnualWorkers()
	if !p.LargeWorker || p.Workers != 2 {
		t.Fatal("retrained grande not returned nextyear")
	}
}

func TestRhineWorkerLossDoesNotGuessLegacyGrayAndLimitsPhysicalSupply(t *testing.T) {
	r, p := visitorFixture("summer-01")
	r.GrayWorkerOwner = p.ID
	r.GrayWorkerTracked = false
	p.Workers++
	for _, c := range r.permanentWorkerLossChoices(p, false) {
		if strings.HasSuffix(c.ID, ":regular") {
			t.Fatal("ambiguous legacy gray guessed")
		}
	}
	p.GrandeRemoved = true
	p.LargeWorker = false
	p.TotalWorkers = 5
	p.Workers = 5
	r.GrayWorkerOwner = ""
	if err := r.trainWorker(p, "regular", 0); err == nil {
		t.Fatal("created sixth non-grande")
	}
	if err := r.trainWorker(p, "grande", 0); err != nil {
		t.Fatal(err)
	}
}

func TestRhineGrayIdentityPersistsAcrossPlacementAndRetrieval(t *testing.T) {
	r, p := visitorFixture("summer-01")
	r.Config.Visitors = "rhine"
	r.Context = nil
	r.Choices = nil
	r.Phase = "summer"
	r.TurnID = p.ID
	r.claimGrayWorker(p)
	permanent := p.TotalWorkers
	regular := p.availableRegularWorkers()
	if err := r.Apply(p.ID, Action{Type: "place", Space: "gain_coin", Gray: true}); err != nil {
		t.Fatal(err)
	}
	seat := r.space("gain_coin").Occupied[0]
	if !seat.Gray || p.GrayWorkerAvailable || p.availableRegularWorkers() != regular {
		t.Fatal("placement consumed wrong identity")
	}
	b, _ := json.Marshal(r)
	var restored Room
	_ = json.Unmarshal(b, &restored)
	r = &restored
	p = r.Player(p.ID)
	r.releaseSeat(seat)
	if !p.GrayWorkerAvailable || p.TotalWorkers != permanent || p.availableRegularWorkers() != regular {
		t.Fatal("retrieval converted gray into permanent")
	}
	for _, c := range r.permanentWorkerLossChoices(p, false) {
		if c.Seat != nil && c.Seat.Gray {
			t.Fatal("gray listed for sacrifice")
		}
	}
}

func TestRhineTutorTrainingPaysPointAndAcademy(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-tutor", func(r *Room, p *Player) {
		r.Config.SpecialWorkers = true
		r.Config.Structures = true
		r.SpecialWorkerPool = []string{"chef"}
		r.Players[1].Buildings = append(r.Players[1].Buildings, "academy")
	})
	q := r.Players[1]
	coins := q.Coins
	rhineChoose(t, r, "train", Action{SpecialWorker: "chef"})
	if p.Coins != 47 || p.VP != 9 || !p.hasSpecialWorker("chef") || p.specialWorkerReady("chef", r.Year) || q.Coins != coins+1 {
		t.Fatal("Tutor special+Academy price/point/readiness")
	}
}
