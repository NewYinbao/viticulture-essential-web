package game

import (
	"fmt"
	"strings"
)

// Choice is persisted; only the head is actionable. Never put secret cards in Options.
type Choice struct {
	OptionFields map[string][]map[string]any `json:"optionFields,omitempty"`
	Cards        []Card                      `json:"cards,omitempty"`
	Title        string                      `json:"title,omitempty"`
	Description  string                      `json:"description,omitempty"`
	ID           string                      `json:"id"`
	PlayerID     string                      `json:"playerId"`
	Kind         string                      `json:"kind"`
	Options      []string                    `json:"options"`
	Count        int                         `json:"count"`
	Visitor      *VisitorStep                `json:"visitor,omitempty"`
	Schema       []ChoiceField               `json:"schema,omitempty"`
	Labels       map[string]string           `json:"labels,omitempty"`
	// SpecialBonus identifies the printed bonus being resolved by Politico.
	// It is public and persisted because a complex bonus may remain pending
	// across a browser reconnect or process restart.
	ActionSpace  string   `json:"actionSpace,omitempty"`
	ActionSlot   int      `json:"actionSlot,omitempty"`
	CardIDs      []string `json:"cardIds,omitempty"`
	SpecialBonus string   `json:"specialBonus,omitempty"`
}

// ActionContext reserves a serializable continuation for the visitor state machine.
type ActionContext struct {
	TriggerSpace  string              `json:"triggerSpace,omitempty"`
	Parent        *ActionContext      `json:"parent,omitempty"`
	ParentChoices []Choice            `json:"parentChoices,omitempty"`
	LastWorker    bool                `json:"lastWorker,omitempty"`
	CoinAfter     int                 `json:"coinAfter,omitempty"`
	TriggerSeat   *Seat               `json:"triggerSeat,omitempty"`
	PassAfter     bool                `json:"passAfter,omitempty"`
	Held          []Card              `json:"held,omitempty"`
	Escrow        map[string][]Card   `json:"escrow,omitempty"`
	ActorID       string              `json:"actorId"`
	SourceCardID  string              `json:"sourceCardId"`
	Space         string              `json:"space"`
	Step          string              `json:"step"`
	Remaining     int                 `json:"remaining"`
	ReturnTurnID  string              `json:"returnTurnId"`
	Data          map[string]string   `json:"data"`
	PendingAction *Action             `json:"pendingAction,omitempty"`
	PendingBonus  bool                `json:"pendingBonus,omitempty"`
	WorkerType    string              `json:"workerType,omitempty"`
	SpecialStage  string              `json:"specialStage,omitempty"`
	PlannedWorker *PlannedWorkerState `json:"plannedWorkerState,omitempty"`
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
	if c.Kind == "messenger" {
		return r.resolveMessenger(r.Player(id), c, a)
	}
	if strings.HasPrefix(c.Kind, "special_") {
		return r.resolveSpecialChoice(r.Player(id), c, a)
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
	case "structure_draft":
		if err := r.resolveStructureDraft(p, c, a); err != nil {
			return err
		}
	case "tuscany_draw", "tuscany_influence", "tuscany_trade", "tuscany_upkeep", "tuscany_next_wake":
		if err := r.tuscanyChoice(p, c, a); err != nil {
			return err
		}
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
	case "structure_visitor_bonus":
		if err := r.resolveStructureVisitorBonus(p, a.Option, a); err != nil {
			return err
		}
	case "structure_influence":
		if a.Option == "place" {
			if len(a.Influence) != 1 {
				return fmt.Errorf("须放置或移动一颗星")
			}
			if err := r.influence(p, a.Influence[0]); err != nil {
				return err
			}
		}
	case "structure_label_factory":
		if a.Option == "use" {
			if p.Coins < 3 {
				return fmt.Errorf("酒标工厂需要3金币")
			}
			p.Coins -= 3
			p.VP += 2
		}
	case "structure_mercado":
		if a.Option == "use" {
			allowed := false
			for _, cardID := range c.CardIDs {
				if cardID == a.CardID {
					allowed = true
				}
			}
			if !allowed {
				return fmt.Errorf("市场只能交付本次抽到的订单")
			}
			if err := r.perform(p, Action{Space: "fill_order", CardID: a.CardID, WineIDs: a.WineIDs}); err != nil {
				return err
			}
		}
	case "structure_barn":
		if a.Option == "take" {
			if err := r.visitorCards(p, a.CardIDs, 2, false); err != nil {
				return err
			}
			p.VP++
		}
	case "structure_fermentation":
		if a.Option == "make" {
			recipe := a.Recipes
			if len(recipe) != 1 {
				return fmt.Errorf("请选择酿造1瓶酒的葡萄配方")
			}
			if err := makeActionWines(p, a, 1); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("CHOICE_NOT_IMPLEMENTED: %s", c.Kind)
	}
	// An optional structure may consume resources promised to a later mandatory
	// visitor step. Reject that choice atomically instead of stranding the queue.
	for _, pending := range r.Choices {
		if pending.Visitor != nil {
			if err := r.visitorContinuationFeasible(r.Player(pending.PlayerID), pending.Visitor); err != nil {
				return err
			}
		}
	}
	r.Choices = r.Choices[1:]
	if c.Kind == "papa" && len(r.Choices) == 0 {
		if r.beginStructureDraft() {
			return nil
		}
	}
	if c.Kind == "structure_draft" && len(r.Choices) == 0 {
		if r.continueStructureDraft() {
			return nil
		}
	}
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
	case "finish_year_finalize":
		r.finishYearFinalize()
	case "tuscany_pass":
		r.tuscanyNext()
	case "tuscany_planner_pass":
		r.startNextPlanner()
	case "ee_barn":
		r.enterEEFall(r.ordered())
	default:
		if r.Context != nil {
			if r.Context.Remaining > 0 {
				r.Context.Remaining--
				r.visitorPrompt(r.Player(r.Context.ActorID), "", "second", []string{"play", "skip"})
				r.TurnID = r.Choices[0].PlayerID
				return nil
			}
			r.finishVisitorContext()
		}
	}
	return nil
}
