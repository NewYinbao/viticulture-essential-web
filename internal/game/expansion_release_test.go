package game

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Tests release readiness through the real catalogue/deck/start APIs. No gate
// bypass or synthetic replacement cards are allowed in this matrix.
func TestReleasedExpansionMatrix(t *testing.T) {
	boardOnly := map[string]bool{"rhine-summer-oldGeneral": true, "rhine-summer-structureReorganizer": true, "rhine-winter-influencer": true, "rhine-winter-lobbyist": true}
	baseIDs := map[string]bool{}
	for _, card := range Catalog() {
		baseIDs[card.ID] = true
	}
	for _, board := range []string{"ee", "tuscany"} {
		for _, visitors := range []string{"ee", "ee_moor", "rhine"} {
			for _, structures := range []bool{false, true} {
				for _, workers := range []bool{false, true} {
					c := ExpansionConfig{board, structures, workers, visitors}
					t.Run(fmt.Sprintf("%s/%s/structures_%t/workers_%t", board, visitors, structures, workers), func(t *testing.T) {
						cards, err := CatalogForConfig(c)
						if err != nil {
							t.Fatal(err)
						}
						ids, counts := map[string]bool{}, map[string]int{}
						for _, card := range cards {
							if ids[card.ID] {
								t.Fatal("duplicate ID", card.ID)
							}
							ids[card.ID], counts[card.Type] = true, counts[card.Type]+1
							visitor := card.Type == "summer" || card.Type == "winter"
							if visitor && !card.Implemented {
								t.Fatal("missing effect", card.ID)
							}
							if visitor && visitors == "rhine" && !strings.HasPrefix(card.ID, "rhine-") {
								t.Fatal("Rhine must replace all other visitors", card.ID)
							}
							if visitor && visitors != "rhine" && strings.HasPrefix(card.ID, "rhine-") {
								t.Fatal("Rhine leaked into another deck", card.ID)
							}
							if visitors == "ee" && strings.HasPrefix(card.ID, "moor-") {
								t.Fatal("Moor leaked into EE", card.ID)
							}
							if (card.Type == "vine" || card.Type == "order") && !baseIDs[card.ID] {
								t.Fatal("visitor selection replaced production deck", card.ID)
							}
						}
						for _, card := range Catalog() {
							visitor := card.Type == "summer" || card.Type == "winter"
							if visitors == "rhine" && visitor {
								continue
							}
							if !ids[card.ID] {
								t.Fatal("EE card omitted from compatible catalogue", card.ID)
							}
						}
						want := 38
						if visitors == "ee_moor" {
							want = 58
						}
						if visitors == "rhine" && board == "tuscany" {
							want = 40
						}
						if counts["summer"] != want || counts["winter"] != want {
							t.Fatal("visitor counts", counts, "want each", want)
						}
						if counts["structure"] != map[bool]int{false: 0, true: 36}[structures] {
							t.Fatal("structure deck gating", counts)
						}
						for id := range boardOnly {
							if ids[id] != (visitors == "rhine" && board == "tuscany") {
								t.Fatal("board-only card must depend solely on Tuscany board", id, c)
							}
						}
						r := expansionLobby()
						if err := r.Apply(r.HostID, Action{Type: "configure", Revision: r.Revision, Config: &c}); err != nil {
							t.Fatal(err)
						}
						r = restoreT(t, r)
						if err := r.InitDecks(); err != nil {
							t.Fatal(err)
						}
						for typ, n := range counts {
							for i := 0; i < n; i++ {
								card, ok := r.take(typ)
								if !ok || !ids[card.ID] {
									t.Fatal("deck identity lost", typ, i)
								}
								delete(ids, card.ID)
								r.discard(card)
							}
						}
						if len(ids) != 0 {
							t.Fatal("catalogue cards not present in real deck", ids)
						}
						r = restoreT(t, r)
						for typ, n := range counts {
							seen := map[string]bool{}
							for i := 0; i < n; i++ {
								card, ok := r.take(typ)
								if !ok || seen[card.ID] {
									t.Fatal("recycle duplicated or lost cards", typ)
								}
								seen[card.ID] = true
							}
							if _, ok := r.take(typ); ok {
								t.Fatal("deck is not finite", typ)
							}
						}
						if err := r.Apply(r.HostID, Action{Type: "start", Revision: r.Revision}); err != nil {
							t.Fatal(err)
						}
						r = restoreT(t, r)
						if len(r.SpecialWorkerPool) != map[bool]int{false: 0, true: 2}[workers] {
							t.Fatal("special pool gating", r.SpecialWorkerPool)
						}
						before, _ := json.Marshal(r)
						if r.Apply(r.HostID, Action{Type: "configure", Revision: r.Revision, Config: &c}) == nil {
							t.Fatal("started config unlocked")
						}
						after, _ := json.Marshal(r)
						if string(before) != string(after) {
							t.Fatal("locked configure mutated state")
						}
						for _, p := range r.Players {
							v := r.View(p.ID)
							if v["config"].(ExpansionConfig) != c {
								t.Fatal("restored View configuration")
							}
							if _, ok := v["influenceRegions"]; ok != (board == "tuscany") {
								t.Fatal("influence projection gating")
							}
							if v["legal"].(map[string]bool)["canConfigure"] {
								t.Fatal("View configure unlocked")
							}
							if !v["ruleSupport"].(map[string]any)["playable"].(bool) {
								t.Fatal("playable View has obsolete gate")
							}
						}
					})
				}
			}
		}
	}
}

func TestReleasedExpansionStillRejectsMissingVisitorEffect(t *testing.T) {
	id := "rhine-summer-plantDealer"
	saved, ok := rhineDefinitions[id]
	if !ok {
		t.Fatal("fixture requires real implemented Rhine card")
	}
	delete(rhineDefinitions, id)
	defer func() { rhineDefinitions[id] = saved }()
	c := ExpansionConfig{Board: "tuscany", Visitors: "rhine"}
	if c.Playable() == nil {
		t.Fatal("missing effect accepted")
	}
	r := expansionLobby()
	r.Config = c
	before, _ := json.Marshal(r)
	if err := r.Apply(r.HostID, Action{Type: "start", Revision: r.Revision}); err == nil {
		t.Fatal("missing effect started")
	}
	after, _ := json.Marshal(r)
	if string(before) != string(after) {
		t.Fatal("missing effect start mutated room")
	}
	v := r.View(r.HostID)
	if v["expansionAvailability"].(map[string]bool)["rhine"] || v["legal"].(map[string]bool)["canStart"] {
		t.Fatal("unavailable deck advertised as playable")
	}
}
