package game

import "fmt"

type RhineStarMove struct {
	PlayerID string `json:"playerId"`
	From     string `json:"from"`
	To       string `json:"to,omitempty"`
}

func rhineStarCount(p *Player) int {
	n := 0
	for _, count := range p.Influence {
		n += count
	}
	return n
}
func rhineRegion(id string) bool {
	for _, region := range tuscanyRegions {
		if region.ID == id {
			return true
		}
	}
	return false
}

func (r *Room) rhineBoardOperation(p *Player, id string, op RhineOperation, a Action, values map[string]string) error {
	if !r.tuscany() {
		return fmt.Errorf("此访客需要Tuscany主板")
	}
	switch op.Kind {
	case "influence_place":
		if len(a.Influence) != 1 {
			return fmt.Errorf("须放置或移动1颗自己的星")
		}
		return r.influence(p, a.Influence[0])
	case "general_responses":
		for _, q := range r.Players {
			if q.ID != p.ID && rhineStarCount(q) == 6 {
				r.rhinePrompt(q, id, []RhineOperation{{Kind: "return_stars", Amount: 3, Target: q.ID}}, nil)
			}
		}
	case "return_stars", "lobbyist", "move_stars":
		count := len(a.RhineStars)
		if op.Kind == "return_stars" && count != op.Amount {
			return fmt.Errorf("须收回恰好%d颗自己的星", op.Amount)
		}
		if op.Kind == "move_stars" && count > op.Amount {
			return fmt.Errorf("至多移动%d颗星", op.Amount)
		}
		if op.Kind == "lobbyist" {
			if count < 1 {
				return fmt.Errorf("至少选择1颗要收回的星")
			}
			if err := visitorPay(p, count, 0); err != nil {
				return err
			}
			for _, q := range r.Players {
				if q.VP >= 20 {
					if err := visitorPay(p, 0, 3); err != nil {
						return err
					}
					break
				}
			}
		}
		// Validate every source against a snapshot, so the same physical star cannot
		// be moved twice by using its newly reached region as a second source.
		sources := map[string]map[string]int{}
		for _, move := range a.RhineStars {
			owner := r.Player(move.PlayerID)
			if owner == nil || !rhineRegion(move.From) {
				return fmt.Errorf("无效星星所有者或来源")
			}
			if op.Kind == "return_stars" && owner.ID != p.ID {
				return fmt.Errorf("只能收回自己的星")
			}
			if sources[owner.ID] == nil {
				sources[owner.ID] = map[string]int{}
			}
			sources[owner.ID][move.From]++
			if sources[owner.ID][move.From] > owner.Influence[move.From] {
				return fmt.Errorf("来源区域星星不足或重复")
			}
			if op.Kind == "move_stars" && (!rhineRegion(move.To) || move.From == move.To) {
				return fmt.Errorf("请选择不同的目标区域")
			}
		}
		for _, move := range a.RhineStars {
			owner := r.Player(move.PlayerID)
			owner.Influence[move.From]--
			if op.Kind == "move_stars" {
				owner.Influence[move.To]++
				// The card's player performs the move. Banquet Hall belongs to
				// that mover, even when the moved star belongs to an opponent.
				if r.Config.Structures && has(p, "banquet_hall") {
					for _, region := range tuscanyRegions {
						if region.ID == move.To {
							r.applyInfluenceReward(p, region.Reward)
							break
						}
					}
				}
			}
		}
	}
	return nil
}

func (r *Room) rhineDestroy(p *Player, building string) (int, error) {
	if cost, ok := costs[building]; ok {
		if !has(p, building) {
			return 0, fmt.Errorf("没有该建筑")
		}
		// The medium cellar prerequisite constrains construction; it does not
		// make a built large cellar or previously obtained wine disappear.
		for i, b := range p.Buildings {
			if b == building {
				p.Buildings = append(p.Buildings[:i], p.Buildings[i+1:]...)
				return cost, nil
			}
		}
	}
	d, ok := structureDef(building)
	if !ok || !r.Config.Structures {
		return 0, fmt.Errorf("无效可拆建筑")
	}
	if err := r.destroyStructure(p, building); err != nil {
		return 0, err
	}
	return d.Cost, nil
}

func (r *Room) rhinePlan(p *Player, a Action) error {
	if r.Context == nil || r.Context.TriggerSeat == nil {
		return fmt.Errorf("未找到打出此牌的工人")
	}
	seat := *r.Context.TriggerSeat
	source := triggerSpace(r.Context)
	if seat.PlayerID != p.ID {
		return fmt.Errorf("只能移动本次自己的工人")
	}
	fromIndex, seatIndex, toIndex := -1, -1, -1
	for i, space := range r.Spaces {
		if space.ID == a.Space && r.plannerFutureSpace(space) && space.Season != "any" {
			toIndex = i
		}
		if space.ID == source {
			for j, existing := range space.Occupied {
				if existing == seat {
					fromIndex, seatIndex = i, j
					break
				}
			}
		}
	}
	if fromIndex < 0 || toIndex < 0 {
		return fmt.Errorf("请选择未来季节的公共行动，且触发工人必须仍在原处")
	}
	a.Large, a.WorkerType = seat.Large, seat.WorkerType
	slot, _, err := placement(&r.Spaces[toIndex], p, a)
	if err != nil {
		return err
	}
	if err := r.paySoldatoToll(a.Space, p.ID); err != nil {
		return err
	}
	r.cancelPlanned(source, seat)
	r.Spaces[fromIndex].Occupied = append(r.Spaces[fromIndex].Occupied[:seatIndex], r.Spaces[fromIndex].Occupied[seatIndex+1:]...)
	seat.Slot = slot
	r.Spaces[toIndex].Occupied = append(r.Spaces[toIndex].Occupied, seat)
	r.Planned = append(r.Planned, PlannedPlacement{PlayerID: p.ID, Space: a.Space, Slot: slot, Large: seat.Large, WorkerType: seat.WorkerType, Gray: seat.Gray})
	return nil
}

func (r *Room) rhineVirtuoso(p *Player, a Action) error {
	var space *Space
	for i := range r.Spaces {
		if r.Spaces[i].ID == a.Space {
			space = &r.Spaces[i]
			break
		}
	}
	if space == nil || space.Season == "any" || seasonIndex(space.Season) < 0 || seasonIndex(space.Season) >= seasonIndex(r.Phase) {
		return fmt.Errorf("只能执行本年更早季节的公共行动")
	}
	key := ""
	if a.Slot > 0 && !a.DeclineBonus {
		if r.tuscany() {
			key = space.BonusSlots[a.Slot]
		} else if a.Slot == 1 {
			key = map[string]string{"tour": "coin", "build": "discount", "sell_grapes": "vp", "summer_visitor": "visitor", "draw_vine": "draw_vine", "plant": "plant"}[space.ID]
		}
		if key == "" {
			return fmt.Errorf("所选位置没有印刷奖励")
		}
	}
	a.BonusOverride = key
	a.WorkerType = ""
	a.Large = false
	return r.startNestedAction(p, a, key != "", nil)
}
