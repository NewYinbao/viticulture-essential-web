package game

import (
	"fmt"
	"strings"
)

func (r *Room) workerCanPlace(p *Player, a Action) error {
	if a.Gray {
		if a.Large || a.WorkerType != "" || !p.GrayWorkerAvailable || r.GrayWorkerOwner != p.ID {
			return fmt.Errorf("灰色临时工不可用")
		}
		return nil
	}
	if a.WorkerType == "" {
		if a.Large {
			if !p.LargeWorker {
				return fmt.Errorf("大工人已使用")
			}
		} else if p.availableRegularWorkers() <= 0 {
			return fmt.Errorf("没有可用普通工人")
		}
		return nil
	}
	if !specialWorkerEnabled(r) || !r.specialWorkerAvailable(a.WorkerType) {
		return fmt.Errorf("特殊工人不可用")
	}
	if a.Large {
		return fmt.Errorf("特殊工人不能作为大工人")
	}
	p.ensureSpecialMaps()
	if !p.specialWorkerReady(a.WorkerType, r.Year) {
		return fmt.Errorf("特殊工人未训练、今年已使用或尚未可用")
	}
	return nil
}

func (r *Room) consumeWorker(p *Player, a Action) {
	if a.Gray {
		p.GrayWorkerAvailable = false
		p.Workers--
		return
	}
	if a.WorkerType == "" {
		if a.Large {
			p.LargeWorker = false
		} else {
			p.Workers--
		}
		return
	}
	p.ensureSpecialMaps()
	p.SpecialWorkerUsed[a.WorkerType] = true
}

func (r *Room) releaseSeat(seat Seat) {
	p := r.Player(seat.PlayerID)
	if p == nil {
		return
	}
	if seat.Gray {
		if r.GrayWorkerOwner == p.ID && !p.GrayWorkerAvailable {
			p.GrayWorkerAvailable = true
			p.Workers++
		}
	} else if seat.WorkerType != "" {
		p.ensureSpecialMaps()
		p.SpecialWorkerUsed[seat.WorkerType] = false
	} else if seat.Large {
		p.LargeWorker = true
	} else {
		p.Workers++
	}
}

func (r *Room) specialWorkerAt(space string, id string) []Seat {
	for _, s := range r.Spaces {
		if s.ID == space {
			out := []Seat{}
			for _, seat := range s.Occupied {
				if seat.PlayerID != id {
					out = append(out, seat)
				}
			}
			return out
		}
	}
	return nil
}

func (r *Room) soldatoToll(space, playerID string) int {
	n := 0
	for _, s := range r.Spaces {
		if s.ID != space {
			continue
		}
		for _, seat := range s.Occupied {
			if seat.PlayerID != playerID && seat.WorkerType == "soldato" {
				n++
			}
		}
	}
	return n
}

func (r *Room) paySoldatoToll(space, playerID string) error {
	n := r.soldatoToll(space, playerID)
	p := r.Player(playerID)
	if n == 0 {
		return nil
	}
	if p == nil || p.Coins < n {
		return fmt.Errorf("向同一行动的Soldato支付%d金币", n)
	}
	p.Coins -= n
	for i := range r.Spaces {
		if r.Spaces[i].ID != space {
			continue
		}
		for _, seat := range r.Spaces[i].Occupied {
			if seat.WorkerType == "soldato" && seat.PlayerID != playerID {
				if q := r.Player(seat.PlayerID); q != nil {
					q.Coins++
				}
			}
		}
	}
	return nil
}

func (r *Room) seasonAllowed(s *Space, a Action) bool {
	if s == nil || s.Season == "any" || s.Season == r.Phase {
		return true
	}
	if a.WorkerType == "" {
		return false
	}
	idx, cur := seasonIndex(s.Season), seasonIndex(r.Phase)
	if a.WorkerType == "traveler" {
		if r.tuscany() {
			return idx >= 0 && cur >= 0 && idx < cur
		}
		return r.Phase == "winter" && s.Season == "summer"
	}
	if a.WorkerType == "messenger" {
		if r.tuscany() {
			return idx >= 0 && cur >= 0 && idx > cur
		}
		return r.Phase == "summer" && s.Season == "winter"
	}
	return false
}

func (r *Room) bumpForChef(s *Space, p *Player, a Action) error {
	if a.WorkerType != "chef" || s.Season == "any" {
		return nil
	}
	for i, seat := range s.Occupied {
		if seat.PlayerID == p.ID || seat.Slot != a.Slot && a.Slot != 0 {
			continue
		}
		if seat.WorkerType == "chef" {
			if a.Slot == 0 {
				continue
			}
			return fmt.Errorf("Chef不能顶替Chef")
		}
		for j := len(r.MessengerPlans) - 1; j >= 0; j-- {
			plan := r.MessengerPlans[j]
			if plan.PlayerID == seat.PlayerID && plan.Space == s.ID && plan.Slot == seat.Slot {
				r.MessengerPlans = append(r.MessengerPlans[:j], r.MessengerPlans[j+1:]...)
			}
		}
		r.releaseSeat(seat)
		s.Occupied = append(s.Occupied[:i], s.Occupied[i+1:]...)
		return nil
	}
	return nil
}

func (r *Room) specialBonusOptions(s *Space) []string {
	seen := map[string]bool{}
	out := []string{}
	if s.BonusSlots != nil {
		// Farmer explicitly ignores player-count limits: every printed bonus
		// on the action is available, including a slot hidden in a 2-player
		// setup.  Normal placement still uses s.Capacity.
		for slot := 1; slot <= 3; slot++ {
			if b := normalizeBonusKey(s.BonusSlots[slot]); b != "" && !seen[b] {
				out = append(out, b)
				seen[b] = true
			}
		}
	} else if b := normalizeBonusKey(spaceBonuses[s.ID]); b != "" {
		out = append(out, b)
	}
	return out
}

func (r *Room) queueSpecialPreChoice(p *Player, a Action, bonus bool, s *Space) bool {
	if !specialWorkerEnabled(r) || a.WorkerType == "" || r.Context == nil || !isBoardSpace(s) {
		return false
	}
	if a.WorkerType == "farmer" {
		opts := r.specialBonusOptions(s)
		if len(opts) > 0 {
			r.Context.PendingAction = cloneAction(a)
			r.Context.PendingBonus = bonus
			r.Context.SpecialStage = "farmer"
			r.enqueue(p.ID, "special_farmer", append(opts, "skip"), 1)
			return true
		}
	}
	if a.WorkerType == "professore" {
		opts := []string{"skip"}
		for _, sp := range r.Spaces {
			if sp.Season != r.Phase || !isBoardSpace(&sp) {
				continue
			}
			for _, seat := range sp.Occupied {
				if seat.PlayerID == p.ID && !seat.Large && seat.WorkerType == "" && !(r.Context.TriggerSeat != nil && r.Context.Space == sp.ID && r.Context.TriggerSeat.Slot == seat.Slot) {
					opts = append(opts, sp.ID+":"+fmt.Sprint(seat.Slot))
				}
			}
		}
		if len(opts) > 1 {
			r.Context.PendingAction = cloneAction(a)
			r.Context.PendingBonus = bonus
			r.Context.SpecialStage = "professore"
			r.enqueue(p.ID, "special_professore", opts, 1)
			return true
		}
	}
	return false
}

func cloneAction(a Action) *Action { b := a; return &b }

func isBoardSpace(s *Space) bool { return s != nil && s.Season != "any" }

func (r *Room) trainWorker(p *Player, id string, discount int) error {
	if id == "grande" {
		return r.trainGrande(p, discount)
	}
	if p.TotalWorkers-p.ownedGrande() >= 5 {
		return fmt.Errorf("普通和特殊工人合计上限为5名；可重新培训失去的大工人")
	}
	if p.TotalWorkers >= 6 {
		return fmt.Errorf("工人上限为6名")
	}
	toll := 0
	if id == "" || id == "regular" {
		for _, q := range r.Players {
			if q.ID != p.ID && has(q, "academy") {
				toll++
			}
		}
		cost := 4 - discount
		cost += toll
		if p.Coins < cost {
			return fmt.Errorf("培训需要%d金币", cost)
		}
		p.Coins -= cost
		for _, q := range r.Players {
			if q.ID != p.ID && has(q, "academy") {
				q.Coins++
			}
		}
		p.TotalWorkers++
		p.Trained++
		return nil
	}
	if !specialWorkerEnabled(r) || !r.specialWorkerAvailable(id) {
		return fmt.Errorf("特殊工人不在本局牌池")
	}
	if p.hasSpecialWorker(id) {
		return fmt.Errorf("每类特殊工人每位玩家只能训练一次")
	}
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			toll++
		}
	}
	cost := 5 - discount + toll
	if p.Coins < cost {
		return fmt.Errorf("特殊工人培训需要%d金币", cost)
	}
	p.ensureSpecialMaps()
	p.Coins -= cost
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			q.Coins++
		}
	}
	p.TotalWorkers++
	p.Trained++
	p.SpecialWorkers = append(p.SpecialWorkers, id)
	p.SpecialWorkerReady[id] = r.Year + 1
	p.SpecialWorkerUsed[id] = false
	return nil
}

// trainWorkerNow is the School variant of training. The placed worker has
// already been spent to activate the structure, so the newly trained worker
// is immediately available for another placement this year. Academy tolls
// and the regular/special worker costs are still settled by trainWorker.
func (r *Room) trainWorkerNow(p *Player, id string) error {
	if err := r.trainWorker(p, id, 4); err != nil {
		return err
	}
	if id == "grande" {
		p.LargeWorker = true
	} else if id == "" || id == "regular" {
		p.Workers++
	}
	if id != "" && id != "regular" && id != "grande" {
		p.ensureSpecialMaps()
		p.SpecialWorkerReady[id] = r.Year
	}
	return nil
}

func (r *Room) merchantEligible(p *Player) bool {
	if p == nil || !specialWorkerEnabled(r) {
		return false
	}
	for _, q := range r.Players {
		if q.ID == p.ID {
			continue
		}
		if r.tuscany() {
			// In winter, a player who passed is represented by year_end rather
			// than a fourth season.  That state is still after the current
			// season and therefore satisfies Merchant's timing condition.
			if q.Season != "year_end" && q.Season != "ready" && seasonIndex(q.Season) <= seasonIndex(r.Phase) {
				return false
			}
		} else if !q.Passed {
			return false
		}
	}
	return true
}

func (r *Room) queuePostSpecial(p *Player, a Action, bonus bool) bool {
	if r.Context == nil || r.Context.WorkerType == "" || r.Context.SpecialStage != "" {
		return false
	}
	if !specialWorkerEnabled(r) {
		return false
	}
	plannerState := r.Context.PlannedWorker
	mafiosoEligible := !bonus
	politicoEligible := bonus
	merchantEligible := r.merchantEligible(p)
	innkeeperHandled := false
	if plannerState != nil && r.Context.Step == "planner" {
		mafiosoEligible = plannerState.MafiosoEligible
		politicoEligible = plannerState.PoliticoEligible
		merchantEligible = plannerState.MerchantEligible
		innkeeperHandled = plannerState.PlacementHandled
	}
	if a.WorkerType == "mafioso" && isBoardSpace(r.space(a.Space)) && mafiosoEligible {
		r.Context.SpecialStage = "mafioso"
		r.enqueue(p.ID, "special_mafioso", []string{"repeat", "skip"}, 1)
		r.Choices[len(r.Choices)-1].ActionSpace = a.Space
		return true
	}
	if a.WorkerType == "politico" && politicoEligible {
		r.Context.SpecialStage = "politico"
		bonusKey := normalizeBonusKey(r.placementBonus(a.Space, r.Context.TriggerSeat.Slot))
		if plannerState != nil && r.Context.Step == "planner" {
			bonusKey = plannerState.PoliticoBonus
		}
		r.Context.Data["bonusKey"] = bonusKey
		c := Choice{ID: NewID(), PlayerID: p.ID, Kind: "special_politico", Options: []string{"repeat_bonus", "skip"}, Count: 1, SpecialBonus: r.Context.Data["bonusKey"]}
		r.Choices = append(r.Choices, c)
		return true
	}
	if a.WorkerType == "oracle" && (a.Space == "draw_vine" || a.Space == "draw_order") {
		before := 0
		if r.Context.Data != nil {
			fmt.Sscan(r.Context.Data["oracleBefore"], &before)
		}
		r.draw(p, a.Space[len("draw_"):])
		ids := []string{}
		if before <= len(p.Hand) {
			for _, c := range p.Hand[before:] {
				ids = append(ids, c.ID)
			}
		}
		// A Tuscany slot may already draw a second card.  Oracle still adds
		// one card and the player discards exactly one of all cards drawn by
		// this action, so the choice may contain two or three card IDs.
		if len(ids) >= 2 {
			r.Context.SpecialStage = "oracle"
			r.enqueue(p.ID, "special_oracle", ids, 1)
			return true
		}
	}
	if a.WorkerType == "merchant" && isBoardSpace(r.space(a.Space)) && merchantEligible {
		r.Context.SpecialStage = "merchant"
		r.enqueue(p.ID, "special_merchant", []string{"vine", "order", "summer", "winter"}, 1)
		return true
	}
	if a.WorkerType == "innkeeper" && !innkeeperHandled && r.queueInnkeeperChoice(p, a) {
		return true
	}
	return false
}

func (r *Room) queueInnkeeperChoice(p *Player, a Action) bool {
	if a.WorkerType == "innkeeper" && len(r.specialWorkerAt(a.Space, p.ID)) > 0 && len(r.Players) > 2 && p.Coins > 0 {
		r.Context.SpecialStage = "innkeeper"
		r.enqueue(p.ID, "special_innkeeper", []string{"skip"}, 1)
		c := &r.Choices[len(r.Choices)-1]
		for _, q := range r.Players {
			if q.ID == p.ID || len(r.workerSeats(a.Space, q.ID)) == 0 {
				continue
			}
			for _, typ := range []string{"summer", "winter"} {
				if hasVisitorType(q, typ) {
					option := q.ID + "|" + typ
					c.Options = append(c.Options, option)
					if c.Labels == nil {
						c.Labels = map[string]string{}
					}
					c.Labels[option] = q.Name + "：随机拿1张" + map[string]string{"summer": "夏季", "winter": "冬季"}[typ] + "访客"
				}
			}
		}
		if len(c.Options) == 1 {
			r.Choices = r.Choices[:len(r.Choices)-1]
			r.Context.SpecialStage = ""
			return false
		}
		return true
	}
	return false
}

func normalizeBonusKey(key string) string {
	switch key {
	case "+1 coin":
		return "coin"
	case "+1 vine":
		return "draw_vine"
	case "+1 order":
		return "draw_order"
	case "discount 1":
		return "discount"
	case "plant up to 2":
		return "plant"
	case "harvest up to 2":
		return "harvest"
	case "make up to 3":
		return "make_wine"
	case "+1 VP":
		return "vp"
	case "play up to 2":
		return "visitor"
	default:
		return key
	}
}

func (r *Room) space(id string) *Space {
	for i := range r.Spaces {
		if r.Spaces[i].ID == id {
			return &r.Spaces[i]
		}
	}
	return nil
}

func (r *Room) applyBonusOnly(p *Player, key string) error {
	switch key {
	case "coin":
		p.Coins++
	case "vp":
		p.VP++
	case "draw_vine":
		r.draw(p, "vine")
	case "draw_order":
		r.draw(p, "order")
	case "discount":
		// The printed discount was already applied to the completed action.
		// A second Politico discount is represented by its equivalent refund.
		p.Coins++
	default:
		return fmt.Errorf("该奖励需要选择具体资源")
	}
	return nil
}

var politicoComplexBonuses = map[string]bool{
	"plant": true, "harvest": true, "make_wine": true, "visitor": true,
	"influence": true, "trade": true, "build_tour": true,
}

func politicoChoice(key string) Choice {
	c := Choice{ID: NewID(), Kind: "special_politico", Options: []string{"confirm", "skip"}, Count: 1, SpecialBonus: key}
	c.Labels = map[string]string{"confirm": "确认再次获得该奖励", "skip": "不重复奖励"}
	return c
}

func (r *Room) resolvePoliticoBonus(p *Player, key string, a Action) error {
	if p == nil {
		return fmt.Errorf("Politico玩家不存在")
	}
	if p.Coins < 1 {
		return fmt.Errorf("Politico支付不足")
	}
	// Payment and the selected effect are in the surrounding Room.Apply
	// transaction.  Any invalid resource selection therefore restores both.
	p.Coins--
	switch key {
	case "plant":
		a.Space = "plant"
		return r.performEEPlacement(p, a, true)
	case "harvest":
		a.Space = "harvest"
		return r.performEEPlacement(p, a, true)
	case "make_wine":
		a.Space = "make_wine"
		return r.performEEPlacement(p, a, true)
	case "visitor":
		if a.CardID == "" {
			return fmt.Errorf("请选择要重复打出的访客")
		}
		if a.Color != "summer" && a.Color != "winter" {
			return fmt.Errorf("请选择夏季或冬季访客")
		}
		// This is a new bonus action, not a repeat of the outer Politico
		// placement. Do not leave the outer action in the visitor continuation.
		r.Context.PendingAction = nil
		return r.startVisitor(p, Action{Space: a.Color + "_visitor", CardID: a.CardID}, true)
	case "influence":
		if len(a.Influence) != 1 {
			return fmt.Errorf("须放置或移动一颗星")
		}
		return r.influence(p, a.Influence[0])
	case "trade":
		if len(a.Trades) != 1 {
			return fmt.Errorf("须完成一次交易")
		}
		return r.trade(p, a.Trades[0])
	case "build_tour":
		if a.Mode != "build" && a.Mode != "tour" {
			return fmt.Errorf("须选择建造或导览")
		}
		a.Space = a.Mode
		return r.performEEPlacement(p, a, true)
	default:
		return r.applyBonusOnly(p, key)
	}
}

func (r *Room) resolveSpecialChoice(p *Player, c Choice, a Action) error {
	if r.Context == nil {
		return fmt.Errorf("特殊工人上下文已失效")
	}
	for _, o := range c.Options {
		if o == a.Option {
			goto valid
		}
	}
	return fmt.Errorf("无效特殊工人选项")
valid:
	switch c.Kind {
	case "special_train":
		if a.Option == "regular" {
			r.Context.PendingAction.SpecialWorker = ""
		} else {
			r.Context.PendingAction.SpecialWorker = a.Option
		}
		return r.resumeSpecialPending(p)
	case "special_farmer":
		if a.Option == "skip" {
			r.Context.PendingAction.BonusOverride = ""
		} else {
			r.Context.PendingAction.BonusOverride = a.Option
		}
		if r.Context.Step == "planner_placement" {
			plan := r.activePlannerPlacement()
			if plan == nil || plan.SpecialWorker == nil {
				return fmt.Errorf("Planner特殊工人预约上下文已失效")
			}
			plan.SpecialWorker.FarmerBonus = r.Context.PendingAction.BonusOverride
			return r.finishSpecialContinuation()
		}
		return r.resumeSpecialPending(p)
	case "special_professore":
		if a.Option != "skip" {
			parts := strings.SplitN(a.Option, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("无效Professore目标")
			}
			sp := r.space(parts[0])
			if sp == nil {
				return fmt.Errorf("目标行动不存在")
			}
			for i, seat := range sp.Occupied {
				if seat.PlayerID == p.ID && !seat.Large && seat.WorkerType == "" && fmt.Sprint(seat.Slot) == parts[1] {
					r.releaseSeat(seat)
					sp.Occupied = append(sp.Occupied[:i], sp.Occupied[i+1:]...)
					break
				}
			}
		}
		if r.Context.Step == "planner_placement" {
			return r.finishSpecialContinuation()
		}
		return r.resumeSpecialPending(p)
	case "special_mafioso":
		if a.Option == "repeat" {
			if r.Context.PendingAction == nil {
				return fmt.Errorf("重复行动上下文已失效")
			}
			// The action repeats with fresh selections; its earlier cards and
			// resources may have already been consumed.
			pending := a
			pending.Space = r.Context.PendingAction.Space
			pending.Slot = r.Context.PendingAction.Slot
			pending.WorkerType, pending.BonusOverride = "", ""
			pending.Large, pending.DeclineBonus, pending.BonusFirst = false, false, false
			return r.finishSpecialRepeat(p, pending, false)
		}
		return r.finishSpecialContinuation()
	case "special_politico":
		key := c.SpecialBonus
		if key == "" {
			key = r.Context.Data["bonusKey"]
		}
		if a.Option == "repeat_bonus" {
			if politicoComplexBonuses[key] {
				next := politicoChoice(key)
				next.PlayerID = p.ID
				r.Choices[0] = next
				return nil
			}
			if err := r.resolvePoliticoBonus(p, key, a); err != nil {
				return err
			}
		} else if a.Option == "confirm" {
			if !politicoComplexBonuses[key] {
				return fmt.Errorf("该Politico奖励不需要资源选择")
			}
			if err := r.resolvePoliticoBonus(p, key, a); err != nil {
				return err
			}
		}
		return r.finishSpecialContinuation()
	case "special_oracle":
		if err := r.discardOneOf(p, a.Option, c.Options); err != nil {
			return err
		}
		return r.finishSpecialContinuation()
	case "special_merchant":
		r.draw(p, a.Option)
		return r.finishSpecialContinuation()
	case "special_innkeeper":
		if a.Option != "skip" {
			if p.Coins < 1 {
				return fmt.Errorf("Innkeeper支付不足")
			}
			parts := strings.SplitN(a.Option, "|", 2)
			if len(parts) != 2 || (parts[1] != "summer" && parts[1] != "winter") {
				return fmt.Errorf("无效Innkeeper目标")
			}
			target := r.Player(parts[0])
			if target == nil || target.ID == p.ID || len(r.workerSeats(r.Context.Space, target.ID)) == 0 || !hasVisitorType(target, parts[1]) {
				return fmt.Errorf("没有拥有该季访客的同行动对手")
			}
			p.Coins--
			target.Coins++
			indices := []int{}
			for i, card := range target.Hand {
				if card.Type == parts[1] {
					indices = append(indices, i)
				}
			}
			if len(indices) > 0 {
				i := indices[rnd(len(indices))]
				p.Hand = append(p.Hand, target.Hand[i])
				removeCard(target, i)
			}
		}
		return r.finishSpecialContinuation()
	}
	return fmt.Errorf("未知特殊工人选择")
}

func (r *Room) discardOneOf(p *Player, id string, options []string) error {
	allowed := false
	for _, option := range options {
		if option == id {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("只能弃置本次Oracle抽到的牌")
	}
	idx := -1
	for i, c := range p.Hand {
		if c.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("只能弃置本次Oracle抽到的牌")
	}
	r.discard(p.Hand[idx])
	p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)
	return nil
}

func hasVisitorType(p *Player, typ string) bool {
	for _, c := range p.Hand {
		if c.Type == typ {
			return true
		}
	}
	return false
}
func (r *Room) workerSeats(space, player string) []Seat {
	for _, s := range r.Spaces {
		if s.ID == space {
			out := []Seat{}
			for _, seat := range s.Occupied {
				if seat.PlayerID == player {
					out = append(out, seat)
				}
			}
			return out
		}
	}
	return nil
}

func (r *Room) resumeSpecialPending(p *Player) error {
	a := *r.Context.PendingAction
	bonus := r.Context.PendingBonus
	trainingSelection := len(r.Choices) > 0 && r.Choices[0].Kind == "special_train"
	r.Context.PendingAction = nil
	r.Context.PendingBonus = false
	r.Context.SpecialStage = ""
	if a.Space == "train" {
		if trainingSelection && r.queueSpecialPreChoice(p, a, bonus, r.space("train")) {
			return r.finishSpecialContinuation()
		}
		discount := 0
		if bonus || (a.BonusOverride == "discount" && !a.DeclineBonus) {
			discount = 1
		}
		if err := r.trainWorker(p, a.SpecialWorker, discount); err != nil {
			return err
		}
	} else if a.Space == "school_training" {
		if err := r.trainWorkerNow(p, a.SpecialWorker); err != nil {
			return err
		}
		d, ok := structureDef(a.Building)
		if !ok {
			return fmt.Errorf("学院训练结构上下文无效")
		}
		if err := r.structureActionReward(p, d); err != nil {
			return err
		}
	} else if err := r.performPlacement(p, a, bonus || a.BonusOverride != ""); err != nil {
		return err
	}
	return r.finishSpecialContinuation()
}

func (r *Room) finishSpecialRepeat(p *Player, a Action, bonus bool) error {
	r.Context.PendingAction = nil
	r.Context.SpecialStage = ""
	if err := r.performPlacement(p, a, bonus); err != nil {
		return err
	}
	return r.finishSpecialContinuation()
}

func (r *Room) finishSpecialContinuation() error {
	r.Choices = r.Choices[1:]
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	if r.Context != nil && r.Context.Step == "planner_placement" {
		r.Context.PendingAction = nil
		r.Context.PendingBonus = false
		r.Context.SpecialStage = ""
		r.restoreVisitorParent()
		return nil
	}
	// A Tuscany coin bonus may be deferred while the special-worker choice is
	// still at the head of the queue.  Once that choice is removed, settle the
	// deferred payment before dropping the serializable context; otherwise a
	// valid Farmer/Politico reward disappears at the end of the turn.
	if r.Context != nil && r.Context.CoinAfter > 0 {
		if actor := r.Player(r.Context.ActorID); actor != nil {
			actor.Coins += r.Context.CoinAfter
		}
		r.Context.CoinAfter = 0
	}
	if r.Context != nil && r.Context.Step == "planner" {
		// A special-worker pre/post choice belongs to the reservation
		// continuation, not to the normal turn loop.  Clear the consumed
		// action before restarting the planner queue so queuePostSpecial
		// cannot enqueue the same worker ability a second time.
		r.Context.PendingAction = nil
		r.Context.PendingBonus = false
		r.Context.SpecialStage = ""
		r.Context = nil
		r.startNextPlanner()
		return nil
	}
	r.Context = nil
	r.next()
	return nil
}

func (r *Room) activateMessenger(p *Player) bool {
	if !specialWorkerEnabled(r) || p == nil {
		return false
	}
	idx := -1
	for i, plan := range r.MessengerPlans {
		sp := r.space(plan.Space)
		if plan.PlayerID == p.ID && sp != nil && sp.Season == r.Phase {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	plan := r.MessengerPlans[idx]
	if !r.plannerCanExecute(plan) && !r.rhinePlannerHarvestCanEnable(plan) {
		r.MessengerPlans = append(r.MessengerPlans[:idx], r.MessengerPlans[idx+1:]...)
		r.AddLog(p.Name + " 的信使预约无法执行，未取得行动格奖励")
		r.Context = nil
		r.next()
		return true
	}
	bonus := normalizeBonusKey(r.placementBonus(plan.Space, plan.Slot))
	r.Context = &ActionContext{ActorID: p.ID, Space: plan.Space, TriggerSpace: plan.Space, Step: "messenger", ReturnTurnID: p.ID, TriggerSeat: &Seat{Season: r.Phase, PlayerID: p.ID, Slot: plan.Slot, WorkerType: "messenger"}, WorkerType: "messenger", Data: map[string]string{}}
	r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: p.ID, Kind: "messenger", Options: []string{"execute"}, Count: 1, ActionSpace: plan.Space, ActionSlot: plan.Slot, SpecialBonus: bonus})
	r.TurnID = p.ID
	return true
}

func (r *Room) resolveMessenger(p *Player, c Choice, a Action) error {
	if p == nil || r.Context == nil || r.Context.Step != "messenger" || a.Option != "execute" || a.Space != c.ActionSpace {
		return fmt.Errorf("无效信使预约执行")
	}
	idx := -1
	var plan PlannedPlacement
	for i, candidate := range r.MessengerPlans {
		if candidate.PlayerID == p.ID && candidate.Space == c.ActionSpace && candidate.Slot == c.ActionSlot && candidate.WorkerType == "messenger" {
			idx, plan = i, candidate
			break
		}
	}
	if idx < 0 || r.space(plan.Space) == nil || r.space(plan.Space).Season != r.Phase {
		return fmt.Errorf("信使预约已失效")
	}
	bonus := c.SpecialBonus != "" && !a.DeclineBonus
	a.Space, a.Slot, a.WorkerType, a.Large, a.Gray = plan.Space, plan.Slot, "", false, false
	r.Choices = r.Choices[1:]
	r.MessengerPlans = append(r.MessengerPlans[:idx], r.MessengerPlans[idx+1:]...)
	r.Context.PendingAction = cloneAction(a)
	r.Context.PendingBonus = bonus
	if err := r.performPlacement(p, a, bonus); err != nil {
		return err
	}
	r.AddLog(p.Name + " 执行了信使预约 · " + r.space(plan.Space).Name)
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	r.finishVisitorContext()
	return nil
}
