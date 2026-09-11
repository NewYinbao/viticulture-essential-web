package game

import (
	"fmt"
	"sort"
)

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
			p.Grapes = append(p.Grapes, Grape{ID: NewID(), Color: color, Value: value})
			return
		}
		value--
	}
}

// apply is transactional even for non-HTTP callers. Preserve player pointer identity.
func makeWine(p *Player, indices []int) error {
	return makeWineType(p, indices, "")
}
func makeWineType(p *Player, indices []int, forced string) error {
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
		if forced == "sparkling" || forced == "charmat" {
			typ = "sparkling"
			min = 7
		} else {
			typ = "blush"
			min = 4
		}
	case red == 2 && white == 1:
		if forced == "charmat" {
			return fmt.Errorf("Charmat只能将1红1白桃红酒转为起泡酒")
		}
		typ = "sparkling"
		min = 7
	default:
		return fmt.Errorf("红/白酒需同色1颗；桃红需红白各1；起泡需2红1白")
	}
	if (forced == "blush" || forced == "sparkling") && typ != forced {
		return fmt.Errorf("葡萄配方与所选酒类不符")
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
	p.Wines = append(p.Wines, Wine{ID: NewID(), Type: typ, Value: total})
	if (typ == "blush" || typ == "sparkling") && has(p, "patio") {
		p.Coins += 2
	}
	if total >= 7 && has(p, "penthouse") {
		p.VP++
	}
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
func makeWines(p *Player, recipes [][]int, limit int) error {
	return makeWinesType(p, recipes, limit, "")
}
func makeWinesType(p *Player, recipes [][]int, limit int, forced string) error {
	types := make([]string, len(recipes))
	for i := range types {
		types[i] = forced
	}
	return makeWinesKinds(p, recipes, limit, types)
}

// Each recipe retains its own optional Charmat choice, including mixed batches.
func makeActionWines(p *Player, a Action, limit int) error {
	if len(a.RecipeTypes) > 0 && len(a.RecipeTypes) != len(a.Recipes) {
		return fmt.Errorf("酒类选择与配方数量不符")
	}
	types := make([]string, len(a.Recipes))
	for i := range types {
		if len(a.RecipeTypes) > 0 {
			types[i] = a.RecipeTypes[i]
		} else if a.Mode == "sparkling" {
			if !has(p, "charmat") || len(a.Recipes[i]) != 2 {
				return fmt.Errorf("查玛法罐须使用1红1白配方")
			}
			types[i] = "sparkling"
		}
		if types[i] != "" && types[i] != "blush" && types[i] != "sparkling" {
			return fmt.Errorf("无效酿酒类型")
		}
		if types[i] == "sparkling" && len(a.Recipes[i]) == 2 && !has(p, "charmat") {
			return fmt.Errorf("红白各一酿起泡需要查玛法罐")
		}
	}
	return makeWinesKinds(p, a.Recipes, limit, types)
}
func makeWinesKinds(p *Player, recipes [][]int, limit int, types []string) error {
	if len(recipes) < 1 || len(recipes) > limit {
		return fmt.Errorf("本次须酿1至%d瓶", limit)
	}
	// A multi-bottle action is one transaction. Validate and apply on a copy so
	// a later invalid recipe cannot spend the earlier recipe's grapes.
	trial := *p
	trial.Grapes = append([]Grape(nil), p.Grapes...)
	trial.Wines = append([]Wine(nil), p.Wines...)
	original := append([]Grape{}, trial.Grapes...)
	used := map[int]bool{}
	// Indices in every recipe refer to the same pre-action grape list.
	for recipeIndex, recipe := range recipes {
		indices := []int{}
		for _, i := range recipe {
			if i < 0 || i >= len(original) || used[i] {
				return fmt.Errorf("葡萄重复或无效")
			}
			used[i] = true
			found := -1
			for j, g := range trial.Grapes {
				if g.ID == original[i].ID {
					found = j
					break
				}
			}
			if found < 0 {
				return fmt.Errorf("葡萄已消耗")
			}
			indices = append(indices, found)
		}
		if e := makeWineType(&trial, indices, types[recipeIndex]); e != nil {
			return e
		}
	}
	p.Grapes, p.Wines = trial.Grapes, trial.Wines
	p.Coins, p.VP = trial.Coins, trial.VP
	return nil
}

// Exceptional wines can already be beyond a missing cellar. Only crossing the
// next printed boundary requires that cellar; occupied destinations are checked separately.
func wineAgeBoundary(p *Player, value int) bool {
	return value < 9 && (value != 3 || has(p, "medium_cellar")) && (value != 6 || has(p, "large_cellar"))
}
