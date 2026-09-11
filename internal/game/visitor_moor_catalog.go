package game

import "strings"

const moorSource = "https://stonemaiergames.com/games/viticulture/moor-visitors-expansion/"

type moorDefinition struct {
	ID, Season, English, Name, Summary string
	Options                            []string
}

var moorDefinitions = []moorDefinition{
	{"moor-summer-01", "summer", "Coachman", "马车夫", "免费建造轭架；或已有轭架时抽2张冬季访客与1张订单。", []string{"build", "draw"}},
	{"moor-summer-02", "summer", "Councilman", "议员", "抽1张订单，然后打出1张冬季访客。", []string{"play"}},
	{"moor-summer-03", "summer", "Cultivator", "培育者", "获得1颗价值2葡萄；或弃掉3颗葡萄获得3分。", []string{"grape", "discard"}},
	{"moor-summer-04", "summer", "Fruit Dealer", "水果商", "付4金币将此牌放在田地上；每次收获该田地，获得2金币或1分。", []string{"attach"}},
	{"moor-summer-05", "summer", "Fusel Maker", "杂醇制造者", "每位玩家可获得1颗价值2葡萄；若至少1位对手接受，你获得2分。", []string{"offer"}},
	{"moor-summer-06", "summer", "Grand Sponsor", "大赞助人", "获得7金币或2分。保留此牌：终局收入不足4时失去4分。", []string{"coins", "vp"}},
	{"moor-summer-07", "summer", "Grape Juice Drinker", "葡萄汁饮者", "获得等同起床行号的金币；或弃2颗葡萄获得2分。", []string{"coins", "discard"}},
	{"moor-summer-08", "summer", "Grape Thief", "偷葡萄的人", "为自己收获1块对手田地；对手今年仍可收获该田地。", []string{"harvest"}},
	{"moor-summer-09", "summer", "Guest of Honor", "贵宾", "抽1张订单；若以大工人打出此牌，再获得1分与2金币。", []string{"draw"}},
	{"moor-summer-10", "summer", "Hotelier", "旅馆经营者", "抽4张冬季访客或获得2分。保留此牌：终局葡萄不足3颗时失去4分。", []string{"draw", "vp"}},
	{"moor-summer-11", "summer", "Irrigationist", "灌溉师", "免费建造灌溉设施；或种植1张需要灌溉的葡萄藤并获得1分。", []string{"build", "plant"}},
	{"moor-summer-12", "summer", "Landlord", "地主", "付14金币建造3座建筑；或已有至少6座建筑时获得2分。", []string{"build", "vp"}},
	{"moor-summer-13", "summer", "Mill Operator", "风车操作员", "已有风车时获得1颗价值3葡萄；否则付2金币建造风车。", []string{"grape", "build"}},
	{"moor-summer-14", "summer", "Premium Vintner", "优质酿酒师", "已有至少5颗葡萄时获得2分；或付3金币获得1颗价值5葡萄。", []string{"vp", "grape"}},
	{"moor-summer-15", "summer", "Principal", "校长", "抽3张订单，选择完成其中1张，或保留其中2张并弃掉另1张。", []string{"draw"}},
	{"moor-summer-16", "summer", "Recruiter", "招募员", "按尚未培训的工人数付费，培训1名工人，本年立即可用。", []string{"train"}},
	{"moor-summer-17", "summer", "Restaurateur", "餐馆老板", "已有品酒室时获得1瓶价值1红酒或白酒；否则付3金币建造品酒室。", []string{"wine", "build"}},
	{"moor-summer-18", "summer", "Tourists", "游客", "弃X张手牌，收获X块不同田地，或陈酿X颗不同葡萄各1次。", []string{"harvest", "age"}},
	{"moor-summer-19", "summer", "Wholesale Buyer", "批发买家", "完成1张订单；若以最后1名可用工人打出此牌，额外获得2分。", []string{"fill"}},
	{"moor-summer-20", "summer", "Wine Supplier", "葡萄酒供应商", "付3金币，升级酒窖1次或获得1瓶价值1白酒。", []string{"upgrade", "wine"}},
	{"moor-winter-01", "winter", "Baron", "男爵", "取回2名自己放在先前季节的工人，本年可以再次使用。", []string{"retrieve"}},
	{"moor-winter-02", "winter", "Blush Wine Dealer", "桃红酒商", "付7金币获得1瓶价值4桃红酒，无需中酒窖。", []string{"wine"}},
	{"moor-winter-03", "winter", "Cellar Dweller", "酒窖居民", "陈酿1瓶酒至多3次；或付2金币升级酒窖；或另付1分完成两项。", []string{"age", "upgrade", "both"}},
	{"moor-winter-04", "winter", "Cheesemonger", "奶酪商", "按当前酒的瓶数增加收入；或付9金币获得3分。", []string{"income", "vp"}},
	{"moor-winter-05", "winter", "Collector", "收藏家", "抽2张订单；或按酒窖中的酒类数获得分数，最多4分。", []string{"draw", "vp"}},
	{"moor-winter-06", "winter", "Counterfeiter", "造假者", "获得3金币；或完成1张订单，仅忽略酒的种类要求。", []string{"coins", "fill"}},
	{"moor-winter-07", "winter", "Draftsman", "绘图师", "优惠5金币建造1座建筑，然后将此牌传给左邻座。", []string{"build"}},
	{"moor-winter-08", "winter", "Exotic Fruit Dealer", "异国水果商", "付X金币，获得1颗价值X的红葡萄或白葡萄。", []string{"buy_1", "buy_2", "buy_3", "buy_4", "buy_5", "buy_6", "buy_7", "buy_8", "buy_9"}},
	{"moor-winter-09", "winter", "Fiance", "未婚夫", "获得3金币；或已有小屋时弃1颗葡萄与1瓶酒，获得4分。", []string{"coins", "discard"}},
	{"moor-winter-10", "winter", "Grape Ice Inventor", "葡萄冰发明家", "冬季付1金币获得1颗价值4葡萄；也可作为夏季访客打出，弃1颗葡萄获得4金币。", []string{"resolve"}},
	{"moor-winter-11", "winter", "Grape Juice Presser", "葡萄汁压榨师", "弃1颗葡萄：获得等同价值的金币，或等同售出价格的分数。", []string{"coins", "vp"}},
	{"moor-winter-12", "winter", "Official", "官员", "抽1张葡萄藤，然后打出1张夏季访客。", []string{"play"}},
	{"moor-winter-13", "winter", "Organic Vintner", "有机酿酒师", "种植1张葡萄藤；或将3瓶不同的酒各降级1，获得2分。", []string{"plant", "devalue"}},
	{"moor-winter-14", "winter", "Sparkling Wine Dealer", "起泡酒商", "付9金币获得1瓶价值7起泡酒，无需大酒窖。", []string{"wine"}},
	{"moor-winter-15", "winter", "Trader", "交易商", "弃3张手牌，再按预先选择的种类抽3张牌，并获得2金币。", []string{"trade"}},
	{"moor-winter-16", "winter", "Truss Harvester", "整串收获者", "收获X块田地，然后陈酿X颗不同葡萄各1次。", []string{"harvest"}},
	{"moor-winter-17", "winter", "Varietal Expert", "品种专家", "选择抽3张葡萄藤或3张订单；看牌后保留1张，或付1金币保留全部。", []string{"vine", "order"}},
	{"moor-winter-18", "winter", "Wine Dealer", "葡萄酒商", "付9金币，获得1瓶价值2红酒与1瓶价值2白酒。", []string{"wine"}},
	{"moor-winter-19", "winter", "Wine Enthusiast", "葡萄酒爱好者", "弃1瓶酒获得2分；或失去1分培训1名工人。", []string{"discard", "train"}},
	{"moor-winter-20", "winter", "Wine Technologist", "葡萄酒技术员", "付3金币，获得1瓶价值1红酒或白酒，或陈酿1颗葡萄至多3次。", []string{"wine", "age"}},
}

func init() {
	for _, d := range moorDefinitions {
		visitorDefs[d.ID] = visitorDef{d.Summary, d.Options}
	}
}
func MoorCatalog() []Card {
	cards := make([]Card, 0, 40)
	for _, d := range moorDefinitions {
		cards = append(cards, Card{ID: d.ID, Type: d.Season, Name: d.Name + " · " + d.English, Description: d.Summary, Implemented: true, RuleSource: moorSource})
	}
	return cards
}
func isMoor(id string) bool { return strings.HasPrefix(id, "moor-") }
func visitorInSeason(p *Player, id, season string) bool {
	return cardIndex(p, id, season) >= 0 || (season == "summer" && id == "moor-winter-10" && cardIndex(p, id, "winter") >= 0)
}
func moorHeldCard(id string) bool {
	return id == "moor-summer-04" || id == "moor-summer-06" || id == "moor-summer-10" || id == "moor-winter-07"
}
func moorCard(id string) Card {
	for _, c := range MoorCatalog() {
		if c.ID == id {
			return c
		}
	}
	return Card{}
}
