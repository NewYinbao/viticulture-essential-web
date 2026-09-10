package game

import (
	"fmt"
	"strings"
)

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

func (r *Room) enqueue(id, kind string, options []string, count int) {
	r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: id, Kind: kind, Options: options, Count: count})
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

func (r *Room) resolveChoice(id string, a Action) error {
	c := r.Choices[0]
	if a.Type != "choose" || c.PlayerID != id || c.ID != a.ChoiceID {
		return fmt.Errorf("须由待决玩家提交匹配choiceId的choose")
	}
	if c.Kind == "planner" {
		return r.resolvePlanner(r.Player(id), a)
	}
	if c.Visitor != nil {
		return r.resolveVisitorChoice(id, a)
	}
	p := r.Player(id)
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
