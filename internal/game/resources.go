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
	p.Wines = append(p.Wines, Wine{ID: NewID(), Type: typ, Value: total})
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
	if len(recipes) < 1 || len(recipes) > limit {
		return fmt.Errorf("本次须酿1至%d瓶", limit)
	}
	original := append([]Grape{}, p.Grapes...)
	used := map[int]bool{}
	// Indices in every recipe refer to the same pre-action grape list.
	for _, recipe := range recipes {
		indices := []int{}
		for _, i := range recipe {
			if i < 0 || i >= len(original) || used[i] {
				return fmt.Errorf("葡萄重复或无效")
			}
			used[i] = true
			found := -1
			for j, g := range p.Grapes {
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
		if e := makeWine(p, indices); e != nil {
			return e
		}
	}
	return nil
}

// Exceptional wines can already be beyond a missing cellar. Only crossing the
// next printed boundary requires that cellar; occupied destinations are checked separately.
func wineAgeBoundary(p *Player, value int) bool {
	return value < 9 && (value != 3 || has(p, "medium_cellar")) && (value != 6 || has(p, "large_cellar"))
}
