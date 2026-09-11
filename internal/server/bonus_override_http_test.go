package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"vineyard/internal/game"
)

func TestHTTPDiscardsForgedBonusAndPersistsOnlyLegalReward(t *testing.T) {
	for _, board := range []string{"ee", "tuscany"} {
		t.Run(board, func(t *testing.T) {
			dir := t.TempDir()
			app, err := loadApp(dir)
			if err != nil {
				t.Fatal(err)
			}
			r := eeBegin(t, 3)
			p := r.Player(r.TurnID)
			p.Coins, p.Hand = 0, nil
			action := game.Action{Type: "place", Space: "tour", Slot: 2, Revision: r.Revision, BonusOverride: "coin"}
			if board == "tuscany" {
				r.Config.Board, r.Phase, r.Spaces = "tuscany", "fall", game.NewTuscanySpaces(3)
				for _, player := range r.Players {
					player.Season, player.Passed = "fall", false
				}
				action.Space, action.BonusOverride = "draw_order", "draw_order"
			}
			app.Store.Rooms[r.Code] = r
			setFixtureSession(app, "bonus-boundary-token", r.Code, p.ID)
			server := httptest.NewServer(app.routes())
			defer server.Close()
			body, err := json.Marshal(action)
			if err != nil {
				t.Fatal(err)
			}
			post := func() int {
				req, err := http.NewRequest(http.MethodPost, server.URL+"/api/action", bytes.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer bonus-boundary-token")
				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()
				var payload map[string]any
				if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				return res.StatusCode
			}
			if status := post(); status != http.StatusOK {
				t.Fatalf("legal base action failed: HTTP %d", status)
			}
			if status := post(); status != http.StatusConflict {
				t.Fatalf("replay was not rejected: HTTP %d", status)
			}
			restored, err := loadApp(dir)
			if err != nil {
				t.Fatal(err)
			}
			stored := restored.Store.Rooms[r.Code].Player(p.ID)
			if board == "ee" && stored.Coins != 2 {
				t.Fatalf("forged tour reward persisted: coins=%d", stored.Coins)
			}
			if board == "tuscany" && len(stored.Hand) != 1 {
				t.Fatalf("forged order reward persisted: hand=%d", len(stored.Hand))
			}
		})
	}
}
