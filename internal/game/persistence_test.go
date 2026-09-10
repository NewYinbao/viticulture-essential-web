package game

import (
	"testing"
	"vineyard/internal/store"
)

// persistedRoom exercises the same JSON file writer used by the server while
// keeping rule tests independent of HTTP and session management.
func persistedRoom(t *testing.T, room *Room) *Room {
	t.Helper()
	dir := t.TempDir()
	if err := store.Save(dir, room); err != nil {
		t.Fatal(err)
	}
	restored := new(Room)
	if err := store.Load(dir, restored); err != nil {
		t.Fatal(err)
	}
	return restored
}
