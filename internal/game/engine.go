package game

import (
	"encoding/json"
	"fmt"
)

// Apply is transactional even for non-HTTP callers and preserves player pointers on rollback.
func (r *Room) Apply(id string, a Action) error {
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
	}
	return err
}
func (r *Room) applyUnsafe(id string, a Action) error {
	p := r.Player(id)
	if p == nil {
		return fmt.Errorf("玩家不存在")
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
		r.Year = 1
		r.Spaces = NewSpaces(len(r.Players))
		r.Ruleset = "ee-base-v1"
		r.SpringLeader = rnd(len(r.Players))
		r.InitDecks()
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
	if r.Phase != "summer" && r.Phase != "winter" {
		return fmt.Errorf("请先选择起床顺序")
	}
	if a.Type == "pass" {
		p.Passed = true
		r.AddLog(p.Name + " 结束本季行动")
		r.next()
		return nil
	}
	if a.Type != "place" {
		return fmt.Errorf("未知行动")
	}
	var s *Space
	for i := range r.Spaces {
		if r.Spaces[i].ID == a.Space {
			s = &r.Spaces[i]
		}
	}
	if s == nil || (s.Season != r.Phase && s.Season != "any") {
		return fmt.Errorf("当前季节没有该行动")
	}
	if a.Large {
		if !p.LargeWorker {
			return fmt.Errorf("大工人已使用")
		}
	} else if p.Workers <= 0 {
		return fmt.Errorf("没有可用普通工人")
	}
	slot, bonus, e := placement(s, p, a)
	if e != nil {
		return e
	}
	r.Context = &ActionContext{ActorID: id, Space: a.Space, Step: "perform", ReturnTurnID: id, TriggerSeat: &Seat{PlayerID: id, Slot: slot, Large: a.Large}}
	if e := r.performPlacement(p, a, bonus); e != nil {
		return e
	}
	if len(r.Choices) == 0 {
		r.Context = nil
	}
	if a.Large {
		p.LargeWorker = false
	} else {
		p.Workers--
	}
	s.Occupied = append(s.Occupied, Seat{PlayerID: id, Large: a.Large, Slot: slot})

	r.AddLog(p.Name + " · " + s.Name)
	if len(r.Choices) > 0 {
		r.TurnID = r.Choices[0].PlayerID
		return nil
	}
	r.next()
	return nil
}
