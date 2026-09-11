package game

import (
	"fmt"
	"strconv"
	"strings"
)

func (r *Room) moorPlayable(p *Player, id string) error {
	q := probePlayer(p)
	if i := cardIndexIn(q.Hand, id); i >= 0 {
		removeCard(q, i)
	}
	for _, op := range visitorDefs[id].Options {
		if r.moorOptionReason(q, &VisitorStep{CardID: id, Stage: "effect", ActorID: p.ID}, op) == "" {
			return nil
		}
	}
	return fmt.Errorf("当前无法完成此访客任一效果")
}
func (r *Room) moorOptionReason(p *Player, s *VisitorStep, op string) string {
	if p == nil {
		return "玩家不存在"
	}
	if op == "skip" || s.Stage == "reward" {
		return ""
	}
	coin := func(n int) string {
		if p.Coins < n {
			return fmt.Sprintf("还差%d金币", n-p.Coins)
		}
		return ""
	}
	grape := func(n int) string {
		if len(p.Grapes) < n {
			return fmt.Sprintf("需要%d颗葡萄", n)
		}
		return ""
	}
	wine := func(n int) string {
		if len(p.Wines) < n {
			return fmt.Sprintf("需要%d瓶酒", n)
		}
		return ""
	}
	build := func(b string, cost int) string {
		if has(p, b) {
			return "已拥有该建筑"
		}
		return coin(cost)
	}
	upgrade := func(cost int) string {
		if has(p, "large_cellar") {
			return "酒窖已满级"
		}
		return coin(cost)
	}
	train := func(cost int) string {
		for _, q := range r.Players {
			if q.ID != p.ID && has(q, "academy") {
				cost++
			}
		}
		if p.TotalWorkers >= 6 {
			return "工人已达6名上限"
		}
		return coin(cost)
	}
	plant := func() string {
		if !r.canPlantVisitor(p, false, false) {
			return "没有满足田地与建筑条件的葡萄藤"
		}
		return ""
	}
	fill := func() string {
		if !r.canFillVisitor(p) {
			return "没有可完成的订单"
		}
		return ""
	}
	gain := func(value int, kind string) string {
		q := probePlayer(p)
		if kind != "" {
			if err := moorGainWine(q, kind, value, true); err != nil {
				return err.Error()
			}
			return ""
		}
		if moorGainGrape(q, "red", value) != nil && moorGainGrape(q, "white", value) != nil {
			return "葡萄槽已满"
		}
		return ""
	}
	age := func(n int) string {
		if !moorCanAgeGrapes(p, n) {
			return fmt.Sprintf("需要%d颗可陈酿葡萄", n)
		}
		return ""
	}
	harvest := func() string {
		for _, f := range p.Fields {
			if !f.Sold && !f.Harvested && len(f.Vines) > 0 {
				return ""
			}
		}
		return "没有可收获田地"
	}
	if s.Stage == "private" {
		if op == "all" {
			return coin(1)
		}
		if op == "fill" {
			q := probePlayer(p)
			q.Hand = append([]Card{}, r.Context.Held...)
			if !r.canFillVisitor(q) {
				return "新抽订单暂时无法完成"
			}
		}
		return ""
	}
	if s.Stage == "age" {
		return age(s.MoorCount)
	}
	switch s.CardID {
	case "moor-summer-01":
		if op == "build" {
			return build("yoke", 0)
		}
		if !has(p, "yoke") {
			return "需要轭架"
		}
	case "moor-summer-02", "moor-winter-12":
		kind := "winter"
		if s.CardID == "moor-winter-12" {
			kind = "summer"
		}
		for _, c := range p.Hand {
			if c.ID != s.CardID && visitorInSeason(p, c.ID, kind) && r.moorNestedPlayable(p, c.ID, kind) == nil {
				return ""
			}
		}
		return "没有可打出的对应季节访客"
	case "moor-summer-03":
		if op == "discard" {
			return grape(3)
		}
		return gain(2, "")
	case "moor-summer-04":
		if v := coin(4); v != "" {
			return v
		}
		for _, f := range p.Fields {
			if !f.Sold && !f.FruitDealer && f.Structure == "" {
				return ""
			}
		}
		return "没有可放置水果商的田地"
	case "moor-summer-05":
		if s.Stage == "reply" && op == "grape" {
			return gain(2, "")
		}
	case "moor-summer-07":
		if op == "discard" {
			return grape(2)
		}
	case "moor-summer-08":
		for _, q := range r.Players {
			if q.ID != p.ID {
				for _, f := range q.Fields {
					if !f.Sold && len(f.Vines) > 0 && f.Structure == "" {
						return ""
					}
				}
			}
		}
		return "对手没有可收获田地"
	case "moor-summer-11":
		if op == "build" {
			return build("irrigation", 0)
		}
		q := probePlayer(p)
		q.Hand = nil
		for _, c := range p.Hand {
			if c.Type == "vine" && c.Irrigation {
				q.Hand = append(q.Hand, c)
			}
		}
		if !r.canPlantVisitor(q, false, false) {
			return "需要可种植且要求灌溉的葡萄藤"
		}
	case "moor-summer-12":
		if op == "vp" {
			if moorBuildingCount(p) < 6 {
				return "需要至少6座建筑"
			}
		} else {
			if v := coin(14); v != "" {
				return v
			}
			n := 0
			for b := range costs {
				if !has(p, b) {
					n++
				}
			}
			if r.Config.Structures {
				for _, c := range p.Hand {
					if c.Type == "structure" {
						n++
					}
				}
			}
			if n < 3 {
				return "不足3座可建建筑"
			}
		}
	case "moor-summer-13":
		if op == "grape" {
			if !has(p, "windmill") {
				return "需要风车"
			}
			return gain(3, "")
		}
		return build("windmill", 2)
	case "moor-summer-14":
		if op == "vp" {
			return grape(5)
		}
		if v := coin(3); v != "" {
			return v
		}
		return gain(5, "")
	case "moor-summer-16":
		return train(6 - p.TotalWorkers)
	case "moor-summer-17":
		if op == "build" {
			return build("tasting_room", 3)
		}
		if !has(p, "tasting_room") {
			return "需要品酒室"
		}
		if gain(1, "red") != "" && gain(1, "white") != "" {
			return "价值1的红酒与白酒槽均已满"
		}
	case "moor-summer-18":
		if len(p.Hand) == 0 {
			return "需要弃掉至少1张手牌"
		}
		if op == "harvest" {
			return harvest()
		}
		return age(1)
	case "moor-summer-19":
		return fill()
	case "moor-summer-20":
		if op == "upgrade" {
			return upgrade(3)
		}
		if v := coin(3); v != "" {
			return v
		}
		return gain(1, "white")
	case "moor-winter-01":
		if len(r.moorPreviousSeats(p)) < 2 {
			return "先前季节不足2名可取回工人"
		}
	case "moor-winter-02":
		if v := coin(7); v != "" {
			return v
		}
		return gain(4, "blush")
	case "moor-winter-03":
		if op == "upgrade" {
			return upgrade(2)
		}
		if op == "both" {
			if v := upgrade(2); v != "" {
				return v
			}
			if p.VP <= -5 {
				return "无法支付1分"
			}
		}
		return wine(1)
	case "moor-winter-04":
		if op == "vp" {
			return coin(9)
		}
		return wine(1)
	case "moor-winter-05":
		if op == "vp" {
			return wine(1)
		}
	case "moor-winter-06":
		if op == "fill" {
			q := probePlayer(p)
			for i := range q.Hand {
				if q.Hand[i].Type == "order" {
					for j := range q.Hand[i].Requirements {
						q.Hand[i].Requirements[j].Type = "red"
					}
				}
			}
			for i := range q.Wines {
				q.Wines[i].Type = "red"
			}
			if !r.canFillVisitor(q) {
				return "酒的数量或价值不足以完成订单"
			}
		}
	case "moor-winter-07":
		if !r.canBuildVisitor(p, 5) {
			return "没有能以优惠5金币建造的建筑"
		}
	case "moor-winter-08":
		n, _ := strconv.Atoi(strings.TrimPrefix(op, "buy_"))
		if v := coin(n); v != "" {
			return v
		}
		return gain(n, "")
	case "moor-winter-09":
		if op == "discard" {
			if !has(p, "cottage") {
				return "需要小屋"
			}
			if v := grape(1); v != "" {
				return v
			}
			return wine(1)
		}
	case "moor-winter-10":
		if r.moorSummerPlay() {
			return grape(1)
		}
		if v := coin(1); v != "" {
			return v
		}
		return gain(4, "")
	case "moor-winter-11":
		return grape(1)
	case "moor-winter-13":
		if op == "plant" {
			return plant()
		}
		if len(p.Wines) < 3 {
			return "需要3瓶可降级的酒"
		}
		for i := range p.Wines {
			for j := i + 1; j < len(p.Wines); j++ {
				for k := j + 1; k < len(p.Wines); k++ {
					if moorDevalueWines(probePlayer(p), []string{p.Wines[i].ID, p.Wines[j].ID, p.Wines[k].ID}) == nil {
						return ""
					}
				}
			}
		}
		return "没有3瓶能同时降级的酒"
	case "moor-winter-14":
		if v := coin(9); v != "" {
			return v
		}
		return gain(7, "sparkling")
	case "moor-winter-15":
		if len(p.Hand) < 3 {
			return "需要弃掉3张手牌"
		}
	case "moor-winter-16":
		return harvest()
	case "moor-winter-18":
		if v := coin(9); v != "" {
			return v
		}
		if v := gain(2, "red"); v != "" {
			return v
		}
		return gain(2, "white")
	case "moor-winter-19":
		if op == "discard" {
			return wine(1)
		}
		if p.VP <= -5 {
			return "无法支付1分"
		}
		return train(0)
	case "moor-winter-20":
		if v := coin(3); v != "" {
			return v
		}
		if op == "age" {
			return age(1)
		}
		if gain(1, "red") != "" && gain(1, "white") != "" {
			return "价值1的红酒与白酒槽均已满"
		}
	}
	return ""
}
func moorBuildingCount(p *Player) int {
	// StructureSlots and Fields record locations of the same structures that
	// buildStructure also stores in Buildings; they are not extra buildings.
	seen := map[string]bool{}
	for _, id := range p.Buildings {
		if id != "" {
			seen[id] = true
		}
	}
	for _, s := range p.StructureSlots {
		if s != "" {
			seen[s] = true
		}
	}
	for _, f := range p.Fields {
		if f.Structure != "" {
			seen[f.Structure] = true
		}
	}
	return len(seen)
}
func (r *Room) moorPreviousSeats(p *Player) []map[string]any {
	out := []map[string]any{}
	for _, sp := range r.Spaces {
		for _, seat := range sp.Occupied {
			season := seat.Season
			if season == "" {
				season = sp.Season
			}
			previous := season == "summer" && r.Phase == "winter"
			if r.tuscany() {
				previous = seasonIndex(season) >= 0 && seasonIndex(season) < seasonIndex(r.Phase)
			}
			if seat.PlayerID == p.ID && previous && !r.isTriggerSeat(sp.ID, seat) {
				out = append(out, map[string]any{"space": sp.ID, "slot": seat.Slot, "label": sp.Name, "large": seat.Large, "workerType": seat.WorkerType})
			}
		}
	}
	return out
}
func (r *Room) moorRetrieve(p *Player, a Action) error {
	if len(a.TargetIDs) != 2 || len(a.Fields) != 2 {
		return fmt.Errorf("须取回2名先前季节的工人")
	}
	eligible := r.moorPreviousSeats(p)
	seen := map[string]bool{}
	for i, space := range a.TargetIDs {
		key := fmt.Sprintf("%s:%d", space, a.Fields[i])
		if seen[key] {
			return fmt.Errorf("工人不能重复")
		}
		seen[key] = true
		valid := false
		for _, seat := range eligible {
			if seat["space"] == space && seat["slot"] == a.Fields[i] {
				valid = true
			}
		}
		if !valid {
			return fmt.Errorf("只能取回自己在先前季节的工人")
		}
	}
	for i, space := range a.TargetIDs {
		for k := range r.Spaces {
			sp := &r.Spaces[k]
			if sp.ID != space {
				continue
			}
			for j, seat := range sp.Occupied {
				if seat.PlayerID != p.ID || seat.Slot != a.Fields[i] {
					continue
				}
				r.releaseSeat(seat)
				r.cancelPlanned(sp.ID, seat)
				if sp.ID == "yoke" {
					p.YokeUsed = false
				}
				sp.Occupied = append(sp.Occupied[:j], sp.Occupied[j+1:]...)
				break
			}
		}
	}
	return nil
}

func moorCanAgeGrapes(p *Player, n int) bool {
	if n < 1 || len(p.Grapes) < n {
		return false
	}
	chosen := []int{}
	var visit func(int) bool
	visit = func(start int) bool {
		if len(chosen) == n {
			return moorAgeGrapes(probePlayer(p), chosen, n, 1) == nil
		}
		for i := start; i <= len(p.Grapes)-(n-len(chosen)); i++ {
			chosen = append(chosen, i)
			if visit(i + 1) {
				return true
			}
			chosen = chosen[:len(chosen)-1]
		}
		return false
	}
	return visit(0)
}
