package game

import "strings"

// This evaluator works exclusively on copied current/public state. Draws are
// never performed in a probe: their identities are neither read nor predicted.
func (r *Room) rhineFeasible(p *Player, id string, ops []RhineOperation, values map[string]string) bool {
	if len(ops) == 0 {
		return true
	}
	op, tail := ops[0], ops[1:]
	if op.Kind == "plan_trigger" && (r.Context == nil || r.Context.TriggerSeat == nil) {
		for _, space := range r.Spaces {
			if space.Season == "any" || !r.plannerFutureSpace(space) {
				continue
			}
			for _, large := range []bool{false, true} {
				if _, _, err := placement(&space, p, Action{Large: large}); err == nil && (!large || p.LargeWorker) {
					return true
				}
			}
		}
		return false
	}
	if op.Kind == "choose" {
		for _, branch := range rhineMenuBranches(op.Target, values) {
			combined := append(append([]RhineOperation{}, branch.Operations...), tail...)
			if r.rhineFeasible(p, id, combined, values) {
				return true
			}
		}
		return false
	}
	switch op.Kind {
	case "general_responses":
		return r.rhineFeasible(p, id, tail, values)
	case "draw_new":
		return true // Two discarded cards are supplied by this draw.
	case "draw", "draw_any":
		q := probePlayer(p)
		typ := op.Target
		if op.Kind == "draw_any" {
			typ = "vine"
		}
		for i := 0; i < op.Amount; i++ {
			q.Hand = append(q.Hand, Card{ID: "rhine-probe-draw:" + NewID(), Type: typ})
		}
		return r.rhineFeasible(q, id, tail, values)
	case "draw_fields":
		return rhineOwnFields(p) > 0 && r.rhineFeasible(p, id, tail, values)
	case "if_opponent_vp_draw", "opponents", "exchange_draw":
		return r.rhineFeasible(p, id, tail, values)
	case "discard_new":
		return true
	}
	try := func(a Action) bool {
		room := rhineCloneRoom(r)
		room.Decks = map[string][]Card{}
		room.Discards = map[string][]Card{}
		q := room.Player(p.ID)
		if q == nil {
			q = probePlayer(p)
			room.Players = append(room.Players, q)
		} else {
			*q = *probePlayer(p)
		}
		copied := map[string]string{}
		for k, v := range values {
			copied[k] = v
		}
		if err := room.rhineOperation(q, id, op, a, copied); err != nil {
			return false
		}
		return room.rhineFeasible(q, id, tail, copied)
	}
	if len(rhineFields(op)) == 0 {
		return try(Action{})
	}
	return r.rhineEachInput(p, op, try)
}

func rhineCombinations(n, count int, accept func([]int) bool) bool {
	if count < 0 || count > n {
		return false
	}
	indices := []int{}
	var visit func(int) bool
	visit = func(start int) bool {
		if len(indices) == count {
			return accept(append([]int{}, indices...))
		}
		for i := start; i < n; i++ {
			indices = append(indices, i)
			if visit(i + 1) {
				return true
			}
			indices = indices[:len(indices)-1]
		}
		return false
	}
	return visit(0)
}

func (r *Room) rhineEachInput(p *Player, op RhineOperation, accept func(Action) bool) bool {
	switch op.Kind {
	case "lose_worker":
		for _, worker := range r.permanentWorkerLossChoices(p, op.Target == "available") {
			if accept(Action{WorkerID: worker.ID}) {
				return true
			}
		}
	case "exchange_cards", "exchange_draw":
		return accept(Action{})
	case "virtuoso":
		for _, sp := range r.Spaces {
			if seasonIndex(sp.Season) >= 0 && seasonIndex(sp.Season) < seasonIndex(r.Phase) {
				if accept(Action{Space: sp.ID, Mode: "tour"}) {
					return true
				}
			}
		}
	case "age_grapes", "engineer_age":
		for count := 1; count <= op.Amount; count++ {
			if rhineCombinations(len(p.Grapes), count, func(indices []int) bool { return accept(Action{Grapes: indices}) }) {
				return true
			}
		}
	case "son_field":
		for i := range p.Fields {
			if accept(Action{Field: i}) {
				return true
			}
		}
	case "son_harvest":
		for red := 0; red <= 3; red++ {
			for white := 0; white <= 3-red; white++ {
				if accept(Action{Grapes: []int{red, white}}) {
					return true
				}
			}
		}
	case "plan_trigger":
		for _, sp := range r.Spaces {
			if accept(Action{Space: sp.ID}) {
				return true
			}
		}
	case "sell_vines":
		if accept(Action{}) {
			return true
		}
		for _, card := range p.Hand {
			if card.Type == "vine" && accept(Action{CardIDs: []string{card.ID}}) {
				return true
			}
		}
	case "buy_cards":
		return accept(Action{}) || accept(Action{Colors: []string{"vine"}}) || accept(Action{Colors: []string{"winter"}})
	case "age_two_total":
		for count := 1; count <= 2; count++ {
			if rhineCombinations(len(p.Wines), count, func(indices []int) bool {
				a := Action{}
				for _, i := range indices {
					a.WineIDs = append(a.WineIDs, p.Wines[i].ID)
				}
				return accept(a)
			}) {
				return true
			}
		}
	case "influence_place":
		for _, to := range tuscanyRegions {
			if rhineStarCount(p) < 6 {
				if accept(Action{Influence: []InfluenceMove{{To: to.ID}}}) {
					return true
				}
			} else {
				for from, n := range p.Influence {
					if n > 0 && accept(Action{Influence: []InfluenceMove{{From: from, To: to.ID}}}) {
						return true
					}
				}
			}
		}
	case "move_stars":
		return accept(Action{})
	case "lobbyist":
		for _, owner := range r.Players {
			for from, n := range owner.Influence {
				if n > 0 && accept(Action{RhineStars: []RhineStarMove{{PlayerID: owner.ID, From: from}}}) {
					return true
				}
			}
		}
	case "return_stars":
		a := Action{}
		for from, n := range p.Influence {
			for i := 0; i < n && len(a.RhineStars) < op.Amount; i++ {
				a.RhineStars = append(a.RhineStars, RhineStarMove{PlayerID: p.ID, From: from})
			}
		}
		return accept(a)
	case "destroy", "destroy_wine":
		for _, b := range p.Buildings {
			for _, color := range []string{"red", "white", "blush", "sparkling"} {
				if accept(Action{Building: b, Color: color}) {
					return true
				}
			}
		}
	case "draw_any", "draw_fields":
		return accept(Action{})
	case "discard_cards":
		cards := []Card{}
		for _, c := range p.Hand {
			if op.Target == "" || c.Type == op.Target {
				cards = append(cards, c)
			}
		}
		return rhineCombinations(len(cards), op.Amount, func(indices []int) bool {
			a := Action{}
			for _, i := range indices {
				a.CardIDs = append(a.CardIDs, cards[i].ID)
			}
			return accept(a)
		})
	case "discard_grape", "sell_grapes", "grape_vendor", "make_two_sparkling":
		first, last := op.Amount, op.Amount
		if op.Kind == "sell_grapes" {
			first = 1
		}
		if op.Kind == "make_two_sparkling" {
			first, last = 2, 2
		}
		for count := first; count <= last; count++ {
			if rhineCombinations(len(p.Grapes), count, func(indices []int) bool { return accept(Action{Grapes: indices}) }) {
				return true
			}
		}
	case "discard_wine", "cellarmaster", "wine_orders", "red_wine_reward", "age_wines":
		count := 1
		if op.Kind == "age_wines" {
			count = op.Amount
		}
		return rhineCombinations(len(p.Wines), count, func(indices []int) bool {
			a := Action{}
			for _, i := range indices {
				a.WineIDs = append(a.WineIDs, p.Wines[i].ID)
			}
			return accept(a)
		})
	case "gain_grape", "gain_wine":
		for _, color := range []string{"red", "white", "blush", "sparkling"} {
			if accept(Action{Color: color}) {
				return true
			}
		}
	case "wine_buy":
		for value := 1; value <= 9; value++ {
			for _, color := range []string{"red", "white", "blush", "sparkling"} {
				if accept(Action{Slot: value, Color: color}) {
					return true
				}
			}
		}
	case "build":
		for building := range costs {
			if accept(Action{Building: building}) {
				return true
			}
		}
		if r.Config.Structures {
			for _, c := range p.Hand {
				if c.Type == "structure" && accept(Action{Building: c.StructureID}) {
					return true
				}
			}
		}
	case "train":
		if p.GrandeRemoved && accept(Action{SpecialWorker: "grande"}) {
			return true
		}
		if accept(Action{SpecialWorker: "regular"}) {
			return true
		}
		for _, worker := range r.SpecialWorkerPool {
			if accept(Action{SpecialWorker: worker}) {
				return true
			}
		}
	case "harvest", "harvest_special", "harvest_chemist":
		for count := 1; count <= op.Amount; count++ {
			if rhineCombinations(len(p.Fields), count, func(indices []int) bool { return accept(Action{Fields: indices}) }) {
				return true
			}
		}
	case "plant", "plant_ignore_buildings", "plant_ignore_capacity", "plant_ampelograph", "plant_exact":
		vines := []Card{}
		for _, c := range p.Hand {
			if c.Type == "vine" && !strings.HasPrefix(c.ID, "rhine-probe-draw:") {
				vines = append(vines, c)
			}
		}
		first := 1
		if op.Kind == "plant_exact" {
			first = op.Amount
		}
		for count := first; count <= op.Amount; count++ {
			if rhineCombinations(len(vines), count, func(indices []int) bool {
				a := Action{}
				for _, idx := range indices {
					a.CardIDs = append(a.CardIDs, vines[idx].ID)
				}
				var assign func() bool
				assign = func() bool {
					if len(a.Fields) == len(a.CardIDs) {
						return accept(a)
					}
					for f := range p.Fields {
						a.Fields = append(a.Fields, f)
						if assign() {
							return true
						}
						a.Fields = a.Fields[:len(a.Fields)-1]
					}
					return false
				}
				return assign()
			}) {
				return true
			}
		}
	case "uproot_keep":
		for i, f := range p.Fields {
			for _, c := range f.Vines {
				if accept(Action{CardIDs: []string{c.ID}, Fields: []int{i}}) {
					return true
				}
			}
		}
	case "fill", "fill_any", "fill_premium", "fill_early":
		for _, c := range p.Hand {
			if c.Type != "order" || strings.HasPrefix(c.ID, "rhine-probe-draw:") {
				continue
			}
			if rhineCombinations(len(p.Wines), len(c.Requirements), func(indices []int) bool {
				a := Action{CardID: c.ID}
				for _, idx := range indices {
					a.WineIDs = append(a.WineIDs, p.Wines[idx].ID)
				}
				return accept(a)
			}) {
				return true
			}
		}
	case "make", "make_exceptional", "make_supervisor":
		// Recipes refer to the original grape list. Enumerate legal disjoint batches
		// so Hired Hand can make the two different wines needed by its later order.
		recipes := [][]int{}
		for count := 1; count <= 3; count++ {
			rhineCombinations(len(p.Grapes), count, func(indices []int) bool {
				q := probePlayer(p)
				if op.Kind == "make_exceptional" {
					q.Buildings = append(q.Buildings, "large_cellar")
				}
				if makeWine(q, indices) == nil {
					recipes = append(recipes, indices)
				}
				return false
			})
		}
		var batch func(int, [][]int, map[int]bool) bool
		batch = func(start int, chosen [][]int, used map[int]bool) bool {
			if len(chosen) > 0 && accept(Action{Recipes: chosen}) {
				return true
			}
			if len(chosen) >= op.Amount {
				return false
			}
			for idx := start; idx < len(recipes); idx++ {
				recipe := recipes[idx]
				free := true
				for _, i := range recipe {
					if used[i] {
						free = false
					}
				}
				if !free {
					continue
				}
				for _, i := range recipe {
					used[i] = true
				}
				if batch(idx+1, append(chosen, recipe), used) {
					return true
				}
				for _, i := range recipe {
					delete(used, i)
				}
			}
			return false
		}
		return batch(0, nil, map[int]bool{})
	}
	return false
}

func rhinePendingOperations(s *VisitorStep, option string) []RhineOperation {
	if s.Rhine == nil {
		return rhineDefinitions[s.CardID].Options[option].Operations
	}
	result := append([]RhineOperation{}, s.Rhine.Branches[option]...)
	return append(result, s.Rhine.Operations...)
}

func (r *Room) rhineContinuationFeasible(p *Player, s *VisitorStep) bool {
	for _, choice := range r.Choices {
		if choice.Visitor == nil || choice.Visitor.CardID != s.CardID || choice.PlayerID != p.ID {
			continue
		}
		v := choice.Visitor
		if v.Rhine == nil {
			continue
		}
		possible := false
		for _, option := range choice.Options {
			if r.rhineFeasible(p, s.CardID, rhinePendingOperations(v, option), v.Rhine.Values) {
				possible = true
				break
			}
		}
		if !possible {
			return false
		}
	}
	return true
}

func rhineHasHiddenStep(ops []RhineOperation) bool {
	for _, op := range ops {
		if strings.HasPrefix(op.Kind, "draw") || op.Kind == "wine_orders" || op.Kind == "grape_vendor" && op.Target == "draw" {
			return true
		}
	}
	return false
}
