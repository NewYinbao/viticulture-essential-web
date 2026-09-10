package game

func (r *Room) View(id string) map[string]any {
	ps := make([]Player, 0)
	for _, p := range r.Players {
		q := *p
		q.Hand = nil
		q.HandCount = len(p.Hand)
		q.HandCounts = map[string]int{"vine": 0, "order": 0, "summer": 0, "winter": 0}
		for _, c := range p.Hand {
			q.HandCounts[c.Type]++
		}
		ps = append(ps, q)
	}
	p := r.Player(id)
	hand := []Card{}
	if p != nil {
		hand = p.Hand
	}
	active := r.TurnID == id && len(r.Choices) == 0
	actionReasons, cardReasons, optionReasons := r.availability(id)
	return map[string]any{"actionReasons": actionReasons, "cardReasons": cardReasons, "optionReasons": optionReasons, "grayWorkerOwner": r.GrayWorkerOwner, "triggerPlacement": r.publicTriggerPlacement(), "planned": r.Planned, "revealed": r.Revealed, "ruleset": r.Ruleset, "parentOptions": r.parentOptions(id), "pendingChoice": r.choiceView(id), "pendingCount": len(r.Choices), "deckCounts": r.deckCounts(), "winnerIds": r.WinnerIDs, "revision": r.Revision, "code": r.Code, "phase": r.Phase, "year": r.Year, "turnId": r.TurnID, "hostId": r.HostID, "youId": id, "players": ps, "hand": hand, "spaces": r.Spaces, "wakeSlots": r.WakeSlots, "log": r.Log, "winnerId": r.WinnerID, "rulesNotes": notes, "legal": map[string]bool{"canChoose": len(r.Choices) > 0 && r.Choices[0].PlayerID == id, "canStart": r.Phase == "lobby" && id == r.HostID && len(ps) >= 2, "canWake": r.Phase == "wake" && active, "canPlace": (r.Phase == "summer" || r.Phase == "winter") && active, "canPass": (r.Phase == "summer" || r.Phase == "winter") && active}}
}

func (r *Room) publicTriggerPlacement() any {
	if r.Context == nil || r.Context.TriggerSeat == nil {
		return nil
	}
	return struct {
		Space string
		Seat  Seat
	}{r.Context.Space, *r.Context.TriggerSeat}
}
