package game

import "strings"

// WorkerPlacement exposes only legal worker/slot combinations, never planned
// cards or action inputs. It is a read-only prerequisite preview, not an action.
type WorkerPlacement struct {
	Gray       bool   `json:"gray,omitempty"`
	WorkerType string `json:"workerType"`
	Large      bool   `json:"large"`
	Slot       int    `json:"slot"`
	Bonus      bool   `json:"bonus"`
	Toll       int    `json:"toll"`
}

func (r *Room) workerPlacements(p *Player, s *Space) []WorkerPlacement {
	out := []WorkerPlacement{}
	if p == nil || (strings.HasPrefix(s.ID, "structure_") && !strings.Contains(s.ID, ":"+p.ID)) {
		return out
	}
	workers := []Action{}
	if p.GrayWorkerAvailable {
		workers = append(workers, Action{Gray: true})
	}
	if p.availableRegularWorkers() > 0 {
		workers = append(workers, Action{})
	}
	if p.LargeWorker {
		workers = append(workers, Action{Large: true})
	}
	if specialWorkerEnabled(r) {
		for _, id := range p.SpecialWorkers {
			if r.specialWorkerAvailable(id) && p.specialWorkerReady(id, r.Year) {
				workers = append(workers, Action{WorkerType: id})
			}
		}
	}
	seen := map[WorkerPlacement]bool{}
	for _, worker := range workers {
		if !r.seasonAllowed(s, worker) {
			continue
		}
		capacity := s.Capacity
		if worker.WorkerType == "traveler" && isBoardSpace(s) && s.Season != r.Phase {
			capacity = 3
		}
		for requested := 0; requested <= capacity; requested++ {
			a := worker
			a.Slot = requested
			space := *s
			space.Capacity = capacity
			space.Occupied = append([]Seat(nil), s.Occupied...)
			if a.WorkerType == "chef" && s.Season != "any" {
				for i, seat := range space.Occupied {
					if seat.PlayerID == p.ID || (requested != 0 && seat.Slot != requested) {
						continue
					}
					if seat.WorkerType == "chef" {
						continue
					}
					a.Slot = seat.Slot
					space.Occupied = append(space.Occupied[:i], space.Occupied[i+1:]...)
					break
				}
			}
			// The engine charges after Chef's bump. The returned Soldato must
			// not make an otherwise free placement look unaffordable.
			toll := 0
			if isBoardSpace(s) {
				for _, seat := range space.Occupied {
					if seat.PlayerID != p.ID && seat.WorkerType == "soldato" {
						toll++
					}
				}
				if p.Coins < toll {
					continue
				}
			}
			slot, bonus, err := placement(&space, p, a)
			if err != nil && requested == 0 && !a.Large && hasOpponentSoldato(s, p.ID) {
				// Soldato permits a paid overflow placement when every slot is full.
				full := true
				for i := 1; i <= capacity; i++ {
					used := false
					for _, seat := range space.Occupied {
						if seat.Slot == i {
							used = true
						}
					}
					if !used {
						full = false
					}
				}
				if full {
					slot, bonus, err = 0, false, nil
				}
			}
			if err != nil {
				continue
			}
			choice := WorkerPlacement{Gray: a.Gray, WorkerType: a.WorkerType, Large: a.Large, Slot: slot, Bonus: bonus, Toll: toll}
			if !seen[choice] {
				seen[choice] = true
				out = append(out, choice)
			}
		}
	}
	return out
}
