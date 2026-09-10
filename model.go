package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Card struct {
	Trellis      bool   `json:"trellis"`
	Irrigation   bool   `json:"irrigation"`
	Implemented  bool   `json:"implemented"`
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Red          int    `json:"red"`
	White        int    `json:"white"`
	Requirements []Wine `json:"requirements"`
	Points       int    `json:"points"`
	Income       int    `json:"income"`
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
	Sold      bool   `json:"sold"`
	Index     int    `json:"index"`
	Capacity  int    `json:"capacity"`
	Vines     []Card `json:"vines"`
	Harvested bool   `json:"harvested"`
}
type Player struct {
	Mama         Card     `json:"mama"`
	Papa         Card     `json:"papa"`
	PapaResolved bool     `json:"papaResolved"`
	YokeUsed     bool     `json:"yokeUsed"`
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Coins        int      `json:"coins"`
	VP           int      `json:"vp"`
	Income       int      `json:"income"`
	Workers      int      `json:"workers"`
	LargeWorker  bool     `json:"largeWorker"`
	TotalWorkers int      `json:"totalWorkers"`
	HandCount    int      `json:"handCount"`
	Fields       []Field  `json:"fields"`
	Buildings    []string `json:"buildings"`
	Grapes       []Grape  `json:"grapes"`
	Wines        []Wine   `json:"wines"`
	Hand         []Card   `json:"hand,omitempty"`
	Passed       bool     `json:"passed"`
	Wake         int      `json:"wake"`
	Trained      int      `json:"trained"`
	MillUsed     bool     `json:"millUsed"`
	TastingUsed  bool     `json:"tastingUsed"`
}
type Seat struct {
	Slot     int    `json:"slot"`
	PlayerID string `json:"playerId"`
	Large    bool   `json:"large"`
}
type Space struct {
	Bonus          string `json:"bonus"`
	DisabledReason string `json:"disabledReason"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Season         string `json:"season"`
	Occupied       []Seat `json:"occupied"`
	Capacity       int    `json:"capacity"`
}
type WakeSlot struct {
	Slot     int    `json:"slot"`
	PlayerID string `json:"playerId"`
	Bonus    string `json:"bonus"`
}
type Room struct {
	GrayWorkerOwner string             `json:"grayWorkerOwner"`
	Planned         []PlannedPlacement `json:"planned,omitempty"`
	Revealed        []Card             `json:"revealed,omitempty"`
	Ruleset         string             `json:"ruleset"`
	Decks           map[string][]Card  `json:"decks"`
	Discards        map[string][]Card  `json:"discards"`
	Choices         []Choice           `json:"choices"`
	Context         *ActionContext     `json:"context"`
	Resume          string             `json:"resume"`
	WinnerIDs       []string           `json:"winnerIds"`
	Revision        int                `json:"revision"`
	Code            string             `json:"code"`
	Phase           string             `json:"phase"`
	Year            int                `json:"year"`
	TurnID          string             `json:"turnId"`
	HostID          string             `json:"hostId"`
	Players         []*Player          `json:"players"`
	Spaces          []Space            `json:"spaces"`
	WakeSlots       []WakeSlot         `json:"wakeSlots"`
	Log             []string           `json:"log"`
	WinnerID        string             `json:"winnerId"`
	SpringLeader    int                `json:"springLeader"`
	FinalYear       bool               `json:"finalYear"`
	Updated         time.Time          `json:"updated"`
}
type Session struct {
	Code     string
	PlayerID string `json:"playerId"`
}
type Store struct {
	Rooms    map[string]*Room
	Sessions map[string]Session
}
type App struct {
	mu    sync.Mutex
	Store Store
	dir   string
	subs  map[chan struct{}]string
}
type Action struct {
	DeclineBonus bool     `json:"declineBonus"`
	TargetIDs    []string `json:"targetIds"`
	Buildings    []string `json:"buildings"`
	Colors       []string `json:"colors"`
	ChoiceID     string   `json:"choiceId"`
	Option       string   `json:"option"`
	CardIDs      []string `json:"cardIds"`
	Color        string   `json:"color"`
	Mode         string   `json:"mode"`
	Recipes      [][]int  `json:"recipes"`
	Fields       []int    `json:"fields"`
	Revision     int      `json:"revision"`
	Type         string   `json:"type"`
	Slot         int      `json:"slot"`
	Space        string   `json:"space"`
	Large        bool     `json:"large"`
	CardID       string   `json:"cardId"`
	Field        int      `json:"field"`
	Building     string   `json:"building"`
	Grapes       []int    `json:"grapes"`
	WineIDs      []string `json:"wineIds"`
}

var notes = []string{"EE 本体全卡验收版：2–6 人，76 张访客真实效果、多人响应、预约行动已接入；非官方认证。", "Planner 多个预约按预约先后执行；Organizer 在外层双访客结算后 pass。两项为本实现公开约定，特殊组合仍需进一步独立核对。", "固定藤42、订单36、Mama/Papa各18；有限牌堆、弃牌重洗和父母开局。原创主题插画部分共用。", "不支持 Tuscany、Automa 或旧预览存档迁移。仅用于可信局域网，不要公开存档或会话令牌。"}

func uid() string {
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
func loadApp(dir string) (*App, error) {
	a := &App{dir: dir, subs: make(map[chan struct{}]string), Store: Store{Rooms: map[string]*Room{}, Sessions: map[string]Session{}}}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	b, e := os.ReadFile(filepath.Join(dir, "ee-state-v1.json"))
	if e == nil {
		e = json.Unmarshal(b, &a.Store)
	}
	if e != nil && !os.IsNotExist(e) {
		return nil, e
	}
	if a.Store.Rooms == nil || a.Store.Sessions == nil {
		return nil, fmt.Errorf("invalid save")
	}
	return a, nil
}
func (a *App) save() error {
	b, e := json.Marshal(a.Store)
	if e != nil {
		return e
	}
	p := filepath.Join(a.dir, "ee-state-v1.json")
	if e = os.WriteFile(p+".tmp", b, 0600); e != nil {
		return e
	}
	return os.Rename(p+".tmp", p)
}
func (a *App) notify(code string) {
	for c, r := range a.subs {
		if r == code {
			select {
			case c <- struct{}{}:
			default:
			}
		}
	}
}
func (r *Room) player(id string) *Player {
	for _, p := range r.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}
func (r *Room) log(s string) {
	r.Log = append(r.Log, s)
	if len(r.Log) > 120 {
		r.Log = r.Log[len(r.Log)-120:]
	}
	r.Updated = time.Now()
}
func has(p *Player, b string) bool {
	for _, s := range p.Buildings {
		if s == b {
			return true
		}
	}
	return false
}
func makePlayer(name string) *Player {
	p := &Player{ID: uid(), Name: name, Coins: 0, Workers: 2, LargeWorker: true, TotalWorkers: 3, Buildings: []string{}, Grapes: []Grape{}, Wines: []Wine{}, Hand: []Card{}}
	for i := 0; i < 3; i++ {
		p.Fields = append(p.Fields, Field{Index: i, Capacity: 5 + i, Vines: []Card{}})
	}
	return p
}

// Retained only to compile legacy preview tests. Live EE draws are Room-scoped.
func draw(p *Player, t string) { panic("legacy draw disabled: use Room.draw") }
func spaces(n int) []Space {
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
func (r *Room) view(id string) map[string]any {
	ps := make([]Player, 0)
	for _, p := range r.Players {
		q := *p
		q.Hand = nil
		q.HandCount = len(p.Hand)
		ps = append(ps, q)
	}
	p := r.player(id)
	hand := []Card{}
	if p != nil {
		hand = p.Hand
	}
	active := r.TurnID == id && len(r.Choices) == 0
	return map[string]any{"grayWorkerOwner": r.GrayWorkerOwner, "triggerPlacement": r.publicTriggerPlacement(), "planned": r.Planned, "revealed": r.Revealed, "ruleset": r.Ruleset, "parentOptions": r.parentOptions(id), "pendingChoice": r.choiceView(id), "pendingCount": len(r.Choices), "deckCounts": r.deckCounts(), "winnerIds": r.WinnerIDs, "revision": r.Revision, "code": r.Code, "phase": r.Phase, "year": r.Year, "turnId": r.TurnID, "hostId": r.HostID, "youId": id, "players": ps, "hand": hand, "spaces": r.Spaces, "wakeSlots": r.WakeSlots, "log": r.Log, "winnerId": r.WinnerID, "rulesNotes": notes, "legal": map[string]bool{"canChoose": len(r.Choices) > 0 && r.Choices[0].PlayerID == id, "canStart": r.Phase == "lobby" && id == r.HostID && len(ps) >= 2, "canWake": r.Phase == "wake" && active, "canPlace": (r.Phase == "summer" || r.Phase == "winter") && active, "canPass": (r.Phase == "summer" || r.Phase == "winter") && active}}
}
