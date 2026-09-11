package game

import (
	"fmt"
	"strings"
)

// StructureDef is the complete Tuscany Essential orange-card catalogue.
// RuleSource deliberately points to the printed rule page/card list used for
// the audit; it is persisted through the Card projection for UI/tests.
type StructureDef struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Cost           int    `json:"cost"`
	RequiresWorker bool   `json:"requiresWorker"`
	Category       string `json:"category"`
	Description    string `json:"description"`
	RuleSource     string `json:"ruleSource"`
}

const structureRuleSource = "Tuscany Essential 官方说明书第7页及出版社实际卡面（Boardspace 镜像，2026-09-11逐张复核）"

var structureDefs = []StructureDef{
	{"cask", "木桶", 2, true, "action", "使1瓶酒陈酿2级，并抽1张订单。", structureRuleSource},
	{"aqueduct", "渡槽", 3, false, "enhancement", "种植葡萄藤时忽略棚架与灌溉建筑要求。", structureRuleSource},
	{"wine_cave", "酒窖洞", 2, true, "action", "使至多2瓶酒各陈酿1级，并抽1张订单。", structureRuleSource},
	{"trading_post", "贸易站", 2, true, "action", "执行一次一换一交易；结构奖励可放置或移动1颗影响力星。", structureRuleSource},
	{"shop", "商店", 5, true, "action", "执行一次交付订单；结构奖励可放置或移动1颗影响力星。", structureRuleSource},
	{"wine_press", "压酒机", 4, true, "action", "酿造至多2瓶酒；结构奖励可放置或移动1颗影响力星。", structureRuleSource},
	{"school", "学校", 7, true, "action", "培训1名本年可用工人（普通免费，特殊加1金币）；另计学院费用，之后得1金币。", structureRuleSource},
	{"wine_bar", "酒吧", 4, true, "action", "弃置1瓶酒，得2分。", structureRuleSource},
	{"patio", "庭院", 3, false, "enhancement", "酿造桃红或起泡酒时获得2金币。", structureRuleSource},
	{"ristorante", "餐厅", 8, true, "action", "弃置1瓶酒和1颗红或白葡萄；结构奖励得3金币和3分。", structureRuleSource},
	{"guest_house", "客房", 3, true, "action", "弃置2张访客牌；结构奖励得2分。", structureRuleSource},
	{"cafe", "咖啡馆", 3, true, "action", "弃置1颗红或白葡萄；结构奖励得3金币和1分。", structureRuleSource},
	{"distiller", "蒸馏器", 2, false, "residual", "年末额外使所有葡萄各陈酿1级。", structureRuleSource},
	{"mercado", "市场", 5, false, "enhancement", "抽到订单牌时，可立即交付该订单。", structureRuleSource},
	{"studio", "工作室", 5, false, "enhancement", "之后每建造一座建筑（含固定建筑和结构牌），额外得1分。", structureRuleSource},
	{"barn", "谷仓", 5, false, "enhancement", "每年夏季结束时可弃2张牌得1分。", structureRuleSource},
	{"academy", "学院", 4, false, "enhancement", "每位对手培训工人时向你支付1金币。", structureRuleSource},
	{"gazebo", "凉亭", 3, false, "enhancement", "你导览时可放置或移动1颗影响力星。", structureRuleSource},
	{"workshop", "工坊", 3, false, "enhancement", "之后建造固定建筑或结构牌时少付1金币。", structureRuleSource},
	{"veranda", "阳台", 5, false, "enhancement", "每次交付订单额外得1分。", structureRuleSource},
	{"wine_parlor", "酒廊", 3, false, "enhancement", "每次交付订单额外得2金币。", structureRuleSource},
	{"label_factory", "酒标工厂", 3, true, "action", "支付3金币并交付1张订单；结构奖励得2分。", structureRuleSource},
	{"harvest_machine", "收割机", 2, false, "enhancement", "收获时可改为从所有未收获田地同时收获。", structureRuleSource},
	{"fermentation_tank", "发酵罐", 4, false, "enhancement", "收获至少1块田地后，可酿造1瓶酒。", structureRuleSource},
	{"charmat", "查玛法罐", 3, false, "enhancement", "可用1红葡萄和1白葡萄酿造起泡酒。", structureRuleSource},
	{"inn", "旅店", 4, false, "enhancement", "每打出1张访客牌得1金币。", structureRuleSource},
	{"tap_room", "品酒吧", 5, false, "enhancement", "打出访客牌后可弃1瓶酒得2分。", structureRuleSource},
	{"tavern", "酒馆", 5, false, "enhancement", "打出访客牌后可弃2颗红/白葡萄得3分。", structureRuleSource},
	{"banquet_hall", "宴会厅", 2, false, "enhancement", "移动影响力星时也获得放置影响力星的地区奖励。", structureRuleSource},
	{"penthouse", "顶层阁楼", 3, false, "enhancement", "酿造品质至少7的酒时，每瓶额外得1分。", structureRuleSource},
	{"fountain", "喷泉", 4, false, "enhancement", "对手导览时你得1金币。", structureRuleSource},
	{"mixer", "调酒器", 3, true, "action", "至多酿造1瓶桃红和1瓶起泡酒，并得1分。", structureRuleSource},
	{"storehouse", "仓库", 2, false, "residual", "年末额外使所有酒各陈酿1级。", structureRuleSource},
	{"statue", "雕像", 9, false, "residual", "年末得1分；该分不会触发最终年。", structureRuleSource},
	{"dock", "码头", 3, false, "residual", "年末抽1张订单牌。", structureRuleSource},
	{"silo", "筒仓", 2, false, "residual", "年末抽1张葡萄藤牌。", structureRuleSource},
}

func StructureCatalog() []Card {
	out := make([]Card, 0, len(structureDefs))
	for i, d := range structureDefs {
		source := d.RuleSource + fmt.Sprintf(" https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-%06d.jpg", i+1)
		out = append(out, Card{ID: "structure-" + d.ID, Type: "structure", Name: d.Name, Description: d.Description, Implemented: true, StructureID: d.ID, StructureCost: d.Cost, StructureWorker: d.RequiresWorker, StructureCategory: d.Category, RuleSource: source})
	}
	return out
}

func structureDef(id string) (StructureDef, bool) {
	id = strings.TrimPrefix(id, "structure-")
	for _, d := range structureDefs {
		if d.ID == id {
			return d, true
		}
	}
	return StructureDef{}, false
}

func structureCardID(id string) string { return "structure-" + strings.TrimPrefix(id, "structure-") }

func (p *Player) hasStructure(id string) bool {
	id = strings.TrimPrefix(id, "structure-")
	for _, b := range p.Buildings {
		if strings.TrimPrefix(b, "structure-") == id {
			return true
		}
	}
	for _, b := range p.StructureSlots {
		if strings.TrimPrefix(b, "structure-") == id {
			return true
		}
	}
	for _, f := range p.Fields {
		if strings.TrimPrefix(f.Structure, "structure-") == id && f.Structure != "" {
			return true
		}
	}
	return false
}

func (p *Player) structureCardInHand(id string) (int, StructureDef, bool) {
	d, ok := structureDef(id)
	if !ok {
		return -1, StructureDef{}, false
	}
	for i, c := range p.Hand {
		if c.Type == "structure" && strings.TrimPrefix(c.ID, "structure-") == d.ID {
			return i, d, true
		}
	}
	return -1, StructureDef{}, false
}

func structureActionSpaceID(owner, id string) string { return "structure_action:" + owner + ":" + id }
func structureDestroySpaceID(owner string) string    { return "structure_destroy:" + owner }

func (r *Room) ensureStructureActionSpace(owner string, d StructureDef) {
	if !d.RequiresWorker {
		return
	}
	id := structureActionSpaceID(owner, d.ID)
	for _, s := range r.Spaces {
		if s.ID == id {
			return
		}
	}
	r.Spaces = append(r.Spaces, Space{ID: id, Name: d.Name + " · 私人行动", Description: d.Description, Season: "any", Capacity: 1, Occupied: []Seat{}})
}

func (r *Room) ensureStructureDestroySpace(owner string) {
	id := structureDestroySpaceID(owner)
	for _, s := range r.Spaces {
		if s.ID == id {
			return
		}
	}
	r.Spaces = append(r.Spaces, Space{ID: id, Name: "拆除结构", Description: "弃置一座已建结构；不返还建造胜利分", Season: "any", Capacity: 1, Occupied: []Seat{}})
}

func (r *Room) ensureAllStructureActionSpaces() {
	if !r.Config.Structures {
		return
	}
	for _, p := range r.Players {
		for _, b := range p.StructureSlots {
			if d, ok := structureDef(b); ok {
				r.ensureStructureActionSpace(p.ID, d)
			}
		}
		for _, f := range p.Fields {
			if d, ok := structureDef(f.Structure); ok {
				r.ensureStructureActionSpace(p.ID, d)
			}
		}
		if hasAnyStructure(p) {
			r.ensureStructureDestroySpace(p.ID)
		}
	}
}

func hasAnyStructure(p *Player) bool {
	for _, b := range p.StructureSlots {
		if _, ok := structureDef(b); ok {
			return true
		}
	}
	for _, f := range p.Fields {
		if _, ok := structureDef(f.Structure); ok {
			return true
		}
	}
	return false
}

func (r *Room) structureOwnerAction(id string) (*Player, StructureDef, bool) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 || parts[0] != "structure_action" {
		return nil, StructureDef{}, false
	}
	d, ok := structureDef(parts[2])
	if !ok {
		return nil, StructureDef{}, false
	}
	p := r.Player(parts[1])
	if p == nil || !p.hasStructure(d.ID) {
		return nil, StructureDef{}, false
	}
	return p, d, true
}

func (r *Room) buildStructure(p *Player, a Action, discount int) error {
	if !r.Config.Structures {
		return fmt.Errorf("未启用Tuscany建筑模块")
	}
	i, d, ok := p.structureCardInHand(a.Building)
	if !ok {
		return fmt.Errorf("请选择自己的结构牌")
	}
	if p.hasStructure(d.ID) {
		return fmt.Errorf("已拥有该结构")
	}
	price := d.Cost - discount
	if has(p, "workshop") {
		price--
	}
	if price < 0 {
		price = 0
	}
	if p.Coins < price {
		return fmt.Errorf("金币不足，需要%d", price)
	}
	if a.Mode == "field" {
		if a.Field < 0 || a.Field >= len(p.Fields) {
			return fmt.Errorf("结构目标田地无效")
		}
		f := &p.Fields[a.Field]
		if f.Sold || len(f.Vines) > 0 || f.Structure != "" {
			return fmt.Errorf("只能占用未出售且空的田地")
		}
		f.Structure = d.ID
	} else if a.Mode == "mat" || a.Mode == "" {
		if len(p.StructureSlots) < 2 {
			p.StructureSlots = append(p.StructureSlots, make([]string, 2-len(p.StructureSlots))...)
		}
		slot := a.StructureSlot
		if slot < 0 || slot >= 2 {
			slot = -1
			for j, b := range p.StructureSlots {
				if b == "" {
					slot = j
					break
				}
			}
			if slot < 0 {
				return fmt.Errorf("建筑垫没有空位；可选择空田地")
			}
		}
		if p.StructureSlots[slot] != "" {
			return fmt.Errorf("建筑垫该位置已占用")
		}
		p.StructureSlots[slot] = d.ID
	} else {
		return fmt.Errorf("结构只能放在结构垫或空田地")
	}
	p.Coins -= price
	p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
	p.Buildings = append(p.Buildings, d.ID)
	p.VP++
	if d.ID != "studio" && has(p, "studio") {
		p.VP++
	}
	r.ensureStructureActionSpace(p.ID, d)
	r.ensureStructureDestroySpace(p.ID)
	r.AddLog(fmt.Sprintf("%s 建造结构 %s", p.Name, d.Name))
	return nil
}

func (r *Room) destroyStructure(p *Player, id string) error {
	id = strings.TrimPrefix(id, "structure-")
	if !p.hasStructure(id) {
		return fmt.Errorf("没有该结构")
	}
	for i, b := range p.StructureSlots {
		if b == id {
			p.StructureSlots[i] = ""
			goto removed
		}
	}
	for i := range p.Fields {
		if p.Fields[i].Structure == id {
			p.Fields[i].Structure = ""
			goto removed
		}
	}
removed:
	for i, b := range p.Buildings {
		if b == id {
			p.Buildings = append(p.Buildings[:i], p.Buildings[i+1:]...)
			break
		}
	}
	r.discard(structureCardForID(id))
	for i, s := range r.Spaces {
		if s.ID == structureActionSpaceID(p.ID, id) {
			r.Spaces[i].Occupied = nil
		}
	}
	r.AddLog(fmt.Sprintf("%s 拆除结构 %s", p.Name, id))
	return nil
}

func structureCardForID(id string) Card {
	id = strings.TrimPrefix(id, "structure-")
	if d, ok := structureDef(id); ok {
		return Card{ID: structureCardID(d.ID), Type: "structure", Name: d.Name, Description: d.Description, Implemented: true, StructureID: d.ID, StructureCost: d.Cost, StructureWorker: d.RequiresWorker, StructureCategory: d.Category, RuleSource: d.RuleSource}
	}
	return Card{ID: structureCardID(id), Type: "structure", StructureID: id, Name: dName(id)}
}

func dName(id string) string {
	if d, ok := structureDef(id); ok {
		return d.Name
	}
	return id
}

func ageWineByID(p *Player, id string, steps int) error {
	for i := range p.Wines {
		if p.Wines[i].ID == id {
			trial := append([]Wine(nil), p.Wines...)
			for n := 0; n < steps; n++ {
				if !wineAgeBoundary(p, trial[i].Value) {
					continue
				}
				target := trial[i].Value + 1
				for j, other := range trial {
					if j != i && other.Type == trial[i].Type && other.Value == target {
						return fmt.Errorf("酒窖中已有同类品质%d的酒", target)
					}
				}
				trial[i].Value = target
			}
			p.Wines[i].Value = trial[i].Value
			return nil
		}
	}
	return fmt.Errorf("酒不在自己的酒窖")
}

func removeGrapeByID(p *Player, id, color string, count int) error {
	for count > 0 {
		found := -1
		for i, g := range p.Grapes {
			colorOK := color == "" || g.Color == color || (color == "red_or_white" && (g.Color == "red" || g.Color == "white"))
			if (id == "" || g.ID == id) && colorOK {
				found = i
				break
			}
		}
		if found < 0 {
			return fmt.Errorf("葡萄不足")
		}
		p.Grapes = append(p.Grapes[:found], p.Grapes[found+1:]...)
		count--
		id = ""
	}
	return nil
}

func (r *Room) performStructureAction(owner *Player, d StructureDef, a Action) error {
	finish := func() error { return r.structureActionReward(owner, d) }
	switch d.ID {
	case "cask":
		if len(a.WineIDs) != 1 {
			return fmt.Errorf("木桶须选择1瓶酒")
		}
		if err := ageWineByID(owner, a.WineIDs[0], 2); err != nil {
			return err
		}
		return finish()
	case "wine_cave":
		if len(a.WineIDs) < 1 || len(a.WineIDs) > 2 {
			return fmt.Errorf("酒窖洞须选择1至2瓶酒")
		}
		seen := map[string]bool{}
		trial := *owner
		trial.Wines = append([]Wine(nil), owner.Wines...)
		for _, id := range a.WineIDs {
			if seen[id] {
				return fmt.Errorf("酒窖洞不能重复选择同一瓶酒")
			}
			seen[id] = true
			if err := ageWineByID(&trial, id, 1); err != nil {
				return err
			}
		}
		owner.Wines = trial.Wines
		return finish()
	case "trading_post":
		if len(a.Trades) != 1 {
			return fmt.Errorf("贸易站须提交1次交易")
		}
		if err := r.trade(owner, a.Trades[0]); err != nil {
			return err
		}
		return finish()
	case "shop":
		a.Space = "fill_order"
		if err := r.fillOrderAction(owner, a); err != nil {
			return err
		}
		return finish()
	case "label_factory":
		if owner.Coins < 3 {
			return fmt.Errorf("酒标工厂需要3金币")
		}
		if err := r.fillOrderAction(owner, a); err != nil {
			return err
		}
		owner.Coins -= 3
		owner.VP += 2
		return nil
	case "wine_press":
		if err := makeActionWines(owner, a, 2); err != nil {
			return err
		}
		return finish()
	case "school":
		if owner.TotalWorkers >= 6 {
			return fmt.Errorf("工人上限为6名")
		}
		if (specialWorkerEnabled(r) || owner.GrandeRemoved) && a.SpecialWorker == "" && r.Context != nil {
			opts := r.specialWorkerOptionsWithDiscount(owner, 4)
			if len(opts) > 1 {
				pending := *cloneAction(a)
				pending.Space = "school_training"
				r.Context.PendingAction = &pending
				r.Context.PendingBonus = false
				r.Context.SpecialStage = "train"
				r.enqueue(owner.ID, "special_train", opts, 1)
				r.Choices[len(r.Choices)-1].ActionSpace = "school_training"
				r.TurnID = owner.ID
				return nil
			}
			if len(opts) == 1 {
				a.SpecialWorker = opts[0]
			}
		}
		if a.SpecialWorker == "" {
			a.SpecialWorker = "regular"
		}
		if err := r.trainWorkerNow(owner, a.SpecialWorker); err != nil {
			return err
		}
		return finish()
	case "wine_bar":
		if err := visitorWine(owner, a.WineIDs, 0); err != nil {
			return err
		}
		return finish()
	case "ristorante":
		trial := *owner
		trial.Wines = append([]Wine(nil), owner.Wines...)
		trial.Grapes = append([]Grape(nil), owner.Grapes...)
		if err := visitorWine(&trial, a.WineIDs, 0); err != nil {
			return err
		}
		if err := removeGrapeByID(&trial, a.GrapeID, "red_or_white", 1); err != nil {
			return err
		}
		owner.Wines, owner.Grapes = trial.Wines, trial.Grapes
		return finish()
	case "guest_house":
		if err := r.visitorCards(owner, a.CardIDs, 2, true); err != nil {
			return err
		}
		return finish()
	case "cafe":
		if err := removeGrapeByID(owner, a.GrapeID, "red_or_white", 1); err != nil {
			return err
		}
		return finish()
	case "mixer":
		if len(a.Recipes) < 1 || len(a.Recipes) > 2 {
			return fmt.Errorf("调酒器须酿造1至2瓶")
		}
		for _, recipe := range a.Recipes {
			if len(recipe) != 2 && len(recipe) != 3 {
				return fmt.Errorf("调酒器只能酿造桃红或起泡酒")
			}
			red, white := 0, 0
			for _, i := range recipe {
				if i < 0 || i >= len(owner.Grapes) {
					return fmt.Errorf("调酒器葡萄选择无效")
				}
				if owner.Grapes[i].Color == "red" {
					red++
				} else if owner.Grapes[i].Color == "white" {
					white++
				} else {
					return fmt.Errorf("调酒器只能使用红或白葡萄")
				}
			}
			if !((red == 1 && white == 1) || (red == 2 && white == 1)) {
				return fmt.Errorf("调酒器只能酿造桃红或起泡酒")
			}
		}
		trial := probePlayer(owner)
		before := len(trial.Wines)
		if err := makeActionWines(trial, a, 2); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, wine := range trial.Wines[before:] {
			if seen[wine.Type] {
				return fmt.Errorf("调酒器至多酿1瓶桃红和1瓶起泡酒")
			}
			seen[wine.Type] = true
		}
		owner.Grapes, owner.Wines, owner.Coins, owner.VP = trial.Grapes, trial.Wines, trial.Coins, trial.VP
		return finish()
	default:
		return fmt.Errorf("结构%s没有私人行动", d.Name)
	}
}

func (r *Room) payAcademyToll(p *Player) error {
	toll := 0
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			toll++
		}
	}
	if p.Coins < toll {
		return fmt.Errorf("学院费用需要%d金币", toll)
	}
	for _, q := range r.Players {
		if q.ID != p.ID && has(q, "academy") {
			p.Coins--
			q.Coins++
		}
	}
	return nil
}

// structureActionReward is the printed middle reward on an action structure.
// Influence rewards are optional continuations because the EE board has no
// influence map; they are only offered when the Tuscany board is active.
func (r *Room) structureActionReward(p *Player, d StructureDef) error {
	switch d.ID {
	case "cask", "wine_cave":
		r.draw(p, "order")
	case "trading_post", "shop", "wine_press":
		if r.tuscany() {
			r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: p.ID, Kind: "structure_influence", Options: []string{"place", "skip"}, Count: 1, Labels: map[string]string{"place": "放置或移动1颗影响力星", "skip": "跳过结构奖励"}})
		}
	case "school":
		p.Coins++
	case "wine_bar":
		p.VP += 2
	case "ristorante":
		p.Coins += 3
		p.VP += 3
	case "guest_house":
		p.VP += 2
	case "cafe":
		p.Coins += 3
		p.VP++
	case "mixer":
		p.VP++
	}
	return nil
}

func firstWineRecipe(p *Player) []int {
	for i := range p.Grapes {
		if len([]int{i}) > 0 {
			q := probePlayer(p)
			if makeWine(q, []int{i}) == nil {
				return []int{i}
			}
		}
	}
	for i := range p.Grapes {
		for j := i + 1; j < len(p.Grapes); j++ {
			q := probePlayer(p)
			if makeWine(q, []int{i, j}) == nil {
				return []int{i, j}
			}
			for k := j + 1; k < len(p.Grapes); k++ {
				q = probePlayer(p)
				if makeWine(q, []int{i, j, k}) == nil {
					return []int{i, j, k}
				}
			}
		}
	}
	return nil
}

func (r *Room) queueStructureVisitorBonus(p *Player) bool {
	if !p.hasStructure("tap_room") && !p.hasStructure("tavern") {
		return false
	}
	opts := []string{"skip"}
	hasWine := p.hasStructure("tap_room") && len(p.Wines) > 0
	eligibleGrapes := 0
	for _, g := range p.Grapes {
		if g.Color == "red" || g.Color == "white" {
			eligibleGrapes++
		}
	}
	hasGrapes := p.hasStructure("tavern") && eligibleGrapes >= 2
	if hasWine {
		opts = append(opts, "tap_room")
	}
	if hasGrapes {
		opts = append(opts, "tavern")
	}
	if hasWine && hasGrapes {
		opts = append(opts, "tap_room+tavern")
	}
	if len(opts) == 1 {
		return false
	}
	schema := []ChoiceField{}
	if hasWine {
		schema = append(schema, ChoiceField{Name: "wineIds", Type: "wines", Min: 1, Max: 1})
	}
	if hasGrapes {
		schema = append(schema, ChoiceField{Name: "grapeIds", Type: "grapes", Min: 2, Max: 2})
	}
	r.Choices = append(r.Choices, Choice{ID: NewID(), PlayerID: p.ID, Kind: "structure_visitor_bonus", Options: opts, Count: 1, Schema: schema, Labels: map[string]string{"skip": "不使用访客奖励", "tap_room": "弃1瓶酒得2分", "tavern": "弃2颗红/白葡萄得3分", "tap_room+tavern": "两项都执行"}})
	return true
}

func (r *Room) resolveStructureVisitorBonus(p *Player, option string, a Action) error {
	if option == "skip" {
		return nil
	}
	// Resolve on a copy so a malformed combined selection cannot consume the
	// wine before the grape part fails. IDs are required for both resources;
	// silently taking the first items would violate the hand/resource picker
	// contract after reconnects or reordered views.
	trial := *p
	trial.Wines = append([]Wine(nil), p.Wines...)
	trial.Grapes = append([]Grape(nil), p.Grapes...)
	if option == "tap_room" || option == "tap_room+tavern" {
		if !p.hasStructure("tap_room") {
			return fmt.Errorf("尚未建造品酒吧")
		}
		if err := visitorWine(&trial, a.WineIDs, 0); err != nil {
			return err
		}
		trial.VP += 2
	}
	if option == "tavern" || option == "tap_room+tavern" {
		if !p.hasStructure("tavern") {
			return fmt.Errorf("尚未建造酒馆")
		}
		if len(a.GrapeIDs) != 2 || a.GrapeIDs[0] == a.GrapeIDs[1] {
			return fmt.Errorf("酒馆需要选择2颗不同的红/白葡萄")
		}
		for _, id := range a.GrapeIDs {
			if err := removeGrapeByID(&trial, id, "red_or_white", 1); err != nil {
				return err
			}
		}
		trial.VP += 3
	}
	p.Wines, p.Grapes, p.VP = trial.Wines, trial.Grapes, trial.VP
	return nil
}

// One harvest effect can cover several fields; its enhancement is offered once.
// The choice owns its resources and does not overwrite an outer worker/visitor action.
func (r *Room) queueFermentation(p *Player) {
	if !r.Config.Structures || r.Context == nil || !has(p, "fermentation_tank") {
		return
	}
	for _, c := range r.Choices {
		if c.Kind == "structure_fermentation" && c.PlayerID == p.ID {
			return
		}
	}
	r.enqueue(p.ID, "structure_fermentation", []string{"make", "skip"}, 1)
}
