package game

import (
	"fmt"
	"strings"
)

func init() {
	add := func(n int, text string, opts ...string) {
		visitorDefs[fmt.Sprintf("winter-%02d", n)] = visitorDef{text, opts}
	}
	add(1, "付3金币得红白品质1葡萄各1或交订单额外1分", "buy", "fill")
	add(2, "得3金币抽1夏访客或抽1订单酿至多2酒", "coins", "make")
	add(3, "抽2夏访客或弃品质至少4的酒得3分", "draw", "discard")
	add(4, "全部酒陈酿两次或付3金币升级酒窖一级", "age", "upgrade")
	add(5, "抽2夏访客及1金币或交订单额外1分", "draw", "fill")
	add(6, "得3金币抽1订单或酿至多3酒", "coins", "make")
	add(7, "失1分免费培训或每个拥有6工人的对手得1分", "train", "vp")
	add(8, "酿至多2酒或付2金币培训", "make", "train")
	add(9, "抽1藤及1夏访客或弃2访客得2分", "draw", "discard")
	add(10, "每张剩余手牌得1金币或弃全部手牌(至少1张)得2分", "coins", "discard")
	add(12, "收获至多2田并得2金币或1分", "coins", "vp")
	add(13, "付2金币培训或已有6工人得2分", "train", "vp")
	add(14, "升级酒窖减2金币或陈酿1酒后交订单", "upgrade", "fill")
	add(15, "全部酒陈酿两次或失1分免费升级酒窖", "age", "upgrade")
	add(18, "收获1田并抽1藤，或付1金币建轭", "harvest", "build")
	add(21, "负分得6金币，否则抽藤、夏访客、订单各1", "resolve")
	add(22, "酿至多2酒，每瓶起泡酒得1分", "make")
	add(23, "抽2订单或付3金币培训，失1分可都做", "draw", "train", "both", "draw_train", "train_draw")
	add(24, "收获至多3田，收获3田得2分", "harvest")
	add(26, "酿至多3酒，每种酿出的酒得1分", "make")
	add(27, "选两项：抽1订单、原价升级酒窖、得1分", "draw_upgrade", "upgrade_draw", "draw_vp", "vp_draw", "upgrade_vp", "vp_upgrade")
	add(28, "酿至多2酒或交订单或弃1葡萄得2分", "make", "fill", "discard")
	add(30, "原价建1建筑，建筑至少6座得2分", "build")
	add(31, "付3金币培训立即可用或弃1酒得2分", "train", "discard")
	add(33, "酿至多2瓶品质至少4的酒，无视酒窖限制", "make")
	add(34, "付1金币得1收入或失2收入得2分", "buy", "sell")
	add(36, "弃1酒得4金币，严格高于所有其他酒则得2分", "discard")
}
func (r *Room) visitorWinterEffect(p *Player, s *VisitorStep, a Action) error {
	draw := func(t string, n int) {
		for i := 0; i < n; i++ {
			r.draw(p, t)
		}
	}
	fill := func(bonus int) error {
		a.Space = "fill_order"
		if e := r.perform(p, a); e != nil {
			return e
		}
		p.VP += bonus
		return nil
	}
	switch s.CardID {
	case "winter-01":
		if a.Option == "buy" {
			if e := visitorPay(p, 3, 0); e != nil {
				return e
			}
			addGrape(p, "red", 1)
			addGrape(p, "white", 1)
		} else {
			return fill(1)
		}
	case "winter-02":
		if s.Stage == "make" {
			return r.visitorMake(p, a, 2)
		}
		if a.Option == "coins" {
			p.Coins += 3
			draw("summer", 1)
		} else {
			draw("order", 1)
			r.visitorPrompt(p, s.CardID, "make", []string{"make"})
		}
	case "winter-03":
		if a.Option == "draw" {
			draw("summer", 2)
		} else {
			if e := visitorWine(p, a.WineIDs, 4); e != nil {
				return e
			}
			p.VP += 3
		}
	case "winter-04":
		if a.Option == "age" {
			ageWines(p)
			ageWines(p)
		} else {
			if e := visitorPay(p, 3, 0); e != nil {
				return e
			}
			return r.visitorUpgrade(p, 99)
		}
	case "winter-05":
		if a.Option == "draw" {
			draw("summer", 2)
			p.Coins++
		} else {
			return fill(1)
		}
	case "winter-06":
		if a.Option == "coins" {
			p.Coins += 3
			draw("order", 1)
		} else {
			return r.visitorMake(p, a, 3)
		}
	case "winter-07":
		if a.Option == "train" {
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
			return r.visitorTrain(p, 0, false, a.SpecialWorker)
		}
		for _, q := range r.Players {
			if q != p && q.TotalWorkers == 6 {
				p.VP++
			}
		}
	case "winter-08":
		if a.Option == "make" {
			return r.visitorMake(p, a, 2)
		}
		return r.visitorTrain(p, 2, false, a.SpecialWorker)
	case "winter-09":
		if a.Option == "draw" {
			draw("vine", 1)
			draw("summer", 1)
		} else {
			if e := r.visitorCards(p, a.CardIDs, 2, true); e != nil {
				return e
			}
			p.VP += 2
		}
	case "winter-10":
		if a.Option == "coins" {
			p.Coins += len(p.Hand)
		} else {
			if len(p.Hand) == 0 {
				return fmt.Errorf("至少1手牌")
			}
			for _, c := range p.Hand {
				r.discard(c)
			}
			p.Hand = nil
			p.VP += 2
		}
	case "winter-12":
		if e := r.visitorHarvest(p, a, 2); e != nil {
			return e
		}
		if a.Option == "coins" {
			p.Coins += 2
		} else {
			p.VP++
		}
	case "winter-13":
		if a.Option == "train" {
			return r.visitorTrain(p, 2, false, a.SpecialWorker)
		}
		if p.TotalWorkers != 6 {
			return fmt.Errorf("须已有6工人")
		}
		p.VP += 2
	case "winter-14":
		if s.Stage == "fill" {
			return fill(0)
		}
		if a.Option == "upgrade" {
			return r.visitorUpgrade(p, 2)
		}
		if len(a.WineIDs) != 1 {
			return fmt.Errorf("请选择1酒陈酿")
		}
		found := false
		for i, w := range p.Wines {
			if w.ID == a.WineIDs[0] {
				found = true
				if !wineAgeBoundary(p, w.Value) {
					return fmt.Errorf("酒不能陈酿")
				}
				for _, v := range p.Wines {
					if v.Type == w.Type && v.Value == w.Value+1 {
						return fmt.Errorf("酒槽占用")
					}
				}
				p.Wines[i].Value++
			}
		}
		if !found {
			return fmt.Errorf("无效酒")
		}
		r.visitorPrompt(p, s.CardID, "fill", []string{"fill"})
	case "winter-15":
		if a.Option == "age" {
			ageWines(p)
			ageWines(p)
		} else {
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
			return r.visitorUpgrade(p, 99)
		}
	case "winter-18":
		if a.Option == "build" {
			return r.visitorBuild(p, "yoke", 1)
		}
		a.Fields = []int{a.Field}
		if e := r.visitorHarvest(p, a, 1); e != nil {
			return e
		}
		draw("vine", 1)
	case "winter-21":
		if p.VP < 0 {
			p.Coins += 6
		} else {
			draw("vine", 1)
			draw("summer", 1)
			draw("order", 1)
		}
	case "winter-22", "winter-26", "winter-33":
		n := 2
		if s.CardID == "winter-26" {
			n = 3
		}
		old := len(p.Wines)
		buildings := append([]string{}, p.Buildings...)
		if s.CardID == "winter-33" && !has(p, "large_cellar") {
			p.Buildings = append(p.Buildings, "large_cellar")
		}
		if e := r.visitorMake(p, a, n); e != nil {
			return e
		}
		p.Buildings = buildings
		types := map[string]bool{}
		for _, w := range p.Wines[old:] {
			types[w.Type] = true
			if s.CardID == "winter-22" && w.Type == "sparkling" {
				p.VP++
			}
			if s.CardID == "winter-33" && w.Value < 4 {
				return fmt.Errorf("必须酿出品质至少4的酒")
			}
		}
		if s.CardID == "winter-26" {
			p.VP += len(types)
		}
	case "winter-23":
		if s.Stage == "train" {
			return r.visitorTrain(p, 3, false, a.SpecialWorker)
		}
		if s.Stage == "draw" {
			draw("order", 2)
			return nil
		}
		if a.Option == "both" || a.Option == "draw_train" || a.Option == "train_draw" {
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
			if a.Option == "train_draw" {
				if e := r.visitorTrain(p, 3, false, a.SpecialWorker); e != nil {
					return e
				}
				r.visitorPrompt(p, s.CardID, "draw", []string{"draw"})
			} else {
				draw("order", 2)
				r.visitorPrompt(p, s.CardID, "train", []string{"train"})
			}
		} else if a.Option == "train" {
			return r.visitorTrain(p, 3, false, a.SpecialWorker)
		} else {
			draw("order", 2)
		}
	case "winter-24":
		if e := r.visitorHarvest(p, a, 3); e != nil {
			return e
		}
		if len(a.Fields) == 3 {
			p.VP += 2
		}
	case "winter-27":
		parts := strings.Split(a.Option, "_")
		perform := func(option string) error {
			switch option {
			case "draw":
				draw("order", 1)
			case "vp":
				p.VP++
			case "upgrade":
				return r.visitorUpgrade(p, 0)
			}
			return nil
		}
		if e := perform(parts[0]); e != nil {
			return e
		}
		if len(parts) == 2 {
			r.visitorPrompt(p, s.CardID, "continuation", []string{parts[1]})
		}
	case "winter-28":
		switch a.Option {
		case "make":
			return r.visitorMake(p, a, 2)
		case "fill":
			return fill(0)
		case "discard":
			if e := visitorGrape(p, a.Grapes); e != nil {
				return e
			}
			p.VP += 2
		}
	case "winter-30":
		if e := r.visitorBuild(p, a.Building, 0); e != nil {
			return e
		}
		if len(p.Buildings) >= 6 {
			p.VP += 2
		}
	case "winter-31":
		if a.Option == "train" {
			return r.visitorTrain(p, 3, true, a.SpecialWorker)
		}
		if e := visitorWine(p, a.WineIDs, 1); e != nil {
			return e
		}
		p.VP += 2
	case "winter-34":
		if a.Option == "buy" {
			if e := visitorPay(p, 1, 0); e != nil {
				return e
			}
			if p.Income < 5 {
				p.Income++
			}
		} else {
			if p.Income < 2 {
				return fmt.Errorf("收入不足")
			}
			p.Income -= 2
			p.VP += 2
		}
	case "winter-36":
		if len(a.WineIDs) != 1 {
			return fmt.Errorf("请选择1酒")
		}
		value := 0
		for _, w := range p.Wines {
			if w.ID == a.WineIDs[0] {
				value = w.Value
			}
		}
		best := true
		for _, q := range r.Players {
			for _, w := range q.Wines {
				if w.ID != a.WineIDs[0] && w.Value >= value {
					best = false
				}
			}
		}
		if e := visitorWine(p, a.WineIDs, 1); e != nil {
			return e
		}
		p.Coins += 4
		if best {
			p.VP += 2
		}
	default:
		return r.visitorInteraction(p, s, a)
	}
	return nil
}
