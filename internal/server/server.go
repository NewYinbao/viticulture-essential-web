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
	Rooms     map[string]*game.Room
	Sessions  map[string]Session
	Passwords map[string]PasswordRecord
}
type App struct {
	mu             sync.Mutex
	Store          Store
	dir            string
	subs           map[chan struct{}]string
	authMu         sync.Mutex
	authAttempts   map[string]authAttempt
	authSlots      chan struct{}
	enrollmentKeys map[string]string
}

func loadApp(dir string) (*App, error) {
	a := &App{dir: dir, subs: make(map[chan struct{}]string), Store: Store{Rooms: map[string]*game.Room{}, Sessions: map[string]Session{}}}
	if e := store.Load(dir, &a.Store); e != nil {
		return nil, e
	}
	if a.Store.Rooms == nil || a.Store.Sessions == nil {
		return nil, fmt.Errorf("invalid save")
	}
	if a.Store.Passwords == nil {
		a.Store.Passwords = make(map[string]PasswordRecord)
	}
	a.authAttempts = make(map[string]authAttempt)
	a.authSlots = make(chan struct{}, 2)
	a.enrollmentKeys = make(map[string]string)
	for _, room := range a.Store.Rooms {
		for _, player := range room.Players {
			if _, protected := a.Store.Passwords[player.ID]; !protected {
				a.enrollmentKeys[player.ID] = game.NewID()[:12]
			}
		}
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
