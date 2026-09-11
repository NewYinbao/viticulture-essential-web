package game

import "fmt"

type Trade struct {
	Give    string   `json:"give"`
	Receive string   `json:"receive"`
	CardIDs []string `json:"cardIds,omitempty"`
	Colors  []string `json:"colors,omitempty"`
	GrapeID string   `json:"grapeId,omitempty"`
	Color   string   `json:"color,omitempty"`
}

func (r *Room) trade(p *Player, t Trade) error {
	switch t.Give {
	case "coins":
		if err := visitorPay(p, 3, 0); err != nil {
			return err
		}
	case "vp":
		if err := visitorPay(p, 0, 1); err != nil {
			return err
		}
	case "cards":
		if err := r.visitorCards(p, t.CardIDs, 2, false); err != nil {
			return err
		}
	case "grape":
		found := false
		for i, g := range p.Grapes {
			if g.ID == t.GrapeID {
				p.Grapes = append(p.Grapes[:i], p.Grapes[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("请选择自己的一颗葡萄")
		}
	default:
		return fmt.Errorf("无效交易支付项")
	}
	switch t.Receive {
	case "coins":
		p.Coins += 3
	case "vp":
		p.VP++
	case "cards":
		if len(t.Colors) != 2 {
			return fmt.Errorf("须选择两张牌的颜色")
		}
		for _, color := range t.Colors {
			if !r.cardColor(color) {
				return fmt.Errorf("无效牌堆")
			}
			r.draw(p, color)
		}
	case "grape":
		if t.Color != "red" && t.Color != "white" {
			return fmt.Errorf("须选红或白葡萄")
		}
		addGrape(p, t.Color, 1)
	default:
		return fmt.Errorf("无效交易获得项")
	}
	return nil
}

func (r *Room) cardColor(color string) bool {
	return color == "vine" || color == "order" || color == "summer" || color == "winter" || (color == "structure" && r.Config.Structures)
}

func (r *Room) performTuscanyPlacement(p *Player, a Action, bonus bool) error {
	key := ""
	if bonus && !a.DeclineBonus {
		key = r.placementBonus(a.Space, a.Slot)
	}
	// BonusOverride is a special worker's independently selected bonus.  It
	// remains valid even when the actor declined the printed slot reward.
	if a.BonusOverride != "" {
		key = a.BonusOverride
	}
	if key == "coin" && a.BonusFirst {
		p.Coins++
	}
	afterCoin := key == "coin" && !a.BonusFirst
	switch a.Space {
	case "influence":
		if len(a.Influence) != 1 {
			return fmt.Errorf("须放置或移动一颗星")
		}
		if err := r.influence(p, a.Influence[0]); err != nil {
			return err
		}
		if key == "influence" {
			r.enqueue(p.ID, "tuscany_influence", []string{"place", "skip"}, 1)
		}
	case "trade":
		if len(a.Trades) != 1 {
			return fmt.Errorf("先完成一次交易")
		}
		if err := r.trade(p, a.Trades[0]); err != nil {
			return err
		}
		if key == "trade" {
			r.enqueue(p.ID, "tuscany_trade", []string{"trade", "skip"}, 1)
		}
	case "flip_field":
		if a.Mode != "buy_field" && a.Mode != "sell_field" {
			return fmt.Errorf("须选择买回或出售田地")
		}
		b := a
		b.Space = "sell_grapes"
		if err := r.performEEPlacement(p, b, false); err != nil {
			return err
		}
		if key == "vp" {
			p.VP++
		} else if key == "draw_structure" {
			r.draw(p, "structure")
		}
	case "sell_wine":
		if len(a.WineIDs) != 1 {
			return fmt.Errorf("须出售一瓶酒")
		}
		found := false
		for i, w := range p.Wines {
			if w.ID == a.WineIDs[0] {
				vp := map[string]int{"red": 1, "white": 1, "blush": 2, "sparkling": 4}[w.Type]
				if vp == 0 {
					return fmt.Errorf("无效酒类型")
				}
				p.VP += vp
				p.Wines = append(p.Wines[:i], p.Wines[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("酒不在自己的酒窖")
		}
		if key == "influence" {
			r.enqueue(p.ID, "tuscany_influence", []string{"place", "skip"}, 1)
		} else if key == "draw_structure" {
			r.draw(p, "structure")
		}
	case "build_tour":
		b := a
		// Structure cards use Mode for their mat/field destination. The selected
		// card still unambiguously selects the autumn build branch.
		if a.Mode == "mat" || a.Mode == "field" {
			if _, _, ok := p.structureCardInHand(a.Building); !ok {
				return fmt.Errorf("请选择自己的结构牌")
			}
			b.Space = "build"
		} else {
			if a.Mode != "build" && a.Mode != "tour" {
				return fmt.Errorf("须选择建造或导览")
			}
			b.Space = a.Mode
			if a.Mode == "build" {
				if _, _, ok := p.structureCardInHand(a.Building); ok {
					b.Mode = "mat"
				}
			}
		}
		if err := r.performEEPlacement(p, b, key == "build_tour"); err != nil {
			return err
		}
	default:
		multi := key == "visitor" || key == "plant" || key == "harvest" || key == "make_wine" || key == "discount" || key == "draw_vine" || key == "draw_order" || key == "vp"
		// BonusOverride has already been resolved to the canonical key above.
		// Do not pass it into performEEPlacement: that helper treats any
		// non-empty override as the physical bonus and would double-apply the
		// coin branch (base tour bonus plus deferred coin).
		base := a
		base.BonusOverride = ""
		if err := r.performEEPlacement(p, base, multi); err != nil {
			return err
		}
	}
	if afterCoin {
		if len(r.Choices) > 0 && r.Context != nil {
			r.Context.CoinAfter++
		} else {
			p.Coins++
		}
	}
	return nil
}
