package game

import (
	"encoding/json"
	"testing"
)

func expansionLobby() *Room {
	p, q := NewPlayer("Host"), NewPlayer("Guest")
	return &Room{Phase: "lobby", HostID: p.ID, Players: []*Player{p, q}, Revision: 4}
}

func TestExpansionConfigValidation(t *testing.T) {
	if got := (ExpansionConfig{}).Normalized(); got.Board != "ee" || got.Visitors != "ee" {
		t.Fatal(got)
	}
	for _, board := range []string{"ee", "tuscany"} {
		for _, visitors := range []string{"ee", "ee_moor", "rhine"} {
			for _, structures := range []bool{false, true} {
				for _, workers := range []bool{false, true} {
					c := ExpansionConfig{board, structures, workers, visitors}
					if err := c.Validate(); err != nil {
						t.Fatalf("legal independent combination %+v: %v", c, err)
					}
					available := true
					if (c.Playable() == nil) != available {
						t.Fatalf("released combination is unavailable: %+v", c)
					}
				}
			}
		}
	}
	for _, c := range []ExpansionConfig{{Board: "bordeaux"}, {Board: "ee+tuscany"}, {Visitors: "moor+rhine"}, {Visitors: "ee+rhine"}} {
		if c.Validate() == nil {
			t.Fatalf("invalid accepted %+v", c)
		}
	}
}

func TestExpansionConfigPermissionAndRollback(t *testing.T) {
	for _, tc := range []struct {
		name, actor, phase string
		revision           int
		config             *ExpansionConfig
	}{
		{"non-host", "guest", "lobby", 4, &ExpansionConfig{}},
		{"stale", "host", "lobby", 3, &ExpansionConfig{}},
		{"started", "host", "summer", 4, &ExpansionConfig{}},
		{"missing", "host", "lobby", 4, nil},
		{"unknown", "host", "lobby", 4, &ExpansionConfig{Visitors: "ee+rhine"}},
		{"unknown-board", "host", "lobby", 4, &ExpansionConfig{Board: "bordeaux"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := expansionLobby()
			r.Phase = tc.phase
			id := r.HostID
			if tc.actor == "guest" {
				id = r.Players[1].ID
			}
			before, _ := json.Marshal(r)
			if err := r.Apply(id, Action{Type: "configure", Revision: tc.revision, Config: tc.config}); err == nil {
				t.Fatal("accepted")
			}
			after, _ := json.Marshal(r)
			if string(before) != string(after) {
				t.Fatal("rejection mutated room")
			}
		})
	}
	r := expansionLobby()
	if err := r.Apply(r.HostID, Action{Type: "configure", Revision: 4, Config: &ExpansionConfig{}}); err != nil {
		t.Fatal(err)
	}
	if r.Config.Board != "ee" || r.Config.Visitors != "ee" {
		t.Fatal(r.Config)
	}
}

func TestInvalidExpansionCannotStartFromSave(t *testing.T) {
	for _, c := range []ExpansionConfig{{Visitors: "ee+rhine"}, {Board: "bordeaux"}} {
		r := expansionLobby()
		r.Config = c
		before, _ := json.Marshal(r)
		if r.Apply(r.HostID, Action{Type: "start"}) == nil {
			t.Fatal("invalid module started", c)
		}
		after, _ := json.Marshal(r)
		if string(before) != string(after) {
			t.Fatal("start rejection mutated save")
		}
	}
}

func TestSpecialWorkersConfigurationCanStartAndProjectsPool(t *testing.T) {
	r := expansionLobby()
	c := &ExpansionConfig{Board: "tuscany", SpecialWorkers: true, Visitors: "ee"}
	if err := r.Apply(r.HostID, Action{Type: "configure", Revision: r.Revision, Config: c}); err != nil {
		t.Fatal(err)
	}
	if err := r.Apply(r.HostID, Action{Type: "start", Revision: r.Revision}); err != nil {
		t.Fatal(err)
	}
	if len(r.SpecialWorkerPool) != 2 || r.SpecialWorkerPool[0] == r.SpecialWorkerPool[1] {
		t.Fatalf("special worker pool not initialized: %#v", r.SpecialWorkerPool)
	}
	v := r.View(r.HostID)
	pool, ok := v["specialWorkerPool"].([]string)
	if !ok || len(pool) != 2 {
		t.Fatalf("special worker pool not projected: %#v", v["specialWorkerPool"])
	}
}

func TestExpansionDeckSelectionFailsClosed(t *testing.T) {
	for _, config := range []ExpansionConfig{{Visitors: "moor+rhine"}, {Board: "bordeaux"}, {Visitors: "ee+rhine"}} {
		r := &Room{Config: config, Decks: map[string][]Card{"vine": {{ID: "existing", Type: "vine"}}}}
		before, _ := json.Marshal(r)
		if r.InitDecks() == nil {
			t.Fatal("unsupported deck initialized", config)
		}
		after, _ := json.Marshal(r)
		if string(before) != string(after) {
			t.Fatal("deck failure mutated state")
		}
	}
	r := &Room{}
	if err := r.InitDecks(); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, typ := range []string{"summer", "winter"} {
		if len(r.Decks[typ]) != 38 {
			t.Fatal("default deck changed")
		}
		for i := 0; i < 38; i++ {
			c, ok := r.take(typ)
			if !ok || seen[c.ID] {
				t.Fatal("duplicate/missing card")
			}
			seen[c.ID] = true
			r.discard(c)
		}
		r = restoreT(t, r)
		for i := 0; i < 38; i++ {
			if _, ok := r.take(typ); !ok {
				t.Fatal("discard recycle lost card")
			}
		}
		if _, ok := r.take(typ); ok {
			t.Fatal("infinite synthetic cards")
		}
	}
}

func TestStructuresCatalogueAndFiniteDeck(t *testing.T) {
	r := expansionLobby()
	r.Config = ExpansionConfig{Board: "tuscany", Structures: true, Visitors: "ee"}
	if err := r.InitDecks(); err != nil {
		t.Fatal(err)
	}
	if len(r.Decks["structure"]) != 36 {
		t.Fatalf("structure deck=%d", len(r.Decks["structure"]))
	}
	seen := map[string]bool{}
	for _, c := range StructureCatalog() {
		if seen[c.ID] || c.RuleSource == "" || c.Description == "" || c.StructureCost <= 0 {
			t.Fatalf("invalid structure %v", c)
		}
		seen[c.ID] = true
	}
	for i := 0; i < 36; i++ {
		c, ok := r.take("structure")
		if !ok || seen["drawn:"+c.ID] {
			t.Fatal("finite structure deck")
		}
		seen["drawn:"+c.ID] = true
		r.discard(c)
	}
	if _, ok := r.take("structure"); !ok {
		t.Fatal("discard should recycle once")
	}
}
