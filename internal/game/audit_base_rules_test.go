//go:build ee_rule_audit

package game

import (
	"testing"
)

// Independent normative audit tests. Opt in: go test -tags ee_rule_audit -run TestAudit -v .
// They intentionally fail on the audited baseline; do not weaken these to preserve a preview.
func TestAuditSpringLeaderMustRotateOppositeWakeSelection(t *testing.T) {
	ps := []*Player{NewPlayer("A"), NewPlayer("B"), NewPlayer("C")}
	r := &Room{Players: ps, Year: 1, SpringLeader: 0}
	r.finishYear()
	// Wake selection uses +1 through Players. End-of-year first-player token must use -1.
	if r.SpringLeader != 2 || r.TurnID != ps[2].ID {
		t.Fatalf("FAQ: counter-clockwise first-player rotation required; got leader=%d turn=%s, want C (index 2)", r.SpringLeader, r.TurnID)
	}
}

func TestAuditCottageCannotRevealBeforeSecondColorCommitted(t *testing.T) {
	p, q := NewPlayer("cottage owner"), NewPlayer("other")
	p.Buildings = []string{"cottage"}
	p.Wake, q.Wake = 1, 2
	p.Passed, q.Passed = true, true
	r := &Room{Players: []*Player{p, q}, Phase: "summer", TurnID: q.ID,
		Decks: map[string][]Card{"summer": {{ID: "audit-s", Type: "summer"}}, "winter": {{ID: "audit-w", Type: "winter"}}}}
	r.next()
	if len(r.Choices) == 0 {
		t.Fatal("fall choice missing")
	}
	c := r.Choices[0]
	// A correct implementation can request both colors atomically, in which case this old
	// single-color command must be rejected. Otherwise it may defer drawing until commit.
	err := r.Apply(p.ID, Action{Type: "choose", ChoiceID: c.ID, Option: "summer"})
	if err != nil {
		return
	}
	if len(p.Hand) > 0 && len(r.Choices) > 0 && r.Choices[0].PlayerID == p.ID && r.Choices[0].Kind == "fall" {
		t.Fatal("FAQ: cottage exposes first card in player's hand while second color is still undecided")
	}
}
