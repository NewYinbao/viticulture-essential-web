package game

import (
	"fmt"
	"sort"
)

func (r *Room) spring() {
	if r.tuscany() {
		r.tuscanySpring()
		return
	}
	r.Phase = "wake"
	r.WakeSlots = []WakeSlot{}
	bonuses := []string{"优先行动", "抽葡萄藤", "抽订单", "1金币", "选择夏季或冬季访客", "1胜利分", "本年额外1工人"}
	for i, b := range bonuses {
		r.WakeSlots = append(r.WakeSlots, WakeSlot{Slot: i + 1, Bonus: b})
	}
	for _, p := range r.Players {
		p.Wake = 0
		p.Passed = false
	}
	r.TurnID = r.Players[r.SpringLeader%len(r.Players)].ID
	r.AddLog(fmt.Sprintf("第%d年 · 选择起床顺序", r.Year))
}
func (r *Room) ordered() []*Player {
	p := append([]*Player{}, r.Players...)
	sort.Slice(p, func(i, j int) bool { return p[i].Wake < p[j].Wake })
	return p
}
func (r *Room) next() {
	if r.tuscany() {
		r.tuscanyNext()
		return
	}
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
	if r.Phase == "summer" {
		if r.Config.Structures {
			for _, p := range ps {
				if has(p, "barn") && len(p.Hand) >= 2 {
					r.enqueue(p.ID, "structure_barn", []string{"take", "skip"}, 2)
				}
			}
			if len(r.Choices) > 0 {
				r.Resume = "ee_barn"
				r.TurnID = r.Choices[0].PlayerID
				return
			}
		}
		r.enterEEFall(ps)
	} else {
		r.endYear()
	}
}
func (r *Room) enterEEFall(ps []*Player) {
	r.Phase = "fall"
	for _, p := range ps {
		p.Passed = false
		if has(p, "cottage") {
			r.enqueue(p.ID, "fall", []string{"summer_summer", "summer_winter", "winter_winter"}, 2)
		} else {
			r.enqueue(p.ID, "fall", []string{"summer", "winter"}, 1)
		}
	}
	r.Resume = "winter"
	r.TurnID = r.Choices[0].PlayerID
	r.AddLog("秋季：选择访客颜色，小屋额外任选一张")
}
func (r *Room) finishYear() {
	// Return all placed workers even when this is the final year.
	r.Spaces = NewSpaces(len(r.Players))
	r.MessengerPlans = nil
	r.GrayWorkerOwner = ""
	r.FinalYear = false
	for _, p := range r.Players {
		// The official EE threshold is checked before the Statue's residual
		// point. Statue may score during upkeep, but cannot itself start the
		// final year.
		thresholdReached := p.VP >= 20
		p.Coins += p.Income
		ageGrapes(p)
		if has(p, "distiller") {
			ageGrapes(p)
		}
		ageWines(p)
		if has(p, "storehouse") {
			ageWines(p)
		}
		if has(p, "silo") {
			r.draw(p, "vine")
		}
		if has(p, "dock") {
			r.draw(p, "order")
		}
		if has(p, "statue") {
			p.VP++
		}
		p.resetAnnualWorkers()
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
		if thresholdReached {
			r.FinalYear = true
		}
	}
	if len(r.Choices) > 0 {
		// Residual order draws can create a Mercado choice. Do not advance the
		// year or end the game until every optional delivery has been resolved.
		r.Resume = "finish_year_finalize"
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	r.finishYearFinalize()
}

func (r *Room) finishYearFinalize() {
	if r.FinalYear {
		r.finishGame()
		return
	}
	r.Year++
	r.SpringLeader = (r.SpringLeader + len(r.Players) - 1) % len(r.Players)
	r.spring()
}
func (r *Room) endYear() {
	r.Phase = "year_end"
	r.Resume = "end_year"
	for _, p := range r.ordered() {
		if len(p.Hand) > 7 {
			r.enqueue(p.ID, "discard", nil, len(p.Hand)-7)
		}
	}
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	r.Resume = ""
	r.finishYear()
}
func scoreKey(p *Player) [4]int {
	k := [4]int{p.VP, p.Coins, 0, 0}
	for _, w := range p.Wines {
		k[2] += w.Value
	}
	for _, g := range p.Grapes {
		k[3] += g.Value
	}
	return k
}

func (r *Room) finishGame() {
	if r.Phase == "finished" {
		return
	}
	r.settleMoorContracts()
	r.Phase = "finished"
	r.TurnID = ""
	ps := append([]*Player{}, r.Players...)
	sort.SliceStable(ps, func(i, j int) bool {
		a, b := ps[i], ps[j]
		if a.VP != b.VP {
			return a.VP > b.VP
		}
		if a.Coins != b.Coins {
			return a.Coins > b.Coins
		}
		av, bv := 0, 0
		for _, w := range a.Wines {
			av += w.Value
		}
		for _, w := range b.Wines {
			bv += w.Value
		}
		if av != bv {
			return av > bv
		}
		av, bv = 0, 0
		for _, g := range a.Grapes {
			av += g.Value
		}
		for _, g := range b.Grapes {
			bv += g.Value
		}
		return av > bv
	})
	r.WinnerID = ps[0].ID
	r.WinnerIDs = nil
	for _, p := range ps {
		if scoreKey(p) == scoreKey(ps[0]) {
			r.WinnerIDs = append(r.WinnerIDs, p.ID)
		}
	}
	r.AddLog("游戏结束，最高排名：" + ps[0].Name)
}
