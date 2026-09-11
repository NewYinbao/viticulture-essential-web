package game

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Apply is transactional even for non-HTTP callers and preserves player pointers on rollback.
func (r *Room) Apply(id string, a Action) error {
	// BonusOverride is server-owned continuation state (for example Farmer or
	// Rhine Virtuoso). Never accept it from the public Apply/HTTP boundary.
	a.BonusOverride = ""
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err = r.applyUnsafe(id, a); err != nil {
		var old Room
		_ = json.Unmarshal(b, &old)
		for i, p := range old.Players {
			if existing := r.Player(p.ID); existing != nil {
				*existing = *p
				old.Players[i] = existing
			}
		}
		*r = old
	} else {
		r.rhinePrepareWinterTurn()
	}
	return err
}
func (r *Room) applyUnsafe(id string, a Action) error {
	p := r.Player(id)
	if p == nil {
		return fmt.Errorf("玩家不存在")
	}
	if a.Type == "configure" {
		return r.configure(id, a)
	}
	if len(r.Choices) > 0 {
		return r.resolveChoice(id, a)
	}
	if a.Type == "choose" {
		return fmt.Errorf("没有待决选择")
	}
	if a.Type == "start" {
		if r.Phase != "lobby" || r.HostID != id || (len(r.Players) < 2 || len(r.Players) > 6) {
			return fmt.Errorf("需要房主且至少2名玩家")
		}
		if err := r.Config.Playable(); err != nil {
			return err
		}
		r.Config = r.Config.Normalized()
		r.Year = 1
		r.Spaces = r.spacesForRules()
		r.Ruleset = "ee-base-v1"
		if r.tuscany() {
			r.Ruleset = "tuscany-essential-v1"
		}
		r.SpringLeader = rnd(len(r.Players))
		if err := r.InitDecks(); err != nil {
			return err
		}
		if err := r.initSpecialWorkerPool(); err != nil {
			return err
		}
		r.setupParents()

		return nil
	}
	if r.Phase == "lobby" || r.Phase == "finished" {
		return fmt.Errorf("当前不能行动")
	}
	if r.TurnID != id {
		return fmt.Errorf("现在不是你的回合")
	}
	if a.Type == "wake" {
		if r.Phase != "wake" || a.Slot < 1 || a.Slot > 7 {
			return fmt.Errorf("请选择有效起床格")
		}
		if r.tuscany() {
			return r.tuscanyWake(p, a)
		}
		w := &r.WakeSlots[a.Slot-1]
		if w.PlayerID != "" {
			return fmt.Errorf("该起床格已占用")
		}
		if a.Slot == 5 && a.Color != "summer" && a.Color != "winter" {
			return fmt.Errorf("起床5须选择summer或winter")
		}
		w.PlayerID = id
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
		r.AddLog(fmt.Sprintf("%s 选择起床格%d", p.Name, a.Slot))
		for k := 1; k <= len(r.Players); k++ {
			q := r.Players[(r.SpringLeader+k)%len(r.Players)]
			if q.Wake == 0 {
				r.TurnID = q.ID
				return nil
			}
		}
		r.Phase = "summer"
		r.TurnID = r.ordered()[0].ID
		return nil
	}
	if !r.actionSeason() {
		return fmt.Errorf("请先选择起床顺序")
	}
	if a.Type == "pass" {
		if r.tuscany() {
			r.tuscanyPass(p)
			return nil
		}
		p.Passed = true
		r.AddLog(p.Name + " 结束本季行动")
		r.next()
		return nil
	}
	if a.Type != "place" {
		return fmt.Errorf("未知行动")
	}
	a = r.normalizeWorkerAction(p, a)
	var s *Space
	if strings.HasPrefix(a.Space, "structure_action:") {
		if owner, d, ok := r.structureOwnerAction(a.Space); ok {
			r.ensureStructureActionSpace(owner.ID, d)
		}
	}
	for i := range r.Spaces {
		if r.Spaces[i].ID == a.Space {
			s = &r.Spaces[i]
		}
	}
	if strings.HasPrefix(a.Space, "structure_action:") {
		owner, d, ok := r.structureOwnerAction(a.Space)
		if !ok || owner.ID != id {
			return fmt.Errorf("只能使用自己已建造的结构行动")
		}
		a.Building = d.ID
	}
	if strings.HasPrefix(a.Space, "structure_destroy:") {
		parts := strings.Split(a.Space, ":")
		if len(parts) != 2 || parts[1] != id {
			return fmt.Errorf("只能使用自己的拆除结构行动")
		}
	}
	if s == nil {
		return fmt.Errorf("行动空间不存在")
	}
	if !r.seasonAllowed(s, a) {
		return fmt.Errorf("当前季节没有该行动")
	}
	if a.WorkerType == "chef" && a.Slot == 0 {
		for _, seat := range s.Occupied {
			if seat.PlayerID != id && seat.WorkerType != "chef" {
				a.Slot = seat.Slot
				break
			}
		}
	}
	if err := r.bumpForChef(s, p, a); err != nil {
		return err
	}
	if err := r.workerCanPlace(p, a); err != nil {
		return err
	}
	if isBoardSpace(s) {
		if err := r.paySoldatoToll(a.Space, id); err != nil {
			return err
		}
	}
	placementSpace := s
	if a.WorkerType == "traveler" && isBoardSpace(s) && s.Season != r.Phase {
		// Traveler may use a printed action spot hidden by the player-count
		// setup.  The public space capacity remains unchanged.
		copySpace := *s
		copySpace.Capacity = 3
		placementSpace = &copySpace
	}
	slot, bonus, e := placement(placementSpace, p, a)
	if e != nil && hasOpponentSoldato(s, id) && !a.Large {
		full := true
		for i := 1; i <= placementSpace.Capacity; i++ {
			occupied := false
			for _, seat := range s.Occupied {
				if seat.Slot == i {
					occupied = true
					break
				}
			}
			if !occupied {
				full = false
				break
			}
		}
		if full && a.Slot == 0 {
			slot, bonus, e = 0, false, nil
		}
	}
	if e != nil {
		return e
	}
	// Declining a printed reward must remove that reward from every later
	// continuation as well.  In particular, training discounts and Politico
	// are only available after the actor actually took the printed bonus.
	effectiveBonus := bonus && !a.DeclineBonus
	if r.tuscany() {
		a.Slot = slot
	}
	r.Context = &ActionContext{ActorID: id, Space: a.Space, TriggerSpace: a.Space, Step: "perform", ReturnTurnID: id, TriggerSeat: &Seat{Season: r.Phase, PlayerID: id, Slot: slot, Large: a.Large, WorkerType: a.WorkerType, Gray: a.Gray}, WorkerType: a.WorkerType, PendingAction: cloneAction(a), PendingBonus: effectiveBonus, Data: map[string]string{"oracleBefore": fmt.Sprint(len(p.Hand))}}
	r.consumeWorker(p, a)
	if r.Phase == "winter" && r.Config.Visitors == "rhine" {
		p.RhineWinterActionYear = r.Year
	}
	r.Context.LastWorker = r.moorLastWorker(p)
	s.Occupied = append(s.Occupied, Seat{Season: r.Phase, PlayerID: id, Large: a.Large, Slot: slot, WorkerType: a.WorkerType, Gray: a.Gray})
	if a.WorkerType == "messenger" && s.Season != r.Phase {
		// The reservation commits only the worker's position. Inputs such as
		// cards, grapes and coins are deliberately chosen from the live state
		// when the future-season turn begins.
		r.MessengerPlans = append(r.MessengerPlans, PlannedPlacement{PlayerID: id, Space: a.Space, Slot: slot, WorkerType: a.WorkerType})
		r.Context = nil
		r.AddLog(p.Name + " 预约了 " + s.Name + " 的信使行动")
		r.next()
		return nil
	}
	if a.Space == "train" && specialWorkerEnabled(r) && a.SpecialWorker == "" {
		discount := 0
		if effectiveBonus {
			discount = 1
		}
		if opts := r.specialWorkerOptionsWithDiscount(p, discount); len(opts) > 1 {
			r.Context.PendingAction = cloneAction(a)
			r.Context.PendingBonus = effectiveBonus
			r.Context.SpecialStage = "train"
			r.enqueue(id, "special_train", opts, 1)
			r.TurnID = id
			return nil
		}
	}
	if r.queueSpecialPreChoice(p, a, effectiveBonus, s) {
		r.TurnID = id
		return nil
	}
	if e := r.performPlacement(p, a, effectiveBonus); e != nil {
		return e
	}
	r.AddLog(p.Name + " · " + s.Name)
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	if r.queuePostSpecial(p, a, effectiveBonus) {
		r.TurnID = id
		return nil
	}
	r.Context = nil
	r.next()
	return nil
}
