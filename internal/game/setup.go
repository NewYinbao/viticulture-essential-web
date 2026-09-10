package game

import (
	"strconv"
	"strings"
)

var mamaDraws = []string{"vso", "vsw", "vow", "sow", "vvs", "vvo", "vvw", "vss", "sso", "ssw", "voo", "soo", "oow", "vww", "sww", "oww", "vw", "so"}

type papaGift struct {
	Cash        int
	Gift        string
	Alternative int
}

var papaGifts = []papaGift{{4, "trellis", 2}, {3, "irrigation", 3}, {5, "yoke", 2}, {3, "medium_cellar", 4}, {2, "cottage", 4}, {1, "windmill", 5}, {0, "tasting_room", 6}, {5, "trellis", 1}, {4, "irrigation", 2}, {6, "yoke", 1}, {4, "medium_cellar", 3}, {3, "cottage", 3}, {2, "windmill", 4}, {1, "tasting_room", 5}, {2, "worker", 4}, {3, "worker", 3}, {4, "vp", 3}, {5, "vp", 2}}

func cardNumber(c Card) int {
	n, _ := strconv.Atoi(c.ID[strings.LastIndex(c.ID, "-")+1:])
	return n - 1
}
func (r *Room) setupParents() {
	r.Phase = "setup"
	r.Resume = "spring"
	for _, p := range r.Players {
		p.Coins = 0
		p.Hand = []Card{}
		p.Mama, _ = r.take("mama")
		p.Papa, _ = r.take("papa")
		for _, ch := range mamaDraws[cardNumber(p.Mama)] {
			r.draw(p, map[rune]string{'v': "vine", 's': "summer", 'o': "order", 'w': "winter"}[ch])
		}
		if cardNumber(p.Mama) >= 16 {
			p.Coins += 2
		}
		g := papaGifts[cardNumber(p.Papa)]
		p.Coins += g.Cash
	}
	// Resolve Papa choices clockwise, starting with the first-player marker.
	for i := range r.Players {
		p := r.Players[(r.SpringLeader+i)%len(r.Players)]
		r.enqueue(p.ID, "papa", []string{"gift", "coins"}, 1)
	}
	r.TurnID = r.Choices[0].PlayerID
}

func (r *Room) parentOptions(id string) any {
	p := r.Player(id)
	if p == nil || p.Papa.ID == "" {
		return nil
	}
	g := papaGifts[cardNumber(p.Papa)]
	return map[string]any{"gift": g.Gift, "coins": g.Alternative, "baseCoins": g.Cash}
}
