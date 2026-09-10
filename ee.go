package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:embed ee_cards.json
var cardJSON []byte

// Choice is persisted; only the head is actionable. Never put secret cards in Options.
type Choice struct {
	ID       string            `json:"id"`
	PlayerID string            `json:"playerId"`
	Kind     string            `json:"kind"`
	Options  []string          `json:"options"`
	Count    int               `json:"count"`
	Visitor  *VisitorStep      `json:"visitor,omitempty"`
	Schema   []ChoiceField     `json:"schema,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// ActionContext reserves a serializable continuation for the visitor state machine.
type ActionContext struct {
	TriggerSeat  *Seat             `json:"triggerSeat,omitempty"`
	PassAfter    bool              `json:"passAfter,omitempty"`
	Held         []Card            `json:"held,omitempty"`
	Escrow       map[string][]Card `json:"escrow,omitempty"`
	ActorID      string            `json:"actorId"`
	SourceCardID string            `json:"sourceCardId"`
	Space        string            `json:"space"`
	Step         string            `json:"step"`
	Remaining    int               `json:"remaining"`
	ReturnTurnID string            `json:"returnTurnId"`
	Data         map[string]string `json:"data"`
}

func catalog() []Card {
	var c []Card
	if e := json.Unmarshal(cardJSON, &c); e != nil {
		panic(e)
	}
	for i := range c {
		if d, ok := visitorDefs[c[i].ID]; ok {
			c[i].Implemented = true
			c[i].Description = d.Summary
		}
	}
	return c
}
func shuffle(c []Card) {
	for i := len(c) - 1; i > 0; i-- {
		j := rnd(i + 1)
		c[i], c[j] = c[j], c[i]
	}
}
func (r *Room) initDecks() {
	r.Decks = map[string][]Card{}
	r.Discards = map[string][]Card{}
	for _, c := range catalog() {
		r.Decks[c.Type] = append(r.Decks[c.Type], c)
	}
	for _, d := range r.Decks {
		shuffle(d)
	}
}
func (r *Room) take(t string) (Card, bool) {
	if len(r.Decks[t]) == 0 && len(r.Discards[t]) > 0 {
		r.Decks[t] = r.Discards[t]
		r.Discards[t] = nil
		shuffle(r.Decks[t])
	}
	d := r.Decks[t]
	if len(d) == 0 {
		return Card{}, false
	}
	c := d[len(d)-1]
	r.Decks[t] = d[:len(d)-1]
	return c, true
}
func (r *Room) draw(p *Player, t string) {
	if c, ok := r.take(t); ok {
		p.Hand = append(p.Hand, c)
	}
}
func (r *Room) discard(c Card) {
	if r.Discards == nil {
		r.Discards = map[string][]Card{}
	}
	r.Discards[c.Type] = append(r.Discards[c.Type], c)
}
func (r *Room) deckCounts() map[string]map[string]int {
	out := map[string]map[string]int{}
	for t, d := range r.Decks {
		out[t] = map[string]int{"deck": len(d), "discard": len(r.Discards[t])}
	}
	return out
}
func (r *Room) enqueue(id, kind string, options []string, count int) {
	r.Choices = append(r.Choices, Choice{ID: uid(), PlayerID: id, Kind: kind, Options: options, Count: count})
}
func (r *Room) choiceView(id string) any {
	if len(r.Choices) == 0 {
		return nil
	}
	c := r.Choices[0]
	if c.PlayerID != id {
		return map[string]string{"playerId": c.PlayerID, "kind": c.Kind}
	}
	return c
}

var mamaDraws = []string{"vso", "vsw", "vow", "sow", "vvs", "vvo", "vvw", "vss", "sso", "ssw", "voo", "soo", "oow", "vww", "sww", "oww", "vw", "so"}

type papaGift struct {
	Cash        int
	Gift        string
	Alternative int
}

var papaGifts = []papaGift{{4, "trellis", 2}, {3, "irrigation", 3}, {5, "yoke", 2}, {3, "medium_cellar", 4}, {2, "cottage", 4}, {1, "windmill", 5}, {0, "tasting_room", 6}, {5, "trellis", 1}, {4, "irrigation", 2}, {6, "yoke", 1}, {4, "medium_cellar", 3}, {3, "cottage", 3}, {2, "windmill", 4}, {1, "tasting_room", 5}, {2, "worker", 4}, {3, "worker", 3}, {4, "vp", 3}, {5, "vp", 2}}

func cardNumber(c Card) int {
	n, _ := strconv.Atoi(c.ID[strings.LastIndex(c.ID, "-")+1:])
	return n - 1
}
func (r *Room) setupParents() {
	r.Phase = "setup"
	r.Resume = "spring"
	for _, p := range r.Players {
		p.Coins = 0
		p.Hand = []Card{}
		p.Mama, _ = r.take("mama")
		p.Papa, _ = r.take("papa")
		for _, ch := range mamaDraws[cardNumber(p.Mama)] {
			r.draw(p, map[rune]string{'v': "vine", 's': "summer", 'o': "order", 'w': "winter"}[ch])
		}
		if cardNumber(p.Mama) >= 16 {
			p.Coins += 2
		}
		g := papaGifts[cardNumber(p.Papa)]
		p.Coins += g.Cash
		r.enqueue(p.ID, "papa", []string{"gift", "coins"}, 1)
	}
	r.TurnID = r.Choices[0].PlayerID
}
func (r *Room) resolveChoice(id string, a Action) error {
	c := r.Choices[0]
	if a.Type != "choose" || c.PlayerID != id || c.ID != a.ChoiceID {
		return fmt.Errorf("须由待决玩家提交匹配choiceId的choose")
	}
	if c.Kind == "planner" {
		return r.resolvePlanner(r.player(id), a)
	}
	if c.Visitor != nil {
		return r.resolveVisitorChoice(id, a)
	}
	p := r.player(id)
	if c.Kind != "discard" {
		ok := false
		for _, o := range c.Options {
			if o == a.Option {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("无效选项")
		}
	}
	switch c.Kind {
	case "papa":
		g := papaGifts[cardNumber(p.Papa)]
		if a.Option == "coins" {
			p.Coins += g.Alternative
		} else {
			switch g.Gift {
			case "vp":
				p.VP++
			case "worker":
				p.TotalWorkers++
				p.Workers++
			default:
				p.Buildings = append(p.Buildings, g.Gift)
			}
		}
		p.PapaResolved = true
	case "fall":
		for _, color := range strings.Split(a.Option, "_") {
			r.draw(p, color)
		}
	case "discard":
		if len(a.CardIDs) != c.Count {
			return fmt.Errorf("必须弃置%d张手牌", c.Count)
		}
		selected := map[string]bool{}
		for _, id := range a.CardIDs {
			if selected[id] {
				return fmt.Errorf("重复卡牌")
			}
			found := false
			for _, card := range p.Hand {
				if card.ID == id {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("非自己的手牌")
			}
			selected[id] = true
		}
		out := []Card{}
		for _, card := range p.Hand {
			if selected[card.ID] {
				r.discard(card)
			} else {
				out = append(out, card)
			}
		}
		p.Hand = out
	default:
		return fmt.Errorf("CHOICE_NOT_IMPLEMENTED: %s", c.Kind)
	}
	r.Choices = r.Choices[1:]
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	resume := r.Resume
	r.Resume = ""
	switch resume {
	case "spring":
		r.spring()
	case "winter":
		r.Phase = "winter"
		r.startNextPlanner()
	case "end_year":
		r.finishYear()
	}
	return nil
}
func (r *Room) endYear() {
	r.Phase = "year_end"
	r.Resume = "end_year"
	for _, p := range r.ordered() {
		if len(p.Hand) > 7 {
			r.enqueue(p.ID, "discard", nil, len(p.Hand)-7)
		}
	}
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	r.Resume = ""
	r.finishYear()
}
func scoreKey(p *Player) [4]int {
	k := [4]int{p.VP, p.Coins, 0, 0}
	for _, w := range p.Wines {
		k[2] += w.Value
	}
	for _, g := range p.Grapes {
		k[3] += g.Value
	}
	return k
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

func (r *Room) parentOptions(id string) any {
	p := r.player(id)
	if p == nil || p.Papa.ID == "" {
		return nil
	}
	g := papaGifts[cardNumber(p.Papa)]
	return map[string]any{"gift": g.Gift, "coins": g.Alternative, "baseCoins": g.Cash}
}

var spaceBonuses = map[string]string{"draw_vine": "+1 vine", "draw_order": "+1 order", "tour": "+1 coin", "build": "discount 1", "plant": "plant up to 2", "harvest": "harvest up to 2", "make_wine": "make up to 3", "train": "discount 1", "fill_order": "+1 VP", "sell_grapes": "+1 VP", "summer_visitor": "play up to 2", "winter_visitor": "play up to 2"}

// Slot 1 is the printed bonus, available only with 3+ players. Two-player slot 1 has no bonus.
// Slot -1 is grande overflow and never blocks or grants a printed bonus.
func placement(s *Space, p *Player, a Action) (int, bool, error) {
	if s.ID == "gain_coin" {
		return -1, false, nil
	}
	if s.ID == "yoke" {
		if !has(p, "yoke") || p.YokeUsed {
			return 0, false, fmt.Errorf("需要未使用的轭")
		}
		return 1, false, nil
	}
	occupied := map[int]bool{}
	for _, seat := range s.Occupied {
		if seat.Slot > 0 {
			occupied[seat.Slot] = true
		}
	}
	slot := a.Slot
	if slot == 0 {
		for i := 1; i <= s.Capacity; i++ {
			if !occupied[i] {
				slot = i
				break
			}
		}
		if slot == 0 && a.Large {
			slot = -1
		}
	}
	if slot == -1 {
		if !a.Large {
			return 0, false, fmt.Errorf("仅大工人可溢出")
		}
		for i := 1; i <= s.Capacity; i++ {
			if !occupied[i] {
				return 0, false, fmt.Errorf("有空格时大工人须占普通格")
			}
		}
		return -1, false, nil
	}
	if slot < 1 || slot > s.Capacity || occupied[slot] {
		return 0, false, fmt.Errorf("行动格已占用或不可用")
	}
	return slot, slot == 1 && s.Capacity >= 2, nil
}
func (r *Room) performPlacement(p *Player, a Action, bonus bool) error {
	bonus = bonus && !a.DeclineBonus
	switch a.Space {
	case "summer_visitor", "winter_visitor":
		return r.startVisitor(p, a, bonus)
	case "gain_coin":
		p.Coins++
		return nil
	case "yoke":
		if a.Mode == "uproot" {
			if e := uproot(p, a); e != nil {
				return e
			}
		} else if a.Mode == "harvest" {
			b := a
			b.Space = "harvest"
			if e := r.perform(p, b); e != nil {
				return e
			}
		} else {
			return fmt.Errorf("轭须选harvest或uproot")
		}
		p.YokeUsed = true
		return nil
	case "sell_grapes":
		if a.Mode == "buy_field" || a.Mode == "sell_field" {
			if a.Field < 0 || a.Field >= len(p.Fields) {
				return fmt.Errorf("无效田地")
			}
			f := &p.Fields[a.Field]
			if a.Mode == "sell_field" {
				if f.Sold || len(f.Vines) > 0 {
					return fmt.Errorf("仅能卖出未种植且未出售田地")
				}
				f.Sold = true
				p.Coins += f.Capacity
			} else {
				if !f.Sold || p.Coins < f.Capacity {
					return fmt.Errorf("田地未出售或金币不足")
				}
				f.Sold = false
				p.Coins -= f.Capacity
			}
			if bonus {
				p.VP++
			}
			return nil
		}
	case "plant":
		if a.Mode == "uproot" {
			return uproot(p, a)
		}
		if len(a.CardIDs) > 0 {
			limit := 1
			if bonus {
				limit = 2
			}
			if len(a.CardIDs) > limit || len(a.Fields) != len(a.CardIDs) {
				return fmt.Errorf("种植数量或田地数量无效")
			}
			for i, id := range a.CardIDs {
				b := a
				b.CardID = id
				b.Field = a.Fields[i]
				if e := r.perform(p, b); e != nil {
					return e
				}
			}
			return nil
		}
	case "harvest":
		if len(a.Fields) > 0 {
			limit := 1
			if bonus {
				limit = 2
			}
			if len(a.Fields) > limit {
				return fmt.Errorf("收获数量超过上限")
			}
			for _, f := range a.Fields {
				b := a
				b.Field = f
				if e := r.perform(p, b); e != nil {
					return e
				}
			}
			return nil
		}
	case "make_wine":
		limit := 2
		if bonus {
			limit = 3
		}
		if len(a.Recipes) == 0 {
			a.Recipes = [][]int{a.Grapes}
		}
		return makeWines(p, a.Recipes, limit)
	}
	// Grant discounts before the validated build/train; HTTP applies this only on a copy.
	if bonus && (a.Space == "build" || a.Space == "train") {
		p.Coins++
	}
	if e := r.perform(p, a); e != nil {
		return e
	}
	if bonus {
		switch a.Space {
		case "draw_vine":
			r.draw(p, "vine")
		case "draw_order":
			r.draw(p, "order")
		case "tour":
			p.Coins++
		case "fill_order", "sell_grapes":
			p.VP++
		}
	}
	return nil
}
func uproot(p *Player, a Action) error {
	if a.Field < 0 || a.Field >= len(p.Fields) {
		return fmt.Errorf("无效田地")
	}
	f := &p.Fields[a.Field]
	for i, c := range f.Vines {
		if c.ID == a.CardID {
			p.Hand = append(p.Hand, c)
			f.Vines = append(f.Vines[:i], f.Vines[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("请选择该田地的葡萄藤")
}
