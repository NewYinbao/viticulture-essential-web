package game

import "fmt"

// The attachment resolves before the reserved action. Test only the current
// field's public yield so a make action is not discarded before its harvest.
func (r *Room) rhinePlannerHarvestCanEnable(plan PlannedPlacement) bool {
	if r.Config.Visitors != "rhine" || r.Phase != "winter" {
		return false
	}
	p := r.Player(plan.PlayerID)
	if p == nil || p.RhineSonField == nil || p.RhineWinterActionYear == r.Year || p.RhineWinterOfferYear == r.Year {
		return false
	}
	for red := 0; red <= 3; red++ {
		for white := 0; white <= 3-red; white++ {
			q := rhineCloneRoom(r)
			qp := q.Player(p.ID)
			q.Context = &ActionContext{ActorID: p.ID, Step: "rhine_first_winter"}
			q.Choices = nil
			if q.rhineSonOperation(qp, RhineOperation{Kind: "son_harvest"}, Action{Grapes: []int{red, white}}) == nil && q.plannerCanExecute(plan) {
				return true
			}
		}
	}
	return false
}

func (r *Room) rhineFinishSkippedPlannerHarvest(ctx *ActionContext) bool {
	if ctx == nil || ctx.Step != "rhine_first_winter" || ctx.Parent == nil {
		return false
	}
	if ctx.Parent.Step == "messenger" {
		idx := -1
		var plan PlannedPlacement
		for i, candidate := range r.MessengerPlans {
			if candidate.PlayerID == ctx.Parent.ActorID && candidate.Space == ctx.Parent.Space {
				idx, plan = i, candidate
				break
			}
		}
		if idx < 0 || r.plannerCanExecute(plan) {
			return false
		}
		r.MessengerPlans = append(r.MessengerPlans[:idx], r.MessengerPlans[idx+1:]...)
		if p := r.Player(plan.PlayerID); p != nil {
			r.AddLog(p.Name + " 的信使预约无法执行，未取得行动格奖励")
		}
		r.Context = nil
		r.Choices = nil
		r.TurnID = plan.PlayerID
		r.next()
		return true
	}
	if ctx.Parent.Step != "planner" || len(r.Planned) == 0 || r.plannerCanExecute(r.Planned[0]) {
		return false
	}
	// Declining the optional harvest may leave the reservation impossible. The
	// ordinary planner expiry path retains/exhausts its worker and moves on.
	r.Context = nil
	r.Choices = nil
	r.startNextPlanner()
	return true
}

// Son-in-law is a permanent field attachment. Its optional harvest is offered
// before the first winter action, including a reserved action, so its grapes can
// be used by that action. Choosing it never consumes a worker or advances turn.
func (r *Room) rhinePrepareWinterTurn() {
	if r.Config.Visitors != "rhine" || r.Phase != "winter" {
		return
	}
	reserved := r.Context != nil && (r.Context.Step == "planner" || r.Context.Step == "messenger") && len(r.Choices) > 0 && r.Choices[0].Kind == r.Context.Step
	if !reserved && (r.Context != nil || len(r.Choices) > 0) {
		return
	}
	p := r.Player(r.TurnID)
	if p == nil {
		return
	}
	alreadyActed := p.RhineWinterActionYear == r.Year
	if reserved {
		p.RhineWinterActionYear = r.Year
	}
	if p.RhineSonField == nil || alreadyActed || p.RhineWinterOfferYear == r.Year {
		return
	}
	p.RhineWinterOfferYear = r.Year
	if !reserved && p.Workers <= 0 && !p.LargeWorker {
		ready := false
		for _, worker := range p.SpecialWorkers {
			if p.specialWorkerReady(worker, r.Year) {
				ready = true
			}
		}
		if !ready {
			return
		}
	}
	idx := *p.RhineSonField
	if idx < 0 || idx >= len(p.Fields) || validateHarvestField(p, idx) != nil || p.Fields[idx].Structure != "" {
		return
	}
	parent, choices := r.Context, r.Choices
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Step: "rhine_first_winter", SourceCardID: "rhine-summer-sonInLaw"}
	if reserved {
		r.Context.Parent, r.Context.ParentChoices = parent, choices
		r.Choices = nil
	}
	r.rhineMenu(p, "rhine-summer-sonInLaw", "son_harvest", nil, nil)
}

func (r *Room) rhineSonOperation(p *Player, op RhineOperation, a Action) error {
	if op.Kind == "son_field" {
		if p.RhineSonField != nil {
			return fmt.Errorf("此牌已放置在田地")
		}
		if a.Field < 0 || a.Field >= len(p.Fields) || p.Fields[a.Field].Sold || p.Fields[a.Field].Structure != "" {
			return fmt.Errorf("须放在自己的葡萄田")
		}
		field := a.Field
		p.RhineSonField = &field
		return nil
	}
	if p.RhineSonField == nil || r.Context == nil || r.Context.Step != "rhine_first_winter" {
		return fmt.Errorf("不是女婿的首次冬季行动时机")
	}
	idx := *p.RhineSonField
	if err := validateHarvestField(p, idx); err != nil {
		return err
	}
	if len(a.Grapes) != 2 {
		return fmt.Errorf("请选择收获的红、白葡萄价值")
	}
	red, white := 0, 0
	for _, vine := range p.Fields[idx].Vines {
		red += vine.Red
		white += vine.White
	}
	rvalue, wvalue := a.Grapes[0], a.Grapes[1]
	if rvalue < 0 || wvalue < 0 || rvalue > red || wvalue > white || rvalue+wvalue != min(3, red+white) {
		return fmt.Errorf("收获总值须为田地产量与3中的较小值")
	}
	addGrape(p, "red", rvalue)
	addGrape(p, "white", wvalue)
	p.Fields[idx].Harvested = true
	r.queueFermentation(p)
	return nil
}
