package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyEnrollmentCodeRestartAndFailedSave(t *testing.T) {
	dir := t.TempDir()
	a, _ := loadApp(dir)
	r := testRoom()
	begin(t, r)
	a.Store.Rooms[r.Code] = r
	a.Store.Sessions["legacy-owner"] = Session{Code: r.Code, PlayerID: r.HostID}
	if err := a.save(); err != nil {
		t.Fatal(err)
	}
	a, _ = loadApp(dir)
	oldCode := a.enrollmentKeys[r.HostID]
	a, _ = loadApp(dir)
	code := a.enrollmentKeys[r.HostID]
	if len(code) != 12 || code == oldCode {
		t.Fatal("enrollment code did not rotate after restart")
	}
	before, _ := json.Marshal(a.Store)
	w := request(a.routes(), "POST", "/api/action", "legacy-owner", map[string]any{"type": "place", "space": "gain_coin", "revision": r.Revision})
	if w.Code != 403 {
		t.Fatalf("legacy action allowed: %d", w.Code)
	}
	after, _ := json.Marshal(a.Store)
	if string(before) != string(after) {
		t.Fatal("unenrolled action mutated store")
	}
	w = request(a.routes(), "POST", "/api/password", "legacy-owner", map[string]string{"password": testPassword, "enrollmentCode": oldCode})
	if w.Code != 403 {
		t.Fatal("previous boot's code accepted")
	}
	a.dir = filepath.Join(dir, "unwritable-file")
	if err := os.WriteFile(a.dir, []byte("file blocks directory"), 0600); err != nil {
		t.Fatal(err)
	}
	w = request(a.routes(), "POST", "/api/password", "legacy-owner", map[string]string{"password": testPassword, "enrollmentCode": code})
	if w.Code != 500 {
		t.Fatalf("save failure status %d", w.Code)
	}
	after, _ = json.Marshal(a.Store)
	if string(before) != string(after) || a.enrollmentKeys[r.HostID] != code {
		t.Fatal("failed save consumed credentials or code")
	}
	a.dir = dir
	w = request(a.routes(), "POST", "/api/password", "legacy-owner", map[string]string{"password": testPassword, "enrollmentCode": code})
	if w.Code != 200 || a.enrollmentKeys[r.HostID] != "" {
		t.Fatalf("enrollment retry failed %d", w.Code)
	}
	a, _ = loadApp(dir)
	if a.enrollmentKeys[r.HostID] != "" {
		t.Fatal("already protected seat received new enrollment code")
	}
}
