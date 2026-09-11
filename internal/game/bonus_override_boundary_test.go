package game

import (
	"encoding/json"
	"testing"
)

func clientAction(t *testing.T, body string) Action {
	t.Helper()
	var a Action
	if err := json.Unmarshal([]byte(body), &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestApplyIgnoresClientBonusOverride(t *testing.T) {
	t.Run("two player EE tour", func(t *testing.T) {
		r := eeBegin(t, 2)
		p := r.Player(r.TurnID)
		p.Coins = 0
		if err := r.Apply(p.ID, clientAction(t, `{"type":"place","space":"tour","slot":1,"bonusOverride":"coin"}`)); err != nil {
			t.Fatal(err)
		}
		if p.Coins != 2 {
			t.Fatalf("client forged the hidden two-player tour reward: coins=%d", p.Coins)
		}
	})

	t.Run("two player EE training", func(t *testing.T) {
		r := eeBegin(t, 2)
		p := r.Players[0]
		r.Phase, r.TurnID = "winter", p.ID
		p.Coins = 4
		if err := r.Apply(p.ID, clientAction(t, `{"type":"place","space":"train","slot":1,"bonusOverride":"discount"}`)); err != nil {
			t.Fatal(err)
		}
		if p.Coins != 0 {
			t.Fatalf("client forged the hidden two-player training discount: coins=%d", p.Coins)
		}
	})

	t.Run("Messenger continuation", func(t *testing.T) {
		r := specialWorkerRoom("messenger")
		r.Config.Board = "ee"
		r.Spaces = NewSpaces(3)
		p := r.Players[0]
		r.Phase, r.TurnID = "winter", p.ID
		p.Coins = 4
		r.MessengerPlans = []PlannedPlacement{{PlayerID: p.ID, Space: "train", Slot: 2, WorkerType: "messenger"}}
		r.space("train").Occupied = []Seat{{Season: "summer", PlayerID: p.ID, Slot: 2, WorkerType: "messenger"}}
		if !r.activateMessenger(p) || len(r.Choices) != 1 {
			t.Fatalf("Messenger continuation was not activated: %+v", r.Choices)
		}
		choiceID := r.Choices[0].ID
		body := `{"type":"choose","choiceId":"` + choiceID + `","option":"execute","space":"train","bonusOverride":"discount"}`
		if err := r.Apply(p.ID, clientAction(t, body)); err != nil {
			t.Fatal(err)
		}
		if p.Coins != 0 {
			t.Fatalf("client forged a Messenger training discount: coins=%d", p.Coins)
		}
	})

	t.Run("two player Tuscany non-bonus slot", func(t *testing.T) {
		r := tuscanyRoom(2)
		p := r.Players[0]
		r.Phase, p.Season, r.TurnID = "spring", "spring", p.ID
		p.Coins = 0
		if err := r.Apply(p.ID, clientAction(t, `{"type":"place","space":"tour","slot":1,"bonusOverride":"coin"}`)); err != nil {
			t.Fatal(err)
		}
		if p.Coins != 2 {
			t.Fatalf("client forged a Tuscany non-bonus-slot reward: coins=%d", p.Coins)
		}
	})
}
