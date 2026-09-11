package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestChefPreviewAccountsForSoldatoRemovedByBump(t *testing.T) {
	for _, remaining := range []bool{false, true} {
		r := specialWorkerRoom("chef", "soldato")
		p, q := r.Players[0], r.Players[1]
		readySpecial(r, p, "chef")
		p.Workers, p.LargeWorker, p.Coins = 0, false, 0
		s := r.space("tour")
		s.Capacity = 2
		s.Occupied = []Seat{{PlayerID: q.ID, Slot: 1, WorkerType: "soldato"}}
		if remaining {
			s.Occupied = append(s.Occupied, Seat{PlayerID: r.Players[2].ID, Slot: 2, WorkerType: "soldato"})
		}
		before, _ := json.Marshal(r)
		options := r.workerPlacements(p, s)
		if placementHas(options, "chef", 1) == remaining {
			t.Fatalf("remaining=%v wrong toll preview: %+v", remaining, options)
		}
		after, _ := json.Marshal(r)
		if !bytes.Equal(before, after) {
			t.Fatal("preview mutated original room")
		}
		err := r.Apply(p.ID, Action{Type: "place", Space: "tour", WorkerType: "chef", Slot: 1})
		if (err != nil) != remaining {
			t.Fatalf("remaining=%v engine mismatch: %v", remaining, err)
		}
		if !remaining && p.Coins != 2 {
			t.Fatal("removed Soldato still collected toll")
		}
	}
}

func TestChefAutomaticSlotSkipsOtherChefs(t *testing.T) {
	for _, bump := range []bool{false, true} {
		r := specialWorkerRoom("chef", "oracle")
		p, q := r.Players[0], r.Players[1]
		readySpecial(r, p, "chef")
		p.Workers, p.LargeWorker = 0, false
		s := r.space("tour")
		s.Capacity = 2
		s.Occupied = []Seat{{PlayerID: q.ID, Slot: 1, WorkerType: "chef"}}
		if bump {
			s.Occupied = append(s.Occupied, Seat{PlayerID: q.ID, Slot: 2})
		}
		count := q.Workers
		if err := r.Apply(p.ID, Action{Type: "place", Space: "tour", WorkerType: "chef"}); err != nil {
			t.Fatalf("bump=%v auto selection hit protected Chef: %v", bump, err)
		}
		if bump && q.Workers != count+1 {
			t.Fatal("bumpable worker was not returned")
		}
		if len(s.Occupied) != 2 || s.Occupied[0].WorkerType != "chef" || s.Occupied[0].PlayerID != q.ID || s.Occupied[1].PlayerID != p.ID || s.Occupied[1].Slot != 2 {
			t.Fatalf("wrong final seats: %+v", s.Occupied)
		}
	}
}
