package game

import (
	"fmt"
	"strings"
)

func init() {
	add := func(id, text string, opts ...string) { visitorDefs[id] = visitorDef{text, opts} }
	add("summer-05", "三选二，顺序自选：得1分、原价建造、种1藤", "vp_build", "build_vp", "vp_plant", "plant_vp", "build_plant", "plant_build")
	add("summer-11", "抽1藤后可种1藤，或交换不同田地的两藤", "draw", "swap")
	add("summer-22", "原价建造与种1藤；价值4藤得1分，顺序自选", "build_plant", "plant_build")
	add("summer-23", "对手合计给3访客，否则退还提交并抽3冬访客", "resolve")
	add("summer-32", "付2金币取回其他行动上至多2工人", "retrieve")
	add("winter-19", "取两个不同弃牌堆顶牌，不能取回自身", "take")
	add("winter-20", "三选二，顺序自选：收获1田、酿至多2酒、交1订单", "harvest_make", "make_harvest", "harvest_fill", "fill_harvest", "make_fill", "fill_make")
	add("winter-29", "收获至多2田或酿至多3酒；失1分可都做", "harvest", "make", "harvest_make", "make_harvest")
	add("winter-32", "执行夏季1行动，不放工人且不拿格奖励", "action")
	add("winter-37", "公开四类牌顶各1，取2其余弃置", "reveal")
}

func (r *Room) visitorSubaction(p *Player, s *VisitorStep, option string, a Action) error {
	switch option {
	case "vp":
		p.VP++
	case "build":
		return r.visitorBuild(p, a.Building, 0)
	case "plant":
		if len(a.CardIDs) != 1 {
			return fmt.Errorf("须种1藤")
		}
		value := 0
		for _, c := range p.Hand {
			if c.ID == a.CardIDs[0] {
				value = c.Red + c.White
			}
		}
		if e := r.visitorPlant(p, a, 1, false, false); e != nil {
			return e
		}
		if s.CardID == "summer-22" && value == 4 {
			p.VP++
		}
	case "harvest":
		n := 1
		if s.CardID == "winter-29" {
			n = 2
		} else if len(a.Fields) != 1 {
			return fmt.Errorf("须收获1田")
		}
		return r.visitorHarvest(p, a, n)
	case "make":
		n := 2
		if s.CardID == "winter-29" {
			n = 3
		}
		return r.visitorMake(p, a, n)
	case "fill":
		a.Space = "fill_order"
		return r.perform(p, a)
	default:
		return fmt.Errorf("无效子动作")
	}
	return nil
}
func (r *Room) visitorSequence(p *Player, s *VisitorStep, a Action) error {
	switch s.CardID {
	case "summer-05", "summer-22", "winter-20", "winter-29":
		parts := strings.Split(a.Option, "_")
		if len(parts) == 2 {
			if s.CardID == "winter-29" {
				if e := visitorPay(p, 0, 1); e != nil {
					return e
				}
			}
			if e := r.visitorSubaction(p, s, parts[0], a); e != nil {
				return e
			}
			r.visitorPrompt(p, s.CardID, "continuation", []string{parts[1]})
			return nil
		}
		return r.visitorSubaction(p, s, a.Option, a)
	case "summer-11":
		if s.Stage == "plant" {
			if len(a.CardIDs) == 0 && a.CardID == "" {
				return nil
			} // Optional after hidden draw; never softlock.
			return r.visitorPlant(p, a, 1, false, false)
		}
		if a.Option == "draw" {
			r.draw(p, "vine")
			r.visitorPrompt(p, s.CardID, "plant", []string{"plant"})
			return nil
		}
		if len(a.Fields) != 2 || len(a.CardIDs) != 2 || a.Fields[0] == a.Fields[1] {
			return fmt.Errorf("须选择不同田地上的2藤")
		}
		indices := []int{-1, -1}
		for i, f := range a.Fields {
			if f < 0 || f >= len(p.Fields) || p.Fields[f].Sold {
				return fmt.Errorf("无效田地")
			}
			for j, c := range p.Fields[f].Vines {
				if c.ID == a.CardIDs[i] {
					indices[i] = j
				}
			}
			if indices[i] < 0 {
				return fmt.Errorf("藤不在田地")
			}
		}
		f, g := &p.Fields[a.Fields[0]], &p.Fields[a.Fields[1]]
		f.Vines[indices[0]], g.Vines[indices[1]] = g.Vines[indices[1]], f.Vines[indices[0]]
		for _, field := range []*Field{f, g} {
			total := 0
			for _, c := range field.Vines {
				total += c.Red + c.White
			}
			if total > field.Capacity {
				return fmt.Errorf("交换后超田地容量")
			}
		}
	case "summer-23":
		if s.Stage == "effect" {
			r.Context.Escrow = map[string][]Card{}
			r.visitorCycle(p, s, false, []string{"give", "skip"})
			r.visitorPrompt(p, s.CardID, "settle", []string{"resolve"})
			return nil
		}
		if s.Stage == "settle" {
			total := 0
			for _, cards := range r.Context.Escrow {
				total += len(cards)
			}
			for id, cards := range r.Context.Escrow {
				q := r.Player(id)
				if total == 3 {
					q = p
				}
				q.Hand = append(q.Hand, cards...)
			}
			if total != 3 {
				for i := 0; i < 3; i++ {
					r.draw(p, "winter")
				}
			}
			r.Context.Escrow = nil
			return nil
		}
		if a.Option == "skip" {
			return nil
		}
		total := 0
		for _, cards := range r.Context.Escrow {
			total += len(cards)
		}
		if len(a.CardIDs) == 0 || total+len(a.CardIDs) > 3 {
			return fmt.Errorf("合计最多3访客")
		}
		seen := map[string]bool{}
		for _, id := range a.CardIDs {
			i := cardIndex(p, id, "summer")
			if i < 0 {
				i = cardIndex(p, id, "winter")
			}
			if i < 0 || seen[id] {
				return fmt.Errorf("无效访客")
			}
			seen[id] = true
			// Empty escrow is omitted by JSON snapshots and HTTP candidate cloning.
			if r.Context.Escrow == nil {
				r.Context.Escrow = make(map[string][]Card)
			}
			r.Context.Escrow[p.ID] = append(r.Context.Escrow[p.ID], p.Hand[i])
			removeCard(p, i)
		}
	case "summer-32":
		// Parallel action/slot arrays. Persisted TriggerSeat identifies only the
		// actual outer placement, including Manager nesting and Planner execution.
		if len(a.TargetIDs) == 0 || len(a.TargetIDs) > 2 || len(a.Fields) != len(a.TargetIDs) {
			return fmt.Errorf("至多2个行动/槽位")
		}
		if e := visitorPay(p, 2, 0); e != nil {
			return e
		}
		for k, id := range a.TargetIDs {
			found := false
			for i := range r.Spaces {
				sp := &r.Spaces[i]
				if sp.ID != id {
					continue
				}
				for j, seat := range sp.Occupied {
					if seat.PlayerID == p.ID && seat.Slot == a.Fields[k] {
						if r.isTriggerSeat(id, seat) {
							return fmt.Errorf("不能取回打出此牌的工人")
						}
						r.releaseSeat(seat)
						r.cancelPlanned(sp.ID, seat)
						if sp.ID == "yoke" {
							p.YokeUsed = false
						}
						sp.Occupied = append(sp.Occupied[:j], sp.Occupied[j+1:]...)
						found = true
						break
					}
				}
			}
			if !found {
				return fmt.Errorf("无效自己的工人")
			}
		}
	case "winter-19":
		if len(a.Colors) != 2 || a.Colors[0] == a.Colors[1] {
			return fmt.Errorf("须选两个不同弃牌堆")
		}
		for _, typ := range a.Colors {
			if typ != "vine" && typ != "order" && typ != "summer" && typ != "winter" {
				return fmt.Errorf("无效弃牌堆")
			}
			d := r.Discards[typ]
			if len(d) == 0 {
				return fmt.Errorf("弃牌堆为空")
			}
			card := d[len(d)-1]
			if card.ID == s.CardID {
				return fmt.Errorf("不能取回自己")
			}
			p.Hand = append(p.Hand, card)
			r.Discards[typ] = d[:len(d)-1]
		}
		for _, card := range r.Context.Held {
			r.discard(card)
		}
		r.Context.Held = nil
	case "winter-32":
		allowed := false
		for _, sp := range r.Spaces {
			previous := sp.Season == "summer"
			if r.tuscany() {
				previous = seasonIndex(sp.Season) >= 0 && seasonIndex(sp.Season) < seasonIndex(r.Phase)
			}
			if sp.ID == a.Space && previous {
				allowed = true
			}
		}
		if !allowed {
			return fmt.Errorf("须选择夏季行动")
		}
		if a.Space == "summer_visitor" {
			if cardIndex(p, a.CardID, "summer") < 0 {
				return fmt.Errorf("须为自己的夏访客")
			}
			return r.queueVisitor(p, a.CardID)
		}
		return r.performPlacement(p, a, false)
	case "winter-37":
		if s.Stage == "effect" {
			for _, typ := range []string{"vine", "summer", "order", "winter"} {
				if c, ok := r.take(typ); ok {
					r.Revealed = append(r.Revealed, c)
				}
			}
			r.visitorPrompt(p, s.CardID, "take", []string{"take"})
			return nil
		}
		if len(a.CardIDs) != 2 || a.CardIDs[0] == a.CardIDs[1] {
			return fmt.Errorf("须取2张不同公开牌")
		}
		selected := map[string]bool{}
		for _, id := range a.CardIDs {
			found := false
			for _, c := range r.Revealed {
				if c.ID == id {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("须为本次公开牌")
			}
			selected[id] = true
		}
		for _, c := range r.Revealed {
			if selected[c.ID] {
				p.Hand = append(p.Hand, c)
			} else {
				r.discard(c)
			}
		}
		r.Revealed = nil
	default:
		return fmt.Errorf("EE_VISITOR_NOT_IMPLEMENTED: %s", s.CardID)
	}
	return nil
}
