package game

import (
	"fmt"
)

func init() {
	add := func(n int, text string, opts ...string) {
		visitorDefs[fmt.Sprintf("summer-%02d", n)] = visitorDef{text, opts}
	}
	add(1, "每块自有空田得2金币或每块已种田得1分", "coins", "vp")
	add(2, "付9金币得3分或失2分得6金币", "buy", "sell")
	add(3, "抽2冬访客或弃品质至少7的酒得4分", "draw", "discard")
	add(4, "建造减2金币，原价5/6建筑额外1分", "build")
	add(6, "得4金币或收获1田", "coins", "harvest")
	add(7, "得3金币或酿至多2酒", "coins", "make")
	add(8, "失3分得9金币或付6金币得2分", "sell", "buy")
	add(9, "种至多2藤得1金币或拔除弃置1藤得2分", "plant", "uproot")
	add(10, "付2金币得品质1葡萄或弃1葡萄得2金币及1分", "buy", "discard")
	add(12, "建造减3金币或每座原价4建筑得1分", "build", "vp")
	add(13, "失1分免费建原价2/3建筑或失2分免费建任意建筑", "small", "any")
	add(14, "得4金币或抽1订单及1冬访客", "coins", "draw")
	add(15, "弃2手牌得4金币或弃4手牌得3分", "coins", "vp")
	add(16, "付4金币抽3冬访客或弃1酒及3访客得3分", "draw", "discard")
	add(19, "无视建筑种1藤或拔除弃置2藤得3分", "plant", "uproot")
	add(20, "弃2手牌抽四类各1", "exchange")
	add(24, "无视建筑种1藤或拔除弃置1藤得2分", "plant", "uproot")
	add(25, "种1藤，总共至少6藤得2分", "plant")
	add(26, "弃1葡萄得1收入或弃1酒得2收入", "grape", "wine")
	add(27, "种1藤可超容量，仍需建筑", "plant")
	add(28, "建造减3金币或种至多2藤；失1分可依次做两项", "build", "plant", "both", "build_plant", "plant_build")
	add(30, "种1藤，该田至少3不同藤种得2分", "plant")
	add(34, "抽2藤或得3金币，失1分可做两项", "draw", "coins", "both")
	add(35, "得3金币或建造减1金币或种至多2藤", "coins", "build", "plant")
	add(36, "付8金币免费建2座建筑", "build")
	add(38, "向至多3个不同对手各付2金币，每个得1分", "pay")
}
func (r *Room) visitorEffect(p *Player, s *VisitorStep, a Action) error {
	if s.CardID == "summer-29" || s.CardID == "summer-33" || s.CardID == "winter-11" {
		return r.visitorSpecial(p, s, a)
	}
	draw := func(t string, n int) {
		for i := 0; i < n; i++ {
			r.draw(p, t)
		}
	}
	switch s.CardID {
	case "summer-01":
		for _, f := range p.Fields {
			if !f.Sold {
				if a.Option == "coins" && len(f.Vines) == 0 {
					p.Coins += 2
				}
				if a.Option == "vp" && len(f.Vines) > 0 {
					p.VP++
				}
			}
		}
	case "summer-02":
		if a.Option == "buy" {
			if e := visitorPay(p, 9, 0); e != nil {
				return e
			}
			p.VP += 3
		} else {
			if e := visitorPay(p, 0, 2); e != nil {
				return e
			}
			p.Coins += 6
		}
	case "summer-03":
		if a.Option == "draw" {
			draw("winter", 2)
		} else {
			if e := visitorWine(p, a.WineIDs, 7); e != nil {
				return e
			}
			p.VP += 4
		}
	case "summer-04":
		if e := r.visitorBuild(p, a.Building, 2); e != nil {
			return e
		}
		if costs[a.Building] >= 5 {
			p.VP++
		}
	case "summer-06":
		if a.Option == "coins" {
			p.Coins += 4
		} else {
			a.Fields = []int{a.Field}
			return r.visitorHarvest(p, a, 1)
		}
	case "summer-07":
		if a.Option == "coins" {
			p.Coins += 3
		} else {
			return r.visitorMake(p, a, 2)
		}
	case "summer-08":
		if a.Option == "buy" {
			if e := visitorPay(p, 6, 0); e != nil {
				return e
			}
			p.VP += 2
		} else {
			if e := visitorPay(p, 0, 3); e != nil {
				return e
			}
			p.Coins += 9
		}
	case "summer-09":
		if a.Option == "plant" {
			if e := r.visitorPlant(p, a, 2, false, false); e != nil {
				return e
			}
			p.Coins++
		} else {
			if e := r.visitorUproot(p, a, 1); e != nil {
				return e
			}
			p.VP += 2
		}
	case "summer-10":
		if a.Option == "buy" {
			if a.Color != "red" && a.Color != "white" {
				return fmt.Errorf("须选red/white")
			}
			if e := visitorPay(p, 2, 0); e != nil {
				return e
			}
			addGrape(p, a.Color, 1)
		} else {
			if e := visitorGrape(p, a.Grapes); e != nil {
				return e
			}
			p.Coins += 2
			p.VP++
		}
	case "summer-12":
		if a.Option == "build" {
			return r.visitorBuild(p, a.Building, 3)
		}
		for _, b := range p.Buildings {
			if costs[b] == 4 {
				p.VP++
			}
		}
	case "summer-13":
		loss := 2
		if a.Option == "small" {
			loss = 1
			if costs[a.Building] != 2 && costs[a.Building] != 3 {
				return fmt.Errorf("须为原价2或3建筑")
			}
		}
		if e := visitorPay(p, 0, loss); e != nil {
			return e
		}
		return r.visitorBuild(p, a.Building, 99)
	case "summer-14":
		if a.Option == "coins" {
			p.Coins += 4
		} else {
			draw("order", 1)
			draw("winter", 1)
		}
	case "summer-15":
		n := 2
		if a.Option == "vp" {
			n = 4
		}
		if e := r.visitorCards(p, a.CardIDs, n, false); e != nil {
			return e
		}
		if n == 2 {
			p.Coins += 4
		} else {
			p.VP += 3
		}
	case "summer-16":
		if a.Option == "draw" {
			if e := visitorPay(p, 4, 0); e != nil {
				return e
			}
			draw("winter", 3)
		} else {
			if e := visitorWine(p, a.WineIDs, 1); e != nil {
				return e
			}
			if e := r.visitorCards(p, a.CardIDs, 3, true); e != nil {
				return e
			}
			p.VP += 3
		}
	case "summer-19", "summer-24":
		if a.Option == "plant" {
			if len(a.CardIDs) != 1 && a.CardID == "" {
				return fmt.Errorf("须种1藤")
			}
			return r.visitorPlant(p, a, 1, true, false)
		}
		n, vp := 1, 2
		if s.CardID == "summer-19" {
			n = 2
			vp = 3
		}
		if e := r.visitorUproot(p, a, n); e != nil {
			return e
		}
		p.VP += vp
	case "summer-20":
		if e := r.visitorCards(p, a.CardIDs, 2, false); e != nil {
			return e
		}
		for _, t := range []string{"vine", "summer", "order", "winter"} {
			draw(t, 1)
		}
	case "summer-25", "summer-30":
		if len(a.CardIDs) != 1 {
			return fmt.Errorf("须种1藤")
		}
		if e := r.visitorPlant(p, a, 1, false, false); e != nil {
			return e
		}
		if s.CardID == "summer-25" {
			n := 0
			for _, f := range p.Fields {
				n += len(f.Vines)
			}
			if n >= 6 {
				p.VP += 2
			}
		} else {
			names := map[string]bool{}
			for _, c := range p.Fields[a.Fields[0]].Vines {
				names[c.Name] = true
			}
			if len(names) >= 3 {
				p.VP += 2
			}
		}
	case "summer-26":
		gain := 1
		if a.Option == "grape" {
			if e := visitorGrape(p, a.Grapes); e != nil {
				return e
			}
		} else {
			if e := visitorWine(p, a.WineIDs, 1); e != nil {
				return e
			}
			gain = 2
		}
		p.Income += gain
		if p.Income > 5 {
			p.Income = 5
		}
	case "summer-27":
		if len(a.CardIDs) != 1 && a.CardID == "" {
			return fmt.Errorf("须种1藤")
		}
		return r.visitorPlant(p, a, 1, false, true)
	case "summer-28":
		if s.Stage == "plant" {
			return r.visitorPlant(p, a, 2, false, false)
		}
		if s.Stage == "build" {
			return r.visitorBuild(p, a.Building, 3)
		}
		if a.Option == "both" || a.Option == "build_plant" || a.Option == "plant_build" {
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
			if a.Option == "plant_build" {
				if e := r.visitorPlant(p, a, 2, false, false); e != nil {
					return e
				}
				r.visitorPrompt(p, s.CardID, "build", []string{"build"})
			} else {
				if e := r.visitorBuild(p, a.Building, 3); e != nil {
					return e
				}
				r.visitorPrompt(p, s.CardID, "plant", []string{"plant"})
			}
		} else if a.Option == "build" {
			return r.visitorBuild(p, a.Building, 3)
		} else {
			return r.visitorPlant(p, a, 2, false, false)
		}
	case "summer-34":
		if a.Option == "both" {
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
		}
		if a.Option != "coins" {
			draw("vine", 2)
		}
		if a.Option != "draw" {
			p.Coins += 3
		}
	case "summer-35":
		switch a.Option {
		case "coins":
			p.Coins += 3
		case "build":
			return r.visitorBuild(p, a.Building, 1)
		case "plant":
			return r.visitorPlant(p, a, 2, false, false)
		}
	case "summer-36":
		if len(a.Buildings) != 2 {
			return fmt.Errorf("须选择2建筑")
		}
		if e := visitorPay(p, 8, 0); e != nil {
			return e
		}
		for _, b := range a.Buildings {
			if e := r.visitorBuild(p, b, 99); e != nil {
				return e
			}
		}
	case "summer-38":
		if len(a.TargetIDs) == 0 || len(a.TargetIDs) > 3 {
			return fmt.Errorf("最多3个对手")
		}
		seen := map[string]bool{}
		for _, id := range a.TargetIDs {
			q := r.Player(id)
			if q == nil || q == p || seen[id] {
				return fmt.Errorf("无效或重复对手")
			}
			seen[id] = true
			if e := visitorPay(p, 2, 0); e != nil {
				return e
			}
			q.Coins += 2
			p.VP++
		}
	default:
		return r.visitorWinterEffect(p, s, a)
	}
	return nil
}
