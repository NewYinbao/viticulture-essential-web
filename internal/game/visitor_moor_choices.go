package game

import (
	"fmt"
	"strings"
)

func moorField(name, typ string, min, max int) map[string]any {
	return map[string]any{"name": name, "type": typ, "min": min, "max": max}
}
func (r *Room) decorateMoorChoice(c *Choice, p *Player) {
	if r.Config.Visitors != "ee_moor" || c.Visitor == nil || !isMoor(c.Visitor.CardID) {
		return
	}
	s := c.Visitor
	card := moorCard(s.CardID)
	c.Title = card.Name
	c.Description = card.Description
	c.OptionFields = map[string][]map[string]any{}
	c.Labels = map[string]string{}
	color := func() map[string]any { return moorField("color", "color", 1, 1) }
	grapes := func(n int) map[string]any { return moorField("grapes", "grapes", n, n) }
	wines := func(n int) map[string]any { return moorField("wineIds", "wines", n, n) }
	cards := func(n, max int) map[string]any { return moorField("cardIds", "cards", n, max) }
	fields := func(n, max int) map[string]any { return moorField("fields", "fields", n, max) }
	number := func() map[string]any {
		f := moorField("slot", "moorNumber", 1, 3)
		f["label"] = "陈酿次数"
		return f
	}
	building := func(n int, discount int) map[string]any {
		f := moorField("buildings", "moorBuildings", n, n)
		if n == 1 {
			f["name"] = "building"
		}
		f["discount"] = discount
		return f
	}
	order := func() []map[string]any {
		f := moorField("cardId", "cards", 1, 1)
		f["filter"] = "order"
		return []map[string]any{f, moorField("wineIds", "wines", 1, 4)}
	}
	plant := func() map[string]any { return moorField("plant", "plant", 1, 1) }
	training := func() map[string]any {
		f := moorField("specialWorker", "moorTraining", 1, 1)
		opts := []map[string]string{{"id": "regular", "label": "普通工人"}}
		if r.Config.SpecialWorkers {
			for _, id := range r.SpecialWorkerPool {
				if !p.hasSpecialWorker(id) {
					label := id
					for _, w := range SpecialWorkerCatalog() {
						if w.ID == id {
							label = w.Name
						}
					}
					opts = append(opts, map[string]string{"id": id, "label": label + "（额外1金币）"})
				}
			}
		}
		f["options"] = opts
		return f
	}
	labels := map[string]string{"build": "建造", "draw": "抽牌", "play": "选择并打出访客", "grape": "获得葡萄", "discard": "弃资源换分", "attach": "付4金币 · 放到田地", "offer": "邀请每位玩家", "coins": "获得金币", "vp": "获得分数", "harvest": "收获", "plant": "种植", "train": "培训工人", "wine": "获得葡萄酒", "age": "陈酿", "fill": "完成订单", "upgrade": "升级酒窖", "retrieve": "取回2名工人", "income": "增加收入", "both": "付1分 · 完成两项", "devalue": "3瓶酒各降级1 · +2分", "trade": "弃3抽3 · +2金币", "vine": "抽3张葡萄藤", "order": "抽3张订单", "keep": "保留所选牌", "all": "付1金币 · 保留3张", "resolve": "执行效果", "skip": "跳过"}
	for _, op := range c.Options {
		c.Labels[op] = labels[op]
		fs := []map[string]any{}
		if s.Stage == "private" {
			if op == "keep" {
				n := 1
				if s.CardID == "moor-summer-15" {
					n = 2
				}
				fs = append(fs, moorField("cardIds", "moorPrivateCards", n, n))
			}
			if op == "fill" {
				fs = append(fs, moorField("cardId", "moorPrivateCards", 1, 1), moorField("wineIds", "wines", 1, 4))
			}
			c.OptionFields[op] = fs
			continue
		}
		if s.Stage == "reward" {
			c.OptionFields[op] = fs
			continue
		}
		switch s.CardID {
		case "moor-summer-02", "moor-winter-12":
			f := moorField("cardId", "cards", 1, 1)
			f["filter"] = "winter"
			if s.CardID == "moor-winter-12" {
				f["filter"] = "summer"
			}
			fs = append(fs, f)
		case "moor-summer-03":
			if op == "grape" {
				fs = append(fs, color())
			} else {
				fs = append(fs, grapes(3))
			}
		case "moor-summer-04":
			fs = append(fs, moorField("field", "moorAnyField", 1, 1))
		case "moor-summer-05":
			if s.Stage == "reply" && op == "grape" {
				fs = append(fs, color())
			}
		case "moor-summer-07":
			if op == "discard" {
				fs = append(fs, grapes(2))
			}
		case "moor-summer-08":
			fs = append(fs, moorField("opponentField", "moorOpponentField", 1, 1))
		case "moor-summer-11":
			if op == "plant" {
				f := plant()
				f["irrigationOnly"] = true
				fs = append(fs, f)
			}
		case "moor-summer-12":
			if op == "build" {
				fs = append(fs, building(3, 99))
			}
		case "moor-summer-13", "moor-summer-14":
			if op == "grape" {
				fs = append(fs, color())
			}
		case "moor-summer-16":
			fs = append(fs, training())
		case "moor-summer-17":
			if op == "wine" {
				fs = append(fs, color())
			}
		case "moor-summer-18":
			max := len(p.Grapes)
			if op == "harvest" {
				max = len(p.Fields)
			}
			if len(p.Hand) < max {
				max = len(p.Hand)
			}
			fs = append(fs, cards(1, max))
			if op == "harvest" {
				fs = append(fs, fields(1, 3))
			} else {
				fs = append(fs, moorField("grapes", "grapes", 1, max))
			}
		case "moor-summer-19":
			fs = order()
		case "moor-winter-01":
			f := moorField("seats", "moorSeats", 2, 2)
			f["options"] = r.moorPreviousSeats(p)
			fs = append(fs, f)
		case "moor-winter-03":
			if op != "upgrade" {
				fs = append(fs, wines(1), number())
				if op == "both" {
					fs = append(fs, moorField("mode", "moorAgeOrder", 1, 1))
				}
			}
		case "moor-winter-06":
			if op == "fill" {
				fs = order()
			}
		case "moor-winter-07":
			fs = append(fs, building(1, 5))
		case "moor-winter-08":
			fs = append(fs, color())
			c.Labels[op] = "付" + strings.TrimPrefix(op, "buy_") + "金币 · 同价值葡萄"
		case "moor-winter-09":
			if op == "discard" {
				fs = append(fs, grapes(1), wines(1))
			}
		case "moor-winter-10":
			if r.Context.Space == "summer_visitor" {
				fs = append(fs, grapes(1))
				c.Labels[op] = "弃1颗葡萄 · +4金币"
			} else {
				fs = append(fs, color())
				c.Labels[op] = "付1金币 · 获得价值4葡萄"
			}
		case "moor-winter-11":
			fs = append(fs, grapes(1))
		case "moor-winter-13":
			if op == "plant" {
				fs = append(fs, plant())
			} else {
				fs = append(fs, wines(3))
			}
		case "moor-winter-15":
			fs = append(fs, cards(3, 3), moorField("colors", "moorDrawColors", 3, 3))
		case "moor-winter-16":
			if s.Stage == "age" {
				fs = append(fs, grapes(s.MoorCount))
			} else {
				fs = append(fs, fields(1, 3))
			}
		case "moor-winter-19":
			if op == "discard" {
				fs = append(fs, wines(1))
			} else {
				fs = append(fs, training())
			}
		case "moor-winter-20":
			if op == "wine" {
				fs = append(fs, color())
			} else {
				fs = append(fs, grapes(1), number())
			}
		}
		if c.Labels[op] == "" {
			c.Labels[op] = fmt.Sprint(op)
		}
		c.OptionFields[op] = fs
	}
}
