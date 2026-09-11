package game

import (
	"encoding/json"
	"fmt"
)

// PlannedPlacement contains only public placement information, never future hand selections.
type PlannedPlacement struct {
	Gray          bool                `json:"gray,omitempty"`
	PlayerID      string              `json:"playerId"`
	Space         string              `json:"space"`
	Slot          int                 `json:"slot"`
	Large         bool                `json:"large"`
	WorkerType    string              `json:"workerType,omitempty"`
	Action        *Action             `json:"action,omitempty"`
	SpecialWorker *PlannedWorkerState `json:"specialWorkerState,omitempty"`
}

// PlannedWorkerState records decisions and conditions fixed when Planner puts a
// special worker on the board. It is persisted privately until the future action.
type PlannedWorkerState struct {
	PlacementHandled bool   `json:"placementHandled,omitempty"`
	FarmerBonus      string `json:"farmerBonus,omitempty"`
	MafiosoEligible  bool   `json:"mafiosoEligible,omitempty"`
	PoliticoEligible bool   `json:"politicoEligible,omitempty"`
	PoliticoBonus    string `json:"politicoBonus,omitempty"`
	MerchantEligible bool   `json:"merchantEligible,omitempty"`
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
		if (!r.tuscany() && r.Phase != "summer") || (r.tuscany() && (!r.actionSeason() || r.Phase == "winter")) {
			return fmt.Errorf("Planner需要可用工人及未来冬季")
		}
		free := false
		workers := []Action{}
		if p.GrayWorkerAvailable {
			workers = append(workers, Action{Gray: true})
		}
		if p.availableRegularWorkers() > 0 {
			workers = append(workers, Action{})
		}
		if p.LargeWorker {
			workers = append(workers, Action{Large: true})
		}
		if specialWorkerEnabled(r) {
			for _, id := range p.SpecialWorkers {
				if p.specialWorkerReady(id, r.Year) && r.specialWorkerAvailable(id) {
					workers = append(workers, Action{WorkerType: id})
				}
			}
		}
		for i := range r.Spaces {
			sp := &r.Spaces[i]
			if !r.plannerFutureSpace(*sp) {
				continue
			}
			for _, worker := range workers {
				if r.workerCanPlace(p, worker) == nil {
					if _, _, e := placement(sp, p, worker); e == nil {
						free = true
					}
				}
				if free {
					break
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
			row := p.Wake
			if r.tuscany() && p.Season == "ready" {
				row = p.NextWake
			}
			if row == i {
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
		actor := r.Player(s.ActorID)
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
		if !r.tuscany() && a.Slot == 5 && a.Color != "summer" && a.Color != "winter" {
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
		if r.tuscany() {
			r.tuscanySeasonReward(p, r.Phase)
			r.Context.PassAfter = true
			return nil
		}
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
		a = r.normalizeWorkerAction(p, a)
		if err := r.workerCanPlace(p, a); err != nil {
			return err
		}
		for i := range r.Spaces {
			sp := &r.Spaces[i]
			if sp.ID != a.Space || !r.plannerFutureSpace(*sp) {
				continue
			}
			slot, bonus, e := placement(sp, p, a)
			if e != nil {
				return e
			}
			a.Slot = slot
			r.consumeWorker(p, a)
			seat := Seat{Season: r.Phase, PlayerID: p.ID, Slot: slot, Large: a.Large, WorkerType: a.WorkerType, Gray: a.Gray}
			sp.Occupied = append(sp.Occupied, seat)
			state := r.plannerWorkerState(p, a, bonus)
			r.Planned = append(r.Planned, PlannedPlacement{PlayerID: p.ID, Space: sp.ID, Slot: slot, Large: a.Large, WorkerType: a.WorkerType, Gray: a.Gray, SpecialWorker: state})
			if r.queuePlannerPlacementWorker(p, a, bonus, sp, seat) {
				return nil
			}
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
	if r.rhineFinishSkippedPlannerHarvest(ctx) {
		return
	}
	if ctx.Parent != nil {
		r.restoreVisitorParent()
		return
	}
	if ctx.Step == "rhine_first_winter" {
		r.Context = nil
		r.TurnID = ctx.ReturnTurnID
		return
	}
	if ctx.CoinAfter > 0 {
		r.Player(ctx.ActorID).Coins += ctx.CoinAfter
	}
	if ctx.PassAfter {
		if !r.tuscany() {
			r.Player(ctx.ActorID).Passed = true
		}
	}
	if ctx.PendingAction != nil && r.queuePostSpecial(r.Player(ctx.ActorID), *ctx.PendingAction, ctx.PendingBonus) {
		r.TurnID = r.Choices[0].PlayerID
		return
	}
	r.Context = nil
	if r.tuscany() && ctx.PassAfter {
		r.TurnID = ctx.ActorID
		r.tuscanyPassWithResume(r.Player(ctx.ActorID), ctx.Step == "planner")
		return
	}
	if ctx.Step == "planner" {
		r.startNextPlanner()
		return
	}
	r.TurnID = ctx.ReturnTurnID
	r.next()
}
func (r *Room) startNextPlanner() {
	for len(r.Planned) > 0 {
		if r.tuscany() {
			idx := -1
			for i, p := range r.Planned {
				for _, sp := range r.Spaces {
					if sp.ID == p.Space && sp.Season == r.Phase {
						idx = i
						break
					}
				}
				if idx >= 0 {
					break
				}
			}
			if idx < 0 {
				break
			}
			plan := r.Planned[idx]
			copy(r.Planned[1:idx+1], r.Planned[:idx])
			r.Planned[0] = plan
		}
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
		if !present || (!r.plannerCanExecute(plan) && !r.rhinePlannerHarvestCanEnable(plan)) {
			r.Planned = r.Planned[1:]
			continue
		}
		r.Context = r.plannerContext(plan)
		r.enqueue(plan.PlayerID, "planner", []string{"execute"}, 1)
		r.Choices[len(r.Choices)-1].SpecialBonus = plannerBonus(plan)
		r.TurnID = plan.PlayerID
		return
	}
	r.TurnID = r.ordered()[0].ID
	if r.activateMessenger(r.Player(r.TurnID)) {
		return
	}
	if r.Player(r.TurnID).Passed {
		r.next()
	}
}

func (r *Room) plannerContext(plan PlannedPlacement) *ActionContext {
	return &ActionContext{
		ActorID:       plan.PlayerID,
		ReturnTurnID:  r.ordered()[0].ID,
		Space:         plan.Space,
		TriggerSpace:  plan.Space,
		Step:          "planner",
		TriggerSeat:   &Seat{Season: r.Phase, PlayerID: plan.PlayerID, Slot: plan.Slot, Large: plan.Large, WorkerType: plan.WorkerType, Gray: plan.Gray},
		WorkerType:    plan.WorkerType,
		PlannedWorker: plan.SpecialWorker,
		Data:          map[string]string{},
	}
}

func plannerAction(plan PlannedPlacement, a Action) Action {
	a.Space = plan.Space
	a.Slot = plan.Slot
	a.WorkerType = plan.WorkerType
	a.BonusOverride = plannerBonus(plan)
	return a
}

func plannerBonus(plan PlannedPlacement) string {
	if plan.SpecialWorker != nil && plan.WorkerType == "farmer" {
		return plan.SpecialWorker.FarmerBonus
	}
	return ""
}

func (r *Room) resolvePlanner(p *Player, a Action) error {
	if (!r.tuscany() && r.Phase != "winter") || len(r.Planned) == 0 || r.Context == nil || r.Context.Step != "planner" {
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
			if r.tuscany() {
				if sp.Season != r.Phase {
					return fmt.Errorf("预约尚未到执行季节")
				}
				bonus = sp.BonusSlots[plan.Slot] != ""
			}
		}
	}
	a = plannerAction(plan, a)
	r.Choices = r.Choices[1:]
	r.Planned = r.Planned[1:]
	r.Context.PendingAction = cloneAction(a)
	r.Context.PendingBonus = bonus
	if (plan.SpecialWorker == nil || !plan.SpecialWorker.PlacementHandled) && r.queueSpecialPreChoice(p, a, bonus, r.space(plan.Space)) {
		r.TurnID = p.ID
		return nil
	}
	if e := r.performPlacement(p, a, bonus); e != nil {
		return e
	}
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
	} else {
		r.finishVisitorContext()
	}
	return nil
}

func (r *Room) plannerWorkerState(p *Player, a Action, bonus bool) *PlannedWorkerState {
	if !specialWorkerEnabled(r) || a.WorkerType == "" {
		return nil
	}
	state := &PlannedWorkerState{}
	switch a.WorkerType {
	case "farmer", "professore", "innkeeper":
		state.PlacementHandled = true
	case "mafioso":
		state.MafiosoEligible = !bonus
	case "politico":
		state.PoliticoEligible = bonus
		if bonus {
			state.PoliticoBonus = normalizeBonusKey(r.placementBonus(a.Space, a.Slot))
		}
	case "merchant":
		state.MerchantEligible = r.merchantEligible(p)
	}
	return state
}

func (r *Room) queuePlannerPlacementWorker(p *Player, a Action, bonus bool, sp *Space, seat Seat) bool {
	if a.WorkerType != "farmer" && a.WorkerType != "professore" && a.WorkerType != "innkeeper" {
		return false
	}
	parent := r.Context
	r.Context = &ActionContext{
		ActorID:      p.ID,
		ReturnTurnID: p.ID,
		Space:        sp.ID,
		TriggerSpace: sp.ID,
		TriggerSeat:  &seat,
		Step:         "planner_placement",
		WorkerType:   a.WorkerType,
		Parent:       parent,
		Data:         map[string]string{},
	}
	queued := false
	if a.WorkerType == "innkeeper" {
		queued = r.queueInnkeeperChoice(p, a)
	} else {
		queued = r.queueSpecialPreChoice(p, a, bonus, sp)
	}
	if !queued {
		r.Context = parent
	}
	return queued
}

func (r *Room) activePlannerPlacement() *PlannedPlacement {
	if r.Context == nil || r.Context.Step != "planner_placement" || r.Context.TriggerSeat == nil {
		return nil
	}
	seat := r.Context.TriggerSeat
	for i := len(r.Planned) - 1; i >= 0; i-- {
		plan := &r.Planned[i]
		if plan.PlayerID == r.Context.ActorID && plan.Space == r.Context.Space && plan.Slot == seat.Slot && plan.Large == seat.Large && plan.WorkerType == seat.WorkerType && plan.Gray == seat.Gray {
			return plan
		}
	}
	return nil
}

// Test minimal legal actions on isolated copies; an impossible reservation is
// exhausted without a reward, never refunded as a usable worker.
func (r *Room) plannerCanExecute(plan PlannedPlacement) bool {
	p := r.Player(plan.PlayerID)
	if p == nil {
		return false
	}
	bonus := false
	for _, sp := range r.Spaces {
		if sp.ID == plan.Space {
			bonus = plan.Slot == 1 && sp.Capacity >= 2
			if r.tuscany() {
				bonus = sp.BonusSlots[plan.Slot] != ""
			}
		}
	}
	probe := func(a Action) bool {
		data, _ := json.Marshal(r)
		var q Room
		_ = json.Unmarshal(data, &q)
		q.Choices = nil
		q.Context = q.plannerContext(plan)
		a = plannerAction(plan, a)
		return q.performPlacement(q.Player(p.ID), a, bonus) == nil
	}
	switch plan.Space {
	case "draw_order", "draw_vine", "draw_structure", "tour":
		return probe(Action{})
	case "train":
		for _, id := range append([]string{"regular", "grande"}, r.SpecialWorkerPool...) {
			if probe(Action{SpecialWorker: id}) {
				return true
			}
		}
	case "build":
		for id := range costs {
			if probe(Action{Building: id}) {
				return true
			}
		}
		if r.Config.Structures {
			for _, card := range p.Hand {
				if card.Type != "structure" || card.StructureID == "" {
					continue
				}
				for slot := 0; slot < 2; slot++ {
					if probe(Action{Building: card.StructureID, Mode: "mat", StructureSlot: slot}) {
						return true
					}
				}
				for field := range p.Fields {
					if probe(Action{Building: card.StructureID, Mode: "field", Field: field}) {
						return true
					}
				}
			}
		}
	case "build_tour":
		return probe(Action{Mode: "tour"})
	case "plant":
		for _, c := range p.Hand {
			if c.Type != "vine" {
				continue
			}
			for field := range p.Fields {
				if probe(Action{CardID: c.ID, Field: field}) {
					return true
				}
			}
		}
	case "influence":
		for _, to := range tuscanyRegions {
			if probe(Action{Influence: []InfluenceMove{{To: to.ID}}}) {
				return true
			}
			for from, count := range p.Influence {
				if count > 0 && probe(Action{Influence: []InfluenceMove{{From: from, To: to.ID}}}) {
					return true
				}
			}
		}
	case "trade":
		trades := []Trade{
			{Give: "coins", Receive: "vp"},
			{Give: "coins", Receive: "cards", Colors: []string{"vine", "order"}},
			{Give: "coins", Receive: "grape", Color: "red"},
			{Give: "vp", Receive: "coins"},
			{Give: "vp", Receive: "cards", Colors: []string{"vine", "order"}},
			{Give: "vp", Receive: "grape", Color: "red"},
		}
		for i := 0; i < len(p.Hand); i++ {
			for j := i + 1; j < len(p.Hand); j++ {
				trades = append(trades,
					Trade{Give: "cards", CardIDs: []string{p.Hand[i].ID, p.Hand[j].ID}, Receive: "coins"},
					Trade{Give: "cards", CardIDs: []string{p.Hand[i].ID, p.Hand[j].ID}, Receive: "vp"},
					Trade{Give: "cards", CardIDs: []string{p.Hand[i].ID, p.Hand[j].ID}, Receive: "grape", Color: "red"},
				)
			}
		}
		for _, grape := range p.Grapes {
			trades = append(trades,
				Trade{Give: "grape", GrapeID: grape.ID, Receive: "coins"},
				Trade{Give: "grape", GrapeID: grape.ID, Receive: "vp"},
				Trade{Give: "grape", GrapeID: grape.ID, Receive: "cards", Colors: []string{"vine", "order"}},
			)
		}
		for _, trade := range trades {
			if probe(Action{Trades: []Trade{trade}}) {
				return true
			}
		}
	case "flip_field":
		for field := range p.Fields {
			for _, mode := range []string{"sell_field", "buy_field"} {
				if probe(Action{Mode: mode, Field: field}) {
					return true
				}
			}
		}
	case "sell_wine":
		for _, w := range p.Wines {
			if probe(Action{WineIDs: []string{w.ID}}) {
				return true
			}
		}
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
	case "summer_visitor", "winter_visitor":
		for _, c := range p.Hand {
			if c.Type != "winter" {
				continue
			}
			if probe(Action{CardID: c.ID}) {
				return true
			}
			// Tuscany permits the printed lira before the action. A reservation
			// that needs that lira is executable, not an exhausted worker.
			if r.tuscany() && bonus && r.placementBonus(plan.Space, plan.Slot) == "coin" && probe(Action{CardID: c.ID, BonusFirst: true}) {
				return true
			}
		}
	}
	return false
}

func (r *Room) plannerFutureSpace(sp Space) bool {
	if r.tuscany() {
		return seasonIndex(sp.Season) > seasonIndex(r.Phase)
	}
	return sp.Season == "winter"
}
func (r *Room) cancelPlanned(space string, seat Seat) {
	out := r.Planned[:0]
	for _, p := range r.Planned {
		if p.Space == space && p.PlayerID == seat.PlayerID && p.Slot == seat.Slot && p.Large == seat.Large && p.WorkerType == seat.WorkerType && p.Gray == seat.Gray {
			continue
		}
		out = append(out, p)
	}
	r.Planned = out
}
