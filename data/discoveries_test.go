package data

import (
	"encoding/json"
	"fmt"
	"testing"

	"labsim/internal/content"
	"labsim/internal/experiment"
)

func TestEveryDiscoveryIsCataloguedWithAFact(t *testing.T) {
	var doc struct {
		Bonus       int
		Discoveries []struct{ ID, Science, Name string }
	}
	if err := json.Unmarshal(Discoveries, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Bonus <= 0 {
		t.Fatal("discovery bonus must be positive (P-2)")
	}
	cat := map[string]bool{}
	for _, d := range doc.Discoveries {
		cat[d.ID] = true
	}
	txt, _ := content.Load(Text)
	var ids []string
	for _, k := range experiment.SlimeKinds {
		ids = append(ids, "slime."+k)
	}
	for _, m := range experiment.Materials(true) {
		ids = append(ids, "circuit."+m.ID)
		if !txt.Has("mat." + m.ID) {
			t.Errorf("no name for material %s", m.ID)
		}
	}
	for _, s := range experiment.Slides(true) {
		ids = append(ids, "micro."+s.ID)
		for _, k := range []string{"micro.slide." + s.ID, "micro.hint." + s.ID} {
			if !txt.Has(k) {
				t.Errorf("missing text %s", k)
			}
		}
	}
	for _, o := range experiment.Options {
		if !txt.Has("micro.opt." + o) {
			t.Errorf("missing option text %s", o)
		}
	}
	for _, p := range experiment.Planets(true) {
		ids = append(ids, "gravity."+p.ID)
		if !txt.Has("planet." + p.ID) {
			t.Errorf("no name for planet %s", p.ID)
		}
	}
	for _, c := range experiment.Crystals() {
		ids = append(ids, "crystal."+c.ID)
		need(t, txt, "powder."+c.ID)
	}
	for _, r := range experiment.Rides() {
		ids = append(ids, "gforce."+r.ID)
		need(t, txt, "ride."+r.ID)
	}
	for _, q := range experiment.TriviaBank() {
		ids = append(ids, "trivia."+q.ID)
		need(t, txt, "trivia."+q.ID+".q")
		for i := 0; i < q.N; i++ {
			need(t, txt, fmt.Sprintf("trivia.%s.a%d", q.ID, i))
		}
	}
	for _, f := range experiment.Flowers {
		ids = append(ids, "garden."+f)
		need(t, txt, "flower."+f)
	}
	ids = append(ids, "garden.full")
	for _, c := range experiment.RobotChallenges(true) {
		ids = append(ids, "robot."+c.ID)
		need(t, txt, "robot.ch."+c.ID)
		need(t, txt, "robot.hint."+c.ID)
		need(t, txt, "robot.goal."+c.ID)
	}
	for _, slot := range experiment.RobotSlots {
		need(t, txt, "robot.slot."+slot)
		for _, p := range experiment.RobotParts[slot] {
			need(t, txt, "robot.part."+p)
		}
	}
	for _, id := range ids {
		if !cat[id] {
			t.Errorf("discovery %q not in discoveries.json", id)
		}
		if !txt.Has("fact." + id) {
			t.Errorf("discovery %q has no fun fact (E-5)", id)
		}
	}
	if len(ids) != len(doc.Discoveries) {
		t.Errorf("catalog has %d discoveries, experiments make %d", len(doc.Discoveries), len(ids))
	}
}

func need(t *testing.T, txt *content.Catalog, key string) {
	t.Helper()
	if !txt.Has(key) {
		t.Errorf("missing text %s", key)
	}
}
