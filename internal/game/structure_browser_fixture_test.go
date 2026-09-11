package game

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Deliberately seeded isolated UI fixtures. These are not natural games.
func TestStructureBrowserFixtureExport(t *testing.T) {
	target := os.Getenv("STRUCTURE_BROWSER_FIXTURES")
	if target == "" {
		t.Skip("fixture export not requested")
	}
	rooms := map[string]*Room{}
	sessions := map[string]any{}
	cases := []map[string]any{}
	for index, d := range structureDefs {
		r := structureRoom()
		p := r.Players[0]
		q := r.Players[1]
		p.Name = "卡面审查"
		q.Name = "对手触发"
		r.Code = fmt.Sprintf("CARD%02d", index+1)
		r.Revision = 1
		r.Ruleset = "tuscany-essential-v1"
		p.Hand = []Card{}
		p.Wines = []Wine{}
		p.Grapes = []Grape{}
		p.Buildings = []string{d.ID, "medium_cellar", "large_cellar"}
		p.StructureSlots = []string{d.ID, ""}
		p.Coins = 20
		p.VP = 0
		order := Card{ID: "fixture-order", Type: "order", Name: "测试订单", Requirements: []Wine{{Type: "red", Value: 1}}, Points: 2, Income: 1}
		vine := Card{ID: "fixture-vine", Type: "vine", Name: "测试葡萄藤", Red: 1}
		r.Decks["order"] = []Card{order}
		r.Decks["vine"] = []Card{vine}
		actor := p
		steps := []map[string]any{}
		expect := map[string]any{}
		step := func(kind, space string, groups ...[]string) {
			m := map[string]any{"kind": kind, "space": space, "groups": groups}
			steps = append(steps, m)
		}
		private := func(groups ...[]string) { step("place", structureActionSpaceID(p.ID, d.ID), groups...) }
		phase := func(s string) { r.Phase = s; p.Season = s; q.Season = s }
		redWine := func() { p.Wines = []Wine{{ID: "wine-red", Type: "red", Value: 1}} }
		grapes := func() {
			p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 4}, {ID: "w1", Color: "white", Value: 3}, {ID: "r2", Color: "red", Value: 2}, {ID: "w2", Color: "white", Value: 2}, {ID: "r3", Color: "red", Value: 3}}
		}
		fill := func(space string) {
			redWine()
			p.Hand = []Card{order}
			phase("winter")
			step("place", space, []string{"cardId", order.ID}, []string{"wines", "wine-red"})
		}
		influence := func() { step("influence", "", []string{"to", "pisa"}) }
		visitor := func() {
			phase("summer")
			p.Hand = []Card{visitorCardT(t, "summer-01")}
			step("place", "summer_visitor", []string{"cardId", "summer-01"})
			step("visitor", "coins")
		}
		switch d.ID {
		case "cask":
			redWine()
			private([]string{"wines", "wine-red"})
			expect["wineValues"] = []int{3}
			expect["hand"] = 1
		case "wine_cave":
			redWine()
			p.Wines = append(p.Wines, Wine{ID: "wine-white", Type: "white", Value: 2})
			private([]string{"wines", "wine-red"}, []string{"wines", "wine-white"})
			expect["wineValues"] = []int{2, 3}
			expect["hand"] = 1
		case "trading_post":
			private([]string{"tuscany:give", "coins"}, []string{"tuscany:receive", "vp"})
			influence()
			expect["coins"] = 19
			expect["vp"] = 1
			expect["pisa"] = 1
		case "shop":
			fill(structureActionSpaceID(p.ID, d.ID))
			influence()
			expect["coins"] = 22
			expect["vp"] = 2
			expect["hand"] = 0
			expect["pisa"] = 1
		case "wine_press":
			grapes()
			private([]string{"grapes", "0"}, []string{"bottle", "2"}, []string{"grapes", "1"})
			influence()
			expect["wines"] = 2
			expect["coins"] = 22
			expect["pisa"] = 1
		case "school":
			p.Coins = 0
			private()
			expect["coins"] = 1
			expect["totalWorkers"] = 4
			expect["workers"] = 2
		case "wine_bar":
			redWine()
			private([]string{"wines", "wine-red"})
			expect["wines"] = 0
			expect["vp"] = 2
		case "ristorante":
			redWine()
			grapes()
			private([]string{"wines", "wine-red"}, []string{"grapeId", "r1"})
			expect["coins"] = 23
			expect["vp"] = 3
			expect["wines"] = 0
			expect["grapes"] = 4
		case "guest_house":
			p.Hand = []Card{visitorCardT(t, "summer-01"), visitorCardT(t, "winter-34")}
			private([]string{"cards", "summer-01"}, []string{"cards", "winter-34"})
			expect["vp"] = 2
			expect["hand"] = 0
		case "cafe":
			grapes()
			private([]string{"grapeId", "r1"})
			expect["coins"] = 23
			expect["vp"] = 1
			expect["grapes"] = 4
		case "mixer":
			grapes()
			private([]string{"grapes", "2"}, []string{"grapes", "3"}, []string{"bottle", "2"}, []string{"grapes", "0"}, []string{"grapes", "1"}, []string{"grapes", "4"})
			expect["vp"] = 1
			expect["wineTypes"] = []string{"blush", "sparkling"}
		case "aqueduct":
			phase("summer")
			vine.Trellis = true
			vine.Irrigation = true
			vine.Red = 3
			vine.White = 3
			p.Hand = []Card{vine}
			step("place", "plant", []string{"cards", vine.ID}, []string{"field-" + vine.ID, "1"})
			expect["planted"] = 1
			expect["hand"] = 0
		case "patio", "penthouse", "charmat":
			phase("fall")
			grapes()
			if d.ID == "patio" {
				step("place", "make_wine", []string{"grapes", "2"}, []string{"grapes", "3"}, []string{"bottle", "2"}, []string{"grapes", "0"}, []string{"grapes", "1"}, []string{"grapes", "4"})
				expect["coins"] = 24
			} else if d.ID == "charmat" {
				step("place", "make_wine", []string{"grapes", "0"}, []string{"grapes", "1"}, []string{"recipeType", "sparkling"})
				expect["wineTypes"] = []string{"sparkling"}
			} else {
				p.Grapes[0].Value = 7
				step("place", "make_wine", []string{"grapes", "0"})
				expect["vp"] = 1
			}
		case "mercado":
			redWine()
			phase("fall")
			step("place", "draw_order")
			step("resource", "structure_mercado", []string{"cardId", order.ID}, []string{"wines", "wine-red"})
			expect["vp"] = 2
			expect["hand"] = 0
			expect["wines"] = 0
		case "studio", "workshop":
			step("place", "build", []string{"building", "trellis"})
			if d.ID == "studio" {
				expect["vp"] = 1
				expect["coins"] = 18
			} else {
				p.Coins = 1
				expect["coins"] = 0
			}
			expect["building"] = "trellis"
		case "barn":
			phase("summer")
			p.Hand = []Card{vine, order}
			step("pass", "")
			step("barn", "", []string{vine.ID}, []string{order.ID})
			expect["vp"] = 1
			expect["hand"] = 0
		case "academy":
			phase("winter")
			actor = q
			r.TurnID = q.ID
			q.Coins = 5
			step("place", "train")
			expect["coins"] = 21
		case "gazebo":
			phase("spring")
			step("place", "tour")
			influence()
			expect["coins"] = 24
			expect["pisa"] = 1
		case "veranda", "wine_parlor":
			fill("fill_order")
			expect["vp"] = 2
			expect["coins"] = 20
			if d.ID == "veranda" {
				expect["vp"] = 3
			} else {
				expect["coins"] = 22
			}
		case "label_factory":
			fill(structureActionSpaceID(p.ID, d.ID))
			expect["coins"] = 17
			expect["vp"] = 4
		case "harvest_machine", "fermentation_tank":
			phase("fall")
			for i := 0; i < 3; i++ {
				p.Fields[i].Vines = []Card{vine}
			}
			if d.ID == "harvest_machine" {
				step("place", "harvest", []string{"mode", "all"})
				expect["harvested"] = 3
			} else {
				step("place", "harvest", []string{"fields", "0"})
				step("resource", "structure_fermentation", []string{"grapes", "0"})
				expect["wines"] = 1
				expect["harvested"] = 1
			}
		case "inn", "tap_room", "tavern":
			redWine()
			grapes()
			visitor()
			if d.ID == "inn" {
				expect["coins"] = 27
			} else if d.ID == "tap_room" {
				step("visitorBonus", "tap_room", []string{"wine", "wine-red"})
				expect["vp"] = 2
				expect["wines"] = 0
			} else {
				step("visitorBonus", "tavern", []string{"grape", "r1"}, []string{"grape", "w1"})
				expect["vp"] = 3
				expect["grapes"] = 3
			}
		case "banquet_hall":
			p.Influence = map[string]int{"siena": 6}
			step("place", "influence", []string{"tuscany:from", "siena"}, []string{"tuscany:to", "pisa"})
			expect["pisa"] = 1
			expect["coins"] = 22
		case "fountain":
			phase("spring")
			actor = q
			r.TurnID = q.ID
			step("place", "tour")
			expect["coins"] = 21
		case "distiller", "storehouse", "statue", "dock", "silo":
			phase("winter")
			redWine()
			p.Grapes = []Grape{{ID: "r1", Color: "red", Value: 1}}
			q.Season = "ready"
			q.Passed = true
			q.NextWake = 3
			step("pass", "")
			if d.ID == "distiller" {
				expect["grapeValues"] = []int{3}
			}
			if d.ID == "storehouse" {
				expect["wineValues"] = []int{3}
			}
			if d.ID == "dock" || d.ID == "silo" {
				expect["hand"] = 1
			}
			if d.ID == "statue" {
				step("choice", "continue")
				step("choice", "2")
				expect["vp"] = 1
				expect["year"] = 2
			}
		default:
			t.Fatalf("unmapped card %s", d.ID)
		}
		r.ensureAllStructureActionSpaces()
		rooms[r.Code] = r
		for _, who := range r.Players {
			token := r.Code + "-" + who.ID
			sessions[token] = map[string]string{"Code": r.Code, "PlayerID": who.ID}
		}
		cases = append(cases, map[string]any{"code": r.Code, "card": d.ID, "ownerId": p.ID, "actorId": actor.ID, "token": r.Code + "-" + actor.ID, "steps": steps, "expect": expect})
	}
	data, e := json.MarshalIndent(map[string]any{"Rooms": rooms, "Sessions": sessions, "cases": cases, "method": "explicit isolated seeded fixtures, never natural games"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(target, data, 0600); e != nil {
		t.Fatal(e)
	}
}
