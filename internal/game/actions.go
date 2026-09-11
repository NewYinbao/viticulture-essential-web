package game

import (
	"fmt"
	"strings"
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
	before := make([]bool, len(p.Fields))
	for i, f := range p.Fields {
		before[i] = f.Harvested
	}
	defer r.moorHarvestRewards(p, before)
	if strings.HasPrefix(a.Space, "structure_destroy:") {
		parts := strings.Split(a.Space, ":")
		if len(parts) != 2 || parts[1] != p.ID {
			return fmt.Errorf("只能使用自己的拆除结构行动")
		}
		a.Space = "destroy_structure"
	}
	if strings.HasPrefix(a.Space, "structure_action:") {
		_, d, ok := r.structureOwnerAction(a.Space)
		if !ok {
			return fmt.Errorf("结构行动不存在")
		}
		a.Space = "structure_action"
		a.Building = d.ID
		return r.performStructureAction(p, d, a)
	}
	switch a.Space {
	case "draw_vine":
		r.draw(p, "vine")
	case "draw_order":
		r.draw(p, "order")
	case "draw_structure":
		if !r.Config.Structures || !r.tuscany() {
			return fmt.Errorf("未启用建筑模块")
		}
		r.draw(p, "structure")
	case "tour":
		p.Coins += 2
		for _, q := range r.Players {
			if q.ID != p.ID && has(q, "fountain") {
				q.Coins++
			}
		}
		if has(p, "tasting_room") && len(p.Wines) > 0 && !p.TastingUsed {
			p.VP++
			p.TastingUsed = true
		}
		if r.tuscany() && has(p, "gazebo") {
			r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: p.ID, Kind: "structure_influence", Options: []string{"place", "skip"}, Count: 1, Labels: map[string]string{"place": "凉亭：放置或移动1颗影响力星", "skip": "跳过凉亭奖励"}})
		}
	case "build":
		if _, _, ok := p.structureCardInHand(a.Building); ok {
			return r.buildStructure(p, a, 0)
		}
		cost, ok := costs[a.Building]
		if !ok {
			return fmt.Errorf("不支持的建筑")
		}
		if has(p, a.Building) {
			return fmt.Errorf("已拥有该建筑")
		}
		if has(p, "workshop") {
			cost = max(0, cost-1)
		}
		if p.Coins < cost {
			return fmt.Errorf("金币不足，需要%d", cost)
		}
		if a.Building == "large_cellar" && !has(p, "medium_cellar") {
			return fmt.Errorf("需要先建中酒窖")
		}
		p.Coins -= cost
		p.Buildings = append(p.Buildings, a.Building)
		if a.Building != "studio" && has(p, "studio") {
			p.VP++
		}
	case "structure_action":
		d, ok := structureDef(a.Building)
		if !ok || !p.hasStructure(d.ID) {
			return fmt.Errorf("只能使用自己已建造的结构行动")
		}
		return r.performStructureAction(p, d, a)
	case "destroy_structure":
		if a.Building == "" {
			return fmt.Errorf("请选择要拆除的结构")
		}
		return r.destroyStructure(p, a.Building)
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
		if c.Trellis && !has(p, "trellis") && !has(p, "aqueduct") {
			return fmt.Errorf("该藤需要棚架")
		}
		if c.Irrigation && !has(p, "irrigation") && !has(p, "aqueduct") {
			return fmt.Errorf("该藤需要灌溉")
		}
		f := &p.Fields[a.Field]
		if f.Sold {
			return fmt.Errorf("已出售田地不能种植")
		}
		if f.Structure != "" {
			return fmt.Errorf("田地已被结构占用，须先拆除结构")
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
		if (a.Mode == "all" || a.HarvestAll) && has(p, "harvest_machine") {
			for i := range p.Fields {
				if !p.Fields[i].Sold && len(p.Fields[i].Vines) > 0 && !p.Fields[i].Harvested {
					if err := r.perform(p, Action{Space: "harvest", Field: i}); err != nil {
						return err
					}
				}
			}
			return nil
		}
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
		r.queueFermentation(p)
	case "train":
		return r.trainWorker(p, a.SpecialWorker, 0)
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
		return makeActionWines(p, a, 2)
	case "fill_order":
		return r.fillOrderAction(p, a)
	default:
		return fmt.Errorf("未知行动空间")
	}
	return nil
}

// fillOrderAction applies the order's normal rewards and passive structure
// enhancements. Label Factory is an explicit private action, so ordinary
// delivery never creates a second, hidden choice.
func (r *Room) fillOrderAction(p *Player, a Action) error {
	i := cardIndex(p, a.CardID, "order")
	if i < 0 {
		return fmt.Errorf("请选择订单")
	}
	c := p.Hand[i]
	if err := fillOrder(p, a); err != nil {
		return err
	}
	r.discard(c)
	if has(p, "veranda") {
		p.VP++
	}
	if has(p, "wine_parlor") {
		p.Coins += 2
	}
	return nil
}
