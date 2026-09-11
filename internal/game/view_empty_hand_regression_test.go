package game

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEmptyPrivateHandRemainsJSONArrayAfterRestore(t *testing.T) {
	r, p := moorFixture("moor-winter-16")
	p.Hand = nil
	r.Choices = nil
	r.Context.Remaining = 1
	r.visitorPrompt(p, "", "second", []string{"play", "skip"})
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var restored Room
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	for _, room := range []*Room{r, &restored} {
		v := room.View(p.ID)
		hand, err := json.Marshal(v["hand"])
		if err != nil || !bytes.Equal(hand, []byte("[]")) {
			t.Fatalf("empty private hand must be [], got %s (%v)", hand, err)
		}
		if v["pendingChoice"].(Choice).Visitor.Stage != "second" {
			t.Fatal("lost second visitor choice")
		}
	}
}
