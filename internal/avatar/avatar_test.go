package avatar

import (
	"errors"
	"strings"
	"testing"
)

const sample = `{"parts":[
 {"id":"hat","name":"Hat","options":[{"id":"none","name":"None"},{"id":"cap","name":"Cap"},{"id":"crown","name":"Crown","price":100},{"id":"beanie","name":"Beanie"}]},
 {"id":"skin","name":"Skin","options":[{"id":"s1","name":"Peach","color":"#fbe0cc"},{"id":"s2","name":"Sand","color":"#f1c7a0"}]}
]}`

func load(t *testing.T) *Catalog {
	t.Helper()
	c, err := Load([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func none(string) bool { return false }

func TestDefaultPicksFirstFree(t *testing.T) {
	c := load(t)
	d := c.Default()
	if d["hat"] != "none" || d["skin"] != "s1" {
		t.Fatalf("Default = %v", d)
	}
}

func TestCycleSkipsUnownedPaid(t *testing.T) {
	c := load(t)
	av := Avatar{"hat": "cap", "skin": "s1"}
	av = c.Cycle(av, "hat", 1, none)
	if av["hat"] != "beanie" {
		t.Fatalf("next after cap = %q, want beanie (crown is locked)", av["hat"])
	}
	av = c.Cycle(av, "hat", 1, none)
	if av["hat"] != "none" {
		t.Fatalf("wraps to %q, want none", av["hat"])
	}
	av = c.Cycle(av, "hat", -1, none)
	if av["hat"] != "beanie" {
		t.Fatalf("backwards = %q", av["hat"])
	}
}

func TestCycleIncludesOwnedPaid(t *testing.T) {
	c := load(t)
	owned := func(k string) bool { return k == Key("hat", "crown") }
	av := c.Cycle(Avatar{"hat": "cap"}, "hat", 1, owned)
	if av["hat"] != "crown" {
		t.Fatalf("got %q, want crown", av["hat"])
	}
}

func TestCycleDoesNotMutateInput(t *testing.T) {
	c := load(t)
	in := Avatar{"hat": "none", "skin": "s1"}
	c.Cycle(in, "hat", 1, none)
	if in["hat"] != "none" {
		t.Fatal("Cycle mutated its input")
	}
}

func TestNormalizeFixesMissingInvalidAndUnowned(t *testing.T) {
	c := load(t)
	av := c.Normalize(Avatar{"hat": "crown", "skin": "nope", "bogus": "x"}, none)
	if av["hat"] != "none" || av["skin"] != "s1" {
		t.Fatalf("Normalize = %v", av)
	}
	if _, ok := av["bogus"]; ok {
		t.Fatal("unknown part kept")
	}
}

func TestOptionLookup(t *testing.T) {
	c := load(t)
	o, ok := c.Option("skin", "s2")
	if !ok || o.Color != "#f1c7a0" || o.Name != "Sand" {
		t.Fatalf("Option = %+v %v", o, ok)
	}
	if _, ok := c.Option("skin", "zzz"); ok {
		t.Fatal("unknown option found")
	}
}

func TestPricedOptions(t *testing.T) {
	c := load(t)
	p := c.Priced()
	if len(p) != 1 || p[0].Part != "hat" || p[0].ID != "crown" || p[0].Price != 100 {
		t.Fatalf("Priced = %+v", p)
	}
}

func TestLoadValidation(t *testing.T) {
	bad := map[string]string{
		"json":       `{`,
		"no free":    `{"parts":[{"id":"hat","name":"Hat","options":[{"id":"a","name":"A","price":5}]}]}`,
		"dup option": `{"parts":[{"id":"hat","name":"Hat","options":[{"id":"a","name":"A"},{"id":"a","name":"B"}]}]}`,
		"dup part":   strings.Replace(sample, `"id":"skin"`, `"id":"hat"`, 1),
		"bad color":  strings.Replace(sample, `#fbe0cc`, `peach`, 1),
		"empty":      `{"parts":[]}`,
	}
	for name, src := range bad {
		if _, err := Load([]byte(src)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}
