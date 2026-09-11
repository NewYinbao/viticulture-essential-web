package game

import "fmt"

func (r *Room) tuscany() bool { return r.Config.Normalized().Board == "tuscany" }

// BonusSlots use physical left-to-right slots. Unlike EE, several leftmost
// Tuscany slots have a printed bonus even at two players.
func NewTuscanySpaces(n int) []Space {
	capacity := 1
	if n >= 3 {
		capacity = 2
	}
	if n >= 5 {
		capacity = 3
	}
	defs := []struct {
		id, name, description, season string
		bonuses                       map[int]string
	}{
		{"draw_vine", "葡萄藤市场", "抽1张葡萄藤", "spring", map[int]string{1: "draw_vine"}},
		{"tour", "酒庄导览", "获得2金币；品酒室每年首次满足条件加1分", "spring", map[int]string{2: "coin"}},
		{"build", "建造建筑", "支付建筑价格建造", "spring", map[int]string{2: "discount"}},
		{"influence", "扩大影响力", "放置1颗影响力星；六颗用完后可移动1颗", "spring", map[int]string{2: "influence"}},
		{"summer_visitor", "夏季访客", "打出1张夏季访客", "summer", map[int]string{1: "coin", 2: "visitor"}},
		{"plant", "种植葡萄藤", "种植1张葡萄藤", "summer", map[int]string{1: "plant"}},
		{"trade", "交换资源", "3金币／1分／2张牌／1颗葡萄：任选一项换另一项", "summer", map[int]string{2: "trade"}},
		{"flip_field", "买卖田地", "按田地面值买回或出售空田地", "summer", map[int]string{1: "vp"}},
		{"draw_order", "获取订单", "抽1张订单", "fall", map[int]string{1: "draw_order"}},
		{"harvest", "收获田地", "收获1块田地，每田每年一次", "fall", map[int]string{1: "harvest", 2: "coin"}},
		{"make_wine", "酿造葡萄酒", "酿造至多2瓶酒", "fall", map[int]string{1: "make_wine"}},
		{"build_tour", "建造或导览", "建造1座建筑或导览获得2金币", "fall", map[int]string{1: "build_tour"}},
		{"winter_visitor", "冬季访客", "打出1张冬季访客", "winter", map[int]string{1: "coin", 2: "visitor"}},
		{"train", "培训工人", "支付4金币培训，下一年可用，最多6名", "winter", map[int]string{1: "discount"}},
		{"sell_wine", "出售葡萄酒", "弃1瓶红／白得1分，桃红2分，起泡4分", "winter", map[int]string{2: "influence"}},
		{"fill_order", "交付订单", "交付1张订单", "winter", map[int]string{2: "vp"}},
	}
	out := make([]Space, 0, 18)
	for _, d := range defs {
		out = append(out, Space{ID: d.id, Name: d.name, Description: d.description, Season: d.season, Capacity: capacity, Occupied: []Seat{}, BonusSlots: d.bonuses})
	}
	return append(out, Space{ID: "gain_coin", Name: "获得1金币", Description: "各季节均可使用，不限人数", Season: "any", Capacity: 0}, Space{ID: "yoke", Name: "轭：收获或拔藤", Description: "自家行动；每年一次", Season: "any", Capacity: 1})
}

func (r *Room) spacesForRules() []Space {
	var spaces []Space
	if r.tuscany() {
		spaces = NewTuscanySpaces(len(r.Players))
	} else {
		spaces = NewSpaces(len(r.Players))
	}
	if r.Config.Structures && r.tuscany() {
		season := "summer"
		season = "spring"
		spaces = append(spaces, Space{ID: "draw_structure", Name: "获取结构牌", Description: "抽1张Tuscany结构牌", Season: season, Capacity: 1, Occupied: []Seat{}})
	}
	return spaces
}

func (r *Room) actionSeason() bool {
	return r.Phase == "summer" || r.Phase == "winter" || (r.tuscany() && (r.Phase == "spring" || r.Phase == "fall"))
}

func (r *Room) placementBonus(space string, slot int) string {
	if r.Config.Structures && r.tuscany() && slot > 0 && (space == "flip_field" || space == "sell_wine") {
		return "draw_structure"
	}
	for _, s := range r.Spaces {
		if s.ID == space {
			if s.BonusSlots != nil {
				return s.BonusSlots[slot]
			}
			// The EE board stores its single printed reward in spaceBonuses
			// instead of BonusSlots. Special workers remain independently
			// configurable with the EE board, so Politico must resolve that
			// reward through the same authoritative table.
			if slot == 1 && s.Capacity >= 2 {
				return spaceBonuses[s.ID]
			}
			return ""
		}
	}
	return ""
}

type InfluenceRegion struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Points int    `json:"points"`
	Reward string `json:"reward"`
}

// Tuscany Essential's influence map is printed on both board sides. Side 2
// adds the structure-card draw in Lucca; the region rewards themselves remain
// the same authoritative table.
var tuscanyRegions = []InfluenceRegion{
	{"lucca", "卢卡", 1, "structure"},
	{"pisa", "比萨", 1, "coins2"},
	{"firenze", "佛罗伦萨", 2, "winter"},
	{"livorno", "利沃诺", 2, "summer"},
	{"siena", "锡耶纳", 2, "coin"},
	{"arezzo", "阿雷佐", 2, "order"},
	{"grosseto", "格罗塞托", 1, "vine"},
}

type InfluenceMove struct {
	From string `json:"from,omitempty"`
	To   string `json:"to"`
}

func (r *Room) influence(p *Player, move InfluenceMove) error {
	var region *InfluenceRegion
	for i := range tuscanyRegions {
		if tuscanyRegions[i].ID == move.To {
			region = &tuscanyRegions[i]
		}
	}
	if region == nil {
		return fmt.Errorf("无效影响力区域")
	}
	n := 0
	for _, count := range p.Influence {
		n += count
	}
	if p.Influence == nil {
		p.Influence = map[string]int{}
	}
	if n < 6 {
		if move.From != "" {
			return fmt.Errorf("须先放完六颗影响力星才能移动")
		}
		p.Influence[move.To]++
		r.applyInfluenceReward(p, region.Reward)
	} else {
		if move.From == move.To || p.Influence[move.From] <= 0 {
			return fmt.Errorf("须移动自己另一地区的星")
		}
		p.Influence[move.From]--
		p.Influence[move.To]++
		if has(p, "banquet_hall") {
			r.applyInfluenceReward(p, region.Reward)
		}
	}
	return nil
}

func (r *Room) applyInfluenceReward(p *Player, reward string) {
	switch reward {
	case "coins2":
		p.Coins += 2
	case "coin":
		p.Coins++
	default:
		r.draw(p, reward)
	}
}

func (r *Room) scoreInfluence() {
	for _, region := range tuscanyRegions {
		best, tied := 0, false
		var winner *Player
		for _, p := range r.Players {
			n := p.Influence[region.ID]
			if n == 0 {
				continue
			}
			if n > best {
				best = n
				winner = p
				tied = false
			} else if n == best {
				tied = true
			}
		}
		if winner != nil && !tied {
			winner.VP += region.Points
		}
	}
}
