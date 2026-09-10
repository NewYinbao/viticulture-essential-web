package game

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMandatoryVisitorCannotStartEmpty(t *testing.T) {
	ids := []string{"summer-02", "summer-04", "summer-05", "summer-08", "summer-10", "summer-13", "summer-15", "summer-16", "summer-19", "summer-20", "summer-22", "summer-24", "summer-25", "summer-26", "summer-27", "summer-30", "summer-32", "summer-36", "winter-01", "winter-13", "winter-14", "winter-16", "winter-19", "winter-20", "winter-30", "winter-31", "winter-34", "winter-36", "winter-37"}
	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			r, p := visitorFixture(id)
			p.Coins = 0
			p.VP = -5
			p.Hand = nil
			p.Wines = nil
			p.Grapes = nil
			p.Income = 0
			p.Buildings = nil
			if id == "summer-04" {
				p.Buildings = []string{"yoke", "trellis"}
			}
			for i := range p.Fields {
				p.Fields[i].Vines = nil
			}
			r.Decks = map[string][]Card{}
			r.Discards = map[string][]Card{}
			for _, c := range Catalog() {
				if c.ID == id {
					p.Hand = append(p.Hand, c)
				}
			}
			before, _ := json.Marshal(r)
			if e := r.Apply(p.ID, Action{Type: "place", Space: strings.Split(id, "-")[0] + "_visitor", CardID: id, Slot: 2}); e == nil {
				t.Fatal("stranded visitor accepted")
			}
			after, _ := json.Marshal(r)
			if !bytes.Equal(before, after) {
				t.Fatal("failed placement mutated card/worker/state")
			}
		})
	}
}

func TestVisitorFirstStepRequiresCompletableRemainder(t *testing.T) {
	cases := []struct {
		id, option string
		setup      func(*Player)
		edit       func(*Action)
	}{
		{"summer-05", "vp_build", func(p *Player) { p.Coins = 0 }, nil},
		{"summer-05", "vp_plant", func(p *Player) { p.Hand = nil }, nil},
		{"summer-22", "build_plant", func(p *Player) { p.Hand = nil }, nil},
		{"summer-22", "plant_build", func(p *Player) { p.Coins = 0 }, nil},
		{"summer-28", "plant_build", func(p *Player) {
			p.Buildings = []string{"yoke", "trellis", "irrigation", "cottage", "windmill", "tasting_room", "medium_cellar", "large_cellar"}
		}, nil},
		{"winter-20", "harvest_fill", func(p *Player) { p.Wines = nil }, nil},
		{"winter-20", "make_fill", func(p *Player) { p.Hand = nil }, nil},
		{"winter-20", "fill_harvest", func(p *Player) {
			for i := range p.Fields {
				p.Fields[i].Harvested = true
			}
		}, nil},
		{"winter-14", "fill", func(p *Player) { p.Hand = nil }, nil},
		{"winter-23", "draw_train", func(p *Player) { p.Coins = 2 }, nil},
		{"winter-23", "both", func(p *Player) { p.TotalWorkers = 6 }, nil},
		{"winter-27", "draw_upgrade", func(p *Player) { p.Coins = 0 }, nil},
		{"winter-27", "vp_upgrade", func(p *Player) { p.Buildings = append(p.Buildings, "medium_cellar", "large_cellar") }, nil},
	}
	for _, tc := range cases {
		t.Run(tc.id+tc.option, func(t *testing.T) {
			r, p := visitorFixture(tc.id)
			playVisitorFixture(t, r, p, tc.id)
			c := r.Choices[0]
			a := prepareVisitorAction(r, c, tc.option)
			tc.setup(p)
			if tc.edit != nil {
				tc.edit(&a)
			}
			before, _ := json.Marshal(r)
			if e := r.Apply(p.ID, a); e == nil {
				t.Fatal("impossible continuation accepted")
			}
			after, _ := json.Marshal(r)
			if !bytes.Equal(before, after) {
				t.Fatal("first step, VP cost, random draw or choice committed")
			}
			// Forced continuations cannot be escaped by a forged skip.
			a.Option = "skip"
			if e := r.Apply(p.ID, a); e == nil {
				t.Fatal("free skip accepted")
			}
		})
	}
}

func TestSequenceCapabilityUsesPostStepResources(t *testing.T) {
	r, p := visitorFixture("summer-22")
	p.Coins = 2
	p.Buildings = nil
	p.Hand = []Card{{ID: "summer-22", Type: "summer"}, {ID: "needs-trellis", Type: "vine", Red: 1, Trellis: true}}
	playVisitorFixture(t, r, p, "summer-22")
	c := r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "build_plant", Building: "trellis"}); e != nil {
		t.Fatal(e)
	}
	c = r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "plant", CardIDs: []string{"needs-trellis"}, Fields: []int{0}}); e != nil {
		t.Fatal(e)
	}
	if r.Context != nil {
		t.Fatal("not completed")
	}
}

func TestMakeFillStartWithTwoNewWines(t *testing.T) {
	r, p := visitorFixture("winter-20")
	p.Wines = nil
	p.Grapes = []Grape{{ID: "red", Color: "red", Value: 1}, {ID: "white", Color: "white", Value: 1}}
	for i := range p.Fields {
		p.Fields[i].Vines = nil
	}
	p.Hand = []Card{{ID: "winter-20", Type: "winter"}, {ID: "two", Type: "order", Requirements: []Wine{{Type: "red", Value: 1}, {Type: "white", Value: 1}}}}
	playVisitorFixture(t, r, p, "winter-20")
	c := r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "make_fill", Recipes: [][]int{{0}, {1}}}); e != nil {
		t.Fatal(e)
	}
	c = r.Choices[0]
	if e := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "fill", CardID: "two", WineIDs: []string{p.Wines[0].ID, p.Wines[1].ID}}); e != nil {
		t.Fatal(e)
	}
}

func TestHiddenDrawGuardIndependentOfDeckOrder(t *testing.T) {
	for _, id := range []string{"winter-23", "winter-27"} {
		r, p := visitorFixture(id)
		playVisitorFixture(t, r, p, id)
		p.Coins = 0
		opt := "draw_train"
		if id == "winter-27" {
			opt = "draw_upgrade"
		}
		c := r.Choices[0]
		a := Action{Type: "choose", ChoiceID: c.ID, Option: opt}
		b, _ := json.Marshal(r)
		e := r.Apply(p.ID, a)
		if e == nil {
			t.Fatal("hidden first step allowed")
		}
		after, _ := json.Marshal(r)
		if !bytes.Equal(b, after) {
			t.Fatal("draw leaked")
		}
		for typ, d := range r.Decks {
			for i, j := 0, len(d)-1; i < j; i, j = i+1, j-1 {
				d[i], d[j] = d[j], d[i]
			}
			r.Decks[typ] = d
		}
		e2 := r.Apply(p.ID, a)
		if e2 == nil || e.Error() != e2.Error() {
			t.Fatal("error reveals deck identity")
		}
	}
}
