package game

import "testing"

// Printed School card: train one worker for free, usable this year; +1 lira.
// Publisher card face mirrored by Boardspace:
// https://raw.githubusercontent.com/ddyer0/boardspace.net/main/client/boardspace-java/boardspace-games/viticulture/images/structurecards-cropped/TuscanyEssStructureCards-Page-000007.jpg
func TestSchoolPrintedFreeTrainingAndSpecialSurcharge(t *testing.T) {
	for _, special := range []bool{false, true} {
		for _, academy := range []bool{false, true} {
			r := specialWorkerRoom("oracle", "farmer")
			r.Config.Structures = true
			p, q := r.Players[0], r.Players[1]
			p.Buildings = []string{"school"}
			p.StructureSlots = []string{"school", ""}
			if academy {
				q.Buildings = []string{"academy"}
			}
			p.Coins = 0
			cost := 0
			id := "regular"
			if special {
				cost++
				id = "oracle"
			}
			if academy {
				cost++
			}
			p.Coins = cost
			beforeWorkers, beforeOther := p.Workers, q.Coins
			a := Action{Type: "place", Space: structureActionSpaceID(p.ID, "school"), SpecialWorker: id}
			if cost > 0 {
				p.Coins = cost - 1
				rejectUnchangedT(t, r, p.ID, a)
				p.Coins = cost
			}
			mustApplyT(t, r, p.ID, a)
			if p.Coins != 1 || p.TotalWorkers != 4 {
				t.Fatalf("school cost/reward: special=%v academy=%v p=%+v", special, academy, p)
			}
			if special {
				if p.Workers != beforeWorkers-1 || !p.specialWorkerReady(id, r.Year) {
					t.Fatal("special not immediately ready or extra ordinary worker")
				}
			} else if p.Workers != beforeWorkers {
				t.Fatal("ordinary trained worker not ready")
			}
			if q.Coins-beforeOther != cost-boolInt(special) {
				t.Fatal("Academy did not receive its fee")
			}
		}
	}
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func TestSchoolZeroCoinOffersRegularAndOneCoinOffersSpecial(t *testing.T) {
	r := specialWorkerRoom("oracle", "farmer")
	r.Config.Structures = true
	p := r.Players[0]
	p.Buildings = []string{"school"}
	p.StructureSlots = []string{"school", ""}
	p.Coins = 1
	mustApplyT(t, r, p.ID, Action{Type: "place", Space: structureActionSpaceID(p.ID, "school")})
	if len(r.Choices) != 1 || r.Choices[0].ActionSpace != "school_training" {
		t.Fatal("School choice lacks immediate-training identity")
	}
	r = restoreT(t, r)
	p = r.Player(p.ID)
	choiceT(t, r, Action{Option: "oracle"})
	if p.Coins != 1 || !p.specialWorkerReady("oracle", r.Year) {
		t.Fatal("restored School did not charge only special surcharge")
	}
}
