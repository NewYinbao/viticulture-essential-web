package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

func (r *Room) spring() {
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
	r.log(fmt.Sprintf("第%d年 · 选择起床顺序", r.Year))
}
func (r *Room) ordered() []*Player {
	p := append([]*Player{}, r.Players...)
	sort.Slice(p, func(i, j int) bool { return p[i].Wake < p[j].Wake })
	return p
}
func (r *Room) next() {
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
			return
		}
	}
	if r.Phase == "summer" {
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
		r.log("秋季：选择访客颜色，小屋额外任选一张")
	} else {
		r.endYear()
	}
}
func (r *Room) finishYear() {
	// Return all placed workers even when this is the final year.
	r.Spaces = spaces(len(r.Players))
	r.GrayWorkerOwner = ""
	for _, p := range r.Players {
		p.Coins += p.Income
		ageGrapes(p)
		ageWines(p)
		p.Workers = p.TotalWorkers - 1
		p.LargeWorker = true
		p.Trained = 0
		p.MillUsed = false
		p.TastingUsed = false
		p.YokeUsed = false
		for i := range p.Fields {
			p.Fields[i].Harvested = false
		}
		if p.VP >= 20 {
			r.FinalYear = true
		}
	}
	if r.FinalYear {
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
		r.log("游戏结束，最高排名：" + ps[0].Name)
		return
	}
	r.Year++
	r.SpringLeader = (r.SpringLeader + len(r.Players) - 1) % len(r.Players)
	r.spring()
}
func cellar(p *Player) int {
	if has(p, "large_cellar") {
		return 9
	}
	if has(p, "medium_cellar") {
		return 6
	}
	return 3
}
func ageGrapes(p *Player) {
	sort.Slice(p.Grapes, func(i, j int) bool { return p.Grapes[i].Value > p.Grapes[j].Value })
	for i := range p.Grapes {
		g := &p.Grapes[i]
		if g.Value >= 9 {
			continue
		}
		free := true
		for _, h := range p.Grapes {
			if h.Color == g.Color && h.Value == g.Value+1 {
				free = false
			}
		}
		if free {
			g.Value++
		}
	}
}
func ageWines(p *Player) {
	sort.Slice(p.Wines, func(i, j int) bool { return p.Wines[i].Value > p.Wines[j].Value })
	for i := range p.Wines {
		w := &p.Wines[i]
		if !wineAgeBoundary(p, w.Value) {
			continue
		}
		free := true
		for _, h := range p.Wines {
			if h.Type == w.Type && h.Value == w.Value+1 {
				free = false
			}
		}
		if free {
			w.Value++
		}
	}
}
func addGrape(p *Player, color string, value int) {
	if value > 9 {
		value = 9
	}
	for value > 0 {
		taken := false
		for _, g := range p.Grapes {
			if g.Color == color && g.Value == value {
				taken = true
			}
		}
		if !taken {
			p.Grapes = append(p.Grapes, Grape{ID: uid(), Color: color, Value: value})
			return
		}
		value--
	}
}

// apply is transactional even for non-HTTP callers. Preserve player pointer identity.
func (r *Room) apply(id string, a Action) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err = r.applyUnsafe(id, a); err != nil {
		var old Room
		_ = json.Unmarshal(b, &old)
		for i, p := range old.Players {
			if existing := r.player(p.ID); existing != nil {
				*existing = *p
				old.Players[i] = existing
			}
		}
		*r = old
	}
	return err
}
func (r *Room) applyUnsafe(id string, a Action) error {
	p := r.player(id)
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}
	if len(r.Choices) > 0 {
		return r.resolveChoice(id, a)
	}
	if a.Type == "choose" {
		return fmt.Errorf("没有待决选择")
	}
	if a.Type == "start" {
		if r.Phase != "lobby" || r.HostID != id || (len(r.Players) < 2 || len(r.Players) > 6) {
			return fmt.Errorf("需要房主且至少2名玩家")
		}
		r.Year = 1
		r.Spaces = spaces(len(r.Players))
		r.Ruleset = "ee-base-v1"
		r.SpringLeader = rnd(len(r.Players))
		r.initDecks()
		r.setupParents()

		return nil
	}
	if r.Phase == "lobby" || r.Phase == "finished" {
		return fmt.Errorf("当前不能行动")
	}
	if r.TurnID != id {
		return fmt.Errorf("现在不是你的回合")
	}
	if a.Type == "wake" {
		if r.Phase != "wake" || a.Slot < 1 || a.Slot > 7 {
			return fmt.Errorf("请选择有效起床格")
		}
		w := &r.WakeSlots[a.Slot-1]
		if w.PlayerID != "" {
			return fmt.Errorf("该起床格已占用")
		}
		if a.Slot == 5 && a.Color != "summer" && a.Color != "winter" {
			return fmt.Errorf("起床5须选择summer或winter")
		}
		w.PlayerID = id
		p.Wake = a.Slot
		switch a.Slot {
		case 2:
			r.draw(p, "vine")
		case 3:
			r.draw(p, "order")
		case 4:
			p.Coins++
		case 5:
			r.draw(p, a.Color)
		case 6:
			p.VP++
		case 7:
			r.claimGrayWorker(p)
		}
		r.log(fmt.Sprintf("%s 选择起床格%d", p.Name, a.Slot))
		for k := 1; k <= len(r.Players); k++ {
			q := r.Players[(r.SpringLeader+k)%len(r.Players)]
			if q.Wake == 0 {
				r.TurnID = q.ID
				return nil
			}
		}
		r.Phase = "summer"
		r.TurnID = r.ordered()[0].ID
		return nil
	}
	if r.Phase != "summer" && r.Phase != "winter" {
		return fmt.Errorf("请先选择起床顺序")
	}
	if a.Type == "pass" {
		p.Passed = true
		r.log(p.Name + " 结束本季行动")
		r.next()
		return nil
	}
	if a.Type != "place" {
		return fmt.Errorf("未知行动")
	}
	var s *Space
	for i := range r.Spaces {
		if r.Spaces[i].ID == a.Space {
			s = &r.Spaces[i]
		}
	}
	if s == nil || (s.Season != r.Phase && s.Season != "any") {
		return fmt.Errorf("当前季节没有该行动")
	}
	if a.Large {
		if !p.LargeWorker {
			return fmt.Errorf("大工人已使用")
		}
	} else if p.Workers <= 0 {
		return fmt.Errorf("没有可用普通工人")
	}
	slot, bonus, e := placement(s, p, a)
	if e != nil {
		return e
	}
	r.Context = &ActionContext{ActorID: id, Space: a.Space, Step: "perform", ReturnTurnID: id, TriggerSeat: &Seat{PlayerID: id, Slot: slot, Large: a.Large}}
	if e := r.performPlacement(p, a, bonus); e != nil {
		return e
	}
	if len(r.Choices) == 0 {
		r.Context = nil
	}
	if a.Large {
		p.LargeWorker = false
	} else {
		p.Workers--
	}
	s.Occupied = append(s.Occupied, Seat{PlayerID: id, Large: a.Large, Slot: slot})

	r.log(p.Name + " · " + s.Name)
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	r.next()
	return nil
}
func cardIndex(p *Player, id, t string) int {
	for i, c := range p.Hand {
		if c.ID == id && c.Type == t {
			return i
		}
	}
	return -1
}
func removeCard(p *Player, i int) { p.Hand = append(p.Hand[:i], p.Hand[i+1:]...) }

var costs = map[string]int{"yoke": 2, "trellis": 2, "irrigation": 3, "medium_cellar": 4, "large_cellar": 6, "cottage": 4, "windmill": 5, "tasting_room": 6}

func perform(p *Player, a Action) error { return (&Room{}).perform(p, a) } // legacy test adapter
func (r *Room) perform(p *Player, a Action) error {
	switch a.Space {
	case "draw_vine":
		r.draw(p, "vine")
	case "draw_order":
		r.draw(p, "order")
	case "tour":
		p.Coins += 2
		if has(p, "tasting_room") && len(p.Wines) > 0 && !p.TastingUsed {
			p.VP++
			p.TastingUsed = true
		}
	case "build":
		cost, ok := costs[a.Building]
		if !ok {
			return fmt.Errorf("不支持的建筑")
		}
		if has(p, a.Building) {
			return fmt.Errorf("已拥有该建筑")
		}
		if p.Coins < cost {
			return fmt.Errorf("金币不足，需要%d", cost)
		}
		if a.Building == "large_cellar" && !has(p, "medium_cellar") {
			return fmt.Errorf("需要先建中酒窖")
		}
		p.Coins -= cost
		p.Buildings = append(p.Buildings, a.Building)
	case "plant":
		i := cardIndex(p, a.CardID, "vine")
		if i < 0 {
			return fmt.Errorf("请选择自己的葡萄藤手牌")
		}
		if a.Field < 0 || a.Field >= len(p.Fields) {
			return fmt.Errorf("田地无效")
		}
		c := p.Hand[i]
		v := c.Red + c.White
		if c.Trellis && !has(p, "trellis") {
			return fmt.Errorf("该藤需要棚架")
		}
		if c.Irrigation && !has(p, "irrigation") {
			return fmt.Errorf("该藤需要灌溉")
		}
		f := &p.Fields[a.Field]
		if f.Sold {
			return fmt.Errorf("已出售田地不能种植")
		}
		total := v
		for _, c := range f.Vines {
			total += c.Red + c.White
		}
		if total > f.Capacity {
			return fmt.Errorf("超出田地容量")
		}
		f.Vines = append(f.Vines, c)
		removeCard(p, i)
		if has(p, "windmill") && !p.MillUsed {
			p.VP++
			p.MillUsed = true
		}
	case "harvest":
		if a.Field < 0 || a.Field >= len(p.Fields) {
			return fmt.Errorf("田地无效")
		}
		f := &p.Fields[a.Field]
		if f.Harvested || len(f.Vines) == 0 {
			return fmt.Errorf("该田地已收获或未种植")
		}
		red, white := 0, 0
		for _, c := range f.Vines {
			red += c.Red
			white += c.White
		}
		addGrape(p, "red", red)
		addGrape(p, "white", white)
		f.Harvested = true
	case "train":
		if p.TotalWorkers >= 6 {
			return fmt.Errorf("工人上限6名")
		}
		if p.Coins < 4 {
			return fmt.Errorf("培训需要4金币")
		}
		p.Coins -= 4
		p.TotalWorkers++
		p.Trained++
	case "summer_visitor", "winter_visitor":
		return fmt.Errorf("EE_VISITOR_NOT_IMPLEMENTED: 真实访客效果尚未实现，不允许演示通用效果")
	case "sell_grapes":
		if len(a.Grapes) == 0 {
			return fmt.Errorf("请选择葡萄")
		}
		seen := map[int]bool{}
		cash := 0
		for _, i := range a.Grapes {
			if i < 0 || i >= len(p.Grapes) || seen[i] {
				return fmt.Errorf("葡萄选择无效")
			}
			seen[i] = true
			cash += (p.Grapes[i].Value + 2) / 3
		}
		out := []Grape{}
		for i, g := range p.Grapes {
			if !seen[i] {
				out = append(out, g)
			}
		}
		p.Grapes = out
		p.Coins += cash
	case "make_wine":
		if len(a.Recipes) == 0 {
			a.Recipes = [][]int{a.Grapes}
		}
		return makeWines(p, a.Recipes, 2)
	case "fill_order":
		i := cardIndex(p, a.CardID, "order")
		if i < 0 {
			return fmt.Errorf("请选择订单")
		}
		c := p.Hand[i]
		if e := fillOrder(p, a); e != nil {
			return e
		}
		r.discard(c)
	default:
		return fmt.Errorf("未知行动空间")
	}
	return nil
}
func makeWine(p *Player, indices []int) error {
	seen := map[int]bool{}
	red, white, total := 0, 0, 0
	for _, i := range indices {
		if i < 0 || i >= len(p.Grapes) || seen[i] {
			return fmt.Errorf("葡萄选择无效")
		}
		seen[i] = true
		g := p.Grapes[i]
		total += g.Value
		if g.Color == "red" {
			red++
		} else {
			white++
		}
	}
	typ := ""
	min := 1
	switch {
	case red == 1 && white == 0:
		typ = "red"
	case red == 0 && white == 1:
		typ = "white"
	case red == 1 && white == 1:
		typ = "blush"
		min = 4
	case red == 2 && white == 1:
		typ = "sparkling"
		min = 7
	default:
		return fmt.Errorf("红/白酒需同色1颗；桃红需红白各1；起泡需2红1白")
	}
	if total > cellar(p) {
		total = cellar(p)
	}
	for total >= min {
		free := true
		for _, w := range p.Wines {
			if w.Type == typ && w.Value == total {
				free = false
			}
		}
		if free {
			break
		}
		total--
	}
	if total < min {
		return fmt.Errorf("品质不足、酒窖未升级或酒槽已满")
	}
	out := []Grape{}
	for i, g := range p.Grapes {
		if !seen[i] {
			out = append(out, g)
		}
	}
	p.Grapes = out
	p.Wines = append(p.Wines, Wine{ID: uid(), Type: typ, Value: total})
	return nil
}
func fillOrder(p *Player, a Action) error {
	i := cardIndex(p, a.CardID, "order")
	if i < 0 {
		return fmt.Errorf("请选择自己的订单")
	}
	c := p.Hand[i]
	if len(a.WineIDs) != len(c.Requirements) {
		return fmt.Errorf("请选择订单需要的酒")
	}
	used := map[string]bool{}
	reqs := append([]Wine{}, c.Requirements...)
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].Value > reqs[j].Value })
	for _, req := range reqs {
		found := false
		for _, id := range a.WineIDs {
			if used[id] {
				continue
			}
			for _, w := range p.Wines {
				if w.ID == id && w.Type == req.Type && w.Value >= req.Value {
					used[id] = true
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("葡萄酒不符合订单")
		}
	}
	out := []Wine{}
	for _, w := range p.Wines {
		if !used[w.ID] {
			out = append(out, w)
		}
	}
	p.Wines = out
	p.VP += c.Points
	p.Income += c.Income
	if p.Income > 5 {
		p.Income = 5
	}
	removeCard(p, i)
	return nil
}
