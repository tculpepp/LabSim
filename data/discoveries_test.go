package data

import (
	"encoding/json"
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
