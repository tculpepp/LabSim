package data

import (
	"testing"

	"labsim/internal/geom"
	"labsim/internal/items"
)

func TestItemsAndSpawns(t *testing.T) {
	cat, err := items.Load(Items)
	if err != nil {
		t.Fatalf("items.json: %v", err)
	}
	if _, ok := cat.Get("star_dust"); !ok {
		t.Fatal("star dust (designer's map item) missing")
	}
	m := loadMap(t)
	rooms := map[string]geom.Rect{}
	for _, r := range m.Rooms {
		rooms[r.ID] = r.Bounds
	}
	rules, err := items.LoadRules(Spawns, cat, rooms)
	if err != nil {
		t.Fatalf("spawns.json: %v", err)
	}
	want := 0
	for _, r := range rules {
		want += r.Count
	}
	// Every pickup must fit somewhere walkable, on any day.
	free := func(x, y float64) bool { return !m.Collides(geom.Centered(x, y, 28, 28), nil) }
	for day := 1; day <= 30; day++ {
		if got := len(items.Generate(rules, rooms, day, free)); got != want {
			t.Fatalf("day %d: %d of %d pickups placed", day, got, want)
		}
	}
}
