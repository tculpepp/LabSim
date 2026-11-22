// Package world loads the lab map and handles collision, doors, and camera math.
package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"labsim/internal/geom"
)

const (
	WallThickness    = 8.0
	DefaultDoorWidth = 80.0
)

var ErrInvalidMap = errors.New("invalid map")

type Side string

const (
	Top    Side = "top"
	Bottom Side = "bottom"
	Left   Side = "left"
	Right  Side = "right"
)

// DoorSpec describes a door opening in a room wall. Offset is the distance
// from the start of the side (top/left) to the door center; 0 means centered.
type DoorSpec struct {
	Side   Side    `json:"side"`
	Offset float64 `json:"offset,omitempty"`
	Width  float64 `json:"width,omitempty"`
}

type Room struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Color       string     `json:"color"`
	Bounds      geom.Rect  `json:"bounds"`
	Locked      bool       `json:"locked,omitempty"`
	UnlockPrice int        `json:"unlockPrice,omitempty"`
	Doors       []DoorSpec `json:"doors"`
}

// Door is a built door opening, spanning the wall thickness.
type Door struct {
	RoomID string
	Side   Side
	Rect   geom.Rect
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// UnlockedFunc reports whether a locked room has been unlocked by the player.
type UnlockedFunc func(roomID string) bool

// Map is the lab layout. Walls are generated from room bounds and doors.
type Map struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Spawn  Point   `json:"spawn"`
	Rooms  []Room  `json:"rooms"`

	walls     []geom.Rect
	doors     []Door
	stations  []Station
	obstacles []geom.Rect
}

func invalid(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidMap, fmt.Sprintf(format, a...))
}

// Load parses and validates a map from JSON.
func Load(b []byte) (*Map, error) {
	var m Map
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMap, err)
	}
	if err := m.build(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Map) build() error {
	if m.Width <= 0 || m.Height <= 0 {
		return invalid("world size must be positive")
	}
	worldRect := geom.Rect{W: m.Width, H: m.Height}
	seen := map[string]bool{}
	for i, r := range m.Rooms {
		if r.ID == "" {
			return invalid("room %d has no id", i)
		}
		if seen[r.ID] {
			return invalid("duplicate room id %q", r.ID)
		}
		seen[r.ID] = true
		if r.Bounds.W < 4*WallThickness || r.Bounds.H < 4*WallThickness {
			return invalid("room %q is too small", r.ID)
		}
		if !worldRect.ContainsRect(r.Bounds) {
			return invalid("room %q is outside the world", r.ID)
		}
		if _, err := ParseHexColor(r.Color); err != nil {
			return invalid("room %q: %v", r.ID, err)
		}
		for _, prev := range m.Rooms[:i] {
			if prev.Bounds.Intersects(r.Bounds) {
				return invalid("rooms %q and %q overlap", prev.ID, r.ID)
			}
		}
		walls, doors, err := roomWalls(r)
		if err != nil {
			return err
		}
		m.walls = append(m.walls, walls...)
		m.doors = append(m.doors, doors...)
	}
	if !worldRect.Contains(m.Spawn.X, m.Spawn.Y) {
		return invalid("spawn is outside the world")
	}
	t := WallThickness
	m.walls = append(m.walls,
		geom.Rect{X: -t, Y: -t, W: m.Width + 2*t, H: t},
		geom.Rect{X: -t, Y: m.Height, W: m.Width + 2*t, H: t},
		geom.Rect{X: -t, Y: 0, W: t, H: m.Height},
		geom.Rect{X: m.Width, Y: 0, W: t, H: m.Height},
	)
	return nil
}

type span struct{ lo, hi float64 }

func sideLength(b geom.Rect, s Side) float64 {
	if s == Left || s == Right {
		return b.H
	}
	return b.W
}

func roomWalls(r Room) ([]geom.Rect, []Door, error) {
	b := r.Bounds
	gaps := map[Side][]span{}
	var doors []Door
	for _, d := range r.Doors {
		switch d.Side {
		case Top, Bottom, Left, Right:
		default:
			return nil, nil, invalid("room %q: bad door side %q", r.ID, d.Side)
		}
		w := d.Width
		if w == 0 {
			w = DefaultDoorWidth
		}
		length := sideLength(b, d.Side)
		c := d.Offset
		if c == 0 {
			c = length / 2
		}
		g := span{c - w/2, c + w/2}
		if g.lo < WallThickness || g.hi > length-WallThickness {
			return nil, nil, invalid("room %q: %s door does not fit", r.ID, d.Side)
		}
		for _, o := range gaps[d.Side] {
			if g.lo < o.hi && o.lo < g.hi {
				return nil, nil, invalid("room %q: overlapping %s doors", r.ID, d.Side)
			}
		}
		gaps[d.Side] = append(gaps[d.Side], g)
		doors = append(doors, Door{RoomID: r.ID, Side: d.Side, Rect: sideRect(b, d.Side, g)})
	}
	var walls []geom.Rect
	for _, s := range []Side{Top, Bottom, Left, Right} {
		gs := gaps[s]
		sort.Slice(gs, func(i, j int) bool { return gs[i].lo < gs[j].lo })
		pos, length := 0.0, sideLength(b, s)
		for _, g := range gs {
			if g.lo > pos {
				walls = append(walls, sideRect(b, s, span{pos, g.lo}))
			}
			pos = g.hi
		}
		if pos < length {
			walls = append(walls, sideRect(b, s, span{pos, length}))
		}
	}
	return walls, doors, nil
}

func sideRect(b geom.Rect, s Side, g span) geom.Rect {
	t := WallThickness
	switch s {
	case Top:
		return geom.Rect{X: b.X + g.lo, Y: b.Y, W: g.hi - g.lo, H: t}
	case Bottom:
		return geom.Rect{X: b.X + g.lo, Y: b.Y + b.H - t, W: g.hi - g.lo, H: t}
	case Left:
		return geom.Rect{X: b.X, Y: b.Y + g.lo, W: t, H: g.hi - g.lo}
	default:
		return geom.Rect{X: b.X + b.W - t, Y: b.Y + g.lo, W: t, H: g.hi - g.lo}
	}
}

// Walls returns all wall segments, including the world border.
func (m *Map) Walls() []geom.Rect { return m.walls }

// Doors returns all door openings.
func (m *Map) Doors() []Door { return m.doors }

func (m *Map) Room(id string) *Room {
	for i := range m.Rooms {
		if m.Rooms[i].ID == id {
			return &m.Rooms[i]
		}
	}
	return nil
}

// RoomAt returns the room containing the point, or nil for hallways.
func (m *Map) RoomAt(x, y float64) *Room {
	for i := range m.Rooms {
		if m.Rooms[i].Bounds.Contains(x, y) {
			return &m.Rooms[i]
		}
	}
	return nil
}

// DoorLocked reports whether a door currently blocks the player.
func (m *Map) DoorLocked(d Door, unlocked UnlockedFunc) bool {
	r := m.Room(d.RoomID)
	if r == nil || !r.Locked {
		return false
	}
	return unlocked == nil || !unlocked(d.RoomID)
}

// Solids returns every rect the player cannot pass: walls plus locked doors.
func (m *Map) Solids(unlocked UnlockedFunc) []geom.Rect {
	out := make([]geom.Rect, 0, len(m.walls)+len(m.obstacles)+len(m.doors))
	out = append(out, m.walls...)
	out = append(out, m.obstacles...)
	for _, d := range m.doors {
		if m.DoorLocked(d, unlocked) {
			out = append(out, d.Rect)
		}
	}
	return out
}

func (m *Map) Collides(box geom.Rect, unlocked UnlockedFunc) bool {
	for _, s := range m.Solids(unlocked) {
		if box.Intersects(s) {
			return true
		}
	}
	return false
}

// Move moves box by (dx, dy), sliding along walls. Large moves are split into
// sub-steps so fast movement cannot tunnel through thin walls.
func (m *Map) Move(box geom.Rect, dx, dy float64, unlocked UnlockedFunc) geom.Rect {
	solids := m.Solids(unlocked)
	steps := int(math.Ceil(math.Max(math.Abs(dx), math.Abs(dy)) / (WallThickness / 2)))
	if steps < 1 {
		steps = 1
	}
	sx, sy := dx/float64(steps), dy/float64(steps)
	for i := 0; i < steps; i++ {
		box = slide(box, sx, 0, solids)
		box = slide(box, 0, sy, solids)
	}
	return box
}

func slide(box geom.Rect, dx, dy float64, solids []geom.Rect) geom.Rect {
	if dx == 0 && dy == 0 {
		return box
	}
	box.X += dx
	box.Y += dy
	for _, s := range solids {
		if !box.Intersects(s) {
			continue
		}
		switch {
		case dx > 0:
			box.X = s.X - box.W
		case dx < 0:
			box.X = s.X + s.W
		case dy > 0:
			box.Y = s.Y - box.H
		case dy < 0:
			box.Y = s.Y + s.H
		}
	}
	return box
}

// DoorNear returns a door within reach of box, or nil.
func (m *Map) DoorNear(box geom.Rect, reach float64) *Door {
	area := box.Expand(reach)
	for i := range m.doors {
		if area.Intersects(m.doors[i].Rect) {
			d := m.doors[i]
			return &d
		}
	}
	return nil
}

// CameraFor returns the top-left camera offset that centers (tx, ty),
// clamped to the world. A world smaller than the view is centered.
func CameraFor(tx, ty, viewW, viewH, worldW, worldH float64) (float64, float64) {
	return camAxis(tx, viewW, worldW), camAxis(ty, viewH, worldH)
}

func camAxis(t, view, world float64) float64 {
	if world <= view {
		return (world - view) / 2
	}
	return math.Max(0, math.Min(t-view/2, world-view))
}

// ParseHexColor parses "#rrggbb".
func ParseHexColor(s string) ([3]uint8, error) {
	var c [3]uint8
	if len(s) != 7 || s[0] != '#' {
		return c, fmt.Errorf("color %q must look like #rrggbb", s)
	}
	for i := range c {
		v, err := strconv.ParseUint(s[1+2*i:3+2*i], 16, 8)
		if err != nil {
			return c, fmt.Errorf("color %q: %v", s, err)
		}
		c[i] = uint8(v)
	}
	return c, nil
}
