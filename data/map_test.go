package data

import (
	"testing"

	"labsim/internal/geom"
	"labsim/internal/world"
)

const playerSize = 24.0

func loadMap(t *testing.T) *world.Map {
	t.Helper()
	m, err := world.Load(Map)
	if err != nil {
		t.Fatalf("map.json invalid: %v", err)
	}
	return m
}

// M-2: the 10 rooms the designer confirmed.
func TestMapHasConfirmedRooms(t *testing.T) {
	m := loadMap(t)
	want := []string{"main_hall", "office", "shop", "chemistry", "physics",
		"greenhouse", "space_lab", "storage", "my_lab", "robot_workshop"}
	if len(m.Rooms) != len(want) {
		t.Fatalf("rooms = %d, want %d", len(m.Rooms), len(want))
	}
	for _, id := range want {
		if m.Room(id) == nil {
			t.Errorf("missing room %q", id)
		}
	}
	ws := m.Room("robot_workshop")
	if ws == nil || !ws.Locked || ws.UnlockPrice <= 0 {
		t.Fatalf("robot workshop must be locked with a price: %+v", ws)
	}
}

func TestSpawnInMainHallAndFree(t *testing.T) {
	m := loadMap(t)
	if r := m.RoomAt(m.Spawn.X, m.Spawn.Y); r == nil || r.ID != "main_hall" {
		t.Fatalf("spawn room = %v", r)
	}
	if m.Collides(geom.Centered(m.Spawn.X, m.Spawn.Y, playerSize, playerSize), nil) {
		t.Fatal("spawn is inside a wall")
	}
}

// Every room center must be reachable on foot from spawn (locked rooms once unlocked).
func TestAllRoomsReachable(t *testing.T) {
	m := loadMap(t)
	for _, unlockedAll := range []bool{false, true} {
		var unlocked world.UnlockedFunc = func(string) bool { return unlockedAll }
		reached := flood(m, unlocked)
		for _, r := range m.Rooms {
			cx, cy := r.Bounds.Center()
			got := reached[cell(cx, cy)]
			want := !r.Locked || unlockedAll
			if got != want {
				t.Errorf("unlocked=%v room %q reachable=%v want %v", unlockedAll, r.ID, got, want)
			}
		}
	}
}

const step = 8.0

type key struct{ x, y int }

func cell(x, y float64) key { return key{int(x / step), int(y / step)} }

func flood(m *world.Map, unlocked world.UnlockedFunc) map[key]bool {
	seen := map[key]bool{}
	start := cell(m.Spawn.X, m.Spawn.Y)
	queue := []key{start}
	seen[start] = true
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range []key{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			n := key{c.x + d.x, c.y + d.y}
			if seen[n] {
				continue
			}
			cx, cy := (float64(n.x)+0.5)*step, (float64(n.y)+0.5)*step
			if m.Collides(geom.Centered(cx, cy, playerSize, playerSize), unlocked) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	return seen
}
