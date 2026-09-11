package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func rejectRhineUnchanged(t *testing.T, r *Room, p *Player, a Action) {
	t.Helper()
	a.Type, a.ChoiceID = "choose", r.Choices[0].ID
	before, _ := json.Marshal(r)
	if err := r.Apply(p.ID, a); err == nil {
		t.Fatal("zero-effect visitor accepted")
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected choice changed room")
	}
}

func TestRhineLobbyistNeedsOneActualStar(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-lobbyist", func(r *Room, p *Player) {
		r.Config.Board = "tuscany"
		r.Players[1].Influence = map[string]int{"pisa": 1}
	})
	rejectRhineUnchanged(t, r, p, Action{Option: "resolve"})
	for _, q := range r.Players {
		q.Influence = nil
	}
	if r.rhineOptionReason(p, r.Choices[0].Visitor, "resolve") == "" {
		t.Fatal("empty map should not advertise a playable Lobbyist")
	}
}

func TestRhineEstateAgentNeedsAnOwnedField(t *testing.T) {
	r, p := rhineFixture(t, "rhine-winter-estateAgent", nil)
	for i := range p.Fields {
		p.Fields[i].Sold = true
	}
	if r.rhineOptionReason(p, r.Choices[0].Visitor, "draw") == "" {
		t.Fatal("all fields sold should make Estate Agent unavailable")
	}
	rejectRhineUnchanged(t, r, p, Action{Option: "draw"})
}

func TestRhineVineTraderMustSellOrBuyButMayDoEither(t *testing.T) {
	for _, sell := range []bool{false, true} {
		r, p := rhineFixture(t, "rhine-summer-vineTrader", nil)
		a := Action{Option: "trade"}
		if sell {
			a.CardIDs = []string{"v0"}
		}
		moorChoose(t, r, a)
		if sell {
			moorChoose(t, r, Action{Option: "continue"})
		} else {
			rejectRhineUnchanged(t, r, p, Action{Option: "continue"})
			moorChoose(t, r, Action{Option: "continue", Colors: []string{"vine"}})
		}
	}
	r, p := rhineFixture(t, "rhine-summer-vineTrader", nil)
	p.Coins, p.Hand = 0, nil
	if r.rhineOptionReason(p, r.Choices[0].Visitor, "trade") == "" {
		t.Fatal("no vines and no purchase funds should make trade unavailable")
	}
}
