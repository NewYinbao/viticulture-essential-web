package game

// The gray meeple is a unique annual component, not a repeatable row reward.
func (r *Room) claimGrayWorker(p *Player) {
	if r.GrayWorkerOwner == "" {
		r.GrayWorkerOwner = p.ID
		r.GrayWorkerTracked = true
		p.GrayWorkerAvailable = true
		p.Workers++
	}
}

func (r *Room) isTriggerSeat(space string, seat Seat) bool {
	if r.Context == nil || r.Context.TriggerSeat == nil || triggerSpace(r.Context) != space {
		return false
	}
	s := r.Context.TriggerSeat
	return s.PlayerID == seat.PlayerID && s.Slot == seat.Slot && s.Large == seat.Large && s.WorkerType == seat.WorkerType && s.Gray == seat.Gray
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
