package game

// The gray meeple is a unique annual component, not a repeatable row reward.
func (r *Room) claimGrayWorker(p *Player) {
	if r.GrayWorkerOwner == "" {
		r.GrayWorkerOwner = p.ID
		p.Workers++
	}
}

func (r *Room) isTriggerSeat(space string, seat Seat) bool {
	return r.Context != nil && r.Context.TriggerSeat != nil && r.Context.Space == space && *r.Context.TriggerSeat == seat
}

func (r *Room) hasRetrievableWorker(p *Player) bool {
	for _, sp := range r.Spaces {
		for _, seat := range sp.Occupied {
			if seat.PlayerID == p.ID && !r.isTriggerSeat(sp.ID, seat) {
				return true
			}
		}
	}
	return false
}
