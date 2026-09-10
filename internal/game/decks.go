package game

import (
	"encoding/json"
)

func Catalog() []Card {
	var c []Card
	if e := json.Unmarshal(cardJSON, &c); e != nil {
		panic(e)
	}
	for i := range c {
		if d, ok := visitorDefs[c[i].ID]; ok {
			c[i].Implemented = true
			c[i].Description = d.Summary
		}
	}
	return c
}
func shuffle(c []Card) {
	for i := len(c) - 1; i > 0; i-- {
		j := rnd(i + 1)
		c[i], c[j] = c[j], c[i]
	}
}
func (r *Room) InitDecks() {
	r.Decks = map[string][]Card{}
	r.Discards = map[string][]Card{}
	for _, c := range Catalog() {
		r.Decks[c.Type] = append(r.Decks[c.Type], c)
	}
	for _, d := range r.Decks {
		shuffle(d)
	}
}
func (r *Room) take(t string) (Card, bool) {
	if len(r.Decks[t]) == 0 && len(r.Discards[t]) > 0 {
		r.Decks[t] = r.Discards[t]
		r.Discards[t] = nil
		shuffle(r.Decks[t])
	}
	d := r.Decks[t]
	if len(d) == 0 {
		return Card{}, false
	}
	c := d[len(d)-1]
	r.Decks[t] = d[:len(d)-1]
	return c, true
}
func (r *Room) draw(p *Player, t string) {
	if c, ok := r.take(t); ok {
		p.Hand = append(p.Hand, c)
	}
}
func (r *Room) discard(c Card) {
	if r.Discards == nil {
		r.Discards = map[string][]Card{}
	}
	r.Discards[c.Type] = append(r.Discards[c.Type], c)
}
func (r *Room) deckCounts() map[string]map[string]int {
	out := map[string]map[string]int{}
	for t, d := range r.Decks {
		out[t] = map[string]int{"deck": len(d), "discard": len(r.Discards[t])}
	}
	return out
}
