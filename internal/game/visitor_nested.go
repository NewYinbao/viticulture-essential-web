package game

import "fmt"

// startNestedVisitor retains the complete parent continuation, including private
// cards and already queued respondents. The stack is persisted with the room.
func (r *Room) startNestedVisitor(p *Player, id string, after []Choice) error {
	if r.Context == nil {
		return fmt.Errorf("子访客需要父行动")
	}
	parent, queued := r.Context, r.Choices
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Space: visitorCardSeason(p, id) + "_visitor", TriggerSeat: parent.TriggerSeat, TriggerSpace: triggerSpace(parent), LastWorker: parent.LastWorker, Parent: parent, ParentChoices: append(append([]Choice{}, after...), queued...)}
	r.Choices = nil
	if err := r.queueVisitor(p, id); err != nil {
		r.Context = parent
		r.Choices = queued
		return err
	}
	return nil
}

func (r *Room) startNestedAction(p *Player, a Action, bonus bool, after []Choice) error {
	if r.Context == nil {
		return fmt.Errorf("子行动需要父行动")
	}
	parent, queued := r.Context, r.Choices
	r.Context = &ActionContext{ActorID: p.ID, ReturnTurnID: p.ID, Space: a.Space, TriggerSeat: parent.TriggerSeat, TriggerSpace: triggerSpace(parent), LastWorker: parent.LastWorker, Parent: parent, ParentChoices: append(append([]Choice{}, after...), queued...)}
	r.Choices = nil
	if err := r.performPlacement(p, a, bonus); err != nil {
		r.Context = parent
		r.Choices = queued
		return err
	}
	if len(r.Choices) == 0 {
		r.visitorPrompt(p, "", "nested_action_done", []string{"continue"})
	}
	return nil
}

func (r *Room) finishNestedParentEffect(p *Player, s *VisitorStep, parent *ActionContext) error {
	child, choices := r.Context, r.Choices
	r.Context = parent
	r.Choices = child.ParentChoices
	if err := r.visitorContinuationFeasible(p, s); err != nil {
		r.Context = child
		r.Choices = choices
		return err
	}
	if s.Stage == "effect" {
		r.queueStructureVisitorBonus(p)
	}
	child.ParentChoices = r.Choices
	r.Context = child
	r.Choices = choices
	return nil
}
func (r *Room) restoreVisitorParent() {
	child := r.Context
	r.Context = child.Parent
	r.Choices = child.ParentChoices
	if child.PassAfter {
		r.Context.PassAfter = true
	}
	if child.CoinAfter > 0 {
		r.Player(child.ActorID).Coins += child.CoinAfter
	}
	if len(r.Choices) == 0 && r.Context.Remaining > 0 {
		r.Context.Remaining--
		r.visitorPrompt(r.Player(r.Context.ActorID), "", "second", []string{"play", "skip"})
	}
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	r.finishVisitorContext()
}

func triggerSpace(ctx *ActionContext) string {
	if ctx == nil {
		return ""
	}
	if ctx.TriggerSpace != "" {
		return ctx.TriggerSpace
	}
	return ctx.Space
}

func visitorCardSeason(p *Player, id string) string {
	for _, c := range p.Hand {
		if c.ID == id {
			return c.Type
		}
	}
	return "summer"
}
