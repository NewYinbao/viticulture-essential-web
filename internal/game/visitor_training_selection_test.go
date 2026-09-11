package game

import (
	"encoding/json"
	"testing"
)

func TestEEVisitorsTrainChosenSpecialWorker(t *testing.T) {
	for _, tc := range []struct {
		id   string
		cost int
		now  bool
	}{
		{"winter-07", 0, false}, {"winter-08", 2, false}, {"winter-13", 2, false}, {"winter-23", 3, false}, {"winter-31", 3, true},
	} {
		t.Run(tc.id, func(t *testing.T) {
			r, p := visitorFixture(tc.id)
			r.Config.SpecialWorkers = true
			r.Config.Structures = true
			r.SpecialWorkerPool = []string{"oracle", "chef"}
			r.Players[1].Buildings = append(r.Players[1].Buildings, "academy")
			if err := r.Apply(p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 2, CardID: tc.id}); err != nil {
				t.Fatal(err)
			}
			workers, total, coins := p.Workers, p.TotalWorkers, p.Coins
			before, _ := json.Marshal(r)
			if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "train", SpecialWorker: "unlisted"}); err == nil {
				t.Fatal("unlisted special worker accepted")
			}
			after, _ := json.Marshal(r)
			if string(before) != string(after) {
				t.Fatal("invalid worker changed room")
			}
			if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "train", SpecialWorker: "oracle"}); err != nil {
				t.Fatal(err)
			}
			ready := r.Year + 1
			if tc.now {
				ready = r.Year
			}
			if p.Coins != coins-tc.cost-2 || p.TotalWorkers != total+1 || p.Workers != workers || p.SpecialWorkerReady["oracle"] != ready {
				t.Fatalf("chosen worker/cost/readiness mismatch: %+v", p)
			}
		})
	}
}

func TestRecruiterRespondentChoosesTheirOwnSpecialWorker(t *testing.T) {
	r, p := visitorFixture("winter-38")
	r.Config.SpecialWorkers = true
	r.SpecialWorkerPool = []string{"oracle", "chef"}
	if err := r.Apply(p.ID, Action{Type: "place", Space: "winter_visitor", Slot: 2, CardID: "winter-38"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: r.Choices[0].ID, Option: "resolve"}); err != nil {
		t.Fatal(err)
	}
	for len(r.Choices) > 0 {
		c := r.Choices[0]
		q := r.Player(c.PlayerID)
		before := q.Coins
		if err := r.Apply(q.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "train", SpecialWorker: "chef"}); err != nil {
			t.Fatal(err)
		}
		if q.Coins != before-2 || !q.hasSpecialWorker("chef") {
			t.Fatal("response trained wrong worker/seat")
		}
	}
}
