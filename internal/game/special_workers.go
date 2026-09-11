package game

import "fmt"

// specialWorkerDefs is the 11-card Tuscany Essential special-worker module.
// The wording follows the local publisher rules evidence; mechanics live in
// special_worker_actions.go.
var specialWorkerDefs = []SpecialWorker{
	{ID: "farmer", Name: "Farmer", Description: "将此工人放到行动板时，可从该行动的所有奖励中选择一个，即使不在奖励格；所有奖励均不受人数限制。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Farmer"},
	{ID: "mafioso", Name: "Mafioso", Description: "放到无奖励行动格并完成行动后，可再执行一次该行动；不得违反每年一次等规则。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Mafioso"},
	{ID: "chef", Name: "Chef", Description: "可放到对手工人占据的行动格，将该工人退回对手待命区；不能顶替 Chef。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Chef"},
	{ID: "innkeeper", Name: "Innkeeper", Description: "放置时可支付1金币给同一行动上的对手，从其手牌随机拿一张指定季节访客；2人局不可用。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Innkeeper"},
	{ID: "professore", Name: "Professore", Description: "放置时可从本季行动板取回自己的一个普通工人，该工人本年可再次放置。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Professore"},
	{ID: "soldato", Name: "Soldato", Description: "对手向有 Soldato 的同一行动放工人时，须支付1金币给该行动上的所有者；即使行动已满也可进入。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Soldato"},
	{ID: "politico", Name: "Politico", Description: "放到奖励行动格并完成行动、取得奖励后，可支付1金币再次取得该奖励。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Politico"},
	{ID: "oracle", Name: "Oracle", Description: "用 Oracle 抽牌时，同色多抽1张，再从本次抽到的牌中弃1张；每回合最多额外抽1张。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Oracle"},
	{ID: "merchant", Name: "Merchant", Description: "所有对手都已进入下一季后，Merchant 放到行动板并完成行动后，可抽任意类型牌1张。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Merchant"},
	{ID: "traveler", Name: "Traveler", Description: "可放到本年较早季节的任一空行动格，不受人数限制，并立即执行该行动。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Traveler"},
	{ID: "messenger", Name: "Messenger", Description: "可预定未来季节行动格；到该季第一次轮到你时，用预定行动替代放工人。若无法执行，不能取得奖励且回合结束。", RuleSource: "Tuscany Essential 规则 PDF 第6页 Messenger"},
}

func SpecialWorkerCatalog() []SpecialWorker {
	out := make([]SpecialWorker, len(specialWorkerDefs))
	copy(out, specialWorkerDefs)
	return out
}

func specialWorkerDef(id string) (SpecialWorker, bool) {
	for _, w := range specialWorkerDefs {
		if w.ID == id {
			return w, true
		}
	}
	return SpecialWorker{}, false
}

func specialWorkerEnabled(r *Room) bool {
	return r != nil && r.Config.Normalized().SpecialWorkers
}

func (r *Room) initSpecialWorkerPool() error {
	if !specialWorkerEnabled(r) {
		return nil
	}
	if len(r.SpecialWorkerPool) == 0 {
		pool := make([]string, 0, len(specialWorkerDefs))
		for _, w := range specialWorkerDefs {
			if len(r.Players) == 2 && w.ID == "innkeeper" {
				continue
			}
			pool = append(pool, w.ID)
		}
		for i := len(pool) - 1; i > 0; i-- {
			j := rnd(i + 1)
			pool[i], pool[j] = pool[j], pool[i]
		}
		r.SpecialWorkerPool = pool[:2]
	}
	if len(r.SpecialWorkerPool) != 2 {
		return fmt.Errorf("特殊工人牌堆必须恰好公开2类")
	}
	seen := map[string]bool{}
	for _, id := range r.SpecialWorkerPool {
		if seen[id] {
			return fmt.Errorf("特殊工人牌堆重复")
		}
		if _, ok := specialWorkerDef(id); !ok {
			return fmt.Errorf("未知特殊工人：%s", id)
		}
		if len(r.Players) == 2 && id == "innkeeper" {
			return fmt.Errorf("2人局不可使用Innkeeper")
		}
		seen[id] = true
	}
	return nil
}

func (r *Room) specialWorkerAvailable(id string) bool {
	for _, v := range r.SpecialWorkerPool {
		if v == id {
			return true
		}
	}
	return false
}

func (p *Player) hasSpecialWorker(id string) bool {
	for _, v := range p.SpecialWorkers {
		if v == id {
			return true
		}
	}
	return false
}

func (p *Player) specialWorkerReady(id string, year int) bool {
	return p.hasSpecialWorker(id) && p.SpecialWorkerReady[id] <= year && !p.SpecialWorkerUsed[id]
}

func (p *Player) ensureSpecialMaps() {
	if p.SpecialWorkerUsed == nil {
		p.SpecialWorkerUsed = map[string]bool{}
	}
	if p.SpecialWorkerReady == nil {
		p.SpecialWorkerReady = map[string]int{}
	}
}

func (r *Room) specialWorkerOptions(p *Player) []string {
	return r.specialWorkerOptionsWithDiscount(p, 0)
}

func (r *Room) specialWorkerOptionsWithDiscount(p *Player, discount int) []string {
	if p == nil {
		return nil
	}
	ids := []string{"regular"}
	if p.GrandeRemoved {
		ids = append(ids, "grande")
	}
	if specialWorkerEnabled(r) {
		ids = append(ids, r.SpecialWorkerPool...)
	}
	out := []string{}
	for _, id := range ids {
		room := rhineCloneRoom(r)
		q := room.Player(p.ID)
		if q == nil {
			q = probePlayer(p)
			room.Players = append(room.Players, q)
		}
		if room.trainWorker(q, id, discount) == nil {
			out = append(out, id)
		}
	}
	return out
}
