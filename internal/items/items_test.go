package items

import (
	"errors"
	"testing"

	"labsim/internal/geom"
)

const catJSON = `{"items":[
 {"id":"star_dust","name":"Star Dust","desc":"Sparkly.","icon":"jar","color":"#ffe27a","sell":3},
 {"id":"battery","name":"Battery","desc":"Zap.","icon":"battery","color":"#ffb347","sell":5}
]}`

func catalog(t *testing.T) *Catalog {
	t.Helper()
	c, err := Load([]byte(catJSON))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCatalogGet(t *testing.T) {
	c := catalog(t)
	it, ok := c.Get("battery")
	if !ok || it.Name != "Battery" || it.Sell != 5 {
		t.Fatalf("Get = %+v %v", it, ok)
	}
	if _, ok := c.Get("nope"); ok {
		t.Fatal("unknown item found")
	}
	if len(c.All()) != 2 {
		t.Fatal("All")
	}
}

func TestCatalogValidation(t *testing.T) {
	for name, src := range map[string]string{
		"json":      `{`,
		"dup":       `{"items":[{"id":"a","name":"A","icon":"jar","color":"#ffffff"},{"id":"a","name":"B","icon":"jar","color":"#ffffff"}]}`,
		"no name":   `{"items":[{"id":"a","icon":"jar","color":"#ffffff"}]}`,
		"bad color": `{"items":[{"id":"a","name":"A","icon":"jar","color":"yellow"}]}`,
		"neg sell":  `{"items":[{"id":"a","name":"A","icon":"jar","color":"#ffffff","sell":-1}]}`,
	} {
		if _, err := Load([]byte(src)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestLoadRulesValidates(t *testing.T) {
	c := catalog(t)
	rooms := map[string]geom.Rect{"storage": {X: 0, Y: 0, W: 400, H: 300}}
	if _, err := LoadRules([]byte(`{"rules":[{"room":"storage","item":"battery","count":2}]}`), c, rooms); err != nil {
		t.Fatal(err)
	}
	for name, src := range map[string]string{
		"unknown item": `{"rules":[{"room":"storage","item":"nope","count":1}]}`,
		"unknown room": `{"rules":[{"room":"attic","item":"battery","count":1}]}`,
		"zero count":   `{"rules":[{"room":"storage","item":"battery","count":0}]}`,
	} {
		if _, err := LoadRules([]byte(src), c, rooms); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestGenerateDeterministicAndInsideRooms(t *testing.T) {
	rooms := map[string]geom.Rect{"storage": {X: 100, Y: 100, W: 400, H: 300}, "lab": {X: 600, Y: 100, W: 300, H: 300}}
	rules := []Rule{{Room: "storage", Item: "star_dust", Count: 4}, {Room: "lab", Item: "battery", Count: 2}}
	ok := func(float64, float64) bool { return true }
	a := Generate(rules, rooms, 3, ok)
	b := Generate(rules, rooms, 3, ok)
	if len(a) != 6 {
		t.Fatalf("got %d pickups, want 6", len(a))
	}
	keys := map[string]bool{}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("same day gave different pickups")
		}
		if keys[a[i].Key] {
			t.Fatalf("duplicate key %q", a[i].Key)
		}
		keys[a[i].Key] = true
		r := rooms["storage"]
		if a[i].ItemID == "battery" {
			r = rooms["lab"]
		}
		if !r.Expand(-SpawnMargin).Contains(a[i].X, a[i].Y) {
			t.Errorf("pickup %+v outside its room margin", a[i])
		}
	}
	c := Generate(rules, rooms, 4, ok)
	if c[0].Key == a[0].Key || (c[0].X == a[0].X && c[0].Y == a[0].Y) {
		t.Fatal("new day should give new keys and new spots")
	}
}

func TestGenerateRespectsBlockedSpots(t *testing.T) {
	rooms := map[string]geom.Rect{"storage": {X: 0, Y: 0, W: 400, H: 300}}
	rules := []Rule{{Room: "storage", Item: "star_dust", Count: 3}}
	// Only the right half of the room is free.
	got := Generate(rules, rooms, 1, func(x, _ float64) bool { return x > 200 })
	for _, p := range got {
		if p.X <= 200 {
			t.Fatalf("pickup placed on blocked spot: %+v", p)
		}
	}
	// Nowhere free: nothing spawns, no hang.
	if n := len(Generate(rules, rooms, 1, func(float64, float64) bool { return false })); n != 0 {
		t.Fatalf("spawned %d on fully blocked room", n)
	}
}

func TestGenerateSpreadsPickups(t *testing.T) {
	rooms := map[string]geom.Rect{"storage": {X: 0, Y: 0, W: 400, H: 300}}
	got := Generate([]Rule{{Room: "storage", Item: "star_dust", Count: 4}}, rooms, 9, func(float64, float64) bool { return true })
	for i := range got {
		for j := i + 1; j < len(got); j++ {
			dx, dy := got[i].X-got[j].X, got[i].Y-got[j].Y
			if dx*dx+dy*dy < MinSpacing*MinSpacing {
				t.Fatalf("pickups %d and %d too close", i, j)
			}
		}
	}
}
