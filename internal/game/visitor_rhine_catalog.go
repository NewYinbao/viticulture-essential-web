package game

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed cards/rhine_cards.json
var rhineJSON []byte
var rhineCardFacts []rhineDefinition

func init() {
	var cards []struct {
		ID, Type, Name, Description, RuleSource string
		Board                                   bool
	}
	if err := json.Unmarshal(rhineJSON, &cards); err != nil {
		panic(err)
	}
	for _, c := range cards {
		rhineCardFacts = append(rhineCardFacts, rhineDefinition{Card: Card{ID: c.ID, Type: c.Type, Name: c.Name, Description: c.Description, RuleSource: c.RuleSource}, Board: c.Board})
	}
	for _, line := range strings.Split(strings.TrimSpace(rhinePrograms+"\n"+rhineMorePrograms), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			panic("bad Rhine program: " + line)
		}
		id, option, label := parts[0], parts[1], parts[2]
		d, ok := rhineDefinitions[id]
		if !ok {
			for _, f := range rhineCardFacts {
				if f.Card.ID == id {
					d = f
					ok = true
					break
				}
			}
		}
		if !ok {
			panic("missing Rhine fact " + id)
		}
		if d.Options == nil {
			d.Options = map[string]rhineOption{}
		}
		var ops []RhineOperation
		for _, token := range strings.Split(parts[3], ";") {
			p := strings.Split(token, ":")
			op := RhineOperation{Kind: p[0]}
			if len(p) > 1 {
				op.Amount, _ = strconv.Atoi(p[1])
			}
			if len(p) > 2 {
				op.Target = p[2]
			}
			ops = append(ops, op)
		}
		d.Options[option] = rhineOption{Label: label, Operations: ops}
		rhineDefinitions[id] = d
	}
	rhineRegisterPairs()
	for id, d := range rhineDefinitions {
		options := []string{}
		for key := range d.Options {
			options = append(options, key)
		}
		sort.Strings(options)
		visitorDefs[id] = visitorDef{Summary: d.Card.Description, Options: options}
	}
}

// Each line is one complete printed branch. Order is significant: costs precede
// rewards, and a following resource selection becomes a persisted continuation.
const rhinePrograms = `
rhine-summer-agent|discard|弃2手牌，得5金币|discard_cards:2;coins:5
rhine-summer-agent|winter|付2金币，抽2冬访客|pay:2;draw:2:winter
rhine-summer-agent|vine|付2金币，抽2藤|pay:2;draw:2:vine
rhine-summer-botanist|grape|得价值2葡萄|gain_grape:2
rhine-summer-botanist|draw|弃1葡萄，抽4藤|discard_grape:1;draw:4:vine
rhine-summer-cicerone|coins|得4金币|coins:4
rhine-summer-cicerone|harvest|收获1田|harvest:1
rhine-summer-cicerone|draw|弃1葡萄，抽3订单|discard_grape:1;draw:3:order
rhine-summer-docent|coins|得3金币|coins:3
rhine-summer-docent|make|酿至多3酒|make:3
rhine-summer-embezzler|vp_coins|失2分，得6金币|lose_vp:2;coins:6
rhine-summer-embezzler|vp_cards|失2分，抽3订单|lose_vp:2;draw:3:order
rhine-summer-embezzler|cards_coins|弃3手牌，得6金币|discard_cards:3;coins:6
rhine-summer-embezzler|cards_orders|弃3手牌，抽3订单|discard_cards:3;draw:3:order
rhine-summer-fortuneTeller|draw|抽2冬访客，检查额外订单|draw:2:winter;if_opponent_vp_draw:5:order
rhine-summer-freelancer|draw|抽1任意类型牌|draw_any:1
rhine-summer-freelancer|build|失2分，免费建1建筑|lose_vp:2;build:99
rhine-summer-friendlyHelper|upgrade|酒窖升一级，减3金币|upgrade:3
rhine-summer-friendlyHelper|wine|得价值1的酒|gain_wine:1
rhine-summer-grapeBuyer|small|付3金币，得红1白1葡萄|pay:3;gain_grape:1:red;gain_grape:1:white
rhine-summer-grapeBuyer|large|付5金币，得红4白4葡萄|pay:5;gain_grape:4:red;gain_grape:4:white
rhine-summer-grapeMerchant|wine|弃2葡萄，得桃红6（无需中酒窖）|discard_grape:2;gain_wine_exceptional:6:blush
rhine-summer-miller|grape|已有风车：得价值3葡萄|require_has:0:windmill;gain_grape:3
rhine-summer-miller|vp|已有风车：得1分|require_has:0:windmill;vp:1
rhine-summer-miller|build|付2金币建风车|require_missing:0:windmill;build_fixed:3:windmill
rhine-summer-peasant|plant|抽1订单，再忽略建筑种1藤|draw:1:order;plant_ignore_buildings:1
rhine-summer-peasant|coins|抽1订单，得2金币|draw:1:order;coins:2
rhine-summer-philanthropist|coins|得3金币|coins:3
rhine-summer-philanthropist|draw|抽1订单和1冬访客|draw:1:order;draw:1:winter
rhine-summer-premiumWineDealer|coins|得3金币|coins:3
rhine-summer-premiumWineDealer|sparkling|付9金币，得起泡7（无需大酒窖）|pay:9;gain_wine_exceptional:7:sparkling
rhine-summer-premiumWineDealer|blush|付9金币，得桃红7（无需大酒窖）|pay:9;gain_wine_exceptional:7:blush
rhine-summer-reorganizer|coins|弃2手牌，得5金币|discard_cards:2;coins:5
rhine-summer-reorganizer|grape|弃2手牌，得价值2葡萄|discard_cards:2;gain_grape:2
rhine-summer-sculptor|coins|每块自有田得1金币|field_coins:1
rhine-summer-sculptor|wine|失1年收入，得桃红4|lose_income:1;gain_wine:4:blush
rhine-summer-sculptor|plant|种至多2藤|plant:2
rhine-summer-sommelier|discard|已有品酒室：弃1酒，得1分|require_has:0:tasting_room;discard_wine:1;vp:1
rhine-summer-sommelier|build|付2金币建品酒室|require_missing:0:tasting_room;build_fixed:4:tasting_room
rhine-summer-subsidizer|build|建1建筑，减2金币|build:2
rhine-summer-subsidizer|plant|种至多2藤|plant:2
rhine-summer-subsidizer|ignore|忽略前置建筑种1藤|plant_ignore_buildings:1
rhine-summer-traveller|coins|得3金币；品酒室另得3|coins:3;if_has_coins:3:tasting_room
rhine-summer-traveller|harvest|收获至多2田；品酒室另得3|harvest:2;if_has_coins:3:tasting_room
rhine-summer-wineLover|income|已有品酒室：得1年收入|require_has:0:tasting_room;income:1
rhine-summer-wineLover|grape|已有品酒室：得价值4葡萄|require_has:0:tasting_room;gain_grape:4
rhine-summer-wineLover|build|失2年收入，免费建品酒室并得价值1酒|require_missing:0:tasting_room;lose_income:2;build_fixed:99:tasting_room;gain_wine:1
rhine-summer-wineTrader|grape|得价值1葡萄|gain_grape:1
rhine-summer-wineTrader|red|付6金币，得红酒7（无需大酒窖）|pay:6;gain_wine_exceptional:7:red
rhine-summer-wineTrader|white|付6金币，得白酒7（无需大酒窖）|pay:6;gain_wine_exceptional:7:white
rhine-summer-writer|income|弃1酒，得2年收入|discard_wine:1;income:2
rhine-summer-writer|grape|失2年收入，得价值4葡萄|lose_income:2;gain_grape:4
rhine-summer-writer|wine|失2年收入，得价值4酒|lose_income:2;gain_wine:4
rhine-winter-advertiser|grape|弃1葡萄，得1分及1年收入|discard_grape:1;vp:1;income:1
rhine-winter-advertiser|wine|弃1酒，得1分及1年收入|discard_wine:1;vp:1;income:1
rhine-winter-advertiser|draw|失2年收入，抽3订单|lose_income:2;draw:3:order
rhine-winter-bargainer|draw|抽3订单|draw:3:order
rhine-winter-cellarman|grapes|付4金币，得红1白1葡萄|pay:4;gain_grape:1:red;gain_grape:1:white
rhine-winter-cellarman|fill|完成1订单，再得3金币|fill:1;coins:3
rhine-winter-eliteOenologist|age|所有酒陈酿两次|age_all:2
rhine-winter-eliteOenologist|upgrade|酒窖升一级，减4金币|upgrade:4
rhine-winter-endorser|grape_income|弃1葡萄，得1分及1年收入|discard_grape:1;vp:1;income:1
rhine-winter-endorser|wine_income|弃1酒，得1分及1年收入|discard_wine:1;vp:1;income:1
rhine-winter-endorser|grape_orders|弃1葡萄，抽3订单|discard_grape:1;draw:3:order
rhine-winter-endorser|wine_orders|弃1酒，抽3订单|discard_wine:1;draw:3:order
rhine-winter-enthusiast|draw|抽2夏访客|draw:2:summer
rhine-winter-enthusiast|fill|完成1订单，再抽1订单|fill:1;draw:1:order
rhine-winter-grapeWhisperer|harvest|收获至多2田，得2金币|harvest:2;coins:2
rhine-winter-harvestExpert|harvest|收获1田|harvest:1
rhine-winter-harvestExpert|vine|抽1藤|draw:1:vine
rhine-winter-harvestExpert|build|付1金币建轭|build_fixed:1:yoke
rhine-winter-lecturer|make|酿至多3酒|make:3
rhine-winter-lecturer|train|付3金币培训1工人|train:3
rhine-winter-lovebirds|draw|得1金币并抽1夏访客|coins:1;draw:1:summer
rhine-winter-lovebirds|make|抽1订单，再酿至多2酒|draw:1:order;make:2
rhine-winter-researcher|draw|抽2订单|draw:2:order
rhine-winter-researcher|train|付3金币培训1工人|train:3
rhine-winter-rhineSailor|exchange|弃3手牌，抽3冬访客并得1金币|discard_cards:3;draw:3:winter;coins:1
rhine-winter-schoolTeacher|train|付4金币培训当年可用工人|train:4:now
rhine-winter-schoolTeacher|wine|弃1酒，得2分|discard_wine:1;vp:2
rhine-winter-shipper|make|酿至多3酒|make:3
rhine-winter-shipper|fill|完成1订单|fill:1
rhine-winter-shipper|coins|得3金币|coins:3
rhine-winter-uncertifiedOenologist|age|所有酒陈酿两次|age_all:2
rhine-winter-uncertifiedOenologist|upgrade|失1分，免费升级酒窖|lose_vp:1;upgrade:99
rhine-winter-winterAgent|draw|抽2夏访客，检查额外订单|draw:2:summer;if_opponent_vp_draw:5:order
rhine-winter-zymologist|make|酿至多2瓶至少价值4的酒，可忽略酒窖|make_exceptional:2
`

func rhineLookup(id string) (rhineDefinition, error) {
	d, ok := rhineDefinitions[id]
	if !ok {
		return d, fmt.Errorf("Rhine访客未实现：%s", id)
	}
	return d, nil
}

const rhineMorePrograms = `
rhine-winter-trainer|train|付3金币培训，当年可用|train:3:now
rhine-winter-trainer|lose|移除1名待命永久工人，得2分|lose_worker:0:available;vp:2
rhine-winter-tutor|train|失1分并付1金币，培训1工人|lose_vp:1;train:1
rhine-winter-tutor|exchange|移除1名永久工人，弃X手牌后抽X牌|lose_worker;exchange_cards;exchange_draw
rhine-winter-lovebirds|make_draw|先酿至多2酒，再抽1订单|make:2;draw:1:order
rhine-winter-theologian|grape_harvest|先得价值1葡萄，再收获1田|gain_grape:1;harvest:1
rhine-winter-laborer|harvest|收获至多2田|harvest:2
rhine-winter-laborer|make|酿至多3酒|make:3
rhine-winter-laborer|harvest_make|失1分，先收获至多2田，再酿至多3酒|lose_vp:1;harvest:2;make:3
rhine-winter-laborer|make_harvest|失1分，先酿至多3酒，再收获至多2田|lose_vp:1;make:3;harvest:2
rhine-winter-supervisor|make|酿至多2酒，每瓶新酿起泡酒得1分|make_supervisor:2
rhine-winter-virtuoso|action|执行前季公共行动，可选任一奖励，无需放工人|virtuoso
rhine-summer-wineEngineer|wine|付2金币，得价值4酒|pay:2;gain_wine:4
rhine-summer-wineEngineer|age|每次付2金币，陈酿1或2葡萄，至多3次|engineer_start;pay:2;engineer_age:2;choose:0:engineer
rhine-summer-sonInLaw|field|永久放在1田，每年首次冬季行动额外收获|son_field
rhine-summer-administrator|plan|移动本次工人到未来季节，季初执行|plan_trigger
rhine-summer-vineTrader|trade|每弃1藤得3金币，再购买藤或冬访客|sell_vines;buy_cards
rhine-winter-cheapBuyer|fill|1或2瓶酒合计增加2价值，再完成订单|age_two_total;fill:1
rhine-summer-brickMason|build|付9金币，建恰好2建筑|pay:9;build:99;build:99
rhine-summer-brickMason|destroy|拆1建筑，得1分并抽1订单|destroy;vp:1;draw:1:order
rhine-summer-dismantler|destroy|拆1建筑，得标价对应价值的酒|destroy_wine
rhine-summer-earlyBuyer|fill|完成1订单，得2金币；5分以内改得4金币|fill_early
rhine-summer-greenskeeper|normal|无人达到10分：得2金币并种1藤|require_vp_threshold:10:no;coins:2;plant:1
rhine-summer-greenskeeper|extra|已有玩家10分：得2金币，种1藤再超田地总值种1藤|require_vp_threshold:10:yes;coins:2;plant:1;plant_ignore_capacity:1
rhine-summer-oldGeneral|resolve|放置1星；放满6星的对手各收回3星|influence_place:1;general_responses
rhine-summer-structureReorganizer|draw|抽1任意牌，可拆1建筑后免费建1建筑|draw_any:1;choose:0:structure_destroy
rhine-winter-influencer|resolve|放置1星，再移动至多3颗任意星|influence_place:1;move_stars:3
rhine-winter-lobbyist|resolve|每付1金币收回1星；有玩家20分时失3分|lobbyist
rhine-summer-accountant|resolve|抽藤、夏访客、冬访客各1并得1金币|draw:1:vine;draw:1:summer;draw:1:winter;coins:1;opponents:0:accountant
rhine-summer-ampelograph|plant|种1藤，可超田地总值|plant_ampelograph:1
rhine-summer-banker|resolve|得5金币，对手可失1分得3金币|coins:5;opponents:0:banker
rhine-summer-grapeMerchant|sell|卖至多3葡萄，卖价三倍|sell_grapes:3
rhine-summer-plantDealer|draw|抽4藤，从中弃2藤|draw_new:4:vine;discard_new:2
rhine-summer-plantReorganizer|replant|抽1藤和1任意牌，再拔1藤并种1藤|draw:1:vine;draw_any:1;uproot_keep:1;plant:1
rhine-winter-brideToBe|coins|得3金币|coins:3
rhine-winter-brideToBe|make|任意2葡萄酿1起泡酒|make_two_sparkling:1
rhine-winter-bureaucrat|resolve|严格最低分得5金币，否则付1金币抽3种牌|bureaucrat
rhine-winter-cellarmaster|discard|弃1酒得4金币，最高价值另得2分|cellarmaster
rhine-winter-chemist|plant|种恰好2藤|plant_exact:2
rhine-winter-chemist|harvest|收获1田，葡萄各减1，得1分|harvest_chemist:1
rhine-winter-duchess|income|付1金币，得1年收入|pay:1;income:1
rhine-winter-duchess|fill|失2年收入，完成恰好2订单|lose_income:2;fill:1;fill:1
rhine-winter-estateAgent|draw|按自有田地数抽任意类型牌|draw_fields
rhine-winter-grapeVendor|coins|弃1葡萄，按价值得金币|grape_vendor:1:coins;if_has_coins:4:tasting_room
rhine-winter-grapeVendor|draw|弃1葡萄，按半价值向上取整抽冬访客|grape_vendor:1:draw;if_has_coins:4:tasting_room
rhine-winter-middleman|draw|抽3藤，然后弃1藤或付2金币|draw:3:vine;choose:0:middleman
rhine-winter-middleman|fill|忽略酒种完成1订单|fill_any:1
rhine-winter-premiumBuyer|fill|完成1订单，每瓶都高2价值另得2分|fill_premium:1
rhine-winter-skeptic|upgrade|酒窖升一级，减3金币|upgrade:3
rhine-winter-skeptic|fill|陈酿2瓶酒，完成1订单|age_wines:2;fill:1
rhine-winter-specialHarvester|harvest_make|收获1田，再酿1酒|harvest:1;make:1
rhine-winter-specialHarvester|separate|将1田的每张藤分别收获|harvest_special:1
rhine-winter-theologian|harvest|收获1田，额外得价值1葡萄|harvest:1;gain_grape:1
rhine-winter-theologian|fill|完成1订单，可弃红酒换金币或分数|fill:1;choose:0:theologian
rhine-winter-wineStoreOwner|draw|弃1酒，按其价值抽订单|wine_orders:1
rhine-winter-wineStoreOwner|buy|付X金币，得价值X的酒|wine_buy
`

func rhineRegisterPairs() {
	op := func(kind string, n int, target string) rhineOption {
		return rhineOption{Label: map[string]string{"vp": "得1分", "build": "建造", "plant": "种藤", "draw": "抽" + target, "coins": "得金币", "upgrade": "酒窖升级", "harvest": "收获", "make": "酿酒", "fill": "交付订单"}[kind], Operations: []RhineOperation{{Kind: kind, Amount: n, Target: target}}}
	}
	groups := []struct {
		id    string
		items map[string]rhineOption
		tail  []RhineOperation
	}{
		{"rhine-summer-contractor", map[string]rhineOption{"vp": op("vp", 1, ""), "build": op("build", 0, ""), "plant": op("plant", 1, "")}, nil},
		{"rhine-summer-owner", map[string]rhineOption{"build": op("build", 0, ""), "draw": op("draw", 1, "order"), "plant": op("plant", 1, "")}, []RhineOperation{{Kind: "if_has_vp", Amount: 1, Target: "windmill"}}},
		{"rhine-summer-supporter", map[string]rhineOption{"vine": op("draw", 1, "vine"), "coins": op("coins", 2, ""), "order": op("draw", 1, "order")}, nil},
		{"rhine-winter-craftsman", map[string]rhineOption{"draw": op("draw", 1, "order"), "upgrade": op("upgrade", 0, ""), "vp": op("vp", 1, "")}, nil},
		{"rhine-winter-hiredHand", map[string]rhineOption{"harvest": op("harvest", 1, ""), "make": op("make", 2, ""), "coins": op("coins", 2, ""), "fill": op("fill", 1, "")}, nil},
	}
	for _, g := range groups {
		var d rhineDefinition
		for _, f := range rhineCardFacts {
			if f.Card.ID == g.id {
				d = f
				break
			}
		}
		d.Options = map[string]rhineOption{}
		for first, a := range g.items {
			for last, b := range g.items {
				if first == last {
					continue
				}
				ops := append([]RhineOperation{}, a.Operations...)
				ops = append(ops, b.Operations...)
				ops = append(ops, g.tail...)
				d.Options[first+"_"+last] = rhineOption{Label: a.Label + " → " + b.Label, Operations: ops}
			}
		}
		rhineDefinitions[g.id] = d
	}
}
