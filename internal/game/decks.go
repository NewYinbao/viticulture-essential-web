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
func expansionCatalog(config ExpansionConfig) []Card {
	c := Catalog()
	if config.Normalized().Visitors == "rhine" {
		out := []Card{}
		for _, card := range c {
			if card.Type != "summer" && card.Type != "winter" {
				out = append(out, card)
			}
		}
		c = append(out, RhineCatalog(config)...)
	}
	if config.Visitors == "ee_moor" {
		c = append(c, MoorCatalog()...)
	}
	if config.Structures {
		c = append(c, StructureCatalog()...)
	}
	return c
}
func shuffle(c []Card) {
	for i := len(c) - 1; i > 0; i-- {
		j := rnd(i + 1)
		c[i], c[j] = c[j], c[i]
	}
}

// CatalogForConfig refuses missing expansion decks; replacing Rhine with EE is
// never a valid fallback, even for non-HTTP callers.
func CatalogForConfig(config ExpansionConfig) ([]Card, error) {
	if err := config.Playable(); err != nil {
		return nil, err
	}
	return expansionCatalog(config.Normalized()), nil
}

func (r *Room) InitDecks() error {
	cards, err := CatalogForConfig(r.Config)
	if err != nil {
		return err
	}
	r.Decks = map[string][]Card{}
	r.Discards = map[string][]Card{}
	for _, c := range cards {
		r.Decks[c.Type] = append(r.Decks[c.Type], c)
	}
	for _, d := range r.Decks {
		shuffle(d)
	}
	for _, p := range r.Players {
		for _, b := range p.StructureSlots {
			if d, ok := structureDef(b); ok {
				r.ensureStructureActionSpace(p.ID, d)
			}
		}
		for _, f := range p.Fields {
			if d, ok := structureDef(f.Structure); ok {
				r.ensureStructureActionSpace(p.ID, d)
			}
		}
	}
	return nil
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
		// Mercado is a persistent enhancement: it also applies to orders drawn
		// by a Tuscany wake/pass or year-end residual effect, where there is no
		// visitor ActionContext. Queue the choice after any currently resolving
		// choice so the draw itself and the optional delivery remain resumable.
		if t == "order" && has(p, "mercado") {
			drawn := probePlayer(p)
			drawn.Hand = []Card{c}
			if r.canFillVisitor(drawn) {
				r.enqueue(p.ID, "structure_mercado", []string{"use", "skip"}, 1)
				r.Choices[len(r.Choices)-1].CardIDs = []string{c.ID}
			}
		}
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
