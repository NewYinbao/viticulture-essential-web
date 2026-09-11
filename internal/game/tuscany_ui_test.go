package game

import (
	"encoding/json"
	"testing"
)

func TestTuscanyViewPhysicalSlotAndSeasonHints(t *testing.T) {
	r := tuscanyRoom(3)
	p := r.Players[0]
	p.Coins = 1
	before, _ := json.Marshal(r)
	actions, _, _ := r.availability(p.ID)
	if actions["build"] != "" {
		t.Fatal("discount on second spring slot must remain reachable", actions)
	}
	r.Phase = "fall"
	actions, _, _ = r.availability(p.ID)
	if actions["harvest"] == "" {
		t.Fatal("fall hints missing")
	}
	r.Phase = "spring"
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("read-only hints mutated room")
	}
}

func TestTuscanyViewCoinBeforeVisitorHintsAndPrivacy(t *testing.T) {
	r := tuscanyRoom(2)
	r.Phase = "winter"
	p := r.Players[0]
	p.Coins = 0
	p.Income = 0
	p.Hand = []Card{visitorCardT(t, "winter-34")}
	actions, cards, _ := r.availability(p.ID)
	if actions["winter_visitor"] != "" || cards["winter-34"] != "" {
		t.Fatal("coin-first visitor blocked", actions, cards)
	}
	v := r.View(p.ID)
	h := v["visitorCardReasons"].(map[string]map[string]string)
	if h["base"]["winter-34"] == "" || h["coinBefore"]["winter-34"] != "" {
		t.Fatal("slot timing not distinguished", h)
	}
	peer := r.View(r.Players[1].ID)
	if _, ok := peer["visitorCardReasons"]; ok {
		t.Fatal("private card hints exposed to waiting peer")
	}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "winter_visitor" {
			r.Spaces[i].Occupied = []Seat{{PlayerID: r.Players[1].ID, Slot: 1}}
		}
	}
	actions, cards, _ = r.availability(p.ID)
	if actions["winter_visitor"] == "" || cards["winter-34"] == "" {
		t.Fatal("overflow borrowed occupied coin reward")
	}
}
