package game

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAvailabilityIsPrivateAndDoesNotMutate(t *testing.T) {
	r := eeBegin(t, 3)
	p := r.Player(r.TurnID)
	p.Hand = []Card{{ID: "summer-04", Type: "summer", Implemented: true}, {ID: "private-vine", Type: "vine", Trellis: true, Red: 2}}
	p.Coins = 0
	p.Buildings = nil
	before, _ := json.Marshal(r)
	v := r.View(p.ID)
	if !strings.Contains(v["cardReasons"].(map[string]string)["private-vine"], "棚架") {
		t.Fatal("missing planting prerequisite")
	}
	for _, other := range r.Players {
		if other.ID != p.ID {
			o := r.View(other.ID)
			if len(o["cardReasons"].(map[string]string)) != 0 || len(o["optionReasons"].(map[string]string)) != 0 {
				t.Fatal("private prerequisite leak")
			}
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("hints changed authoritative state")
	}
}

func TestVisitorPrerequisitesDoNotBlockFreeOrOptionalEffects(t *testing.T) {
	r := eeRoom(2)
	p := r.Players[0]
	p.Coins = 0
	p.VP = -5
	p.Hand = nil
	p.Grapes = nil
	p.Wines = nil
	tests := []struct {
		id, stage, option string
		blocked           bool
	}{
		{"summer-02", "effect", "buy", true}, {"summer-02", "effect", "sell", true},
		{"summer-07", "effect", "make", true}, {"summer-07", "effect", "coins", false},
		{"summer-11", "plant", "plant", false}, {"winter-10", "effect", "coins", false},
		{"winter-13", "effect", "vp", true}, {"winter-04", "effect", "age", false},
	}
	for _, tc := range tests {
		if got := r.visitorOptionReason(p, &VisitorStep{CardID: tc.id, Stage: tc.stage}, tc.option); (got != "") != tc.blocked {
			t.Errorf("%+v: %q", tc, got)
		}
	}
}

func TestActionAffordabilityUsesAvailableBonusSlot(t *testing.T) {
	r := eeBegin(t, 3)
	r.Phase = "winter"
	p := r.Player(r.TurnID)
	p.Coins = 3
	a, _, _ := r.availability(p.ID)
	if a["train"] != "" {
		t.Fatal("3 coins must cover bonus training", a["train"])
	}
	for i := range r.Spaces {
		if r.Spaces[i].ID == "train" {
			r.Spaces[i].Occupied = []Seat{{Slot: 1, PlayerID: r.Players[0].ID}}
		}
	}
	a, _, _ = r.availability(p.ID)
	if !strings.Contains(a["train"], "1 金币") {
		t.Fatal("occupied bonus must not discount training", a["train"])
	}
}
