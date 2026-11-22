package world

import (
	"errors"
	"strings"
	"testing"

	"labsim/internal/geom"
)

// testMap: one open room with a bottom door, one locked room with a left door.
const testMap = `{
  "width": 600, "height": 400,
  "spawn": {"x": 100, "y": 100},
  "rooms": [
    {"id": "hall", "name": "Hall", "color": "#ffeecc",
     "bounds": {"x": 0, "y": 0, "w": 200, "h": 200},
     "doors": [{"side": "bottom"}]},
    {"id": "vault", "name": "Vault", "color": "#aaaaaa", "locked": true, "unlockPrice": 50,
     "bounds": {"x": 300, "y": 0, "w": 200, "h": 200},
     "doors": [{"side": "left", "offset": 100, "width": 60}]}
  ]
}`

func load(t *testing.T, s string) *Map {
	t.Helper()
	m, err := Load([]byte(s))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return m
}

func noneUnlocked(string) bool { return false }
func allUnlocked(string) bool  { return true }

func TestLoadBuildsRoomsAndDoors(t *testing.T) {
	m := load(t, testMap)
	if len(m.Rooms) != 2 {
		t.Fatalf("rooms = %d", len(m.Rooms))
	}
	doors := m.Doors()
	if len(doors) != 2 {
		t.Fatalf("doors = %d, want 2", len(doors))
	}
	// Hall bottom door: centered, default width 80, at bottom wall.
	want := geom.Rect{X: 60, Y: 200 - WallThickness, W: 80, H: WallThickness}
	if doors[0].Rect != want {
		t.Fatalf("hall door = %+v, want %+v", doors[0].Rect, want)
	}
}

func TestLoadRejectsBadMaps(t *testing.T) {
	cases := map[string]string{
		"bad json":      `{`,
		"dup id":        strings.Replace(testMap, `"id": "vault"`, `"id": "hall"`, 1),
		"overlap":       strings.Replace(testMap, `"x": 300, "y": 0`, `"x": 150, "y": 0`, 1),
		"outside world": strings.Replace(testMap, `"x": 300, "y": 0`, `"x": 500, "y": 0`, 1),
		"door too wide": strings.Replace(testMap, `"width": 60}`, `"width": 500}`, 1),
		"bad side":      strings.Replace(testMap, `"side": "left"`, `"side": "up"`, 1),
		"bad color":     strings.Replace(testMap, `#aaaaaa`, `grey`, 1),
		"spawn outside": strings.Replace(testMap, `"x": 100, "y": 100`, `"x": -5, "y": 100`, 1),
		"zero size":     strings.Replace(testMap, `"width": 600`, `"width": 0`, 1),
		"missing id":    strings.Replace(testMap, `"id": "vault",`, ``, 1),
		"overlap doors": strings.Replace(testMap, `[{"side": "bottom"}]`, `[{"side": "bottom"},{"side": "bottom", "offset": 110}]`, 1),
		"room too small": strings.Replace(testMap, `"w": 200, "h": 200},
     "doors": [{"side": "bottom"}]`, `"w": 20, "h": 200},
     "doors": []`, 1),
	}
	for name, src := range cases {
		if _, err := Load([]byte(src)); !errors.Is(err, ErrInvalidMap) {
			t.Errorf("%s: err = %v, want ErrInvalidMap", name, err)
		}
	}
}

func TestRoomAt(t *testing.T) {
	m := load(t, testMap)
	if r := m.RoomAt(100, 100); r == nil || r.ID != "hall" {
		t.Fatalf("RoomAt hall = %v", r)
	}
	if r := m.RoomAt(250, 100); r != nil {
		t.Fatalf("hallway should be nil, got %v", r.ID)
	}
	if m.Room("vault") == nil || m.Room("nope") != nil {
		t.Fatal("Room lookup wrong")
	}
}

func TestMoveFreeSpace(t *testing.T) {
	m := load(t, testMap)
	box := geom.Centered(100, 100, 24, 24)
	got := m.Move(box, 5, -3, noneUnlocked)
	if got.X != box.X+5 || got.Y != box.Y-3 {
		t.Fatalf("moved to %+v", got)
	}
}

func TestMoveStopsAtWall(t *testing.T) {
	m := load(t, testMap)
	// Player near hall's right wall (inner face at x=192).
	box := geom.Rect{X: 160, Y: 50, W: 24, H: 24}
	got := m.Move(box, 20, 0, noneUnlocked)
	if got.X+got.W != 200-WallThickness {
		t.Fatalf("right edge = %v, want %v", got.X+got.W, 200-WallThickness)
	}
}

func TestMoveSlidesAlongWall(t *testing.T) {
	m := load(t, testMap)
	box := geom.Rect{X: 168, Y: 50, W: 24, H: 24} // touching right wall
	got := m.Move(box, 5, 5, noneUnlocked)
	if got.X != 168 || got.Y != 55 {
		t.Fatalf("slide = %+v, want x=168 y=55", got)
	}
}

func TestMoveThroughDoor(t *testing.T) {
	m := load(t, testMap)
	// Centered on hall bottom door (x 60..140), walking down out of the room.
	box := geom.Rect{X: 88, Y: 170, W: 24, H: 24}
	got := box
	for i := 0; i < 20; i++ {
		got = m.Move(got, 0, 3, noneUnlocked)
	}
	if got.Y <= 200 {
		t.Fatalf("did not exit through door, y=%v", got.Y)
	}
}

func TestLockedDoorBlocksUntilUnlocked(t *testing.T) {
	m := load(t, testMap)
	// In hallway, lined up with vault's left door (y 70..130), walking right.
	box := geom.Rect{X: 260, Y: 88, W: 24, H: 24}
	blocked := m.Move(box, 30, 0, noneUnlocked)
	if blocked.X+blocked.W > 300 {
		t.Fatalf("locked door let player in: %+v", blocked)
	}
	open := m.Move(box, 30, 0, allUnlocked)
	if open.X != 290 {
		t.Fatalf("unlocked door blocked: %+v", open)
	}
}

func TestMoveNoTunnelling(t *testing.T) {
	m := load(t, testMap)
	box := geom.Rect{X: 160, Y: 50, W: 24, H: 24}
	got := m.Move(box, 200, 0, noneUnlocked)
	if got.X+got.W > 200-WallThickness {
		t.Fatalf("tunnelled through wall: %+v", got)
	}
}

func TestWorldEdgesBlock(t *testing.T) {
	m := load(t, testMap)
	box := geom.Rect{X: 580, Y: 380, W: 10, H: 10}
	got := m.Move(box, 50, 50, noneUnlocked)
	if got.X+got.W > 600 || got.Y+got.H > 400 {
		t.Fatalf("escaped world: %+v", got)
	}
}

func TestCollides(t *testing.T) {
	m := load(t, testMap)
	if !m.Collides(geom.Rect{X: 195, Y: 50, W: 10, H: 10}, noneUnlocked) {
		t.Fatal("box on wall should collide")
	}
	if m.Collides(geom.Centered(100, 100, 24, 24), noneUnlocked) {
		t.Fatal("room center should be free")
	}
}

func TestDoorNear(t *testing.T) {
	m := load(t, testMap)
	box := geom.Rect{X: 270, Y: 88, W: 24, H: 24}
	d := m.DoorNear(box, 16)
	if d == nil || d.RoomID != "vault" {
		t.Fatalf("DoorNear = %+v", d)
	}
	if m.DoorNear(geom.Centered(100, 100, 24, 24), 16) != nil {
		t.Fatal("no door expected near hall center")
	}
}

func TestCameraFor(t *testing.T) {
	cases := []struct {
		name           string
		tx, ty         float64
		wantX, wantY   float64
		viewW, viewH   float64
		worldW, worldH float64
	}{
		{"centered", 500, 400, 200, 200, 600, 400, 1000, 1000},
		{"clamped top-left", 10, 10, 0, 0, 600, 400, 1000, 1000},
		{"clamped bottom-right", 990, 990, 400, 600, 600, 400, 1000, 1000},
		{"world smaller than view centers", 50, 50, -100, -50, 300, 200, 100, 100},
	}
	for _, c := range cases {
		x, y := CameraFor(c.tx, c.ty, c.viewW, c.viewH, c.worldW, c.worldH)
		if x != c.wantX || y != c.wantY {
			t.Errorf("%s: camera = %v,%v want %v,%v", c.name, x, y, c.wantX, c.wantY)
		}
	}
}

func TestParseHexColor(t *testing.T) {
	c, err := ParseHexColor("#3060d0")
	if err != nil || c != [3]uint8{0x30, 0x60, 0xd0} {
		t.Fatalf("ParseHexColor = %v, %v", c, err)
	}
	for _, bad := range []string{"", "3060d0", "#3060d", "#zzzzzz"} {
		if _, err := ParseHexColor(bad); err == nil {
			t.Errorf("ParseHexColor(%q) should fail", bad)
		}
	}
}
