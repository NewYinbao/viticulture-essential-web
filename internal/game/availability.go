package game

import (
	"fmt"
	"strings"
)

// Hints are conservative, read-only prerequisites. They never execute an action,
// draw a card, or predict a random result; Apply remains authoritative.
func (r *Room) visitorOptionReason(p *Player, s *VisitorStep, option string) string {
	if option == "skip" || s.Stage == "reward" {
		return ""
	}
	if e := r.visitorBeforeEffect(p, s, Action{Option: option}); e != nil {
		return e.Error()
	}
	id := s.CardID
	if s.Stage == "effect" && id == "summer-28" && (option == "both" || strings.Contains(option, "_")) {
		if p.VP <= -5 {
			return "支付后分数会低于 −5"
		}
		possible := false
		if option == "plant_build" {
			possible = r.canPlantVisitor(p, false, false) && r.canBuildVisitor(p, 3)
		} else {
			for b := range costs {
				q, ok := r.probeAction(p, func(r *Room, q *Player) error { return r.visitorBuild(q, b, 3) })
				if ok && r.canPlantVisitor(q, false, false) {
					possible = true
					break
				}
			}
		}
		if !possible {
			return "无法完成建造与种植两项"
		}
	}
	if s.Stage == "effect" && id == "winter-23" && (option == "both" || strings.Contains(option, "train")) {
		if p.TotalWorkers >= 6 {
			return "工人已达 6 名上限"
		}
		if p.Coins < 3 {
			return fmt.Sprintf("还差 %d 金币", 3-p.Coins)
		}
	}
	if s.Stage == "effect" && (id == "summer-05" || id == "summer-22" || id == "winter-20" || id == "winter-29") {
		if !r.canSequence(p, s, option) {
			return "无法完成所选的全部步骤"
		}
		return ""
	}
	// Payments that precede an effect. Costs already paid by a previous stage
	// must not be charged again in a continuation hint.
	if s.Stage == "effect" {
		cash := map[string]int{"summer-02/buy": 9, "summer-08/buy": 6, "summer-10/buy": 2, "summer-16/draw": 4, "summer-32/retrieve": 2, "summer-36/build": 8, "summer-38/pay": 2, "winter-01/buy": 3, "winter-34/buy": 1}[id+"/"+option]
		if p.Coins < cash {
			return fmt.Sprintf("还差 %d 金币", cash-p.Coins)
		}
		loss := map[string]int{"summer-02/sell": 2, "summer-08/sell": 3, "summer-13/small": 1, "summer-13/any": 2, "summer-34/both": 1, "winter-07/train": 1, "winter-15/upgrade": 1}[id+"/"+option]
		if (id == "summer-28" || id == "winter-23") && (option == "both" || strings.Contains(option, "_")) {
			loss = 1
		}
		if p.VP-loss < -5 {
			return "支付后分数会低于 −5"
		}
	}
	if id == "winter-13" && option == "vp" && p.TotalWorkers != 6 {
		return "需要拥有 6 名工人"
	}
	if id == "winter-34" && option == "sell" && p.Income < 2 {
		return "需要至少 2 点年收入"
	}
	if option == "train" {
		if p.TotalWorkers >= 6 {
			return "工人已达 6 名上限"
		}
		price, ok := map[string]int{"winter-07": 0, "winter-08": 2, "winter-13": 2, "winter-23": 3, "winter-31": 3}[id]
		if ok && p.Coins < price {
			return fmt.Sprintf("还差 %d 金币", price-p.Coins)
		}
	}
	if option == "upgrade" && (id == "winter-04" || id == "winter-14" || id == "winter-15" || id == "winter-27") {
		if has(p, "large_cellar") {
			return "酒窖已达最高等级"
		}
		price := 4
		if has(p, "medium_cellar") {
			price = 6
		}
		switch id {
		case "winter-04":
			price = 3
		case "winter-14":
			price -= 2
		case "winter-15":
			price = 0
		}
		if p.Coins < price {
			return fmt.Sprintf("还差 %d 金币", price-p.Coins)
		}
	}
	if option == "plant" {
		// Horticulturist allows planting zero after its mandatory draw.
		if id == "summer-11" && s.Stage == "plant" {
			return ""
		}
		if !r.canPlantVisitor(p, id == "summer-19" || id == "summer-24", id == "summer-27") {
			return "无可种藤：检查建筑与田地容量"
		}
	}
	if option == "build" {
		if id == "winter-18" {
			if has(p, "yoke") {
				return "已经建有轭"
			}
			if p.Coins < 1 {
				return "还差 1 金币"
			}
			return ""
		}
		discount := map[string]int{"summer-04": 2, "summer-12": 3, "summer-18": 2, "summer-28": 3, "summer-35": 1, "summer-36": 99}[id]
		if !r.canBuildVisitor(p, discount) {
			return "无可建建筑：检查金币与前置酒窖"
		}
	}
	if option == "harvest" || id == "winter-12" {
		if !r.canSubaction(p, s, "harvest") {
			return "没有可收获的田地"
		}
	}
	if option == "make" && !r.canMakeVisitor(p, id == "winter-33", map[bool]int{true: 4, false: 0}[id == "winter-33"]) {
		return "无可酿配方：检查葡萄与酒窖"
	}
	if option == "fill" && id != "winter-14" && !r.canFillVisitor(p) {
		return "现有酒无法满足订单"
	}
	if option == "fill" && id == "winter-14" && s.Stage == "fill" && !r.canFillVisitor(p) {
		return "现有酒无法满足订单"
	}
	if option == "uproot" {
		n := 0
		for _, f := range p.Fields {
			if !f.Sold {
				n += len(f.Vines)
			}
		}
		need := 1
		if id == "summer-19" {
			need = 2
		}
		if n < need {
			return fmt.Sprintf("田上至少需要 %d 张藤", need)
		}
	}
	if option == "retrieve" && id == "summer-32" && !r.hasRetrievableWorker(p) {
		return "没有可收回的已放置工人"
	}
	return ""
}

func (r *Room) cardReason(p *Player, c Card) string {
	switch c.Type {
	case "vine":
		if c.Trellis && !has(p, "trellis") {
			return "需要棚架"
		}
		if c.Irrigation && !has(p, "irrigation") {
			return "需要灌溉"
		}
		q := probePlayer(p)
		q.Hand = []Card{c}
		if !r.canPlantVisitor(q, false, false) {
			return "没有足够容量的田地"
		}
	case "order":
		q := probePlayer(p)
		q.Hand = []Card{c}
		if !r.canFillVisitor(q) {
			return "现有酒未满足订单"
		}
	case "summer", "winter":
		if c.ID == "summer-29" {
			workers := p.Workers
			if p.LargeWorker {
				workers++
			}
			if workers < 2 {
				return "出牌后还需 1 名工人预约"
			}
		}
		if e := r.mandatoryVisitorPlayable(p, c.ID); e != nil {
			return "没有可完成的效果分支"
		}
		if e := r.specialPlayable(p, c.ID); e != nil {
			return e.Error()
		}
	}
	return ""
}

func (r *Room) availability(id string) (map[string]string, map[string]string, map[string]string) {
	actions, cards, options := map[string]string{}, map[string]string{}, map[string]string{}
	p := r.Player(id)
	if p == nil {
		return actions, cards, options
	}
	if len(r.Choices) > 0 {
		c := r.Choices[0]
		if c.PlayerID == id && c.Visitor != nil {
			for _, o := range c.Options {
				options[o] = r.visitorOptionReason(p, c.Visitor, o)
			}
		}
		return actions, cards, options
	}
	if r.TurnID != id || (r.Phase != "summer" && r.Phase != "winter") {
		return actions, cards, options
	}
	for _, c := range p.Hand {
		cards[c.ID] = r.cardReason(p, c)
	}
	for i := range r.Spaces {
		s := &r.Spaces[i]
		if s.Season != r.Phase && s.Season != "any" {
			continue
		}
		reason := ""
		large := p.Workers == 0
		if _, _, err := placement(s, p, Action{Large: large}); err != nil {
			large = true
		}
		_, bonus, err := placement(s, p, Action{Large: large})
		if p.Workers == 0 && !p.LargeWorker {
			reason = "没有待命工人"
		} else if err != nil {
			reason = err.Error()
		} else if large && !p.LargeWorker {
			reason = "普通格已满，大工人已使用"
		}
		if reason == "" {
			discount := 0
			if bonus {
				discount = 1
			}
			switch s.ID {
			case "build":
				if !r.canBuildVisitor(p, discount) {
					reason = "没有可负担的未建建筑"
				}
			case "train":
				if p.TotalWorkers >= 6 {
					reason = "工人已达 6 名上限"
				} else if p.Coins < 4-discount {
					reason = fmt.Sprintf("还差 %d 金币", 4-discount-p.Coins)
				}
			case "plant":
				if !r.canPlantVisitor(p, false, false) {
					reason = "没有可种植的葡萄藤"
				}
			case "harvest":
				if !r.canSubaction(p, &VisitorStep{}, "harvest") {
					reason = "没有可收获的田地"
				}
			case "make_wine":
				if !r.canMakeVisitor(p, false, 0) {
					reason = "没有可酿配方"
				}
			case "fill_order":
				if !r.canFillVisitor(p) {
					reason = "现有酒未满足订单"
				}
			case "summer_visitor", "winter_visitor":
				typ := strings.TrimSuffix(s.ID, "_visitor")
				found := false
				for _, c := range p.Hand {
					if c.Type == typ && cards[c.ID] == "" {
						found = true
					}
				}
				if !found {
					reason = "没有可打出的当季访客"
				}
			}
		}
		actions[s.ID] = reason
	}
	return actions, cards, options
}
