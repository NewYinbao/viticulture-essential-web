package game

import (
	"fmt"
	"strings"
)

func (p *Player) ownedGrande() int {
	if p.GrandeRemoved {
		return 0
	}
	return 1
}
func (p *Player) availableRegularWorkers() int {
	n := p.Workers
	if p.GrayWorkerAvailable {
		n--
	}
	return max(0, n)
}
func (p *Player) resetAnnualWorkers() {
	p.Workers = max(0, p.TotalWorkers-p.ownedGrande()-len(p.SpecialWorkers))
	p.LargeWorker = !p.GrandeRemoved
	p.GrayWorkerAvailable = false
	p.Trained = 0
}

// Worker identities are explicit from acquisition through placement/retrieval.
// Historical saves without a gray marker are not guessed when sacrificing a
// regular meeple: those ambiguous pieces are excluded until the annual reset.
type PermanentWorkerChoice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Space string `json:"space,omitempty"`
	Seat  *Seat  `json:"seat,omitempty"`
}

func (r *Room) permanentWorkerLossChoices(p *Player, availableOnly bool) []PermanentWorkerChoice {
	out := []PermanentWorkerChoice{}
	regularKnown := r.GrayWorkerOwner != p.ID || r.GrayWorkerTracked
	if regularKnown && p.availableRegularWorkers() > 0 {
		out = append(out, PermanentWorkerChoice{ID: "available:regular", Label: "待命普通工人"})
	}
	if !p.GrandeRemoved && p.LargeWorker {
		out = append(out, PermanentWorkerChoice{ID: "available:grande", Label: "待命大工人"})
	}
	for _, id := range p.SpecialWorkers {
		if p.specialWorkerReady(id, r.Year) {
			out = append(out, PermanentWorkerChoice{ID: "available:" + id, Label: "待命特殊工人 · " + id})
		}
	}
	if availableOnly {
		return out
	}
	for _, sp := range r.Spaces {
		for i, seat := range sp.Occupied {
			if seat.PlayerID != p.ID || seat.Gray || (!regularKnown && !seat.Large && seat.WorkerType == "") {
				continue
			}
			// The Tutor may remove the very worker that played it. The action and
			// its already triggered effects still complete in the saved context.
			label := "普通工人"
			if seat.Large {
				label = "大工人"
			}
			if seat.WorkerType != "" {
				label = seat.WorkerType
			}
			copy := seat
			out = append(out, PermanentWorkerChoice{ID: fmt.Sprintf("placed:%s:%d", sp.ID, i), Label: sp.Name + " · " + label, Space: sp.ID, Seat: &copy})
		}
	}
	regularPlaced := 0
	grandePlaced := false
	for _, sp := range r.Spaces {
		for _, seat := range sp.Occupied {
			if seat.PlayerID == p.ID {
				if seat.Large {
					grandePlaced = true
				} else if !seat.Gray && seat.WorkerType == "" {
					regularPlaced++
				}
			}
		}
	}
	if regularKnown && p.TotalWorkers-p.ownedGrande()-len(p.SpecialWorkers)-p.availableRegularWorkers()-regularPlaced > 0 {
		out = append(out, PermanentWorkerChoice{ID: "unready:regular", Label: "培训中普通工人"})
	}
	if !p.GrandeRemoved && !p.LargeWorker && !grandePlaced {
		out = append(out, PermanentWorkerChoice{ID: "unready:grande", Label: "培训中大工人"})
	}
	for _, id := range p.SpecialWorkers {
		if p.SpecialWorkerReady[id] > r.Year {
			out = append(out, PermanentWorkerChoice{ID: "unready:" + id, Label: "培训中特殊工人 · " + id})
		}
	}
	return out
}
func (r *Room) losePermanentWorker(p *Player, id string, availableOnly bool) error {
	var selected *PermanentWorkerChoice
	for _, choice := range r.permanentWorkerLossChoices(p, availableOnly) {
		if choice.ID == id {
			c := choice
			selected = &c
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("请选择可移除的永久工人；灰色临时工不可移除")
	}
	kind := strings.TrimPrefix(strings.TrimPrefix(id, "available:"), "unready:")
	if selected.Seat != nil {
		seat := *selected.Seat
		kind = "regular"
		if seat.Large {
			kind = "grande"
		}
		if seat.WorkerType != "" {
			kind = seat.WorkerType
		}
		sp := r.space(selected.Space)
		found := false
		for i, s := range sp.Occupied {
			if s == seat {
				sp.Occupied = append(sp.Occupied[:i], sp.Occupied[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("工人位置已改变")
		}
		r.cancelPlanned(selected.Space, seat)
		if selected.Space == "yoke" {
			p.YokeUsed = false
		}
		plans := r.MessengerPlans[:0]
		for _, plan := range r.MessengerPlans {
			if plan.PlayerID == p.ID && plan.Space == selected.Space && plan.Slot == seat.Slot && plan.WorkerType == seat.WorkerType && plan.Large == seat.Large && plan.Gray == seat.Gray {
				continue
			}
			plans = append(plans, plan)
		}
		r.MessengerPlans = plans
	} else if kind == "regular" && strings.HasPrefix(id, "available:") {
		p.Workers--
	}
	if kind == "grande" {
		p.GrandeRemoved = true
		p.LargeWorker = false
	} else if kind != "regular" {
		for i, x := range p.SpecialWorkers {
			if x == kind {
				p.SpecialWorkers = append(p.SpecialWorkers[:i], p.SpecialWorkers[i+1:]...)
				break
			}
		}
		delete(p.SpecialWorkerReady, kind)
		delete(p.SpecialWorkerUsed, kind)
	}
	if strings.HasPrefix(id, "unready:") && p.Trained > 0 {
		p.Trained--
	}
	p.TotalWorkers--
	return nil
}

func (r *Room) trainGrande(p *Player, discount int) error {
	if !p.GrandeRemoved {
		return fmt.Errorf("已拥有大工人")
	}
	if p.TotalWorkers >= 6 {
		return fmt.Errorf("工人上限为6名")
	}
	toll := 0
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			toll++
		}
	}
	cost := max(0, 4-discount) + toll
	if p.Coins < cost {
		return fmt.Errorf("培训需要%d金币", cost)
	}
	p.Coins -= cost
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			q.Coins++
		}
	}
	p.GrandeRemoved = false
	p.LargeWorker = false
	p.TotalWorkers++
	p.Trained++
	return nil
}

// Legacy clients submitted an ordinary-worker action for both ordinary-shaped
// pieces. Resolve it to the explicitly tracked gray meeple only when no regular
// worker remains, before making a seat. No historical location is inferred.
func (r *Room) normalizeWorkerAction(p *Player, a Action) Action {
	if !a.Gray && !a.Large && a.WorkerType == "" && p.availableRegularWorkers() == 0 && p.GrayWorkerAvailable {
		a.Gray = true
	}
	return a
}
