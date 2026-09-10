package game

import (
	"fmt"
)

var spaceBonuses = map[string]string{"draw_vine": "+1 vine", "draw_order": "+1 order", "tour": "+1 coin", "build": "discount 1", "plant": "plant up to 2", "harvest": "harvest up to 2", "make_wine": "make up to 3", "train": "discount 1", "fill_order": "+1 VP", "sell_grapes": "+1 VP", "summer_visitor": "play up to 2", "winter_visitor": "play up to 2"}

// Slot 1 is the printed bonus, available only with 3+ players. Two-player slot 1 has no bonus.
// Slot -1 is grande overflow and never blocks or grants a printed bonus.
func placement(s *Space, p *Player, a Action) (int, bool, error) {
	if s.ID == "gain_coin" {
		return -1, false, nil
	}
	if s.ID == "yoke" {
		if !has(p, "yoke") || p.YokeUsed {
			return 0, false, fmt.Errorf("需要未使用的轭")
		}
		return 1, false, nil
	}
	occupied := map[int]bool{}
	for _, seat := range s.Occupied {
		if seat.Slot > 0 {
			occupied[seat.Slot] = true
		}
	}
	slot := a.Slot
	if slot == 0 {
		for i := 1; i <= s.Capacity; i++ {
			if !occupied[i] {
				slot = i
				break
			}
		}
		if slot == 0 && a.Large {
			slot = -1
		}
	}
	if slot == -1 {
		if !a.Large {
			return 0, false, fmt.Errorf("仅大工人可溢出")
		}
		for i := 1; i <= s.Capacity; i++ {
			if !occupied[i] {
				return 0, false, fmt.Errorf("有空格时大工人须占普通格")
			}
		}
		return -1, false, nil
	}
	if slot < 1 || slot > s.Capacity || occupied[slot] {
		return 0, false, fmt.Errorf("行动格已占用或不可用")
	}
	return slot, slot == 1 && s.Capacity >= 2, nil
}
func (r *Room) performPlacement(p *Player, a Action, bonus bool) error {
	bonus = bonus && !a.DeclineBonus
	switch a.Space {
	case "summer_visitor", "winter_visitor":
		return r.startVisitor(p, a, bonus)
	case "gain_coin":
		p.Coins++
		return nil
	case "yoke":
		if a.Mode == "uproot" {
			if e := uproot(p, a); e != nil {
				return e
			}
		} else if a.Mode == "harvest" {
			b := a
			b.Space = "harvest"
			if e := r.perform(p, b); e != nil {
				return e
			}
		} else {
			return fmt.Errorf("轭须选harvest或uproot")
		}
		p.YokeUsed = true
		return nil
	case "sell_grapes":
		if a.Mode == "buy_field" || a.Mode == "sell_field" {
			if a.Field < 0 || a.Field >= len(p.Fields) {
				return fmt.Errorf("无效田地")
			}
			f := &p.Fields[a.Field]
			if a.Mode == "sell_field" {
				if f.Sold || len(f.Vines) > 0 {
					return fmt.Errorf("仅能卖出未种植且未出售田地")
				}
				f.Sold = true
				p.Coins += f.Capacity
			} else {
				if !f.Sold || p.Coins < f.Capacity {
					return fmt.Errorf("田地未出售或金币不足")
				}
				f.Sold = false
				p.Coins -= f.Capacity
			}
			if bonus {
				p.VP++
			}
			return nil
		}
	case "plant":
		if a.Mode == "uproot" {
			return fmt.Errorf("种植行动不能拔藤；请使用轭或允许拔藤的访客")
		}
		if len(a.CardIDs) > 0 {
			limit := 1
			if bonus {
				limit = 2
			}
			if len(a.CardIDs) > limit || len(a.Fields) != len(a.CardIDs) {
				return fmt.Errorf("种植数量或田地数量无效")
			}
			for i, id := range a.CardIDs {
				b := a
				b.CardID = id
				b.Field = a.Fields[i]
				if e := r.perform(p, b); e != nil {
					return e
				}
			}
			return nil
		}
	case "harvest":
		if len(a.Fields) > 0 {
			limit := 1
			if bonus {
				limit = 2
			}
			if len(a.Fields) > limit {
				return fmt.Errorf("收获数量超过上限")
			}
			for _, f := range a.Fields {
				b := a
				b.Field = f
				if e := r.perform(p, b); e != nil {
					return e
				}
			}
			return nil
		}
	case "make_wine":
		limit := 2
		if bonus {
			limit = 3
		}
		if len(a.Recipes) == 0 {
			a.Recipes = [][]int{a.Grapes}
		}
		return makeWines(p, a.Recipes, limit)
	}
	// Grant discounts before the validated build/train; HTTP applies this only on a copy.
	if bonus && (a.Space == "build" || a.Space == "train") {
		p.Coins++
	}
	if e := r.perform(p, a); e != nil {
		return e
	}
	if bonus {
		switch a.Space {
		case "draw_vine":
			r.draw(p, "vine")
		case "draw_order":
			r.draw(p, "order")
		case "tour":
			p.Coins++
		case "fill_order", "sell_grapes":
			p.VP++
		}
	}
	return nil
}
func uproot(p *Player, a Action) error {
	if a.Field < 0 || a.Field >= len(p.Fields) {
		return fmt.Errorf("无效田地")
	}
	f := &p.Fields[a.Field]
	for i, c := range f.Vines {
		if c.ID == a.CardID {
			p.Hand = append(p.Hand, c)
			f.Vines = append(f.Vines[:i], f.Vines[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("请选择该田地的葡萄藤")
}
