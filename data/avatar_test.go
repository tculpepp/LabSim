package data

import (
	"testing"

	"labsim/internal/avatar"
	"labsim/internal/filter"
	"labsim/internal/names"
)

func TestAvatarCatalog(t *testing.T) {
	c, err := avatar.Load(Avatar)
	if err != nil {
		t.Fatalf("avatar.json invalid: %v", err)
	}
	// A-2..A-5: every customizable part the requirements list.
	for _, p := range []string{"body", "skin", "hair", "hairColor", "coat", "shirt", "shoes", "pattern", "eyewear", "hat", "gear", "pet"} {
		if _, ok := c.Default()[p]; !ok {
			t.Errorf("missing part %q", p)
		}
	}
	// The designer chose these four starter pets.
	for _, pet := range []string{"robot", "cat", "slime", "puppy"} {
		o, ok := c.Option("pet", pet)
		if !ok || o.Price != 0 {
			t.Errorf("starter pet %q missing or not free", pet)
		}
	}
}

// S-3: the safe name builder must never produce a blocked name.
func TestEveryPresetNamePassesFilter(t *testing.T) {
	b, err := names.Load(Names)
	if err != nil {
		t.Fatal(err)
	}
	all := b.All()
	if len(all) < 100 {
		t.Fatalf("only %d preset names", len(all))
	}
	for _, n := range all {
		if err := filter.Check(n); err != nil {
			t.Errorf("preset %q: %v", n, err)
		}
	}
}
