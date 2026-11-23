package shop

import (
	"errors"
	"strings"
	"testing"
)

const src = `{"entries":[
 {"id":"spark_kit","kind":"tool","name":"Spark Kit","price":50},
 {"id":"bag_plus","kind":"backpack","name":"Bigger Backpack","price":40,"slots":2},
 {"id":"bag_big","kind":"backpack","name":"Huge Backpack","price":80,"slots":3,"requires":"bag_plus"},
 {"id":"robot_workshop","kind":"room","name":"Robot Workshop","price":135},
 {"id":"plant","kind":"decor","name":"Plant","price":15}
]}`

func load(t *testing.T) *Catalog {
	t.Helper()
	c, err := Load([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestKeys(t *testing.T) {
	c := load(t)
	want := map[string]string{"spark_kit": "tool:spark_kit", "bag_plus": "backpack:bag_plus", "robot_workshop": "room:robot_workshop", "plant": "decor:plant"}
	for id, k := range want {
		e, ok := c.Get(id)
		if !ok || e.Key() != k {
			t.Errorf("%s key = %q, want %q", id, e.Key(), k)
		}
	}
	if (Entry{Kind: KindOutfit, Part: "hat", Option: "crown"}).Key() != "avatar:hat:crown" {
		t.Error("outfit key should match avatar.Key")
	}
}

func TestOfKind(t *testing.T) {
	c := load(t)
	if got := c.OfKind(KindBackpack); len(got) != 2 || got[0].ID != "bag_plus" {
		t.Fatalf("OfKind = %v", got)
	}
}

func TestCheck(t *testing.T) {
	c := load(t)
	owned := map[string]bool{}
	owns := func(k string) bool { return owned[k] }
	kit, _ := c.Get("spark_kit")
	if err := Check(kit, 49, owns); !errors.Is(err, ErrTooPoor) {
		t.Fatalf("poor: %v", err)
	}
	if err := Check(kit, 50, owns); err != nil {
		t.Fatalf("ok: %v", err)
	}
	owned["tool:spark_kit"] = true
	if err := Check(kit, 500, owns); !errors.Is(err, ErrOwned) {
		t.Fatalf("owned: %v", err)
	}
	big, _ := c.Get("bag_big")
	if err := Check(big, 500, owns); !errors.Is(err, ErrNeedsFirst) {
		t.Fatalf("requires: %v", err)
	}
	owned["backpack:bag_plus"] = true
	if err := Check(big, 500, owns); err != nil {
		t.Fatalf("after requirement: %v", err)
	}
}

func TestLoadValidates(t *testing.T) {
	bad := map[string]string{
		"json":         `{`,
		"dup":          strings.Replace(src, `"id":"plant"`, `"id":"spark_kit"`, 1),
		"no price":     strings.Replace(src, `"price":15`, `"price":0`, 1),
		"bad kind":     strings.Replace(src, `"kind":"decor"`, `"kind":"cake"`, 1),
		"no slots":     strings.Replace(src, `"slots":2`, `"slots":0`, 1),
		"bad requires": strings.Replace(src, `"requires":"bag_plus"`, `"requires":"nope"`, 1),
	}
	for name, s := range bad {
		if _, err := Load([]byte(s)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
