package game

func (r *Room) View(id string) map[string]any {
	r.ensureAllStructureActionSpaces()
	ps := make([]Player, 0)
	for _, p := range r.Players {
		q := *p
		q.Hand = nil
		q.HandCount = len(p.Hand)
		q.HandCounts = map[string]int{"vine": 0, "order": 0, "summer": 0, "winter": 0, "structure": 0}
		for _, c := range p.Hand {
			q.HandCounts[c.Type]++
		}
		q.SpecialWorkerUsed = cloneBoolMap(p.SpecialWorkerUsed)
		q.SpecialWorkerReady = cloneIntMap(p.SpecialWorkerReady)
		ps = append(ps, q)
	}
	p := r.Player(id)
	hand := []Card{}
	if p != nil {
		hand = append(hand, p.Hand...)
	}
	active := r.TurnID == id && len(r.Choices) == 0
	actionReasons, cardReasons, optionReasons := r.availability(id)
	v := map[string]any{"actionReasons": actionReasons, "cardReasons": cardReasons, "optionReasons": optionReasons, "grayWorkerOwner": r.GrayWorkerOwner, "triggerPlacement": r.publicTriggerPlacement(), "planned": publicPlans(r.Planned), "messengerPlans": publicPlans(r.MessengerPlans), "revealed": r.Revealed, "ruleset": r.Ruleset, "parentOptions": r.parentOptions(id), "pendingChoice": r.choiceView(id), "pendingCount": len(r.Choices), "deckCounts": r.deckCounts(), "winnerIds": r.WinnerIDs, "revision": r.Revision, "code": r.Code, "phase": r.Phase, "year": r.Year, "turnId": r.TurnID, "hostId": r.HostID, "youId": id, "players": ps, "hand": hand, "spaces": r.Spaces, "wakeSlots": r.WakeSlots, "log": r.Log, "winnerId": r.WinnerID, "rulesNotes": notes, "specialWorkerCatalog": SpecialWorkerCatalog(), "specialWorkerPool": r.SpecialWorkerPool, "legal": map[string]bool{"canChoose": len(r.Choices) > 0 && r.Choices[0].PlayerID == id, "canStart": r.Phase == "lobby" && id == r.HostID && len(ps) >= 2, "canWake": r.Phase == "wake" && active, "canPlace": (r.Phase == "summer" || r.Phase == "winter") && active, "canPass": (r.Phase == "summer" || r.Phase == "winter") && active}}
	v["config"] = r.Config.Normalized()
	capabilities := map[string]bool{"ee": true, "tuscany": true, "structures": true, "specialWorkers": true}
	for _, visitors := range []string{"ee_moor", "rhine"} {
		config := r.Config.Normalized()
		config.Visitors = visitors
		capabilities[visitors] = config.Playable() == nil
	}
	v["expansionAvailability"] = capabilities
	// Public capability status accompanies saved configurations, including
	// development snapshots. Help must never imply an unfinished save is playable.
	support := map[string]any{"playable": true, "reason": ""}
	if err := r.Config.Playable(); err != nil {
		support["playable"], support["reason"] = false, err.Error()
	}
	v["ruleSupport"] = support
	legal := v["legal"].(map[string]bool)
	legal["canConfigure"] = r.Phase == "lobby" && id == r.HostID
	legal["canStart"] = legal["canStart"] && r.Config.Playable() == nil
	legal["canPlace"] = r.actionSeason() && active
	legal["canPass"] = r.actionSeason() && active
	if r.tuscany() {
		v["influenceRegions"] = append([]InfluenceRegion{}, tuscanyRegions...)
		if p != nil && active {
			base, before := map[string]string{}, map[string]string{}
			q := probePlayer(p)
			q.Coins++
			for _, c := range p.Hand {
				base[c.ID], before[c.ID] = r.cardReason(p, c), r.cardReason(q, c)
			}
			v["visitorCardReasons"] = map[string]map[string]string{"base": base, "coinBefore": before}
		}
	}
	if p != nil {
		v["availableSpecialWorkers"] = r.specialWorkerOptions(p)
		placements := map[string][]WorkerPlacement{}
		if active && r.actionSeason() {
			for i := range r.Spaces {
				placements[r.Spaces[i].ID] = r.workerPlacements(p, &r.Spaces[i])
			}
		}
		v["workerPlacements"] = placements
	}
	return v
}

func cloneBoolMap(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := map[string]bool{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneIntMap(in map[string]int) map[string]int {
	if in == nil {
		return nil
	}
	out := map[string]int{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (r *Room) publicTriggerPlacement() any {
	if r.Context == nil || r.Context.TriggerSeat == nil {
		return nil
	}
	return struct {
		Space string
		Seat  Seat
	}{triggerSpace(r.Context), *r.Context.TriggerSeat}
}

// Future action inputs remain private; only the committed placement is public.
func publicPlans(plans []PlannedPlacement) []PlannedPlacement {
	out := append([]PlannedPlacement(nil), plans...)
	for i := range out {
		out[i].Action = nil
		out[i].SpecialWorker = nil
	}
	return out
}
