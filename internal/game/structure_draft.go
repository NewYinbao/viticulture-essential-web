package game

import "fmt"

// beginStructureDraft implements the official EE-board fallback when the
// Tuscany extended board is not used: after Mama/Papa choices, each player
// receives four structure cards and the table drafts one card per round,
// passing the remainder to the player on the right.
func (r *Room) beginStructureDraft() bool {
	if !r.Config.Structures || r.tuscany() || len(r.Players) < 2 {
		return false
	}
	r.StructureDrafts = make(map[string][]Card, len(r.Players))
	r.StructureDraftRound = 1
	for _, p := range r.Players {
		pool := make([]Card, 0, 4)
		for len(pool) < 4 {
			c, ok := r.take("structure")
			if !ok {
				return false
			}
			pool = append(pool, c)
		}
		r.StructureDrafts[p.ID] = pool
	}
	r.queueStructureDraftChoices()
	r.TurnID = r.Choices[0].PlayerID
	r.AddLog("开局草拟结构牌：每人选1张，其余向右传递")
	return true
}

func (r *Room) queueStructureDraftChoices() {
	for _, p := range r.Players {
		pool := r.StructureDrafts[p.ID]
		options := make([]string, 0, len(pool))
		labels := make(map[string]string, len(pool))
		for _, c := range pool {
			options = append(options, c.ID)
			labels[c.ID] = c.Name
		}
		r.Choices = append(r.Choices, Choice{
			ID: NewID(), PlayerID: p.ID, Kind: "structure_draft", Options: options, Count: 1, Labels: labels,
		})
	}
}

func removeDraftCard(pool []Card, id string) (Card, []Card, error) {
	for i, c := range pool {
		if c.ID == id {
			return c, append(pool[:i], pool[i+1:]...), nil
		}
	}
	return Card{}, pool, fmt.Errorf("请选择当前草拟牌中的结构牌")
}

func (r *Room) resolveStructureDraft(p *Player, c Choice, a Action) error {
	pool := r.StructureDrafts[p.ID]
	card, rest, err := removeDraftCard(pool, a.Option)
	if err != nil {
		return err
	}
	r.StructureDrafts[p.ID] = rest
	p.Hand = append(p.Hand, card)
	return nil
}

func (r *Room) continueStructureDraft() bool {
	if r.StructureDraftRound <= 0 || len(r.StructureDrafts) == 0 {
		return false
	}
	if r.StructureDraftRound >= 4 {
		r.StructureDrafts = nil
		r.StructureDraftRound = 0
		r.spring()
		return true
	}
	passed := make(map[string][]Card, len(r.Players))
	for i, p := range r.Players {
		next := r.Players[(i+1)%len(r.Players)]
		passed[next.ID] = append(passed[next.ID], r.StructureDrafts[p.ID]...)
	}
	r.StructureDrafts = passed
	r.StructureDraftRound++
	r.queueStructureDraftChoices()
	r.TurnID = r.Choices[0].PlayerID
	return true
}
