package main

// The gray meeple is a unique annual component, not a repeatable row reward.
func (r *Room) claimGrayWorker(p *Player) {
	if r.GrayWorkerOwner == "" {
		r.GrayWorkerOwner = p.ID
		p.Workers++
	}
}

// Exceptional wines can already be beyond a missing cellar. Only crossing the
// next printed boundary requires that cellar; occupied destinations are checked separately.
func wineAgeBoundary(p *Player, value int) bool {
	return value < 9 && (value != 3 || has(p, "medium_cellar")) && (value != 6 || has(p, "large_cellar"))
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

// Enumerate legal single-wine results from current public resources, never decks.
func (r *Room) eachMadeWine(p *Player, exceptional bool, minimum int, accept func(*Player) bool) bool {
	try := func(ids []int) bool {
		q := probePlayer(p)
		buildings := append([]string{}, q.Buildings...)
		if exceptional {
			q.Buildings = append(q.Buildings, "large_cellar")
		}
		if makeWine(q, ids) != nil {
			return false
		}
		q.Buildings = buildings
		if q.Wines[len(q.Wines)-1].Value < minimum {
			return false
		}
		return accept(q)
	}
	for i := range p.Grapes {
		if try([]int{i}) {
			return true
		}
		for j := i + 1; j < len(p.Grapes); j++ {
			if try([]int{i, j}) {
				return true
			}
			for k := j + 1; k < len(p.Grapes); k++ {
				if try([]int{i, j, k}) {
					return true
				}
			}
		}
	}
	return false
}
func (r *Room) canMakeVisitor(p *Player, exceptional bool, minimum int) bool {
	return r.eachMadeWine(p, exceptional, minimum, func(*Player) bool { return true })
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
