package main

import "fmt"

func init() {
	add := func(id, text string, opts ...string) { visitorDefs[id] = visitorDef{text, opts} }
	add("summer-17", "抽藤、订单、冬访客各1；每个对手可抽1夏访客", "resolve")
	add("summer-18", "所有玩家可减2金币建造，每个参与对手令你得1分", "resolve")
	add("summer-21", "得5金币；对手可失1分得3金币", "resolve")
	add("summer-31", "每个对手可付你2金币，不付则你得1分", "resolve")
	add("summer-37", "所有玩家可种1藤，每个参与对手令你得2金币", "resolve")

	add("winter-16", "弃1葡萄或酒得1分", "grape", "wine")
	add("winter-17", "所有玩家可酿至多2酒，每个参与对手让你抽1藤/夏访客", "resolve")
	add("winter-25", "所有玩家可取回大工人，每个参与对手令你得1分", "resolve")
	add("winter-35", "选至多3对手各给你1夏访客，无法给则你得1分", "select")
	add("winter-38", "所有玩家可付1金币培训，每个参与对手令你得1分", "resolve")
}
func (r *Room) visitorCycle(p *Player, s *VisitorStep, include bool, opts []string) {
	index := 0
	for i, q := range r.Players {
		if q == p {
			index = i
		}
	}
	for k := 0; k < len(r.Players); k++ {
		q := r.Players[(index+k)%len(r.Players)]
		if q == p && !include {
			continue
		}
		r.visitorPrompt(q, s.CardID, "reply", opts)
	}
}
func (r *Room) visitorTransfer(from, to *Player, ids []string, n int, typ string) error {
	if len(ids) != n {
		return fmt.Errorf("须给出%d张手牌", n)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		i := -1
		for j, c := range from.Hand {
			if c.ID == id && (typ == "" || c.Type == typ) {
				i = j
			}
		}
		if i < 0 || seen[id] {
			return fmt.Errorf("无效手牌")
		}
		seen[id] = true
		to.Hand = append(to.Hand, from.Hand[i])
		removeCard(from, i)
	}
	return nil
}
func (r *Room) visitorInteraction(p *Player, s *VisitorStep, a Action) error {
	if s.CardID == "summer-23" {
		return r.visitorSequence(p, s, a)
	}
	actor := r.player(s.ActorID)
	if s.Stage == "reward" {
		r.draw(p, a.Option)
		return nil
	}
	if s.Stage == "reply" {
		if a.Option == "skip" {
			if s.CardID == "summer-31" {
				actor.VP++
			}
			return nil
		}
		switch s.CardID {
		case "summer-17":
			r.draw(p, "summer")
		case "summer-18":
			if e := r.visitorBuild(p, a.Building, 2); e != nil {
				return e
			}
			if p != actor {
				actor.VP++
			}
		case "summer-21":
			if e := visitorPay(p, 0, 1); e != nil {
				return e
			}
			p.Coins += 3
		case "summer-31":
			if e := visitorPay(p, 2, 0); e != nil {
				return e
			}
			actor.Coins += 2
		case "summer-37":
			if len(a.CardIDs) != 1 {
				return fmt.Errorf("须种1藤")
			}
			if e := r.visitorPlant(p, a, 1, false, false); e != nil {
				return e
			}
			if p != actor {
				actor.Coins += 2
			}
		case "winter-17":
			if len(a.Recipes) == 0 {
				return fmt.Errorf("至少酿1酒，或skip")
			}
			if e := r.visitorMake(p, a, 2); e != nil {
				return e
			}
			if p != actor {
				r.visitorPrompt(actor, s.CardID, "reward", []string{"vine", "summer"})
				last := r.Choices[len(r.Choices)-1]
				r.Choices = append([]Choice{last}, r.Choices[:len(r.Choices)-1]...)
			}
		case "winter-25":
			found := false
			for i := range r.Spaces {
				sp := &r.Spaces[i]
				for j, seat := range sp.Occupied {
					if seat.PlayerID == p.ID && seat.Large {
						r.cancelPlanned(sp.ID, seat)
						if sp.ID == "yoke" {
							p.YokeUsed = false
						}
						sp.Occupied = append(sp.Occupied[:j], sp.Occupied[j+1:]...)
						found = true
						break
					}
				}
			}
			if !found || p.LargeWorker {
				return fmt.Errorf("没有已放置的大工人")
			}
			p.LargeWorker = true
			if p != actor {
				actor.VP++
			}
		case "winter-35":
			return r.visitorTransfer(p, actor, a.CardIDs, 1, "summer")
		case "winter-38":
			if e := r.visitorTrain(p, 1, false); e != nil {
				return e
			}
			if p != actor {
				actor.VP++
			}
		}
		return nil
	}
	switch s.CardID {
	case "summer-17":
		r.draw(p, "vine")
		r.draw(p, "order")
		r.draw(p, "winter")
		r.visitorCycle(p, s, false, []string{"draw", "skip"})
	case "summer-18":
		r.visitorCycle(p, s, true, []string{"build", "skip"})
	case "summer-21":
		p.Coins += 5
		r.visitorCycle(p, s, false, []string{"sell", "skip"})
	case "summer-31":
		r.visitorCycle(p, s, false, []string{"pay", "skip"})
	case "summer-37":
		r.visitorCycle(p, s, true, []string{"plant", "skip"})
	case "winter-16":
		if a.Option == "grape" {
			if e := visitorGrape(p, a.Grapes); e != nil {
				return e
			}
		} else {
			if e := visitorWine(p, a.WineIDs, 1); e != nil {
				return e
			}
		}
		p.VP++
	case "winter-17":
		r.visitorCycle(p, s, true, []string{"make", "skip"})
	case "winter-25":
		r.visitorCycle(p, s, true, []string{"retrieve", "skip"})
	case "winter-35":
		if len(a.TargetIDs) > 3 {
			return fmt.Errorf("至多3对手")
		}
		seen := map[string]bool{}
		for _, id := range a.TargetIDs {
			q := r.player(id)
			if q == nil || q == p || seen[id] {
				return fmt.Errorf("无效对手")
			}
			seen[id] = true
			hasCard := false
			for _, c := range q.Hand {
				if c.Type == "summer" {
					hasCard = true
				}
			}
			if hasCard {
				r.visitorPrompt(q, s.CardID, "reply", []string{"give"})
			} else {
				p.VP++
			}
		}
	case "winter-38":
		r.visitorCycle(p, s, true, []string{"train", "skip"})
	default:
		return r.visitorSequence(p, s, a)
	}
	return nil
}
