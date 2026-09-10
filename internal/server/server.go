package server

import (
	"fmt"
	"sync"
	"vineyard/internal/game"
	"vineyard/internal/store"
)

type Session struct {
	Code     string
	PlayerID string `json:"playerId"`
}
type Store struct {
	Rooms    map[string]*game.Room
	Sessions map[string]Session
}
type App struct {
	mu    sync.Mutex
	Store Store
	dir   string
	subs  map[chan struct{}]string
}

func loadApp(dir string) (*App, error) {
	a := &App{dir: dir, subs: make(map[chan struct{}]string), Store: Store{Rooms: map[string]*game.Room{}, Sessions: map[string]Session{}}}
	if e := store.Load(dir, &a.Store); e != nil {
		return nil, e
	}
	if a.Store.Rooms == nil || a.Store.Sessions == nil {
		return nil, fmt.Errorf("invalid save")
	}
	return a, nil
}
func (a *App) save() error { return store.Save(a.dir, a.Store) }
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

// New loads the durable application state.
func New(dir string) (*App, error) { return loadApp(dir) }
