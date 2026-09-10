package game

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func testRoom() *Room {
	a, b := NewPlayer("A"), NewPlayer("B")
	return &Room{Code: "TEST01", Phase: "lobby", HostID: a.ID, Players: []*Player{a, b}, Spaces: []Space{}, WakeSlots: []WakeSlot{}, Log: []string{}}
}
func begin(t *testing.T, r *Room) {
	t.Helper()
	if e := r.Apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	resolveTestChoices(t, r)
	r.SpringLeader = 0
	r.TurnID = r.Players[0].ID
	for i := 1; r.Phase == "wake"; i++ {
		if e := r.Apply(r.TurnID, Action{Type: "wake", Slot: i, Color: "summer"}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestPrivateView(t *testing.T) {
	r := testRoom()
	begin(t, r)
	b, _ := json.Marshal(r.View(r.Players[0].ID))
	s := string(b)
	for _, c := range r.Players[1].Hand {
		if strings.Contains(s, c.ID) {
			t.Fatal("other hand leaked")
		}
	}
	for _, c := range r.Players[0].Hand {
		if !strings.Contains(s, c.ID) {
			t.Fatal("own hand missing")
		}
	}
	if !strings.Contains(s, "\"vp\"") {
		t.Fatal("JSON vp protocol")
	}
}
func TestTurnAndCapacity(t *testing.T) {
	r := testRoom()
	begin(t, r)
	a, b := r.Players[0], r.Players[1]
	if r.Apply(b.ID, Action{Type: "place", Space: "tour"}) == nil {
		t.Fatal("out of turn")
	}
	if e := r.Apply(a.ID, Action{Type: "place", Space: "tour"}); e != nil {
		t.Fatal(e)
	}
	if r.Apply(b.ID, Action{Type: "place", Space: "tour"}) == nil {
		t.Fatal("occupied")
	}
	if e := r.Apply(b.ID, Action{Type: "place", Space: "tour", Large: true}); e != nil {
		t.Fatal(e)
	}
	if b.LargeWorker {
		t.Fatal("large worker not consumed")
	}
}
func TestPlantCapacityAndFacilities(t *testing.T) {
	p := NewPlayer("A")
	c := Card{ID: NewID(), Type: "vine", Red: 4, Trellis: true, Irrigation: true}
	p.Hand = []Card{c}
	if perform(p, Action{Space: "plant", CardID: c.ID, Field: 0}) == nil {
		t.Fatal("requires buildings")
	}
	p.Buildings = []string{"trellis", "irrigation"}
	if e := perform(p, Action{Space: "plant", CardID: c.ID, Field: 0}); e != nil {
		t.Fatal(e)
	}
	c.ID = NewID()
	p.Hand = []Card{c}
	if perform(p, Action{Space: "plant", CardID: c.ID, Field: 0}) == nil {
		t.Fatal("capacity")
	}
	if len(p.Hand) != 1 {
		t.Fatal("failed action consumed card")
	}
}
func TestHarvest(t *testing.T) {
	p := NewPlayer("A")
	p.Fields[0].Vines = []Card{{Red: 2, White: 1}, {Red: 1}}
	if e := perform(p, Action{Space: "harvest", Field: 0}); e != nil {
		t.Fatal(e)
	}
	if len(p.Grapes) != 2 {
		t.Fatal(p.Grapes)
	}
	if perform(p, Action{Space: "harvest", Field: 0}) == nil {
		t.Fatal("double harvest")
	}
}
func TestWineAndOrder(t *testing.T) {
	p := NewPlayer("A")
	p.Grapes = []Grape{{ID: NewID(), Color: "red", Value: 3}, {ID: NewID(), Color: "white", Value: 2}}
	if makeWine(p, []int{0, 0}) == nil {
		t.Fatal("duplicate grape")
	}
	if makeWine(p, []int{0, 1}) == nil {
		t.Fatal("blush without cellar")
	}
	if len(p.Grapes) != 2 {
		t.Fatal("failed wine consumed grapes")
	}
	p.Buildings = []string{"medium_cellar"}
	if e := makeWine(p, []int{0, 1}); e != nil {
		t.Fatal(e)
	}
	if p.Wines[0].Type != "blush" || p.Wines[0].Value != 5 {
		t.Fatal(p.Wines)
	}
	c := Card{ID: NewID(), Type: "order", Requirements: []Wine{{Type: "blush", Value: 4}}, Points: 4, Income: 1}
	p.Hand = []Card{c}
	if e := fillOrder(p, Action{CardID: c.ID, WineIDs: []string{p.Wines[0].ID}}); e != nil {
		t.Fatal(e)
	}
	if p.VP != 4 || p.Income != 1 || len(p.Wines) != 0 || len(p.Hand) != 0 {
		t.Fatal("order rewards")
	}
}
func TestAgeAndTraining(t *testing.T) {
	p := NewPlayer("A")
	p.Grapes = []Grape{{Color: "red", Value: 9}, {Color: "red", Value: 8}, {Color: "white", Value: 2}}
	ageGrapes(p)
	for _, g := range p.Grapes {
		if g.Color == "red" && g.Value != 8 && g.Value != 9 {
			t.Fatal(g)
		}
	}
	if p.Grapes[2].Value != 3 {
		t.Fatal(p.Grapes)
	}
	p.Coins = 4
	if e := perform(p, Action{Space: "train"}); e != nil {
		t.Fatal(e)
	}
	if p.Workers != 2 || p.TotalWorkers != 4 || p.Coins != 0 || p.Trained != 1 {
		t.Fatal("training available too early")
	}
}
func TestYearAndFinish(t *testing.T) {
	r := testRoom()
	begin(t, r)
	for r.Year == 1 {
		resolveTestChoices(t, r)
		if r.Year != 1 {
			break
		}
		if e := r.Apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Phase != "wake" || r.Year != 2 {
		t.Fatal(r.Phase)
	}
	for _, p := range r.Players {
		if p.Workers != 2 || !p.LargeWorker {
			t.Fatal("reset")
		}
	}
	for i := 1; r.Phase == "wake"; i++ {
		if e := r.Apply(r.TurnID, Action{Type: "wake", Slot: i, Color: "summer"}); e != nil {
			t.Fatal(e)
		}
	}
	r.Players[1].VP = 20
	for r.Phase != "finished" {
		resolveTestChoices(t, r)
		if r.Phase == "finished" {
			break
		}
		if e := r.Apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if r.WinnerID != r.Players[1].ID {
		t.Fatal("wrong winner")
	}
}
func TestSixPlayersWake(t *testing.T) {
	r := testRoom()
	for i := 2; i < 6; i++ {
		r.Players = append(r.Players, NewPlayer(fmt.Sprint(i)))
	}
	begin(t, r)
	for _, s := range r.Spaces {
		want := 3
		if s.ID == "gain_coin" {
			want = 0
		}
		if s.ID == "yoke" {
			want = 1
		}
		if s.Capacity != want {
			t.Fatal("6p capacity")
		}
	}
	for n := 0; n < 6; n++ {
		if e := r.Apply(r.TurnID, Action{Type: "pass"}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Phase != "fall" || len(r.Choices) != 6 {
		t.Fatal("six-player fall choices", r.Phase, len(r.Choices))
	}
	resolveTestChoices(t, r)
	if r.Phase != "winter" {
		t.Fatal(r.Phase)
	}
}

func TestNaturalFullGame(t *testing.T) {
	r := testRoom()
	if e := r.Apply(r.HostID, Action{Type: "start"}); e != nil {
		t.Fatal(e)
	}
	acted := map[string]bool{}
	for n := 0; n < 1000 && r.Phase != "finished"; n++ {
		resolveTestChoices(t, r)
		if r.Phase == "finished" {
			break
		}
		p := r.Player(r.TurnID)
		a := Action{Type: "pass"}
		key := fmt.Sprintf("%d-%s-%s", r.Year, r.Phase, p.ID)
		switch r.Phase {
		case "wake":
			slot := 1
			if p.ID == r.HostID {
				slot = 6
			}
			a = Action{Type: "wake", Slot: slot}
		case "summer":
			if !acted[key] {
				a = Action{Type: "place", Space: "tour", Large: true}
				acted[key] = true
			}

		}
		if e := r.Apply(p.ID, a); e != nil {
			t.Fatalf("year=%d phase=%s action=%+v: %v", r.Year, r.Phase, a, e)
		}
	}
	if r.Phase != "finished" || r.Player(r.WinnerID).VP < 20 {
		t.Fatal("natural game did not finish")
	}
	t.Logf("natural full game ended in year %d, winner VP=%d", r.Year, r.Player(r.WinnerID).VP)
}
func TestLargeWorkerUsesFreeSlot(t *testing.T) {
	r := testRoom()
	begin(t, r)
	if e := r.Apply(r.TurnID, Action{Type: "place", Space: "tour", Large: true}); e != nil {
		t.Fatal(e)
	}
	if r.Apply(r.TurnID, Action{Type: "place", Space: "tour"}) == nil {
		t.Fatal("large worker must occupy a free normal slot")
	}
}

// Resolve real setup/fall/discard decisions; never simulate visitor effects.
func resolveTestChoices(t *testing.T, r *Room) {
	t.Helper()
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		a := Action{Type: "choose", ChoiceID: c.ID}
		switch c.Kind {
		case "papa":
			a.Option = "coins"
		case "fall":
			a.Option = "winter"
		case "discard":
			for _, card := range r.Player(c.PlayerID).Hand[:c.Count] {
				a.CardIDs = append(a.CardIDs, card.ID)
			}
		default:
			t.Fatalf("unexpected choice %s", c.Kind)
		}
		if e := r.Apply(c.PlayerID, a); e != nil {
			t.Fatal(e)
		}
	}
}
