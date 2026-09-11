package game

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func rhineExtraFields(op RhineOperation) []map[string]any {
	f := func(name, typ string, min, max int) []map[string]any {
		return []map[string]any{rhineField(name, typ, min, max)}
	}
	switch op.Kind {
	case "lose_worker":
		fs := f("rhineWorker", "rhine_worker_loss", 1, 1)
		fs[0]["availableOnly"] = op.Target == "available"
		return fs
	case "exchange_cards":
		return f("cardIds", "cards", 0, 99)
	case "exchange_draw":
		return f("drawTypes", "rhine_draw", 0, 99)
	case "make_supervisor":
		return f("recipes", "recipes", 1, op.Amount)
	case "virtuoso":
		return f("rhineVirtual", "rhine_virtuoso", 1, 1)
	case "age_grapes", "engineer_age":
		return f("grapes", "grapes", 1, op.Amount)
	case "son_field":
		return f("rhineField", "rhine_field", 1, 1)
	case "son_harvest":
		return f("rhineYield", "rhine_yield", 1, 1)
	case "plan_trigger":
		return f("rhinePlan", "rhine_plan", 1, 1)
	case "sell_vines":
		fs := f("cardIds", "cards", 0, 99)
		fs[0]["filter"] = "vine"
		return fs
	case "buy_cards":
		return f("rhinePurchase", "rhine_purchase", 0, 99)
	case "age_two_total":
		return f("wineIds", "wines", 1, 2)
	case "influence_place":
		return f("influence", "rhine_influence", 1, 1)
	case "return_stars", "move_stars", "lobbyist":
		fs := f("rhineStars", "rhine_stars", 0, op.Amount)
		fs[0]["operation"] = op.Kind
		if op.Kind == "return_stars" {
			fs[0]["min"] = op.Amount
			fs[0]["own"] = true
		}
		if op.Kind == "lobbyist" {
			fs[0]["min"] = 1
			fs[0]["max"] = 36
		}
		return fs
	case "destroy":
		return f("building", "rhine_destroy", 1, 1)
	case "destroy_wine":
		return []map[string]any{rhineField("building", "rhine_destroy", 1, 1), map[string]any{"name": "color", "type": "rhine_wine", "valueFromBuilding": true}}
	case "train":
		fs := f("specialWorker", "trainingWorker", 1, 1)
		fs[0]["cost"] = op.Amount
		fs[0]["now"] = op.Target == "now"
		return fs
	case "sell_grapes", "grape_vendor":
		return f("grapes", "grapes", 1, op.Amount)
	case "plant_ampelograph", "plant_exact":
		return f("plant", "plant", op.Amount, op.Amount)
	case "harvest_special", "harvest_chemist":
		return f("fields", "fields", 1, 1)
	case "fill_any", "fill_premium", "fill_early":
		fs := f("cardId", "cards", 1, 1)
		fs[0]["filter"] = "order"
		return append(fs, rhineField("wineIds", "wines", 1, 3))
	case "discard_new":
		fs := f("cardIds", "cards", op.Amount, op.Amount)
		fs[0]["filter"] = "vine"
		return fs
	case "uproot_keep":
		return f("uproot", "uproot", 1, 1)
	case "cellarmaster", "wine_orders", "red_wine_reward":
		return f("wineIds", "wines", 1, 1)
	case "make_two_sparkling":
		return f("grapes", "grapes", 2, 2)
	case "age_wines":
		return f("wineIds", "wines", op.Amount, op.Amount)
	case "wine_buy":
		fs := f("rhineValue", "rhine_value", 1, 9)
		return append(fs, map[string]any{"name": "color", "type": "rhine_wine", "valueFromPurchase": true})
	case "draw_fields":
		return f("drawTypes", "rhine_draw", 1, 3)
	}
	return nil
}

func rhineMakeExceptional(p *Player, a Action, n, minimum int) error {
	q := probePlayer(p)
	original := append([]string{}, q.Buildings...)
	count := len(q.Wines)
	q.Buildings = append(q.Buildings, "large_cellar")
	if err := makeActionWines(q, a, n); err != nil {
		return err
	}
	q.Buildings = original
	for _, w := range q.Wines[count:] {
		if w.Value < minimum {
			return fmt.Errorf("每瓶新酒须至少价值%d", minimum)
		}
	}
	*p = *q
	return nil
}

func rhineValue(values map[string]string, key string) int {
	n, _ := strconv.Atoi(values[key])
	return n
}
func rhineOwnFields(p *Player) int {
	n := 0
	for _, f := range p.Fields {
		if !f.Sold {
			n++
		}
	}
	return n
}

func (r *Room) rhineExtraOperation(p *Player, id string, op RhineOperation, a Action, values map[string]string) error {
	switch op.Kind {
	case "lose_worker":
		return r.losePermanentWorker(p, a.WorkerID, op.Target == "available")
	case "exchange_cards":
		if err := r.visitorCards(p, a.CardIDs, len(a.CardIDs), false); err != nil {
			return err
		}
		values["exchangeCount"] = strconv.Itoa(len(a.CardIDs))
	case "exchange_draw":
		if len(a.Colors) != rhineValue(values, "exchangeCount") {
			return fmt.Errorf("抽牌数必须等于本次弃牌数")
		}
		for _, color := range a.Colors {
			if color != "vine" && color != "order" && color != "summer" && color != "winter" && !(color == "structure" && r.Config.Structures) {
				return fmt.Errorf("无效抽牌类型")
			}
		}
		for _, color := range a.Colors {
			r.draw(p, color)
		}
	case "make_supervisor":
		before := map[string]bool{}
		for _, wine := range p.Wines {
			before[wine.ID] = true
		}
		if err := makeActionWines(p, a, op.Amount); err != nil {
			return err
		}
		for _, wine := range p.Wines {
			if !before[wine.ID] && wine.Type == "sparkling" {
				p.VP++
			}
		}
		return nil
	case "virtuoso":
		return r.rhineVirtuoso(p, a)
	case "engineer_start":
		values["engineerRemaining"] = "2"
	case "engineer_decrement":
		n := rhineValue(values, "engineerRemaining")
		if n < 1 {
			return fmt.Errorf("最多付费执行3次")
		}
		values["engineerRemaining"] = strconv.Itoa(n - 1)
	case "engineer_age":
		selected := []string{}
		for _, i := range a.Grapes {
			if i < 0 || i >= len(p.Grapes) {
				return fmt.Errorf("无效葡萄")
			}
			selected = append(selected, p.Grapes[i].ID)
		}
		if original := values["engineerGrapes"]; original != "" {
			for _, id := range selected {
				if !strings.Contains(","+original+",", ","+id+",") {
					return fmt.Errorf("只能继续陈酿最初选定的葡萄")
				}
			}
		} else {
			values["engineerGrapes"] = strings.Join(selected, ",")
		}
		return rhineAgeGrapes(p, a.Grapes, op.Amount)
	case "age_grapes":
		return rhineAgeGrapes(p, a.Grapes, op.Amount)
	case "son_field", "son_harvest":
		return r.rhineSonOperation(p, op, a)
	case "plan_trigger":
		return r.rhinePlan(p, a)
	case "sell_vines":
		for _, cid := range a.CardIDs {
			if cardIndex(p, cid, "vine") < 0 {
				return fmt.Errorf("只能出售手中的葡萄藤")
			}
		}
		if err := r.visitorCards(p, a.CardIDs, len(a.CardIDs), false); err != nil {
			return err
		}
		p.Coins += 3 * len(a.CardIDs)
		values["vinesSold"] = strconv.Itoa(len(a.CardIDs))
	case "buy_cards":
		if len(a.Colors) == 0 && rhineValue(values, "vinesSold") == 0 {
			return fmt.Errorf("至少出售或购买1张牌")
		}
		cost := 0
		for _, color := range a.Colors {
			switch color {
			case "vine":
				cost += 2
			case "winter":
				cost += 4
			default:
				return fmt.Errorf("只能购买葡萄藤或冬访客")
			}
		}
		if err := visitorPay(p, cost, 0); err != nil {
			return err
		}
		for _, color := range a.Colors {
			r.draw(p, color)
		}
	case "age_two_total":
		return rhineAgeTwoTotal(p, a.WineIDs)
	case "influence_place", "general_responses", "return_stars", "move_stars", "lobbyist":
		return r.rhineBoardOperation(p, id, op, a, values)
	case "destroy":
		_, err := r.rhineDestroy(p, a.Building)
		return err
	case "destroy_wine":
		cost, err := r.rhineDestroy(p, a.Building)
		if err != nil {
			return err
		}
		return rhineGainWine(p, a.Color, cost, false)
	case "require_vp_threshold":
		reached := false
		for _, q := range r.Players {
			if q.VP >= op.Amount {
				reached = true
			}
		}
		if reached != (op.Target == "yes") {
			return fmt.Errorf("分数门槛不匹配")
		}
	case "require_has":
		if !has(p, op.Target) {
			return fmt.Errorf("须已建造%s", op.Target)
		}
	case "require_missing":
		if has(p, op.Target) {
			return fmt.Errorf("已建造%s", op.Target)
		}
	case "build_fixed":
		return r.visitorBuild(p, op.Target, op.Amount)
	case "if_has_coins":
		if has(p, op.Target) {
			p.Coins += op.Amount
		}
	case "if_has_vp":
		if has(p, op.Target) {
			p.VP += op.Amount
		}
	case "field_coins":
		p.Coins += rhineOwnFields(p) * op.Amount
	case "draw_fields":
		if rhineOwnFields(p) == 0 {
			return fmt.Errorf("至少需要1块未出售的田地")
		}
		return r.rhineOperation(p, id, RhineOperation{Kind: "draw_any", Amount: rhineOwnFields(p)}, a, values)
	case "draw_new":
		before := map[string]bool{}
		for _, c := range p.Hand {
			before[c.ID] = true
		}
		for n := 0; n < op.Amount; n++ {
			r.draw(p, op.Target)
		}
		ids := []string{}
		for _, c := range p.Hand {
			if !before[c.ID] {
				ids = append(ids, c.ID)
			}
		}
		values["drawn"] = strings.Join(ids, ",")
	case "discard_new":
		allowed := "," + values["drawn"] + ","
		for _, cid := range a.CardIDs {
			if !strings.Contains(allowed, ","+cid+",") {
				return fmt.Errorf("只能弃置本次新抽的牌")
			}
		}
		return r.visitorCards(p, a.CardIDs, op.Amount, false)
	case "uproot_keep":
		if len(a.CardIDs) != 1 || len(a.Fields) != 1 {
			return fmt.Errorf("须选择1张已种葡萄藤")
		}
		return uproot(p, Action{CardID: a.CardIDs[0], Field: a.Fields[0]})
	case "plant_ampelograph":
		if err := r.visitorPlant(p, a, 1, false, true); err != nil {
			return err
		}
		fs := a.Fields
		if len(fs) == 0 {
			fs = []int{a.Field}
		}
		f := p.Fields[fs[0]]
		total := 0
		for _, c := range f.Vines {
			total += c.Red + c.White
		}
		if total <= f.Capacity {
			p.Coins += 2
		}
	case "plant_exact":
		if len(a.CardIDs) != op.Amount {
			return fmt.Errorf("须种恰好%d藤", op.Amount)
		}
		return r.visitorPlant(p, a, op.Amount, false, false)
	case "sell_grapes":
		if len(a.Grapes) < 1 || len(a.Grapes) > op.Amount {
			return fmt.Errorf("须选择1至%d葡萄", op.Amount)
		}
		coins := 0
		seen := map[int]bool{}
		for _, i := range a.Grapes {
			if i < 0 || i >= len(p.Grapes) || seen[i] {
				return fmt.Errorf("无效或重复葡萄")
			}
			seen[i] = true
			coins += 3 * ((p.Grapes[i].Value + 2) / 3)
		}
		if err := r.rhineOperation(p, id, RhineOperation{Kind: "discard_grape", Amount: len(a.Grapes)}, a, values); err != nil {
			return err
		}
		p.Coins += coins
	case "grape_vendor":
		if len(a.Grapes) != 1 || a.Grapes[0] < 0 || a.Grapes[0] >= len(p.Grapes) {
			return fmt.Errorf("须选择1葡萄")
		}
		value := p.Grapes[a.Grapes[0]].Value
		if err := visitorGrape(p, a.Grapes); err != nil {
			return err
		}
		if op.Target == "coins" {
			p.Coins += value
		} else {
			for n := 0; n < (value+1)/2; n++ {
				r.draw(p, "winter")
			}
		}
	case "bureaucrat":
		lowest := true
		for _, q := range r.Players {
			if q.ID != p.ID && q.VP <= p.VP {
				lowest = false
			}
		}
		if lowest {
			p.Coins += 5
		} else {
			if err := visitorPay(p, 1, 0); err != nil {
				return err
			}
			r.draw(p, "vine")
			r.draw(p, "winter")
			r.draw(p, "order")
		}
	case "cellarmaster":
		w, err := rhineSelectedWine(p, a.WineIDs)
		if err != nil {
			return err
		}
		highest := true
		for _, q := range r.Players {
			for _, other := range q.Wines {
				if other.Value > w.Value {
					highest = false
				}
			}
		}
		if err := visitorWine(p, a.WineIDs, 1); err != nil {
			return err
		}
		p.Coins += 4
		if highest {
			p.VP += 2
		}
	case "wine_orders":
		w, err := rhineSelectedWine(p, a.WineIDs)
		if err != nil {
			return err
		}
		if err := visitorWine(p, a.WineIDs, 1); err != nil {
			return err
		}
		for n := 0; n < w.Value; n++ {
			r.draw(p, "order")
		}
	case "wine_buy":
		if a.Slot < 1 || a.Slot > 9 {
			return fmt.Errorf("酒的价值须在1至9之间")
		}
		if err := visitorPay(p, a.Slot, 0); err != nil {
			return err
		}
		return rhineGainWine(p, a.Color, a.Slot, false)
	case "red_wine_reward":
		w, err := rhineSelectedWine(p, a.WineIDs)
		if err != nil {
			return err
		}
		if w.Type != "red" {
			return fmt.Errorf("须为红酒")
		}
		if err := visitorWine(p, a.WineIDs, 1); err != nil {
			return err
		}
		if op.Target == "coins" {
			p.Coins += w.Value
		} else {
			p.VP += 2
		}
	case "fill_early":
		reward := 2
		if p.VP <= 5 {
			reward = 4
		}
		a.Space = "fill_order"
		if err := r.perform(p, a); err != nil {
			return err
		}
		p.Coins += reward
	case "fill_any", "fill_premium":
		return r.rhineFill(p, a, op.Kind)
	case "make_two_sparkling":
		return rhineTwoSparkling(p, a.Grapes)
	case "harvest_special", "harvest_chemist":
		return r.rhineHarvest(p, a, op.Kind)
	case "age_wines":
		return rhineAgeWines(p, a.WineIDs, op.Amount)
	case "opponents":
		for offset := 1; offset < len(r.Players); offset++ {
			actorIndex := 0
			for i, q := range r.Players {
				if q.ID == p.ID {
					actorIndex = i
					break
				}
			}
			q := r.Players[(actorIndex+offset)%len(r.Players)]
			r.rhineMenu(q, id, op.Target, nil, values)
		}
	case "none":
		return nil
	default:
		return fmt.Errorf("未实现Rhine操作：%s", op.Kind)
	}
	return nil
}

func rhineSelectedWine(p *Player, ids []string) (Wine, error) {
	if len(ids) == 1 {
		for _, w := range p.Wines {
			if w.ID == ids[0] {
				return w, nil
			}
		}
	}
	return Wine{}, fmt.Errorf("须选择自己的1瓶酒")
}

func rhineTwoSparkling(p *Player, indices []int) error {
	if len(indices) != 2 || indices[0] == indices[1] {
		return fmt.Errorf("须选择2颗不同葡萄")
	}
	q := probePlayer(p)
	sum := 0
	for _, i := range indices {
		if i < 0 || i >= len(q.Grapes) {
			return fmt.Errorf("无效葡萄")
		}
		sum += q.Grapes[i].Value
	}
	if err := rhineGainWine(q, "sparkling", sum, false); err != nil {
		return err
	}
	a, b := indices[0], indices[1]
	if a < b {
		a, b = b, a
	}
	q.Grapes = append(q.Grapes[:a], q.Grapes[a+1:]...)
	q.Grapes = append(q.Grapes[:b], q.Grapes[b+1:]...)
	if has(q, "patio") {
		q.Coins += 2
	}
	if has(q, "penthouse") {
		q.VP++
	}
	*p = *q
	return nil
}

func rhineAgeWines(p *Player, ids []string, n int) error {
	if len(ids) != n {
		return fmt.Errorf("须选择%d瓶不同的酒", n)
	}
	q := probePlayer(p)
	selected := []Wine{}
	seen := map[string]bool{}
	for _, id := range ids {
		w, err := rhineSelectedWine(q, []string{id})
		if err != nil || seen[id] {
			return fmt.Errorf("无效或重复酒")
		}
		seen[id] = true
		selected = append(selected, w)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Value > selected[j].Value })
	for _, w := range selected {
		if err := ageWineByID(q, w.ID, 1); err != nil {
			return err
		}
	}
	*p = *q
	return nil
}

func rhineAgeTwoTotal(p *Player, ids []string) error {
	if len(ids) < 1 || len(ids) > 2 {
		return fmt.Errorf("须选择1或2瓶酒")
	}
	q := probePlayer(p)
	original := 0
	selected := []Wine{}
	seen := map[string]bool{}
	for _, id := range ids {
		w, err := rhineSelectedWine(q, []string{id})
		if err != nil || seen[id] {
			return fmt.Errorf("无效或重复酒")
		}
		seen[id] = true
		selected = append(selected, w)
		original += w.Value
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].Value > selected[j].Value })
	steps := 1
	if len(ids) == 1 {
		steps = 2
	}
	for _, w := range selected {
		if err := ageWineByID(q, w.ID, steps); err != nil {
			return err
		}
	}
	total := 0
	for _, id := range ids {
		w, _ := rhineSelectedWine(q, []string{id})
		total += w.Value
	}
	if total-original != 2 {
		return fmt.Errorf("必须合计增加恰好2价值，检查酒窖和酒槽")
	}
	*p = *q
	return nil
}

func (r *Room) rhineHarvest(p *Player, a Action, kind string) error {
	fields := a.Fields
	if a.HarvestAll {
		if !r.Config.Structures || !has(p, "harvest_machine") {
			return fmt.Errorf("需要收割机才能收获所有田地")
		}
		fields = nil
		for i := range p.Fields {
			if validateHarvestField(p, i) == nil {
				fields = append(fields, i)
			}
		}
		if len(fields) == 0 {
			return fmt.Errorf("没有可收获田地")
		}
	} else if len(fields) != 1 {
		return fmt.Errorf("须选择1田")
	}
	for _, idx := range fields {
		if err := r.rhineHarvestField(p, idx, kind); err != nil {
			return err
		}
	}
	if kind == "harvest_chemist" {
		p.VP++
	}
	r.queueFermentation(p)
	return nil
}

func (r *Room) rhineHarvestField(p *Player, idx int, kind string) error {
	if idx < 0 || idx >= len(p.Fields) {
		return fmt.Errorf("无效田地")
	}
	f := &p.Fields[idx]
	if f.Sold || f.Harvested || f.Structure != "" || len(f.Vines) == 0 {
		return fmt.Errorf("该田不能收获")
	}
	if kind == "harvest_special" {
		for _, c := range f.Vines {
			if c.Red > 0 {
				addGrape(p, "red", c.Red)
			}
			if c.White > 0 {
				addGrape(p, "white", c.White)
			}
		}
	} else {
		red, white := 0, 0
		for _, c := range f.Vines {
			red += c.Red
			white += c.White
		}
		if red > 1 {
			addGrape(p, "red", red-1)
		}
		if white > 1 {
			addGrape(p, "white", white-1)
		}
	}
	f.Harvested = true
	return nil
}

// Backtracking matches each delivered bottle exactly once. Premium bonuses
// require an individual +2 match; surplus totals cannot satisfy that condition.
func rhineWineMatch(wines, reqs []Wine, ignore bool, extra int) bool {
	if len(wines) != len(reqs) {
		return false
	}
	if len(reqs) == 0 {
		return true
	}
	for i, w := range wines {
		if (ignore || w.Type == reqs[0].Type) && w.Value >= reqs[0].Value+extra {
			left := append([]Wine{}, wines[:i]...)
			left = append(left, wines[i+1:]...)
			if rhineWineMatch(left, reqs[1:], ignore, extra) {
				return true
			}
		}
	}
	return false
}
func (r *Room) rhineFill(p *Player, a Action, kind string) error {
	idx := cardIndex(p, a.CardID, "order")
	if idx < 0 {
		return fmt.Errorf("须选择自己的订单")
	}
	order := p.Hand[idx]
	wines := []Wine{}
	seen := map[string]bool{}
	for _, id := range a.WineIDs {
		w, err := rhineSelectedWine(p, []string{id})
		if err != nil || seen[id] {
			return fmt.Errorf("无效或重复酒")
		}
		seen[id] = true
		wines = append(wines, w)
	}
	if !rhineWineMatch(wines, order.Requirements, kind == "fill_any", 0) {
		return fmt.Errorf("交付酒的价值或种类不符合订单")
	}
	premium := kind == "fill_premium" && rhineWineMatch(wines, order.Requirements, false, 2)
	if kind == "fill_any" {
		// Preserve the original public order in the discard pile. Temporarily adapt
		// only its requirements to the selected types for ordinary fill triggers.
		reqs := append([]Wine{}, order.Requirements...)
		sort.Slice(reqs, func(i, j int) bool { return reqs[i].Value > reqs[j].Value })
		sort.Slice(wines, func(i, j int) bool { return wines[i].Value > wines[j].Value })
		for i := range reqs {
			reqs[i].Type = wines[i].Type
		}
		p.Hand[idx].Requirements = reqs
	}
	a.Space = "fill_order"
	err := r.perform(p, a)
	if err != nil {
		return err
	}
	if kind == "fill_any" {
		cards := r.Discards["order"]
		for i := len(cards) - 1; i >= 0; i-- {
			if cards[i].ID == order.ID {
				cards[i] = order
				break
			}
		}
	}
	if premium {
		p.VP += 2
	}
	return nil
}

var rhineMenus = map[string]map[string]rhineOption{
	"engineer": {
		"again": {Label: "再付2金币，陈酿1或2葡萄", Operations: []RhineOperation{{Kind: "pay", Amount: 2}, {Kind: "engineer_age", Amount: 2}, {Kind: "engineer_decrement"}, {Kind: "choose", Target: "engineer"}}},
		"done":  {Label: "完成选择"},
	},
	"son_harvest":       {"harvest": {Label: "女婿：额外收获标记田地（总值至多3）", Operations: []RhineOperation{{Kind: "son_harvest"}}}, "skip": {Label: "不额外收获"}},
	"structure_destroy": {"destroy": {Label: "拆1建筑", Operations: []RhineOperation{{Kind: "destroy"}, {Kind: "choose", Target: "structure_build"}}}, "skip": {Label: "保留建筑"}},
	"structure_build":   {"build": {Label: "免费建1建筑", Operations: []RhineOperation{{Kind: "build", Amount: 99}}}, "skip": {Label: "不建造"}},
	"accountant":        {"draw": {Label: "抽1夏访客", Operations: []RhineOperation{{Kind: "draw", Amount: 1, Target: "summer"}}}, "skip": {Label: "不抽牌"}},
	"banker":            {"sell": {Label: "失1分，得3金币", Operations: []RhineOperation{{Kind: "lose_vp", Amount: 1}, {Kind: "coins", Amount: 3}}}, "skip": {Label: "不兑换"}},
	"middleman":         {"discard": {Label: "弃1藤", Operations: []RhineOperation{{Kind: "discard_cards", Amount: 1, Target: "vine"}}}, "pay": {Label: "付2金币", Operations: []RhineOperation{{Kind: "pay", Amount: 2}}}},
	"theologian":        {"coins": {Label: "弃红酒，按价值得金币", Operations: []RhineOperation{{Kind: "red_wine_reward", Target: "coins"}}}, "vp": {Label: "弃红酒，得2分", Operations: []RhineOperation{{Kind: "red_wine_reward", Target: "vp"}}}, "skip": {Label: "保留红酒"}},
}

func (r *Room) rhineMenu(p *Player, id, key string, tail []RhineOperation, values map[string]string) {
	branches := rhineMenuBranches(key, values)
	options := []string{}
	for option := range branches {
		options = append(options, option)
	}
	sort.Strings(options)
	r.visitorPrompt(p, id, "rhine_menu", options)
	c := &r.Choices[len(r.Choices)-1]
	c.Visitor.Rhine = &RhineStep{Operations: tail, Values: values, Branches: map[string][]RhineOperation{}}
	c.Labels = map[string]string{}
	c.OptionFields = map[string][]map[string]any{}
	for option, branch := range branches {
		c.Visitor.Rhine.Branches[option] = branch.Operations
		c.Labels[option] = branch.Label
		c.OptionFields[option] = rhineInput(append(append([]RhineOperation{}, branch.Operations...), tail...))
	}
	r.rhineDynamicFields(c, p)
}

func rhineMenuBranches(key string, values map[string]string) map[string]rhineOption {
	if key == "engineer" && rhineValue(values, "engineerRemaining") <= 0 {
		return map[string]rhineOption{"done": {Label: "结束陈酿"}}
	}
	return rhineMenus[key]
}

func rhineAgeGrapes(p *Player, indices []int, n int) error {
	if len(indices) < 1 || len(indices) > n {
		return fmt.Errorf("须选择1至%d葡萄", n)
	}
	q := probePlayer(p)
	seen := map[int]bool{}
	selected := append([]int{}, indices...)
	for _, i := range indices {
		if i < 0 || i >= len(q.Grapes) || seen[i] {
			return fmt.Errorf("无效或重复葡萄")
		}
		seen[i] = true
	}
	sort.Slice(selected, func(i, j int) bool { return q.Grapes[selected[i]].Value > q.Grapes[selected[j]].Value })
	aged := 0
	for _, i := range selected {
		g := &q.Grapes[i]
		if g.Value == 9 {
			continue
		}
		blocked := false
		for _, other := range q.Grapes {
			if other.Color == g.Color && other.Value == g.Value+1 {
				blocked = true
			}
		}
		if !blocked {
			g.Value++
			aged++
		}
	}
	if aged == 0 {
		return fmt.Errorf("所选葡萄均不能陈酿")
	}
	*p = *q
	return nil
}

func (r *Room) rhineDynamicFields(c *Choice, p *Player) {
	if p == nil {
		return
	}
	for option, fields := range c.OptionFields {
		ops := rhinePendingOperations(c.Visitor, option)
		for _, f := range fields {
			if f["type"] == "rhine_purchase" && c.Visitor.Rhine != nil && rhineValue(c.Visitor.Rhine.Values, "vinesSold") == 0 {
				f["min"] = 1
			}
			if f["type"] == "grapes" && c.Visitor.Rhine != nil && c.Visitor.Rhine.Values["engineerGrapes"] != "" {
				f["allowedIds"] = strings.Split(c.Visitor.Rhine.Values["engineerGrapes"], ",")
			}
			if f["type"] == "rhine_worker_loss" {
				f["workers"] = r.permanentWorkerLossChoices(p, f["availableOnly"] == true)
			}
			if f["type"] == "rhine_draw" && c.Visitor.Rhine != nil {
				for _, op := range ops {
					if op.Kind == "exchange_draw" {
						n := rhineValue(c.Visitor.Rhine.Values, "exchangeCount")
						f["min"], f["max"] = n, n
					}
				}
			}
			if f["type"] == "rhine_draw" {
				for _, op := range ops {
					if op.Kind == "draw_fields" {
						n := rhineOwnFields(p)
						f["min"], f["max"] = n, n
						break
					}
				}
			}
			if f["type"] == "cards" && c.Visitor.Rhine != nil && c.Visitor.Rhine.Values["drawn"] != "" {
				for _, op := range ops {
					if op.Kind == "discard_new" {
						f["allowedIds"] = strings.Split(c.Visitor.Rhine.Values["drawn"], ",")
						break
					}
				}
			}
		}
	}
}
