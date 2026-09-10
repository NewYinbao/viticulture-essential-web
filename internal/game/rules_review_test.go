package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPapaChoicesStartAtFirstPlayerAndWrap(t *testing.T) {
	for leader := 0; leader < 6; leader++ {
		r := eeRoom(6)
		r.InitDecks()
		r.SpringLeader = leader
		r.setupParents()
		for i, c := range r.Choices {
			if want := r.Players[(leader+i)%6].ID; c.PlayerID != want {
				t.Fatalf("leader %d, choice %d: got %s, want %s", leader, i, c.PlayerID, want)
			}
		}
		if r.TurnID != r.Players[leader].ID {
			t.Fatal("first player must choose first")
		}
	}
}

func TestPublicHandTypesWithoutPrivateCardIdentities(t *testing.T) {
	r := eeRoom(2)
	other := r.Players[1]
	other.Hand = []Card{{ID: "secret-vine-a", Type: "vine"}, {ID: "secret-vine-b", Type: "vine"}, {ID: "secret-visitor", Type: "winter"}}
	v := r.View(r.Players[0].ID)
	p := v["players"].([]Player)[1]
	if p.HandCount != 3 || p.HandCounts["vine"] != 2 || p.HandCounts["winter"] != 1 || p.HandCounts["order"] != 0 {
		t.Fatalf("incorrect public counts: %+v", p.HandCounts)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "secret-") {
		t.Fatal("opponent card identity leaked")
	}
	if other.HandCounts != nil || len(other.Hand) != 3 {
		t.Fatal("view mutated authoritative hand")
	}
}
