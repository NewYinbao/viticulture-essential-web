package game

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Probes only use the actor's current resources and public board. Never draw,
// shuffle, inspect deck identities, refund a committed effect, or add a skip.
func probePlayer(p *Player) *Player {
	b, _ := json.Marshal(p)
	var q Player
	_ = json.Unmarshal(b, &q)
	return &q
}
func (r *Room) probeAction(p *Player, run func(*Room, *Player) error) (*Player, bool) {
	q := probePlayer(p)
	room := *r
	room.Discards = map[string][]Card{}
	room.Context = &ActionContext{ActorID: p.ID}
	room.Choices = nil
	return q, run(&room, q) == nil
}
func (r *Room) canBuildVisitor(p *Player, discount int) bool {
	for b := range costs {
		if _, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorBuild(q, b, discount) }); ok {
			return true
		}
	}
	return false
}
func (r *Room) canPlantVisitor(p *Player, structures, limit bool) bool {
	for _, c := range p.Hand {
		if c.Type != "vine" {
			continue
		}
		for i := range p.Fields {
			if _, ok := r.probeAction(p, func(r *Room, q *Player) error {
				return r.visitorPlant(q, Action{CardIDs: []string{c.ID}, Fields: []int{i}}, 1, structures, limit)
			}); ok {
				return true
			}
		}
	}
	return false
}
func (r *Room) canFillVisitor(p *Player) bool {
	for _, c := range p.Hand {
		if c.Type != "order" {
			continue
		}
		// Select high requirements first; wines of the same type are interchangeable
		// except for value. Minimum sufficient wine preserves all higher values.
		reqs := append([]Wine{}, c.Requirements...)
		for i := range reqs {
			for j := i + 1; j < len(reqs); j++ {
				if reqs[j].Value > reqs[i].Value {
					reqs[i], reqs[j] = reqs[j], reqs[i]
				}
			}
		}
		used := map[string]bool{}
		ids := []string{}
		for _, req := range reqs {
			best := -1
			for i, w := range p.Wines {
				if !used[w.ID] && w.Type == req.Type && w.Value >= req.Value && (best < 0 || w.Value < p.Wines[best].Value) {
					best = i
				}
			}
			if best < 0 {
				break
			}
			used[p.Wines[best].ID] = true
			ids = append(ids, p.Wines[best].ID)
		}
		if len(ids) == len(reqs) {
			if _, ok := r.probeAction(p, func(r *Room, q *Player) error {
				return r.perform(q, Action{Space: "fill_order", CardID: c.ID, WineIDs: ids})
			}); ok {
				return true
			}
		}
	}
	return false
}
func (r *Room) canSubaction(p *Player, s *VisitorStep, op string) bool {
	switch op {
	case "build":
		return r.canBuildVisitor(p, 0)
	case "plant":
		return r.canPlantVisitor(p, false, false)
	case "fill":
		return r.canFillVisitor(p)
	case "harvest":
		for i := range p.Fields {
			if _, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.perform(q, Action{Space: "harvest", Field: i}) }); ok {
				return true
			}
		}
		return false
	case "vp":
		return true
	case "make":
		return r.canMakeVisitor(p, false, 0)
	}
	return false
}
func (r *Room) canMakeThenFill(p *Player, n int) bool {
	if r.canFillVisitor(p) {
		return true
	}
	if n == 0 {
		return false
	}
	hasOrder := false
	for _, c := range p.Hand {
		if c.Type == "order" {
			hasOrder = true
		}
	}
	if !hasOrder {
		return false
	}
	try := func(indices []int) bool {
		q := probePlayer(p)
		return makeWine(q, indices) == nil && r.canMakeThenFill(q, n-1)
	}
	for i := range p.Grapes {
		if try([]int{i}) {
			return true
		}
		for j := i + 1; j < len(p.Grapes); j++ {
			if try([]int{i, j}) {
				return true
			}
			for k := j + 1; k < len(p.Grapes); k++ {
				if try([]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
}
func (r *Room) canSequence(p *Player, s *VisitorStep, option string) bool {
	parts := strings.Split(option, "_")
	if len(parts) != 2 {
		return r.canSubaction(p, s, option)
	}
	if s.CardID == "winter-29" {
		if p.VP <= -5 {
			return false
		}
	}
	first, last := parts[0], parts[1]
	if first == "make" {
		if last == "fill" {
			return r.eachMadeWine(p, false, 0, func(q *Player) bool { return r.canMakeThenFill(q, 1) })
		}
		return r.canSubaction(p, s, first) && r.canSubaction(p, s, last)
	}
	if first == "harvest" && last == "make" {
		for i := range p.Fields {
			q, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorHarvest(q, Action{Fields: []int{i}}, 1) })
			if ok && r.canSubaction(q, s, last) {
				return true
			}
		}
		return false
	}
	if first == "vp" {
		q := probePlayer(p)
		q.VP++
		return r.canSubaction(q, s, last)
	}
	// Deterministic sequence probes must satisfy each mandatory action.
	if first == "fill" || first == "harvest" {
		return r.canSubaction(p, s, first) && r.canSubaction(p, s, last)
	}
	try := func(a Action) bool {
		q, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorSubaction(q, s, first, a) })
		return ok && r.canSubaction(q, s, last)
	}
	if first == "build" {
		for b := range costs {
			if try(Action{Building: b}) {
				return true
			}
		}
	}
	if first == "plant" {
		for _, c := range p.Hand {
			if c.Type == "vine" {
				for i := range p.Fields {
					if try(Action{CardIDs: []string{c.ID}, Fields: []int{i}}) {
						return true
					}
				}
			}
		}
	}
	return false
}
func (r *Room) mandatoryVisitorPlayable(p *Player, id string) error {
	// Exclude the played visitor from discard costs (also for nested/bonus plays).
	p = probePlayer(p)
	for i, c := range p.Hand {
		if c.ID == id {
			removeCard(p, i)
			break
		}
	}
	ok := true
	vines := 0
	for _, f := range p.Fields {
		if !f.Sold {
			vines += len(f.Vines)
		}
	}
	switch id {
	case "summer-09":
		ok = r.canPlantVisitor(p, false, false) || vines > 0
	case "summer-28":
		ok = r.canPlantVisitor(p, false, false) || r.canBuildVisitor(p, 3)
	case "summer-38":
		ok = p.Coins >= 2 && len(r.Players) > 1
	case "winter-12", "winter-24":
		ok = r.canSubaction(p, &VisitorStep{}, "harvest")
	case "winter-18":
		ok = r.canSubaction(p, &VisitorStep{}, "harvest") || (!has(p, "yoke") && p.Coins >= 1)
	case "winter-22", "winter-26":
		ok = r.canMakeVisitor(p, false, 0)
	case "winter-33":
		ok = r.canMakeVisitor(p, true, 4)
	case "winter-08":
		ok = r.canMakeVisitor(p, false, 0) || (p.TotalWorkers < 6 && p.Coins >= 2)
	case "summer-02":
		ok = p.Coins >= 9 || p.VP >= -3
	case "summer-08":
		ok = p.Coins >= 6 || p.VP >= -2
	case "summer-04":
		ok = r.canBuildVisitor(p, 2)
	case "summer-05", "summer-22", "winter-20", "winter-29":
		ok = false
		for _, o := range visitorDefs[id].Options {
			if r.canSequence(p, &VisitorStep{CardID: id}, o) {
				ok = true
				break
			}
		}
	case "summer-10":
		ok = p.Coins >= 2 || len(p.Grapes) > 0
	case "summer-13":
		ok = false
		for b, cost := range costs {
			loss := 2
			if cost == 2 || cost == 3 {
				loss = 1
			}
			if p.VP-loss >= -5 {
				if _, yes := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorBuild(q, b, 99) }); yes {
					ok = true
				}
			}
		}
	case "summer-15", "summer-20":
		ok = len(p.Hand) >= 2
	case "summer-16":
		visitors := 0
		for _, c := range p.Hand {
			if c.Type == "summer" || c.Type == "winter" {
				visitors++
			}
		}
		ok = p.Coins >= 4 || (visitors >= 3 && len(p.Wines) > 0)
	case "summer-19":
		ok = r.canPlantVisitor(p, true, false) || vines >= 2
	case "summer-24":
		ok = r.canPlantVisitor(p, true, false) || vines >= 1
	case "summer-25", "summer-30":
		ok = r.canPlantVisitor(p, false, false)
	case "summer-27":
		ok = r.canPlantVisitor(p, false, true)
	case "summer-26", "winter-16":
		ok = len(p.Grapes) > 0 || len(p.Wines) > 0
	case "summer-32":
		ok = p.Coins >= 2 && r.hasRetrievableWorker(p)
	case "summer-36":
		ok = false
		if p.Coins >= 8 {
			for b := range costs {
				q, yes := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorBuild(q, b, 99) })
				if yes && r.canBuildVisitor(q, 99) {
					ok = true
					break
				}
			}
		}
	case "winter-01":
		ok = p.Coins >= 3 || r.canFillVisitor(p)
	case "winter-13":
		ok = p.TotalWorkers == 6 || (p.TotalWorkers < 6 && p.Coins >= 2)
	case "winter-14":
		_, ok = r.probeAction(p, func(r *Room, q *Player) error { return r.visitorUpgrade(q, 2) })
		if !ok {
			for i, w := range p.Wines {
				if !wineAgeBoundary(p, w.Value) {
					continue
				}
				blocked := false
				for _, v := range p.Wines {
					if v.Type == w.Type && v.Value == w.Value+1 {
						blocked = true
					}
				}
				if !blocked {
					q := probePlayer(p)
					q.Wines[i].Value++
					if r.canFillVisitor(q) {
						ok = true
						break
					}
				}
			}
		}
	case "winter-19":
		n := 0
		for _, typ := range []string{"vine", "summer", "order", "winter"} {
			d := r.Discards[typ]
			if len(d) > 0 && d[len(d)-1].ID != id {
				n++
			}
		}
		ok = n >= 2
	case "winter-30":
		ok = r.canBuildVisitor(p, 0)
	case "winter-31":
		ok = (p.Coins >= 3 && p.TotalWorkers < 6) || len(p.Wines) > 0
	case "winter-34":
		ok = p.Coins >= 1 || p.Income >= 2
	case "winter-36":
		ok = len(p.Wines) > 0
	case "winter-37":
		n := 0
		for _, typ := range []string{"vine", "summer", "order", "winter"} {
			if len(r.Decks[typ])+len(r.Discards[typ]) > 0 {
				n++
			}
		}
		ok = n >= 2
	}
	if !ok {
		return fmt.Errorf("访客没有可完成的强制分支: %s", id)
	}
	return nil
}

// Validate before the first visible/random step. Drawn identities never enter a
// feasibility decision. Other continuations are checked after deterministic
// first-step simulation within apply's transaction, before HTTP save/commit.
func (r *Room) visitorBeforeEffect(p *Player, s *VisitorStep, a Action) error {
	if s.Stage != "effect" {
		return nil
	}
	if s.CardID == "winter-02" && a.Option == "make" && !r.canMakeVisitor(p, false, 0) {
		return fmt.Errorf("抽订单前须能酿至少1酒")
	}
	if s.CardID == "winter-23" && (a.Option == "both" || a.Option == "draw_train") {
		if p.Coins < 3 || p.TotalWorkers >= 6 || p.VP <= -5 {
			return fmt.Errorf("抽牌前须能支付培训及1分成本")
		}
	}
	if s.CardID == "winter-27" && strings.Contains(a.Option, "upgrade") {
		if _, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorUpgrade(q, 0) }); !ok {
			return fmt.Errorf("选两项前须能完成酒窖升级")
		}
	}
	return nil
}
func (r *Room) visitorContinuationFeasible(p *Player, s *VisitorStep) error {
	for _, c := range r.Choices {
		v := c.Visitor
		if v == nil || v.CardID != s.CardID || c.PlayerID != p.ID {
			continue
		}
		ok := true
		switch s.CardID {
		case "summer-05", "summer-22", "winter-20", "winter-29":
			if v.Stage == "continuation" {
				ok = r.canSubaction(p, v, c.Options[0])
			}
		case "summer-28":
			if v.Stage == "build" {
				ok = r.canBuildVisitor(p, 3)
			} else if v.Stage == "plant" {
				ok = r.canPlantVisitor(p, false, false)
			}
		case "winter-14":
			if v.Stage == "fill" {
				ok = r.canFillVisitor(p)
			}
		}
		if !ok {
			return fmt.Errorf("第一步未提交：余下强制步骤无法完成")
		}
	}
	return nil
}

// Enumerate legal single-wine results from current public resources, never decks.
func (r *Room) eachMadeWine(p *Player, exceptional bool, minimum int, accept func(*Player) bool) bool {
	try := func(ids []int) bool {
		q := probePlayer(p)
		buildings := append([]string{}, q.Buildings...)
		if exceptional {
			q.Buildings = append(q.Buildings, "large_cellar")
		}
		if makeWine(q, ids) != nil {
			return false
		}
		q.Buildings = buildings
		if q.Wines[len(q.Wines)-1].Value < minimum {
			return false
		}
		return accept(q)
	}
	for i := range p.Grapes {
		if try([]int{i}) {
			return true
		}
		for j := i + 1; j < len(p.Grapes); j++ {
			if try([]int{i, j}) {
				return true
			}
			for k := j + 1; k < len(p.Grapes); k++ {
				if try([]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
}
func (r *Room) canMakeVisitor(p *Player, exceptional bool, minimum int) bool {
	return r.eachMadeWine(p, exceptional, minimum, func(*Player) bool { return true })
}
