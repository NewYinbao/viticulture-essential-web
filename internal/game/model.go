package game

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"time"
)

type Card struct {
	Trellis           bool   `json:"trellis"`
	Irrigation        bool   `json:"irrigation"`
	Implemented       bool   `json:"implemented"`
	ID                string `json:"id"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Red               int    `json:"red"`
	White             int    `json:"white"`
	Requirements      []Wine `json:"requirements"`
	Points            int    `json:"points"`
	Income            int    `json:"income"`
	StructureID       string `json:"structureId,omitempty"`
	StructureCost     int    `json:"structureCost,omitempty"`
	StructureWorker   bool   `json:"structureWorker,omitempty"`
	StructureCategory string `json:"structureCategory,omitempty"`
	RuleSource        string `json:"ruleSource,omitempty"`
}
type SpecialWorker struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RuleSource  string `json:"ruleSource"`
}
type Grape struct {
	ID    string `json:"id"`
	Color string `json:"color"`
	Value int    `json:"value"`
}
type Wine struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Value int    `json:"value"`
}
type Field struct {
	FruitDealer bool   `json:"fruitDealer,omitempty"`
	Sold        bool   `json:"sold"`
	Index       int    `json:"index"`
	Capacity    int    `json:"capacity"`
	Vines       []Card `json:"vines"`
	Harvested   bool   `json:"harvested"`
	Structure   string `json:"structure,omitempty"`
}
type Player struct {
	GrandeRemoved         bool            `json:"grandeRemoved,omitempty"`
	GrayWorkerAvailable   bool            `json:"grayWorkerAvailable,omitempty"`
	RhineSonField         *int            `json:"rhineSonField,omitempty"`
	RhineWinterActionYear int             `json:"rhineWinterActionYear,omitempty"`
	RhineWinterOfferYear  int             `json:"rhineWinterOfferYear,omitempty"`
	MoorContracts         []Card          `json:"moorContracts,omitempty"`
	Season                string          `json:"season,omitempty"`
	NextWake              int             `json:"nextWake,omitempty"`
	Influence             map[string]int  `json:"influence,omitempty"`
	Mama                  Card            `json:"mama"`
	Papa                  Card            `json:"papa"`
	PapaResolved          bool            `json:"papaResolved"`
	YokeUsed              bool            `json:"yokeUsed"`
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Coins                 int             `json:"coins"`
	VP                    int             `json:"vp"`
	Income                int             `json:"income"`
	Workers               int             `json:"workers"`
	LargeWorker           bool            `json:"largeWorker"`
	TotalWorkers          int             `json:"totalWorkers"`
	HandCount             int             `json:"handCount"`
	HandCounts            map[string]int  `json:"handCounts,omitempty"`
	Fields                []Field         `json:"fields"`
	Buildings             []string        `json:"buildings"`
	StructureSlots        []string        `json:"structureSlots,omitempty"`
	Grapes                []Grape         `json:"grapes"`
	Wines                 []Wine          `json:"wines"`
	Hand                  []Card          `json:"hand,omitempty"`
	Passed                bool            `json:"passed"`
	Wake                  int             `json:"wake"`
	Trained               int             `json:"trained"`
	MillUsed              bool            `json:"millUsed"`
	TastingUsed           bool            `json:"tastingUsed"`
	SpecialWorkers        []string        `json:"specialWorkers,omitempty"`
	SpecialWorkerUsed     map[string]bool `json:"specialWorkerUsed,omitempty"`
	SpecialWorkerReady    map[string]int  `json:"specialWorkerReady,omitempty"`
}
type Seat struct {
	Gray       bool   `json:"gray,omitempty"`
	Season     string `json:"season,omitempty"`
	Slot       int    `json:"slot"`
	PlayerID   string `json:"playerId"`
	Large      bool   `json:"large"`
	WorkerType string `json:"workerType,omitempty"`
}
type Space struct {
	BonusSlots     map[int]string `json:"bonusSlots,omitempty"`
	Bonus          string         `json:"bonus"`
	DisabledReason string         `json:"disabledReason"`
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Season         string         `json:"season"`
	Occupied       []Seat         `json:"occupied"`
	Capacity       int            `json:"capacity"`
}
type WakeSlot struct {
	Slot     int    `json:"slot"`
	PlayerID string `json:"playerId"`
	Bonus    string `json:"bonus"`
}
type Room struct {
	GrayWorkerTracked bool               `json:"grayWorkerTracked,omitempty"`
	FirstPlayerID     string             `json:"firstPlayerId,omitempty"`
	Config            ExpansionConfig    `json:"config"`
	GrayWorkerOwner   string             `json:"grayWorkerOwner"`
	Planned           []PlannedPlacement `json:"planned,omitempty"`
	Revealed          []Card             `json:"revealed,omitempty"`
	Ruleset           string             `json:"ruleset"`
	Decks             map[string][]Card  `json:"decks"`
	Discards          map[string][]Card  `json:"discards"`
	Choices           []Choice           `json:"choices"`
	Context           *ActionContext     `json:"context"`
	Resume            string             `json:"resume"`
	WinnerIDs         []string           `json:"winnerIds"`
	Revision          int                `json:"revision"`
	Code              string             `json:"code"`
	Phase             string             `json:"phase"`
	Year              int                `json:"year"`
	TurnID            string             `json:"turnId"`
	HostID            string             `json:"hostId"`
	Players           []*Player          `json:"players"`
	Spaces            []Space            `json:"spaces"`
	WakeSlots         []WakeSlot         `json:"wakeSlots"`
	Log               []string           `json:"log"`
	WinnerID          string             `json:"winnerId"`
	SpringLeader      int                `json:"springLeader"`
	FinalYear         bool               `json:"finalYear"`
	SpecialWorkerPool []string           `json:"specialWorkerPool,omitempty"`
	MessengerPlans    []PlannedPlacement `json:"messengerPlans,omitempty"`
	// StructureDrafts is private, persisted setup state for EE + structures.
	// Each player sees only the current pool through pendingChoice.Options.
	StructureDrafts     map[string][]Card `json:"structureDrafts,omitempty"`
	StructureDraftRound int               `json:"structureDraftRound,omitempty"`
	Updated             time.Time         `json:"updated"`
}
type Action struct {
	WorkerID      string           `json:"workerId,omitempty"`
	Gray          bool             `json:"gray,omitempty"`
	RhineStars    []RhineStarMove  `json:"rhineStars,omitempty"`
	BonusFirst    bool             `json:"bonusFirst,omitempty"`
	Influence     []InfluenceMove  `json:"influence,omitempty"`
	Trades        []Trade          `json:"trades,omitempty"`
	Config        *ExpansionConfig `json:"config,omitempty"`
	DeclineBonus  bool             `json:"declineBonus"`
	TargetIDs     []string         `json:"targetIds"`
	Buildings     []string         `json:"buildings"`
	Colors        []string         `json:"colors"`
	ChoiceID      string           `json:"choiceId"`
	Option        string           `json:"option"`
	CardIDs       []string         `json:"cardIds"`
	Color         string           `json:"color"`
	Mode          string           `json:"mode"`
	RecipeTypes   []string         `json:"recipeTypes,omitempty"`
	HarvestAll    bool             `json:"harvestAll,omitempty"`
	Recipes       [][]int          `json:"recipes"`
	Fields        []int            `json:"fields"`
	Revision      int              `json:"revision"`
	Type          string           `json:"type"`
	Slot          int              `json:"slot"`
	Space         string           `json:"space"`
	Large         bool             `json:"large"`
	CardID        string           `json:"cardId"`
	Field         int              `json:"field"`
	Building      string           `json:"building"`
	Grapes        []int            `json:"grapes"`
	GrapeID       string           `json:"grapeId,omitempty"`
	GrapeIDs      []string         `json:"grapeIds,omitempty"`
	WineIDs       []string         `json:"wineIds"`
	WorkerType    string           `json:"workerType,omitempty"`
	SpecialWorker string           `json:"specialWorker,omitempty"`
	BonusOverride string           `json:"bonusOverride,omitempty"`
	StructureSlot int              `json:"structureSlot,omitempty"`
}

var notes = []string{"EE 本体全卡验收版：2–6 人，76 张访客真实效果、多人响应、预约行动已接入；非官方认证。", "EE/Tuscany 主板、36张结构牌、11种特殊工人及 EE、EE+Moor、Rhine 访客牌组可按开局配置组合。", "固定藤42、订单36、Mama/Papa各18；所有牌堆有限，抽空后只重洗对应弃牌堆。Rhine 替换其他访客，Tuscany 主板额外加入4张专用 Rhine 牌。原创主题插画部分共用。", "仅支持本项目可审计规则集；Automa、旧预览存档迁移不在当前可玩范围。仅用于可信局域网，不要公开存档或会话令牌。"}

func NewID() string {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func rnd(n int) int {
	x, e := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if e != nil {
		panic(e)
	}
	return int(x.Int64())
}
func (r *Room) Player(id string) *Player {
	for _, p := range r.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}
func (r *Room) AddLog(s string) {
	r.Log = append(r.Log, s)
	if len(r.Log) > 120 {
		r.Log = r.Log[len(r.Log)-120:]
	}
	r.Updated = time.Now()
}
func has(p *Player, b string) bool {
	if p.hasStructure(b) {
		return true
	}
	for _, s := range p.Buildings {
		if s == b {
			return true
		}
	}
	return false
}
func NewPlayer(name string) *Player {
	p := &Player{ID: NewID(), Name: name, Coins: 0, Workers: 2, LargeWorker: true, TotalWorkers: 3, Buildings: []string{}, Grapes: []Grape{}, Wines: []Wine{}, Hand: []Card{}}
	for i := 0; i < 3; i++ {
		p.Fields = append(p.Fields, Field{Index: i, Capacity: 5 + i, Vines: []Card{}})
	}
	return p
}

// Retained only to compile legacy preview tests. Live EE draws are Room-scoped.
func draw(p *Player, t string) { panic("legacy draw disabled: use Room.draw") }
func NewSpaces(n int) []Space {
	cap := 1
	if n >= 3 {
		cap = 2
	}
	if n >= 5 {
		cap = 3
	}
	defs := [][4]string{{"draw_vine", "葡萄藤市场", "抽取1张葡萄藤", "summer"}, {"tour", "酒庄导览", "获得2金币；已建品酒室且有酒时每年额外1分", "summer"}, {"build", "建造建筑", "支付金币建造设施", "summer"}, {"plant", "种植葡萄藤", "选择手牌与田地，须满足容量及设施", "summer"}, {"summer_visitor", "夏季访客", "打出1张夏季访客卡", "summer"}, {"sell_grapes", "出售葡萄", "按品质区间获得金币", "summer"}, {"draw_order", "获取订单", "抽取1张订单", "winter"}, {"harvest", "收获田地", "每块田地每年最多收获一次", "winter"}, {"make_wine", "酿造葡萄酒", "红/白用1颗，桃红用红白各1，起泡用2红1白", "winter"}, {"fill_order", "交付订单", "选择订单与酒，领取分数和永久收入", "winter"}, {"train", "培训工人", "支付4金币，新工人下一年可用，上限6人", "winter"}, {"winter_visitor", "冬季访客", "打出1张冬季访客卡", "winter"}}
	out := []Space{}
	for _, d := range defs {
		out = append(out, Space{ID: d[0], Name: d[1], Description: d[2], Season: d[3], Capacity: cap, Occupied: []Seat{}})
	}
	for i := range out {
		out[i].Bonus = spaceBonuses[out[i].ID]

	}
	out = append(out, Space{ID: "gain_coin", Name: "获得1金币（不限人数）", Season: "any", Capacity: 0}, Space{ID: "yoke", Name: "轭：收获或拔藤", Season: "any", Capacity: 1})
	return out
}
