package game

import (
	"fmt"
)

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
