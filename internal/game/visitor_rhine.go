package game

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RhineOperation is saved on the private visitor choice. Each input operation
// gets its own choice ID; resolved costs and draws are never replayed on reload.
type RhineOperation struct {
	Kind   string `json:"kind"`
	Amount int    `json:"amount,omitempty"`
	Target string `json:"target,omitempty"`
}
type RhineStep struct {
	Operations []RhineOperation            `json:"operations"`
	Values     map[string]string           `json:"values,omitempty"`
	Branches   map[string][]RhineOperation `json:"branches,omitempty"`
}
type rhineOption struct {
	Label      string
	Operations []RhineOperation
}
type rhineDefinition struct {
	Card    Card
	Options map[string]rhineOption
	Board   bool
}

var rhineDefinitions = map[string]rhineDefinition{}

func isRhine(id string) bool {
	return strings.HasPrefix(id, "rhine-summer-") || strings.HasPrefix(id, "rhine-winter-")
}

func RhineCatalog(config ExpansionConfig) []Card {
	var out []Card
	for _, raw := range rhineCardFacts {
		if raw.Board && config.Normalized().Board != "tuscany" {
			continue
		}
		card := raw.Card
		_, card.Implemented = rhineDefinitions[card.ID]
		out = append(out, card)
	}
	return out
}

func rhineField(name, typ string, min, max int) map[string]any {
	return map[string]any{"name": name, "type": typ, "min": min, "max": max}
}
func rhineFields(op RhineOperation) []map[string]any {
	f := func(name, typ string, min, max int) []map[string]any {
		return []map[string]any{rhineField(name, typ, min, max)}
	}
	switch op.Kind {
	case "discard_cards":
		fs := f("cardIds", "cards", op.Amount, op.Amount)
		fs[0]["filter"] = op.Target
		return fs
	case "discard_grape":
		return f("grapes", "grapes", op.Amount, op.Amount)
	case "discard_wine":
		return f("wineIds", "wines", op.Amount, op.Amount)
	case "gain_grape":
		if op.Target == "" {
			return f("color", "color", 1, 1)
		}
	case "gain_wine":
		if op.Target == "" {
			fs := f("color", "rhine_wine", 1, 1)
			fs[0]["value"] = op.Amount
			return fs
		}
	case "plant", "plant_ignore_buildings", "plant_ignore_capacity":
		return f("plant", "plant", 1, op.Amount)
	case "harvest":
		return f("fields", "fields", 1, op.Amount)
	case "make", "make_exceptional":
		return f("recipes", "recipes", 1, op.Amount)
	case "build":
		fs := f("building", "rhine_build", 1, 1)
		fs[0]["discount"] = op.Amount
		return fs
	case "fill":
		fs := f("cardId", "cards", 1, 1)
		fs[0]["filter"] = "order"
		return append(fs, rhineField("wineIds", "wines", 1, 3))
	case "draw_any":
		return f("drawTypes", "rhine_draw", op.Amount, op.Amount)
	}
	return rhineExtraFields(op)
}

func rhineInput(ops []RhineOperation) []map[string]any {
	for _, op := range ops {
		if op.Kind == "draw" || op.Kind == "draw_new" || op.Kind == "choose" {
			return nil
		}
		if fs := rhineFields(op); len(fs) > 0 {
			return fs
		}
	}
	return nil
}
func (r *Room) rhinePrompt(p *Player, id string, ops []RhineOperation, values map[string]string) {
	r.visitorPrompt(p, id, "rhine_continue", []string{"continue"})
	c := &r.Choices[len(r.Choices)-1]
	c.Visitor.Rhine = &RhineStep{Operations: ops, Values: values}
	c.OptionFields = map[string][]map[string]any{"continue": rhineInput(ops)}
	c.Labels = map[string]string{"continue": rhineOperationLabel(ops[0])}
	r.rhineDynamicFields(c, p)
}

func rhineOperationLabel(op RhineOperation) string {
	labels := map[string]string{"discard_cards": "选择要弃置的手牌", "discard_grape": "选择要弃置的葡萄", "discard_wine": "选择要弃置的酒", "gain_grape": "选择葡萄颜色", "gain_wine": "选择酒种", "plant": "种植葡萄藤", "plant_ignore_buildings": "种植：忽略前置建筑", "plant_ignore_capacity": "种植：可超田地总值", "harvest": "选择收获田地", "make": "选择酿酒配方", "make_exceptional": "酿价值至少4的酒", "fill": "选择订单与交付酒", "draw_any": "先确定全部抽牌颜色", "build": "选择建造建筑"}
	if x := labels[op.Kind]; x != "" {
		return x
	}
	return "继续结算"
}

func (r *Room) rhineDecorate(c *Choice) {
	if c.Visitor == nil || !isRhine(c.Visitor.CardID) {
		return
	}
	if c.Visitor.Rhine != nil {
		return
	}
	d, ok := rhineDefinitions[c.Visitor.CardID]
	if !ok {
		return
	}
	c.OptionFields = map[string][]map[string]any{}
	c.Labels = map[string]string{}
	for key, o := range d.Options {
		c.Labels[key] = o.Label
		c.OptionFields[key] = rhineInput(o.Operations)
	}
	r.rhineDynamicFields(c, r.Player(c.PlayerID))
}

func (r *Room) rhineVisitorEffect(p *Player, s *VisitorStep, a Action) error {
	var ops []RhineOperation
	values := map[string]string{}
	if s.Rhine != nil {
		if s.Rhine.Branches != nil {
			branch, ok := s.Rhine.Branches[a.Option]
			if !ok {
				return fmt.Errorf("无效Rhine续步")
			}
			ops = append(ops, branch...)
		}
		ops = append(ops, s.Rhine.Operations...)
		for k, v := range s.Rhine.Values {
			values[k] = v
		}
	} else {
		d, ok := rhineDefinitions[s.CardID]
		if !ok {
			return fmt.Errorf("Rhine访客尚未完成：%s", s.CardID)
		}
		o, ok := d.Options[a.Option]
		if !ok {
			return fmt.Errorf("无效Rhine选项")
		}
		ops = append(ops, o.Operations...)
	}
	consumed := false
	for i, op := range ops {
		if op.Kind == "choose" {
			if op.Target == "engineer" && rhineValue(values, "engineerRemaining") <= 0 {
				continue
			}
			r.rhineMenu(p, s.CardID, op.Target, ops[i+1:], values)
			return nil
		}
		input := len(rhineFields(op)) > 0
		if input && consumed {
			r.rhinePrompt(p, s.CardID, ops[i:], values)
			return nil
		}
		if e := r.rhineOperation(p, s.CardID, op, a, values); e != nil {
			return e
		}
		if input {
			consumed = true
		}
		if op.Kind == "draw" || op.Kind == "draw_new" {
			consumed = true
		}
	}
	return nil
}

func (r *Room) rhineOperation(p *Player, id string, op RhineOperation, a Action, values map[string]string) error {
	switch op.Kind {
	case "coins":
		p.Coins += op.Amount
	case "vp":
		p.VP += op.Amount
	case "income":
		p.Income = min(5, p.Income+op.Amount)
	case "pay":
		return visitorPay(p, op.Amount, 0)
	case "lose_vp":
		return visitorPay(p, 0, op.Amount)
	case "lose_income":
		if p.Income < op.Amount {
			return fmt.Errorf("年收入不足%d", op.Amount)
		}
		p.Income -= op.Amount
	case "draw":
		for n := 0; n < op.Amount; n++ {
			r.draw(p, op.Target)
		}
	case "draw_any":
		if len(a.Colors) != op.Amount {
			return fmt.Errorf("须先选定%d张牌的全部类型", op.Amount)
		}
		for _, color := range a.Colors {
			if color != "vine" && color != "order" && color != "summer" && color != "winter" && !(color == "structure" && r.Config.Structures) {
				return fmt.Errorf("无效抽牌类型")
			}
		}
		for _, color := range a.Colors {
			r.draw(p, color)
		}
	case "discard_cards":
		for _, cid := range a.CardIDs {
			if op.Target != "" && cardIndex(p, cid, op.Target) < 0 {
				return fmt.Errorf("所选手牌类型不符")
			}
		}
		return r.visitorCards(p, a.CardIDs, op.Amount, false)
	case "discard_grape":
		if len(a.Grapes) != op.Amount {
			return fmt.Errorf("须选择%d颗葡萄", op.Amount)
		}
		seen := map[int]bool{}
		for _, i := range a.Grapes {
			if i < 0 || i >= len(p.Grapes) || seen[i] {
				return fmt.Errorf("无效或重复葡萄")
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
	case "discard_wine":
		if op.Amount != 1 {
			return fmt.Errorf("暂不支持此弃酒数量")
		}
		return visitorWine(p, a.WineIDs, 1)
	case "gain_grape":
		color := op.Target
		if color == "" {
			color = a.Color
		}
		if color != "red" && color != "white" {
			return fmt.Errorf("须选择红或白葡萄")
		}
		addGrape(p, color, op.Amount)
	case "gain_wine":
		color := op.Target
		if color == "" {
			color = a.Color
		}
		return rhineGainWine(p, color, op.Amount, false)
	case "gain_wine_exceptional":
		return rhineGainWine(p, op.Target, op.Amount, true)
	case "plant", "plant_ignore_buildings", "plant_ignore_capacity":
		return r.visitorPlant(p, a, op.Amount, op.Kind == "plant_ignore_buildings", op.Kind == "plant_ignore_capacity")
	case "harvest":
		return r.visitorHarvest(p, a, op.Amount)
	case "make":
		return r.visitorMake(p, a, op.Amount)
	case "make_exceptional":
		return rhineMakeExceptional(p, a, op.Amount, 4)
	case "fill":
		a.Space = "fill_order"
		return r.perform(p, a)
	case "build":
		return r.visitorBuild(p, a.Building, op.Amount)
	case "upgrade":
		return r.visitorUpgrade(p, op.Amount)
	case "train":
		return r.visitorTrain(p, op.Amount, op.Target == "now", a.SpecialWorker)
	case "age_all":
		for n := 0; n < op.Amount; n++ {
			ageWines(p)
		}
	case "if_opponent_vp_draw":
		for _, q := range r.Players {
			if q.ID != p.ID && q.VP >= op.Amount {
				r.draw(p, op.Target)
				break
			}
		}
	default:
		return r.rhineExtraOperation(p, id, op, a, values)
	}
	return nil
}

func rhineGainWine(p *Player, typ string, value int, exceptional bool) error {
	minimum := map[string]int{"red": 1, "white": 1, "blush": 4, "sparkling": 7}[typ]
	if minimum == 0 {
		return fmt.Errorf("无效酒种")
	}
	if !exceptional {
		value = min(value, cellar(p))
	}
	value = min(value, 9)
	for ; value >= minimum; value-- {
		free := true
		for _, w := range p.Wines {
			if w.Type == typ && w.Value == value {
				free = false
				break
			}
		}
		if free {
			p.Wines = append(p.Wines, Wine{ID: NewID(), Type: typ, Value: value})
			return nil
		}
	}
	return fmt.Errorf("酒窖、最低价值或空酒槽不满足")
}

func (r *Room) rhineOptionReason(p *Player, s *VisitorStep, option string) string {
	d, ok := rhineDefinitions[s.CardID]
	if !ok {
		return "此访客尚未实现"
	}
	if d.Board && !r.tuscany() {
		return "需要Tuscany主板"
	}
	values := map[string]string{}
	if s.Rhine != nil {
		values = s.Rhine.Values
	}
	ops := rhinePendingOperations(s, option)
	if r.rhineFeasible(p, s.CardID, ops, values) {
		return ""
	}
	if len(ops) > 0 {
		op := ops[0]
		switch op.Kind {
		case "pay":
			if p.Coins < op.Amount {
				return fmt.Sprintf("还差%d金币", op.Amount-p.Coins)
			}
		case "lose_income":
			if p.Income < op.Amount {
				return fmt.Sprintf("需要%d年收入", op.Amount)
			}
		case "lose_vp":
			if p.VP-op.Amount < -5 {
				return "支付后分数低于−5"
			}
		}
	}
	return "现有资源无法完成该分支及其后续步骤"
}

func (r *Room) rhinePlayable(p *Player, id string) error {
	d, ok := rhineDefinitions[id]
	if !ok {
		return fmt.Errorf("Rhine访客尚未实现")
	}
	q := probePlayer(p)
	for i, c := range q.Hand {
		if c.ID == id {
			removeCard(q, i)
			break
		}
	}
	for option := range d.Options {
		if r.rhineOptionReason(q, &VisitorStep{CardID: id, Stage: "effect"}, option) == "" {
			return nil
		}
	}
	return fmt.Errorf("没有可执行的Rhine效果分支")
}

// The JSON copy is useful for deterministic whole-room preflight without sharing
// slices or private continuation maps with the live game.
func rhineCloneRoom(r *Room) *Room {
	b, _ := json.Marshal(r)
	var q Room
	_ = json.Unmarshal(b, &q)
	return &q
}
