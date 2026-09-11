package game

import (
	"fmt"
)

type VisitorStep struct {
	Rhine     *RhineStep `json:"rhine,omitempty"`
	MoorCount int        `json:"moorCount,omitempty"`
	CardID    string     `json:"cardId"`
	Stage     string     `json:"stage"`
	ActorID   string     `json:"actorId"`
}
type ChoiceField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Min  int    `json:"min"`
	Max  int    `json:"max"`
}
type visitorDef struct {
	Summary string
	Options []string
}

var visitorDefs = map[string]visitorDef{}

func (r *Room) visitorPrompt(p *Player, cardID, stage string, options []string) {
	r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: p.ID, Kind: "visitor", Options: options, Count: 1, Visitor: &VisitorStep{CardID: cardID, Stage: stage, ActorID: r.Context.ActorID}})
	r.decorateMoorChoice(&r.Choices[len(r.Choices)-1], p)
	r.rhineDecorate(&r.Choices[len(r.Choices)-1])
}
func (r *Room) startVisitor(p *Player, a Action, bonus bool) error {
	typ := "summer"
	if a.Space == "winter_visitor" {
		typ = "winter"
	}
	if !visitorInSeason(p, a.CardID, typ) {
		return fmt.Errorf("非自己的当季访客")
	}
	if _, ok := visitorDefs[a.CardID]; !ok {
		return fmt.Errorf("EE_VISITOR_NOT_IMPLEMENTED: %s", a.CardID)
	}
	if a.CardID == "summer-29" && r.Context != nil && r.Context.Step == "perform" {
		available := *p
		if a.Large {
			available.LargeWorker = false
		} else {
			available.Workers--
		}
		if e := r.specialPlayable(&available, a.CardID); e != nil {
			return e
		}
	}
	if r.Context == nil {
		r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Space: a.Space}
	}
	r.Context.Remaining = 0
	if bonus {
		r.Context.Remaining = 1
	}
	return r.queueVisitor(p, a.CardID)
}
func (r *Room) queueVisitor(p *Player, id string) error {
	d, ok := visitorDefs[id]
	if !ok {
		return fmt.Errorf("EE_VISITOR_NOT_IMPLEMENTED: %s", id)
	}
	i := cardIndex(p, id, "summer")
	if i < 0 {
		i = cardIndex(p, id, "winter")
	}
	if i < 0 {
		return fmt.Errorf("访客不在手中")
	}
	if e := r.mandatoryVisitorPlayable(p, id); e != nil {
		return e
	}
	if e := r.specialPlayable(p, id); e != nil {
		return e
	}
	c := p.Hand[i]
	removeCard(p, i)
	if id == "winter-19" || moorHeldCard(id) {
		r.Context.Held = append(r.Context.Held, c)
	} else if id != "rhine-summer-sonInLaw" {
		r.discard(c)
	}
	if has(p, "inn") {
		p.Coins++
	}
	r.Context.SourceCardID = id
	r.visitorPrompt(p, id, "effect", d.Options)
	return nil
}
func (r *Room) resolveVisitorChoice(id string, a Action) error {
	c := r.Choices[0]
	p := r.Player(id)
	valid := false
	for _, o := range c.Options {
		if a.Option == o {
			valid = true
		}
	}
	if !valid {
		return fmt.Errorf("无效访客选项")
	}
	r.Choices = r.Choices[1:]
	if c.Visitor.Stage == "nested_action_done" {
		// Continue to the saved parent through the ordinary completion path.
	} else if c.Visitor.Stage == "second" {
		if a.Option == "play" {
			typ := "summer"
			if r.Context.Space == "winter_visitor" {
				typ = "winter"
			}
			if !visitorInSeason(p, a.CardID, typ) {
				return fmt.Errorf("非自己的当季访客")
			}
			if e := r.queueVisitor(p, a.CardID); e != nil {
				return e
			}
		}
	} else {
		if e := r.visitorBeforeEffect(p, c.Visitor, a); e != nil {
			return e
		}
		effectContext := r.Context
		if e := r.visitorEffect(p, c.Visitor, a); e != nil {
			return e
		}
		if r.Context != effectContext {
			if err := r.finishNestedParentEffect(p, c.Visitor, effectContext); err != nil {
				return err
			}
			if len(r.Choices) == 1 && r.Choices[0].Visitor != nil && r.Choices[0].Visitor.Stage == "nested_action_done" {
				r.Choices = nil
				r.finishVisitorContext()
				return nil
			}
			r.TurnID = r.Choices[0].PlayerID
			return nil
		}
		if e := r.visitorContinuationFeasible(p, c.Visitor); e != nil {
			return e
		}
		if c.Visitor.Stage == "effect" {
			if r.queueStructureVisitorBonus(p) {
				r.TurnID = r.Choices[0].PlayerID
				return nil
			}
		}
		if e := r.visitorContinuationFeasible(p, c.Visitor); e != nil {
			return e
		}
	}
	if len(r.Choices) == 0 && r.Context.Remaining > 0 {
		r.Context.Remaining--
		r.visitorPrompt(r.Player(r.Context.ActorID), "", "second", []string{"play", "skip"})
	}
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	r.finishVisitorContext()
	return nil
}
func visitorPay(p *Player, coins, vp int) error {
	if p.Coins < coins || p.VP-vp < -5 {
		return fmt.Errorf("金币不足或分数将低于-5")
	}
	p.Coins -= coins
	p.VP -= vp
	return nil
}
func (r *Room) visitorBuild(p *Player, b string, discount int) error {
	if d, ok := structureDef(b); ok && r.Config.Structures {
		if p.hasStructure(d.ID) {
			return fmt.Errorf("已建造该结构")
		}
		a := Action{Building: structureCardID(d.ID), Mode: "mat", StructureSlot: -1}
		if _, _, ok := p.structureCardInHand(d.ID); !ok {
			return fmt.Errorf("结构牌不在手中")
		}
		if len(p.StructureSlots) >= 2 && p.StructureSlots[0] != "" && p.StructureSlots[1] != "" {
			for i, f := range p.Fields {
				if !f.Sold && len(f.Vines) == 0 && f.Structure == "" {
					a.Mode, a.Field = "field", i
					break
				}
			}
		}
		return r.buildStructure(p, a, discount)
	}
	cost, ok := costs[b]
	if !ok {
		return fmt.Errorf("无效建筑")
	}
	if has(p, "workshop") {
		cost = max(0, cost-1)
	}
	if discount > cost {
		discount = cost
	}
	if p.Coins < cost-discount {
		return fmt.Errorf("金币不足")
	}
	p.Coins += discount
	return r.perform(p, Action{Space: "build", Building: b})
}
func (r *Room) visitorCards(p *Player, ids []string, n int, visitors bool) error {
	if len(ids) != n {
		return fmt.Errorf("须选择%d张手牌", n)
	}
	seen := map[string]bool{}
	selected := make([]Card, 0, n)
	for _, id := range ids {
		i := -1
		for j, c := range p.Hand {
			if c.ID == id {
				i = j
			}
		}
		if i < 0 || seen[id] {
			return fmt.Errorf("无效或重复手牌")
		}
		seen[id] = true
		c := p.Hand[i]
		if visitors && c.Type != "summer" && c.Type != "winter" {
			return fmt.Errorf("须为访客")
		}
		selected = append(selected, c)
	}
	// Commit only after every selected card is validated. This keeps Guest
	// House, Barn, and two-card trades atomic on invalid input.
	for _, c := range selected {
		r.discard(c)
		for i, handCard := range p.Hand {
			if handCard.ID == c.ID {
				removeCard(p, i)
				break
			}
		}
	}
	return nil
}
func visitorWine(p *Player, ids []string, min int) error {
	if len(ids) != 1 {
		return fmt.Errorf("须选择1瓶酒")
	}
	for i, w := range p.Wines {
		if w.ID == ids[0] && w.Value >= min {
			p.Wines = append(p.Wines[:i], p.Wines[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("酒不存在或品质不足")
}
func visitorGrape(p *Player, indices []int) error {
	if len(indices) != 1 || indices[0] < 0 || indices[0] >= len(p.Grapes) {
		return fmt.Errorf("须选择1颗葡萄")
	}
	i := indices[0]
	p.Grapes = append(p.Grapes[:i], p.Grapes[i+1:]...)
	return nil
}
func (r *Room) visitorPlant(p *Player, a Action, n int, noStructures, noLimit bool) error {
	ids, fs := a.CardIDs, a.Fields
	if len(ids) == 0 && a.CardID != "" {
		ids = []string{a.CardID}
		fs = []int{a.Field}
	}
	if len(ids) == 0 || len(ids) > n || len(fs) != len(ids) {
		return fmt.Errorf("种植数量无效")
	}
	for i, id := range ids {
		idx := cardIndex(p, id, "vine")
		if idx < 0 || fs[i] < 0 || fs[i] >= len(p.Fields) {
			return fmt.Errorf("无效藤或田地")
		}
		original := p.Hand[idx]
		if noStructures {
			p.Hand[idx].Trellis = false
			p.Hand[idx].Irrigation = false
		}
		cap := p.Fields[fs[i]].Capacity
		if noLimit {
			p.Fields[fs[i]].Capacity = 1000
		}
		if e := r.perform(p, Action{Space: "plant", CardID: id, Field: fs[i]}); e != nil {
			return e
		}
		p.Fields[fs[i]].Capacity = cap
		vines := p.Fields[fs[i]].Vines
		vines[len(vines)-1] = original
	}
	return nil
}
func (r *Room) visitorUproot(p *Player, a Action, n int) error {
	if len(a.CardIDs) != n || len(a.Fields) != n {
		return fmt.Errorf("须拔除%d藤", n)
	}
	for i, id := range a.CardIDs {
		if e := uproot(p, Action{CardID: id, Field: a.Fields[i]}); e != nil {
			return e
		}
	}
	return r.visitorCards(p, a.CardIDs, n, false)
}
func (r *Room) visitorHarvest(p *Player, a Action, n int) error {
	if a.HarvestAll && has(p, "harvest_machine") {
		return r.performEEPlacement(p, Action{Space: "harvest", Mode: "all"}, false)
	}
	if len(a.Fields) == 0 || len(a.Fields) > n {
		return fmt.Errorf("收获超限")
	}
	for _, f := range a.Fields {
		if e := r.perform(p, Action{Space: "harvest", Field: f}); e != nil {
			return e
		}
	}
	return nil
}
func (r *Room) visitorMake(p *Player, a Action, n int) error {
	if len(a.Recipes) == 0 {
		return fmt.Errorf("至少酿1酒")
	}
	return makeActionWines(p, a, n)
}
func (r *Room) visitorTrain(p *Player, cost int, now bool, workerType ...string) error {
	id := "regular"
	if len(workerType) > 0 && workerType[0] != "" {
		id = workerType[0]
	}
	if cost < 0 {
		return fmt.Errorf("培训基础费用无效")
	}
	if err := r.trainWorker(p, id, 4-cost); err != nil {
		return err
	}
	if now {
		if id == "grande" {
			p.LargeWorker = true
		} else if id == "regular" {
			p.Workers++
		} else {
			p.SpecialWorkerReady[id] = r.Year
		}
	}
	return nil
}
func (r *Room) visitorUpgrade(p *Player, discount int) error {
	b := "medium_cellar"
	if has(p, b) {
		b = "large_cellar"
	}
	return r.visitorBuild(p, b, discount)
}
