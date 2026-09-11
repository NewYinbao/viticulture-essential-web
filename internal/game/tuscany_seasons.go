package game

import (
	"fmt"
	"strconv"
)

var tuscanySeasons = []string{"spring", "summer", "fall", "winter"}

func seasonIndex(s string) int {
	for i, v := range tuscanySeasons {
		if s == v {
			return i
		}
	}
	return -1
}

func (r *Room) tuscanySpring() {
	r.Phase = "wake"
	r.WakeSlots = []WakeSlot{}
	bonuses := []string{"需持起始标记；无奖励", "夏：1金币", "夏：藤；秋：夏访客", "夏：订单；秋：夏访客", "夏：自选访客；秋：冬访客；冬：影响力", "夏：1分；秋：冬访客；冬：葡萄陈酿", "夏：临时工人；秋：任意牌；冬：下年首位"}
	for i, b := range bonuses {
		r.WakeSlots = append(r.WakeSlots, WakeSlot{Slot: i + 1, Bonus: b})
	}
	for _, p := range r.Players {
		p.Wake = 0
		p.Season = "spring"
		p.Passed = false
	}
	r.TurnID = r.Players[r.SpringLeader].ID
}

func (r *Room) tuscanyWake(p *Player, a Action) error {
	if a.Slot < 2 || a.Slot > 7 || r.WakeSlots[a.Slot-1].PlayerID != "" {
		return fmt.Errorf("首年只能选择空的第2–7行")
	}
	r.WakeSlots[a.Slot-1].PlayerID = p.ID
	p.Wake = a.Slot
	for k := 1; k <= len(r.Players); k++ {
		q := r.Players[(r.SpringLeader+k)%len(r.Players)]
		if q.Wake == 0 {
			r.TurnID = q.ID
			return nil
		}
	}
	r.Phase = "spring"
	r.TurnID = r.ordered()[0].ID
	return nil
}

func (r *Room) tuscanyPass(p *Player) {
	r.tuscanyPassWithResume(p, false)
}

func (r *Room) tuscanyPassWithResume(p *Player, planner bool) {
	oldPhase := r.Phase
	p.Passed = true
	if r.Phase == "winter" {
		p.Season = "year_end"
		// Workers return now and open spaces for players still taking winter turns.
		for i := range r.Spaces {
			sp := &r.Spaces[i]
			out := sp.Occupied[:0]
			for _, seat := range sp.Occupied {
				if seat.PlayerID != p.ID {
					out = append(out, seat)
				} else {
					r.cancelPlanned(sp.ID, seat)
				}
			}
			sp.Occupied = out
		}
		p.resetAnnualWorkers()
		// The gray worker returns to the board supply, not the next year's pool.
		// Keep its annual owner until everybody has passed, preventing re-awards.
		ageGrapes(p)
		ageWines(p)
		if has(p, "distiller") {
			ageGrapes(p)
		}
		if has(p, "storehouse") {
			ageWines(p)
		}
		if has(p, "silo") {
			r.draw(p, "vine")
		}
		if has(p, "dock") {
			r.draw(p, "order")
		}
		if len(p.Hand) > 7 {
			r.enqueue(p.ID, "discard", nil, len(p.Hand)-7)
		}
		r.enqueue(p.ID, "tuscany_upkeep", []string{"continue"}, 1)
	} else {
		p.Season = tuscanySeasons[seasonIndex(r.Phase)+1]
		r.tuscanySeasonReward(p, p.Season)
	}
	if oldPhase == "summer" && r.Config.Structures && has(p, "barn") {
		r.enqueue(p.ID, "structure_barn", []string{"take", "skip"}, 2)
	}
	r.AddLog(p.Name + " 结束" + r.Phase + "行动")
	if len(r.Choices) > 0 {
		r.Resume = "tuscany_pass"
		if planner {
			r.Resume = "tuscany_planner_pass"
		}
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	if planner {
		r.startNextPlanner()
	} else {
		r.tuscanyNext()
	}
}

func (r *Room) tuscanySeasonReward(p *Player, season string) {
	switch season {
	case "summer":
		switch p.Wake {
		case 2:
			p.Coins++
		case 3:
			r.draw(p, "vine")
		case 4:
			if r.Config.Structures {
				r.draw(p, "structure")
			} else {
				r.draw(p, "order")
			}
		case 5:
			r.enqueue(p.ID, "tuscany_draw", []string{"summer", "winter"}, 1)
		case 6:
			p.VP++
		case 7:
			if r.Config.Structures {
				r.draw(p, "structure")
			} else {
				r.claimGrayWorker(p)
			}
		}
	case "fall":
		switch p.Wake {
		case 3, 4:
			r.draw(p, "summer")
		case 5, 6:
			r.draw(p, "winter")
		case 7:
			opts := []string{"vine", "order", "summer", "winter"}
			if r.Config.Structures {
				opts = append(opts, "structure")
			}
			r.enqueue(p.ID, "tuscany_draw", opts, 1)
		}
		if has(p, "cottage") {
			r.enqueue(p.ID, "tuscany_draw", []string{"summer", "winter"}, 1)
		}
	case "winter":
		switch p.Wake {
		case 5:
			r.enqueue(p.ID, "tuscany_influence", []string{"place", "skip"}, 1)
		case 6:
			ageGrapes(p)
		case 7:
			r.FirstPlayerID = p.ID
		}
	}
}

func (r *Room) tuscanyNext() {
	ps := r.ordered()
	idx := 0
	for i, p := range ps {
		if p.ID == r.TurnID {
			idx = i
		}
	}
	for k := 1; k <= len(ps); k++ {
		p := ps[(idx+k)%len(ps)]
		if !p.Passed {
			r.TurnID = p.ID
			if r.activateMessenger(p) {
				return
			}
			return
		}
	}
	if r.Phase == "winter" {
		r.tuscanyFinishYear()
		return
	}
	r.Phase = tuscanySeasons[seasonIndex(r.Phase)+1]
	for _, p := range ps {
		p.Passed = false
	}
	r.TurnID = ps[0].ID
	// Resolve season-start Planner placements before the first normal turn.
	// startNextPlanner activates that player's Messenger when the queue drains.
	r.startNextPlanner()
}

func (r *Room) tuscanyNextWakeOptions(p *Player) []string {
	if r.FirstPlayerID == p.ID {
		return []string{"1"}
	}
	out := []string{}
	for slot := 2; slot <= 7; slot++ {
		used := false
		for _, q := range r.Players {
			if q.NextWake == slot {
				used = true
			}
		}
		if !used {
			out = append(out, strconv.Itoa(slot))
		}
	}
	return out
}

func (r *Room) tuscanyChoice(p *Player, c Choice, a Action) error {
	switch c.Kind {
	case "tuscany_draw":
		r.draw(p, a.Option)
	case "tuscany_influence":
		if a.Option == "skip" {
			return nil
		}
		if len(a.Influence) != 1 {
			return fmt.Errorf("须放置或移动一颗星")
		}
		return r.influence(p, a.Influence[0])
	case "tuscany_trade":
		if a.Option == "skip" {
			return nil
		}
		if len(a.Trades) != 1 {
			return fmt.Errorf("须完成一次交易")
		}
		return r.trade(p, a.Trades[0])
	case "tuscany_upkeep":
		p.Coins += p.Income
		r.enqueue(p.ID, "tuscany_next_wake", r.tuscanyNextWakeOptions(p), 1)
	case "tuscany_next_wake":
		valid := false
		for _, s := range r.tuscanyNextWakeOptions(p) {
			if s == a.Option {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("下一年起床位不可用")
		}
		p.NextWake, _ = strconv.Atoi(a.Option)
		p.Season = "ready"
	default:
		return fmt.Errorf("未知 Tuscany 选择")
	}
	return nil
}

func (r *Room) tuscanyFinishYear() {
	// Check the scores after all winter actions, before residual Statue points
	// and influence scoring. The official structure card explicitly says that
	// its end-of-year point cannot trigger the final year.
	r.FinalYear = false
	for _, p := range r.Players {
		if p.VP >= 25 {
			r.FinalYear = true
		}
	}
	if r.FinalYear {
		for _, p := range r.Players {
			if has(p, "statue") {
				p.VP++
			}
		}
		r.scoreInfluence()
		r.finishGame()
		return
	}
	// Statue still pays in a non-final year, but its point is deliberately
	// applied only after the threshold check above. It can then contribute to
	// the next year's threshold as required by the printed rule.
	for _, p := range r.Players {
		if has(p, "statue") {
			p.VP++
		}
	}
	r.MessengerPlans = nil
	r.Year++
	r.GrayWorkerOwner = ""
	r.FirstPlayerID = ""
	r.Spaces = r.spacesForRules()
	for i := range r.WakeSlots {
		r.WakeSlots[i].PlayerID = ""
	}
	for _, p := range r.Players {
		p.Wake = p.NextWake
		p.NextWake = 0
		p.Season = "spring"
		p.Passed = false
		p.MillUsed = false
		p.TastingUsed = false
		p.YokeUsed = false
		p.ensureSpecialMaps()
		for id := range p.SpecialWorkerUsed {
			p.SpecialWorkerUsed[id] = false
		}
		for i := range p.Fields {
			p.Fields[i].Harvested = false
		}
		r.WakeSlots[p.Wake-1].PlayerID = p.ID
	}
	r.Phase = "spring"
	r.TurnID = r.ordered()[0].ID
}
