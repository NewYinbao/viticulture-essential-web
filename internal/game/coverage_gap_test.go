package game

import "testing"

func TestStructureDirectEffectsCoverage(t *testing.T) {
	t.Run("aqueduct ignores both planting prerequisites", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"aqueduct"}
		p.Hand = []Card{{ID: "vine", Type: "vine", Red: 1, Trellis: true, Irrigation: true}}
		if err := r.perform(p, Action{Space: "plant", CardID: "vine", Field: 0}); err != nil {
			t.Fatal(err)
		}
		if len(p.Hand) != 0 || len(p.Fields[0].Vines) != 1 {
			t.Fatalf("aqueduct did not permit the planted vine: %+v", p.Fields[0])
		}
	})

	t.Run("trading post performs a trade and queues influence", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"trading_post"}
		if err := r.performStructureAction(p, mustStructure("trading_post"), Action{Trades: []Trade{{Give: "coins", Receive: "vp"}}}); err != nil {
			t.Fatal(err)
		}
		if p.Coins != 17 || p.VP != 1 || len(r.Choices) != 1 || r.Choices[0].Kind != "structure_influence" {
			t.Fatalf("trading post result mismatch: player=%+v choices=%+v", p, r.Choices)
		}
	})

	t.Run("shop fills an order and queues its influence reward", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"shop"}
		p.Hand = []Card{{ID: "order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
		p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
		if err := r.performStructureAction(p, mustStructure("shop"), Action{CardID: "order", WineIDs: []string{"wine"}}); err != nil {
			t.Fatal(err)
		}
		if len(p.Hand) != 0 || len(p.Wines) != 0 || p.VP != 2 || p.Income != 1 || len(r.Choices) != 1 || r.Choices[0].Kind != "structure_influence" {
			t.Fatalf("shop order/reward mismatch: player=%+v choices=%+v", p, r.Choices)
		}
	})

	t.Run("ristorante consumes exact resources and grants its reward", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"ristorante"}
		p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
		p.Grapes = []Grape{{ID: "grape", Color: "white", Value: 1}}
		if err := r.performStructureAction(p, mustStructure("ristorante"), Action{WineIDs: []string{"wine"}, GrapeID: "grape"}); err != nil {
			t.Fatal(err)
		}
		if len(p.Wines) != 0 || len(p.Grapes) != 0 || p.Coins != 23 || p.VP != 3 {
			t.Fatalf("ristorante result mismatch: %+v", p)
		}
	})

	t.Run("guest house discards two visitors and grants two points", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"guest_house"}
		p.Hand = []Card{{ID: "summer", Type: "summer"}, {ID: "winter", Type: "winter"}}
		if err := r.performStructureAction(p, mustStructure("guest_house"), Action{CardIDs: []string{"summer", "winter"}}); err != nil {
			t.Fatal(err)
		}
		if len(p.Hand) != 0 || p.VP != 2 || len(r.Discards["summer"]) != 1 || len(r.Discards["winter"]) != 1 {
			t.Fatalf("guest house result mismatch: player=%+v discards=%+v", p, r.Discards)
		}
	})

	t.Run("veranda and wine parlor enhance a normal delivery", func(t *testing.T) {
		r := structureRoom()
		p := r.Players[0]
		p.Buildings = []string{"veranda", "wine_parlor"}
		p.Hand = []Card{{ID: "order", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}}
		p.Wines = []Wine{{ID: "wine", Type: "red", Value: 1}}
		coins := p.Coins
		if err := r.fillOrderAction(p, Action{CardID: "order", WineIDs: []string{"wine"}}); err != nil {
			t.Fatal(err)
		}
		if p.VP != 3 || p.Coins != coins+2 || p.Income != 1 {
			t.Fatalf("delivery enhancements mismatch: %+v", p)
		}
	})

	t.Run("fountain pays only the opponent owner on a tour", func(t *testing.T) {
		r := structureRoom()
		actor, owner := r.Players[0], r.Players[1]
		owner.Buildings = []string{"fountain"}
		actorCoins, ownerCoins := actor.Coins, owner.Coins
		if err := r.perform(actor, Action{Space: "tour"}); err != nil {
			t.Fatal(err)
		}
		if actor.Coins != actorCoins+2 || owner.Coins != ownerCoins+1 {
			t.Fatalf("fountain tour payment mismatch: actor=%d owner=%d", actor.Coins, owner.Coins)
		}
	})
}
