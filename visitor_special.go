package main

import (
	"encoding/json"
	"fmt"
)

// PlannedPlacement contains only public placement information, never future hand selections.
type PlannedPlacement struct {
	PlayerID string `json:"playerId"`
	Space    string `json:"space"`
	Slot     int    `json:"slot"`
	Large    bool   `json:"large"`
}

func init() {
	visitorDefs["summer-29"] = visitorDef{"放可用工人到冬季格，冬初执行", []string{"plan"}}
	visitorDefs["summer-33"] = visitorDef{"移到空起床行，领奖后结束本季", []string{"move"}}
	visitorDefs["winter-11"] = visitorDef{"右邻座失1分、给你2手牌或付你3金币", []string{"resolve"}}
}
func (r *Room) rightNeighbor(p *Player) *Player {
	for i, q := range r.Players {
		if q.ID == p.ID && len(r.Players) > 1 {
			return r.Players[(i+len(r.Players)-1)%len(r.Players)]
		}
	}
	return nil
}
func queenOptions(p *Player) []string {
	opts := []string{}
	if p == nil {
		return opts
	}
	if p.VP > -5 {
		opts = append(opts, "vp")
	}
	if len(p.Hand) >= 2 {
		opts = append(opts, "cards")
	}
	if p.Coins >= 3 {
		opts = append(opts, "coins")
	}
	return opts
}

// The FAQ requires the minimum mandatory effect to be possible before playing.
func (r *Room) specialPlayable(p *Player, id string) error {
	switch id {
	case "winter-11":
		if len(queenOptions(r.rightNeighbor(p))) == 0 {
			return fmt.Errorf("Queen右邻座无法完成任一强制分支")
		}
	case "summer-29":
		if r.Phase != "summer" || (p.Workers == 0 && !p.LargeWorker) {
			return fmt.Errorf("Planner需要可用工人及未来冬季")
		}
		free := false
		for i := range r.Spaces {
			sp := &r.Spaces[i]
			if sp.Season != "winter" {
				continue
			}
			if p.Workers > 0 {
				if _, _, e := placement(sp, p, Action{}); e == nil {
					free = true
				}
			}
			if p.LargeWorker {
				if _, _, e := placement(sp, p, Action{Large: true}); e == nil {
					free = true
				}
			}
		}
		if !free {
			return fmt.Errorf("Planner没有可预约的冬季格")
		}

	case "summer-33":
		if len(r.organizerRows()) == 0 {
			return fmt.Errorf("Organizer没有空起床行")
		}
	}
	return nil
}
func (r *Room) organizerRows() []int {
	rows := []int{}
	for i := 1; i <= 7; i++ {
		free := true
		for _, p := range r.Players {
			if p.Wake == i {
				free = false
			}
		}
		if free {
			rows = append(rows, i)
		}
	}
	return rows
}
func (r *Room) visitorSpecial(p *Player, s *VisitorStep, a Action) error {
	switch s.CardID {
	case "winter-11":
		if s.Stage == "effect" {
			q := r.rightNeighbor(p)
			opts := queenOptions(q)
			if len(opts) == 0 {
				return fmt.Errorf("Queen无可执行分支")
			}
			r.visitorPrompt(q, s.CardID, "reply", opts)
			return nil
		}
		actor := r.player(s.ActorID)
		if r.rightNeighbor(actor) != p {
			return fmt.Errorf("仅右邻座响应Queen")
		}
		switch a.Option {
		case "vp":
			return visitorPay(p, 0, 1)
		case "cards":
			return r.visitorTransfer(p, actor, a.CardIDs, 2, "")
		case "coins":
			if e := visitorPay(p, 3, 0); e != nil {
				return e
			}
			actor.Coins += 3
		default:
			return fmt.Errorf("无效Queen分支")
		}
	case "summer-33":
		free := false
		for _, i := range r.organizerRows() {
			if i == a.Slot {
				free = true
			}
		}
		if !free {
			return fmt.Errorf("须选择空起床行")
		}
		if a.Slot == 5 && a.Color != "summer" && a.Color != "winter" {
			return fmt.Errorf("第5行须选summer或winter")
		}
		for i := range r.WakeSlots {
			if r.WakeSlots[i].PlayerID == p.ID {
				r.WakeSlots[i].PlayerID = ""
			}
			if r.WakeSlots[i].Slot == a.Slot {
				r.WakeSlots[i].PlayerID = p.ID
			}
		}
		p.Wake = a.Slot
		switch a.Slot {
		case 2:
			r.draw(p, "vine")
		case 3:
			r.draw(p, "order")
		case 4:
			p.Coins++
		case 5:
			r.draw(p, a.Color)
		case 6:
			p.VP++
		case 7:
			r.claimGrayWorker(p)
		}
		// Do not discard the outer bonus visitor or a Manager continuation.
		r.Context.PassAfter = true
	case "summer-29":
		if e := r.specialPlayable(p, s.CardID); e != nil {
			return e
		}
		if a.Large {
			if !p.LargeWorker {
				return fmt.Errorf("大工人已用")
			}
		} else if p.Workers <= 0 {
			return fmt.Errorf("无可用工人")
		}
		for i := range r.Spaces {
			sp := &r.Spaces[i]
			if sp.ID != a.Space || sp.Season != "winter" {
				continue
			}
			slot, _, e := placement(sp, p, a)
			if e != nil {
				return e
			}
			if a.Large {
				p.LargeWorker = false
			} else {
				p.Workers--
			}
			sp.Occupied = append(sp.Occupied, Seat{PlayerID: p.ID, Slot: slot, Large: a.Large})
			r.Planned = append(r.Planned, PlannedPlacement{PlayerID: p.ID, Space: sp.ID, Slot: slot, Large: a.Large})
			return nil
		}
		return fmt.Errorf("Planner只能预约本年冬季行动")
	}
	return nil
}

// Complete the entire outer placement before passing. Winter-start work must
// finish before the first normal winter turn, not call next from its respondent.
func (r *Room) finishVisitorContext() {
	ctx := r.Context
	if ctx.PassAfter {
		r.player(ctx.ActorID).Passed = true
	}
	r.Context = nil
	if ctx.Step == "planner" {
		r.startNextPlanner()
		return
	}
	r.TurnID = ctx.ReturnTurnID
	r.next()
}
func (r *Room) startNextPlanner() {
	for len(r.Planned) > 0 {
		plan := r.Planned[0]
		present := false
		for _, sp := range r.Spaces {
			if sp.ID == plan.Space {
				for _, seat := range sp.Occupied {
					if seat.PlayerID == plan.PlayerID && seat.Slot == plan.Slot && seat.Large == plan.Large {
						present = true
					}
				}
			}
		}
		if !present || !r.plannerCanExecute(plan) {
			r.Planned = r.Planned[1:]
			continue
		}
		r.Context = &ActionContext{ActorID: plan.PlayerID, ReturnTurnID: r.ordered()[0].ID, Space: plan.Space, Step: "planner", TriggerSeat: &Seat{PlayerID: plan.PlayerID, Slot: plan.Slot, Large: plan.Large}}
		r.enqueue(plan.PlayerID, "planner", []string{"execute"}, 1)
		r.TurnID = plan.PlayerID
		return
	}
	r.TurnID = r.ordered()[0].ID
	if r.player(r.TurnID).Passed {
		r.next()
	}
}
func (r *Room) resolvePlanner(p *Player, a Action) error {
	if r.Phase != "winter" || len(r.Planned) == 0 || r.Context == nil || r.Context.Step != "planner" {
		return fmt.Errorf("无冬初预约")
	}
	plan := r.Planned[0]
	if plan.PlayerID != p.ID || a.Option != "execute" || (a.Space != "" && a.Space != plan.Space) {
		return fmt.Errorf("无效预约执行")
	}
	bonus := false
	for _, sp := range r.Spaces {
		if sp.ID == plan.Space {
			bonus = plan.Slot == 1 && sp.Capacity >= 2
		}
	}
	a.Space = plan.Space
	r.Choices = r.Choices[1:]
	if e := r.performPlacement(p, a, bonus); e != nil {
		return e
	}
	r.Planned = r.Planned[1:]
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
	} else {
		r.finishVisitorContext()
	}
	return nil
}

// Test minimal legal actions on isolated copies; an impossible reservation is
// exhausted without a reward, never refunded as a usable worker.
func (r *Room) plannerCanExecute(plan PlannedPlacement) bool {
	p := r.player(plan.PlayerID)
	bonus := false
	for _, sp := range r.Spaces {
		if sp.ID == plan.Space {
			bonus = plan.Slot == 1 && sp.Capacity >= 2
		}
	}
	probe := func(a Action) bool {
		data, _ := json.Marshal(r)
		var q Room
		_ = json.Unmarshal(data, &q)
		q.Choices = nil
		q.Context = &ActionContext{ActorID: p.ID, Space: plan.Space}
		a.Space = plan.Space
		return q.performPlacement(q.player(p.ID), a, bonus) == nil
	}
	switch plan.Space {
	case "draw_order", "train":
		return probe(Action{})
	case "harvest":
		for i := range p.Fields {
			if probe(Action{Field: i}) {
				return true
			}
		}
	case "make_wine":
		for i := range p.Grapes {
			if probe(Action{Recipes: [][]int{{i}}}) {
				return true
			}
			for j := i + 1; j < len(p.Grapes); j++ {
				if probe(Action{Recipes: [][]int{{i, j}}}) {
					return true
				}
				for k := j + 1; k < len(p.Grapes); k++ {
					if probe(Action{Recipes: [][]int{{i, j, k}}}) {
						return true
					}
				}
			}
		}
	case "fill_order":
		for _, c := range p.Hand {
			if c.Type != "order" {
				continue
			}
			var selectWine func(int, []string) bool
			selectWine = func(start int, ids []string) bool {
				if len(ids) == len(c.Requirements) {
					return probe(Action{CardID: c.ID, WineIDs: ids})
				}
				for i := start; i < len(p.Wines); i++ {
					if selectWine(i+1, append(append([]string{}, ids...), p.Wines[i].ID)) {
						return true
					}
				}
				return false
			}
			if selectWine(0, nil) {
				return true
			}
		}
	case "winter_visitor":
		for _, c := range p.Hand {
			if c.Type == "winter" && probe(Action{CardID: c.ID}) {
				return true
			}
		}
	}
	return false
}
func (r *Room) cancelPlanned(space string, seat Seat) {
	out := r.Planned[:0]
	for _, p := range r.Planned {
		if p.Space == space && p.PlayerID == seat.PlayerID && p.Slot == seat.Slot && p.Large == seat.Large {
			continue
		}
		out = append(out, p)
	}
	r.Planned = out
}
