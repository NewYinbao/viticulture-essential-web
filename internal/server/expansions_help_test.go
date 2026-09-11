package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
	"vineyard/internal/game"
)

// These are persisted development configurations, not a production start-gate
// bypass. Exercise the real private/public HTTP and SSE projections used by help.
func TestExpansionHelpViewMatrix(t *testing.T) {
	a, err := loadApp(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, board := range []string{"ee", "tuscany"} {
		for _, visitors := range []string{"ee", "ee_moor", "rhine"} {
			for _, structures := range []bool{false, true} {
				for _, workers := range []bool{false, true} {
					cfg := game.ExpansionConfig{Board: board, Visitors: visitors, Structures: structures, SpecialWorkers: workers}
					t.Run(fmt.Sprint(cfg), func(t *testing.T) {
						p, q := game.NewPlayer("Host"), game.NewPlayer("Guest")
						p.Hand = []game.Card{{ID: "private-owner", Type: "vine"}}
						q.Hand = []game.Card{{ID: "private-responder", Type: "summer"}}
						r := &game.Room{Code: "HELP", Config: cfg, Phase: "summer", Players: []*game.Player{p, q}, HostID: p.ID, TurnID: q.ID,
							Choices: []game.Choice{{ID: "private-choice", Kind: "visitor", PlayerID: q.ID, Options: []string{"coins", "cards"}, Visitor: &game.VisitorStep{CardID: "winter-11", Stage: "reply", ActorID: p.ID}}}}
						a.Store.Rooms[r.Code] = r
						setFixtureSession(a, "host", r.Code, p.ID)
						setFixtureSession(a, "guest", r.Code, q.ID)
						before, _ := json.Marshal(r)
						h := a.routes()
						srv := httptest.NewServer(h)
						defer srv.Close()
						for _, token := range []string{"host", "guest"} {
							w := request(h, "GET", "/api/state", token, nil)
							if w.Code != 200 {
								t.Fatal(w.Body.String())
							}
							var view map[string]any
							if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
								t.Fatal(err)
							}
							support := view["ruleSupport"].(map[string]any)
							if support["playable"] != (cfg.Playable() == nil) {
								t.Fatal("incorrect playability", support)
							}
							if cfg.Playable() != nil && support["reason"] == "" {
								t.Fatal("missing unfinished warning")
							}
							choice := view["pendingChoice"].(map[string]any)
							if token == "host" {
								if len(choice) != 2 || strings.Contains(w.Body.String(), "private-responder") || strings.Contains(w.Body.String(), "winter-11") {
									t.Fatal("help leaked another response", choice)
								}
							} else if choice["visitor"].(map[string]any)["stage"] != "reply" {
								t.Fatal("lost private stage")
							}
							client := http.Client{Timeout: 3 * time.Second}
							res, err := client.Get(srv.URL + "/api/events?token=" + token)
							if err != nil {
								t.Fatal(err)
							}
							scanner := bufio.NewScanner(res.Body)
							if !scanner.Scan() {
								res.Body.Close()
								t.Fatal("no SSE")
							}
							var event map[string]any
							err = json.Unmarshal([]byte(strings.TrimPrefix(scanner.Text(), "data: ")), &event)
							res.Body.Close()
							if err != nil || !reflect.DeepEqual(event, view) {
								t.Fatal("HTTP/poll/SSE help views diverged", err)
							}
						}
						after, _ := json.Marshal(r)
						if string(before) != string(after) {
							t.Fatal("reading help mutated gameplay")
						}
						if err := a.save(); err != nil {
							t.Fatal(err)
						}
						restored, err := loadApp(a.dir)
						if err != nil || restored.Store.Rooms[r.Code].Config != cfg {
							t.Fatal("configuration not restored", err)
						}
					})
				}
			}
		}
	}
}
