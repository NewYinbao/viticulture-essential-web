package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func (r *Room) moorEffect(p *Player, s *VisitorStep, a Action) error {
	if reason := r.moorOptionReason(p, s, a.Option); reason != "" {
		return fmt.Errorf("%s", reason)
	}
	draw := func(kind string, n int) {
		for i := 0; i < n; i++ {
			r.draw(p, kind)
		}
	}
	gain := func(value int) error { return moorGainGrape(p, a.Color, value) }
	pay := func(n int) error { return visitorPay(p, n, 0) }
	switch s.CardID {
	case "moor-summer-01":
		if a.Option == "build" {
			return r.visitorBuild(p, "yoke", 99)
		}
		draw("winter", 2)
		draw("order", 1)
	case "moor-summer-02", "moor-winter-12":
		season, kind := "winter", "order"
		if s.CardID == "moor-winter-12" {
			season, kind = "summer", "vine"
		}
		if !visitorInSeason(p, a.CardID, season) {
			return fmt.Errorf("请选择自己的%s访客", season)
		}
		// Validate the target before revealing a random card; rollback never enables fishing.
		if err := r.moorNestedPlayable(p, a.CardID, season); err != nil {
			return err
		}
		existingCount := len(r.Choices)
		draw(kind, 1)
		// Resolve the draw's immediate structure triggers before the nested
		// visitor effect, while retaining the precommitted visitor selection.
		immediate := append([]Choice{}, r.Choices[existingCount:]...)
		r.Choices = r.Choices[:existingCount]
		if err := r.startNestedVisitor(p, a.CardID, nil); err != nil {
			return err
		}
		r.Choices = append(immediate, r.Choices...)
		r.Context.Space = season + "_visitor"
		r.decorateMoorChoice(&r.Choices[len(r.Choices)-1], p)
		return nil
	case "moor-summer-03":
		if a.Option == "grape" {
			return gain(2)
		}
		if err := moorDiscardGrapes(p, a.Grapes, 3); err != nil {
			return err
		}
		p.VP += 3
	case "moor-summer-04":
		if s.Stage == "reward" {
			if a.Option == "coins" {
				p.Coins += 2
			} else {
				p.VP++
			}
			return nil
		}
		if a.Field < 0 || a.Field >= len(p.Fields) || p.Fields[a.Field].Sold || p.Fields[a.Field].Structure != "" || p.Fields[a.Field].FruitDealer {
			return fmt.Errorf("请选择未出售且没有结构或水果商的田地")
		}
		if err := pay(4); err != nil {
			return err
		}
		p.Fields[a.Field].FruitDealer = true
		r.moorReleaseHeld(s.CardID)
	case "moor-summer-05":
		if s.Stage == "effect" {
			r.visitorCycle(p, s, true, []string{"grape", "skip"})
			return nil
		}
		if a.Option == "skip" {
			return nil
		}
		if err := gain(2); err != nil {
			return err
		}
		if p.ID != s.ActorID {
			if r.Context.Data == nil {
				r.Context.Data = map[string]string{}
			}
			if r.Context.Data["moor-fusel-awarded"] != "yes" {
				r.Player(s.ActorID).VP += 2
				r.Context.Data["moor-fusel-awarded"] = "yes"
			}
		}
	case "moor-summer-06", "moor-summer-10":
		if a.Option == "vp" {
			p.VP += 2
		} else if a.Option == "coins" {
			p.Coins += 7
		} else {
			draw("winter", 4)
		}
		p.MoorContracts = append(p.MoorContracts, r.moorReleaseHeld(s.CardID))
	case "moor-summer-07":
		if a.Option == "coins" {
			p.Coins += p.Wake
		} else {
			if err := moorDiscardGrapes(p, a.Grapes, 2); err != nil {
				return err
			}
			p.VP += 2
		}
	case "moor-summer-08":
		if len(a.TargetIDs) != 1 {
			return fmt.Errorf("请选择1位对手")
		}
		q := r.Player(a.TargetIDs[0])
		if q == nil || q.ID == p.ID || a.Field < 0 || a.Field >= len(q.Fields) {
			return fmt.Errorf("无效对手田地")
		}
		f := q.Fields[a.Field]
		if f.Sold || len(f.Vines) == 0 || f.Structure != "" {
			return fmt.Errorf("田地没有可收获葡萄藤")
		}
		red, white := 0, 0
		for _, v := range f.Vines {
			red += v.Red
			white += v.White
		}
		if red > 0 {
			addGrape(p, "red", red)
		}
		if white > 0 {
			addGrape(p, "white", white)
		}
	case "moor-summer-09":
		draw("order", 1)
		if r.Context.TriggerSeat != nil && r.Context.TriggerSeat.Large {
			p.VP++
			p.Coins += 2
		}
	case "moor-summer-11":
		if a.Option == "build" {
			return r.visitorBuild(p, "irrigation", 99)
		}
		cid := a.CardID
		if len(a.CardIDs) == 1 {
			cid = a.CardIDs[0]
		}
		idx := cardIndex(p, cid, "vine")
		if idx < 0 || !p.Hand[idx].Irrigation {
			return fmt.Errorf("须种植需要灌溉的葡萄藤")
		}
		if err := r.visitorPlant(p, a, 1, false, false); err != nil {
			return err
		}
		p.VP++
	case "moor-summer-12":
		if a.Option == "vp" {
			p.VP += 2
			return nil
		}
		if len(a.Buildings) != 3 {
			return fmt.Errorf("请选择3座不同建筑，按建造顺序排列")
		}
		if err := pay(14); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, b := range a.Buildings {
			if seen[b] {
				return fmt.Errorf("建筑不能重复")
			}
			seen[b] = true
			if err := r.visitorBuild(p, b, 99); err != nil {
				return err
			}
		}
	case "moor-summer-13":
		if a.Option == "grape" {
			return gain(3)
		}
		return r.moorFixedBuild(p, "windmill", 2)
	case "moor-summer-14":
		if a.Option == "vp" {
			p.VP += 2
			return nil
		}
		if err := pay(3); err != nil {
			return err
		}
		return gain(5)
	case "moor-summer-15":
		if s.Stage == "effect" {
			return r.moorDrawPrivate(p, s, "order", 3, []string{"fill", "keep"})
		}
		if a.Option == "keep" {
			return r.moorKeepPrivate(p, a.CardIDs, 2)
		}
		if !moorContainsCard(r.Context.Held, a.CardID) {
			return fmt.Errorf("只能完成刚抽到的订单")
		}
		i := cardIndexIn(r.Context.Held, a.CardID)
		c := r.Context.Held[i]
		p.Hand = append(p.Hand, c)
		if err := r.fillOrderAction(p, a); err != nil {
			return err
		}
		r.Context.Held = append(r.Context.Held[:i], r.Context.Held[i+1:]...)
		return r.moorKeepPrivate(p, nil, 0)
	case "moor-summer-16":
		return r.moorTrain(p, a.SpecialWorker, 6-p.TotalWorkers, true)
	case "moor-summer-17":
		if a.Option == "wine" {
			return moorGainWine(p, a.Color, 1, true)
		}
		return r.moorFixedBuild(p, "tasting_room", 3)
	case "moor-summer-18":
		n := len(a.CardIDs)
		if n < 1 {
			return fmt.Errorf("至少弃1张手牌")
		}
		if err := r.visitorCards(p, a.CardIDs, n, false); err != nil {
			return err
		}
		if a.Option == "harvest" {
			if len(a.Fields) != n {
				return fmt.Errorf("田地数量须等于弃牌数量")
			}
			return r.visitorHarvest(p, a, n)
		}
		return moorAgeGrapes(p, a.Grapes, n, 1)
	case "moor-summer-19":
		if err := r.fillOrderAction(p, a); err != nil {
			return err
		}
		if r.Context.LastWorker {
			p.VP += 2
		}
	case "moor-summer-20":
		if a.Option == "upgrade" {
			return r.moorUpgrade(p, 3)
		}
		if err := pay(3); err != nil {
			return err
		}
		return moorGainWine(p, "white", 1, true)
	case "moor-winter-01":
		return r.moorRetrieve(p, a)
	case "moor-winter-02":
		if err := pay(7); err != nil {
			return err
		}
		return moorGainWine(p, "blush", 4, true)
	case "moor-winter-03":
		if a.Option == "upgrade" {
			return r.moorUpgrade(p, 2)
		}
		if a.Option == "both" {
			if err := visitorPay(p, 0, 1); err != nil {
				return err
			}
			if a.Mode != "age_first" {
				if err := r.moorUpgrade(p, 2); err != nil {
					return err
				}
			}
		}
		if len(a.WineIDs) != 1 {
			return fmt.Errorf("请选择1瓶酒")
		}
		times := a.Slot
		if times < 1 || times > 3 {
			return fmt.Errorf("陈酿次数须为1至3")
		}
		before := moorWineValue(p, a.WineIDs[0])
		if err := ageWineByID(p, a.WineIDs[0], times); err != nil {
			return err
		}
		if moorWineValue(p, a.WineIDs[0]) == before {
			return fmt.Errorf("所选酒无法陈酿")
		}
		if a.Option == "both" && a.Mode == "age_first" {
			return r.moorUpgrade(p, 2)
		}
	case "moor-winter-04":
		if a.Option == "vp" {
			if err := pay(9); err != nil {
				return err
			}
			p.VP += 3
		} else {
			p.Income += len(p.Wines)
			if p.Income > 5 {
				p.Income = 5
			}
		}
	case "moor-winter-05":
		if a.Option == "draw" {
			draw("order", 2)
		} else {
			types := map[string]bool{}
			for _, w := range p.Wines {
				types[w.Type] = true
			}
			p.VP += len(types)
		}
	case "moor-winter-06":
		if a.Option == "coins" {
			p.Coins += 3
		} else {
			return r.moorFillAnyType(p, a)
		}
	case "moor-winter-07":
		if err := r.visitorBuild(p, a.Building, 5); err != nil {
			return err
		}
		for i, q := range r.Players {
			if q.ID == p.ID {
				next := r.Players[(i+1)%len(r.Players)]
				next.Hand = append(next.Hand, r.moorReleaseHeld(s.CardID))
				break
			}
		}
	case "moor-winter-08":
		n, err := strconv.Atoi(strings.TrimPrefix(a.Option, "buy_"))
		if err != nil || n < 1 || n > 9 {
			return fmt.Errorf("葡萄价值须为1至9")
		}
		if err := pay(n); err != nil {
			return err
		}
		return gain(n)
	case "moor-winter-09":
		if a.Option == "coins" {
			p.Coins += 3
		} else {
			if err := moorDiscardGrapes(p, a.Grapes, 1); err != nil {
				return err
			}
			if err := visitorWine(p, a.WineIDs, 0); err != nil {
				return err
			}
			p.VP += 4
		}
	case "moor-winter-10":
		if r.moorSummerPlay() {
			if err := moorDiscardGrapes(p, a.Grapes, 1); err != nil {
				return err
			}
			p.Coins += 4
		} else {
			if err := pay(1); err != nil {
				return err
			}
			return gain(4)
		}
	case "moor-winter-11":
		if len(a.Grapes) != 1 || a.Grapes[0] < 0 || a.Grapes[0] >= len(p.Grapes) {
			return fmt.Errorf("请选择1颗葡萄")
		}
		value := p.Grapes[a.Grapes[0]].Value
		if err := moorDiscardGrapes(p, a.Grapes, 1); err != nil {
			return err
		}
		if a.Option == "coins" {
			p.Coins += value
		} else {
			p.VP += (value + 2) / 3
		}
	case "moor-winter-13":
		if a.Option == "plant" {
			return r.visitorPlant(p, a, 1, false, false)
		}
		if err := moorDevalueWines(p, a.WineIDs); err != nil {
			return err
		}
		p.VP += 2
	case "moor-winter-14":
		if err := pay(9); err != nil {
			return err
		}
		return moorGainWine(p, "sparkling", 7, true)
	case "moor-winter-15":
		if len(a.Colors) != 3 {
			return fmt.Errorf("请先选择3张牌的种类")
		}
		for _, kind := range a.Colors {
			if kind != "vine" && kind != "summer" && kind != "order" && kind != "winter" && !(kind == "structure" && r.Config.Structures) {
				return fmt.Errorf("无效牌种")
			}
		}
		if err := r.visitorCards(p, a.CardIDs, 3, false); err != nil {
			return err
		}
		for _, kind := range a.Colors {
			draw(kind, 1)
		}
		p.Coins += 2
	case "moor-winter-16":
		if s.Stage == "age" {
			return moorAgeGrapes(p, a.Grapes, s.MoorCount, 1)
		}
		n := len(a.Fields)
		if a.HarvestAll && has(p, "harvest_machine") {
			n = 0
			for i := range p.Fields {
				if validateHarvestField(p, i) == nil {
					n++
				}
			}
		}
		if n < 1 {
			return fmt.Errorf("请选择至少1块田地")
		}
		q, ok := r.probeAction(p, func(rr *Room, pp *Player) error { return rr.visitorHarvest(pp, a, n) })
		if !ok || !moorCanAgeGrapes(q, n) {
			return fmt.Errorf("收获后无法陈酿%d颗不同葡萄", n)
		}
		if err := r.visitorHarvest(p, a, n); err != nil {
			return err
		}
		r.visitorPrompt(p, s.CardID, "age", []string{"age"})
		c := &r.Choices[len(r.Choices)-1]
		c.Visitor.MoorCount = n
		r.decorateMoorChoice(c, p)
	case "moor-winter-17":
		if s.Stage == "effect" {
			return r.moorDrawPrivate(p, s, a.Option, 3, []string{"keep", "all"})
		}
		if a.Option == "all" {
			if err := pay(1); err != nil {
				return err
			}
			ids := []string{}
			for _, c := range r.Context.Held {
				ids = append(ids, c.ID)
			}
			return r.moorKeepPrivate(p, ids, 3)
		}
		return r.moorKeepPrivate(p, a.CardIDs, 1)
	case "moor-winter-18":
		if err := pay(9); err != nil {
			return err
		}
		if err := moorGainWine(p, "red", 2, true); err != nil {
			return err
		}
		return moorGainWine(p, "white", 2, true)
	case "moor-winter-19":
		if a.Option == "discard" {
			if err := visitorWine(p, a.WineIDs, 0); err != nil {
				return err
			}
			p.VP += 2
		} else {
			if err := visitorPay(p, 0, 1); err != nil {
				return err
			}
			return r.moorTrain(p, a.SpecialWorker, 0, false)
		}
	case "moor-winter-20":
		if err := pay(3); err != nil {
			return err
		}
		if a.Option == "wine" {
			return moorGainWine(p, a.Color, 1, true)
		}
		if a.Slot < 1 || a.Slot > 3 {
			return fmt.Errorf("陈酿次数须为1至3")
		}
		return moorAgeGrapes(p, a.Grapes, 1, a.Slot)
	default:
		return fmt.Errorf("未知Moor牌效 %s", s.CardID)
	}
	return nil
}

func cardIndexIn(cards []Card, id string) int {
	for i, c := range cards {
		if c.ID == id {
			return i
		}
	}
	return -1
}
func moorContainsCard(cards []Card, id string) bool { return cardIndexIn(cards, id) >= 0 }
func (r *Room) moorReleaseHeld(id string) Card {
	i := cardIndexIn(r.Context.Held, id)
	if i < 0 {
		return moorCard(id)
	}
	c := r.Context.Held[i]
	r.Context.Held = append(r.Context.Held[:i], r.Context.Held[i+1:]...)
	return c
}
func moorGainGrape(p *Player, color string, value int) error {
	if color != "red" && color != "white" {
		return fmt.Errorf("请选择红葡萄或白葡萄")
	}
	before := len(p.Grapes)
	addGrape(p, color, value)
	if len(p.Grapes) == before {
		return fmt.Errorf("对应葡萄槽已满")
	}
	return nil
}
func moorGainWine(p *Player, kind string, value int, ignoreCellar bool) error {
	min := 1
	switch kind {
	case "red", "white":
	case "blush":
		min = 4
	case "sparkling":
		min = 7
	default:
		return fmt.Errorf("无效葡萄酒种类")
	}
	if !ignoreCellar && value > cellar(p) {
		value = cellar(p)
	}
	for value >= min {
		free := true
		for _, w := range p.Wines {
			if w.Type == kind && w.Value == value {
				free = false
			}
		}
		if free {
			p.Wines = append(p.Wines, Wine{ID: NewID(), Type: kind, Value: value})
			return nil
		}
		value--
	}
	return fmt.Errorf("对应酒槽已满")
}
func moorDiscardGrapes(p *Player, indices []int, n int) error {
	if len(indices) != n {
		return fmt.Errorf("须选择%d颗不同葡萄", n)
	}
	seen := map[int]bool{}
	for _, i := range indices {
		if i < 0 || i >= len(p.Grapes) || seen[i] {
			return fmt.Errorf("葡萄选择无效或重复")
		}
		seen[i] = true
	}
	out := []Grape{}
	for i, g := range p.Grapes {
		if !seen[i] {
			out = append(out, g)
		}
	}
	p.Grapes = out
	return nil
}
func moorAgeGrapes(p *Player, indices []int, n, steps int) error {
	if len(indices) != n || n < 1 {
		return fmt.Errorf("须陈酿%d颗不同葡萄", n)
	}
	seen := map[int]bool{}
	trial := append([]Grape{}, p.Grapes...)
	order := append([]int{}, indices...)
	for _, i := range order {
		if i < 0 || i >= len(trial) || seen[i] {
			return fmt.Errorf("葡萄选择无效或重复")
		}
		seen[i] = true
	}
	sort.Slice(order, func(i, j int) bool { return trial[order[i]].Value > trial[order[j]].Value })
	for _, i := range order {
		old := trial[i].Value
		for step := 0; step < steps && trial[i].Value < 9; step++ {
			next := trial[i].Value + 1
			blocked := false
			for j, g := range trial {
				if j != i && g.Color == trial[i].Color && g.Value == next {
					blocked = true
				}
			}
			if blocked {
				break
			}
			trial[i].Value = next
		}
		if old == trial[i].Value {
			return fmt.Errorf("所选葡萄无法陈酿")
		}
	}
	p.Grapes = trial
	return nil
}
func moorWineValue(p *Player, id string) int {
	for _, w := range p.Wines {
		if w.ID == id {
			return w.Value
		}
	}
	return -1
}
func moorDevalueWines(p *Player, ids []string) error {
	if len(ids) != 3 {
		return fmt.Errorf("须选择3瓶不同的酒")
	}
	seen := map[string]bool{}
	trial := append([]Wine{}, p.Wines...)
	for _, id := range ids {
		if seen[id] {
			return fmt.Errorf("不能重复选择酒")
		}
		seen[id] = true
		found := false
		for i, w := range trial {
			if w.ID == id {
				min := 1
				if w.Type == "blush" {
					min = 4
				}
				if w.Type == "sparkling" {
					min = 7
				}
				if w.Value <= min {
					return fmt.Errorf("酒的价值已达下限")
				}
				trial[i].Value--
				found = true
			}
		}
		if !found {
			return fmt.Errorf("无效酒")
		}
	}
	occupied := map[string]bool{}
	for _, w := range trial {
		key := fmt.Sprintf("%s-%d", w.Type, w.Value)
		if occupied[key] {
			return fmt.Errorf("降级后的酒槽被占用")
		}
		occupied[key] = true
	}
	p.Wines = trial
	return nil
}
func (r *Room) moorFixedBuild(p *Player, b string, cost int) error {
	return r.visitorBuild(p, b, costs[b]-cost)
}
func (r *Room) moorUpgrade(p *Player, cost int) error {
	b := "medium_cellar"
	if has(p, b) {
		b = "large_cellar"
	}
	return r.moorFixedBuild(p, b, cost)
}
func (r *Room) moorTrain(p *Player, id string, cost int, now bool) error {
	if err := r.trainWorker(p, id, 4-cost); err != nil {
		return err
	}
	if now {
		if id == "grande" {
			p.LargeWorker = true
		} else if id != "" && id != "regular" {
			p.SpecialWorkerReady[id] = r.Year
		} else {
			p.Workers++
		}
	}
	return nil
}
func (r *Room) moorDrawPrivate(p *Player, s *VisitorStep, kind string, n int, opts []string) error {
	for i := 0; i < n; i++ {
		c, ok := r.take(kind)
		if !ok {
			return fmt.Errorf("牌堆不足%d张", n)
		}
		r.Context.Held = append(r.Context.Held, c)
	}
	r.visitorPrompt(p, s.CardID, "private", opts)
	r.Choices[len(r.Choices)-1].Cards = append([]Card{}, r.Context.Held...)
	return nil
}
func (r *Room) moorKeepPrivate(p *Player, ids []string, n int) error {
	if len(ids) != n {
		return fmt.Errorf("须保留%d张新抽到的牌", n)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] || !moorContainsCard(r.Context.Held, id) {
			return fmt.Errorf("只能选择不同的新抽牌")
		}
		seen[id] = true
	}
	for _, c := range r.Context.Held {
		if seen[c.ID] {
			p.Hand = append(p.Hand, c)
		} else {
			r.discard(c)
		}
	}
	r.Context.Held = nil
	return nil
}
func (r *Room) moorFillAnyType(p *Player, a Action) error {
	idx := cardIndex(p, a.CardID, "order")
	if idx < 0 {
		return fmt.Errorf("请选择订单")
	}
	c := p.Hand[idx]
	if len(a.WineIDs) != len(c.Requirements) {
		return fmt.Errorf("酒的数量不符订单")
	}
	reqs := append([]Wine{}, c.Requirements...)
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].Value > reqs[j].Value })
	used := map[string]bool{}
	for _, req := range reqs {
		best := ""
		value := 10
		for _, id := range a.WineIDs {
			for _, w := range p.Wines {
				if w.ID == id && !used[id] && w.Value >= req.Value && w.Value < value {
					best = id
					value = w.Value
				}
			}
		}
		if best == "" {
			return fmt.Errorf("酒的价值不足或重复")
		}
		used[best] = true
	}
	wines := []Wine{}
	for _, w := range p.Wines {
		if !used[w.ID] {
			wines = append(wines, w)
		}
	}
	p.Wines = wines
	removeCard(p, idx)
	r.discard(c)
	p.VP += c.Points
	p.Income += c.Income
	if p.Income > 5 {
		p.Income = 5
	}
	if has(p, "veranda") {
		p.VP++
	}
	if has(p, "wine_parlor") {
		p.Coins += 2
	}
	return nil
}
func (r *Room) settleMoorContracts() {
	for _, p := range r.Players {
		for _, c := range p.MoorContracts {
			if (c.ID == "moor-summer-06" && p.Income < 4) || (c.ID == "moor-summer-10" && len(p.Grapes) < 3) {
				p.VP -= 4
				if p.VP < -5 {
					p.VP = -5
				}
			}
		}
	}
}

func (r *Room) moorLastWorker(p *Player) bool {
	if p.Workers > 0 || p.LargeWorker {
		return false
	}
	for _, id := range p.SpecialWorkers {
		if p.specialWorkerReady(id, r.Year) {
			return false
		}
	}
	return true
}
func (r *Room) moorHarvestRewards(p *Player, before []bool) {
	if r.Context == nil {
		return
	}
	if r.Context.Data == nil {
		r.Context.Data = map[string]string{}
	}
	for i, f := range p.Fields {
		key := fmt.Sprintf("moor-fruit:%s:%d", p.ID, i)
		if i < len(before) && !before[i] && f.Harvested && f.FruitDealer && r.Context.Data[key] != "yes" {
			r.Context.Data[key] = "yes"
			r.visitorPrompt(p, "moor-summer-04", "reward", []string{"coins", "vp"})
		}
	}
}

func (r *Room) moorSummerPlay() bool {
	if r.Context != nil && r.Context.Space != "" {
		return r.Context.Space == "summer_visitor"
	}
	return r.Phase == "summer"
}
func (r *Room) moorNestedPlayable(p *Player, id, season string) error {
	room := *r
	room.Context = &ActionContext{ActorID: p.ID, Space: season + "_visitor"}
	return room.mandatoryVisitorPlayable(p, id)
}
