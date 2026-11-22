package world

import (
	"encoding/json"
	"fmt"
	"math"

	"labsim/internal/geom"
)

const (
	DefaultStationW = 48.0
	DefaultStationH = 32.0
	doorClearance   = 30.0 // stations must leave doorways walkable
)

// Station is a piece of furniture or a character the player can use:
// experiment benches, NPCs, the wardrobe, the decorating table.
type Station struct {
	ID    string  `json:"id"`
	Kind  string  `json:"kind"`
	Room  string  `json:"room"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w,omitempty"`
	H     float64 `json:"h,omitempty"`
	Label string  `json:"label,omitempty"`
}

func (s Station) Rect() geom.Rect { return geom.Centered(s.X, s.Y, s.W, s.H) }

// LoadStations parses {"stations": [...]}, filling default sizes.
func LoadStations(b []byte) ([]Station, error) {
	var doc struct {
		Stations []Station `json:"stations"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("%w: stations: %v", ErrInvalidMap, err)
	}
	for i := range doc.Stations {
		s := &doc.Stations[i]
		if s.W == 0 {
			s.W = DefaultStationW
		}
		if s.H == 0 {
			s.H = DefaultStationH
		}
	}
	return doc.Stations, nil
}

// AddStations validates stations against the map and makes them solid.
func (m *Map) AddStations(list []Station) error {
	for _, s := range list {
		if s.ID == "" || s.Kind == "" {
			return invalid("station needs id and kind: %+v", s)
		}
		if m.Station(s.ID) != nil {
			return invalid("duplicate station %q", s.ID)
		}
		r := m.Room(s.Room)
		if r == nil {
			return invalid("station %q: unknown room %q", s.ID, s.Room)
		}
		rect := s.Rect()
		if !r.Bounds.Expand(-WallThickness).ContainsRect(rect) {
			return invalid("station %q is not inside %q", s.ID, s.Room)
		}
		for _, d := range m.doors {
			if rect.Intersects(d.Rect.Expand(doorClearance)) {
				return invalid("station %q blocks a door of %q", s.ID, d.RoomID)
			}
		}
		for _, o := range m.stations {
			if rect.Intersects(o.Rect()) {
				return invalid("stations %q and %q overlap", o.ID, s.ID)
			}
		}
		m.stations = append(m.stations, s)
		m.obstacles = append(m.obstacles, rect)
	}
	return nil
}

func (m *Map) Stations() []Station { return m.stations }

func (m *Map) Station(id string) *Station {
	for i := range m.stations {
		if m.stations[i].ID == id {
			return &m.stations[i]
		}
	}
	return nil
}

// StationNear returns the closest station within reach of box, or nil.
func (m *Map) StationNear(box geom.Rect, reach float64) *Station {
	area := box.Expand(reach)
	bx, by := box.Center()
	var best *Station
	bestD := math.Inf(1)
	for i := range m.stations {
		s := &m.stations[i]
		if !area.Intersects(s.Rect()) {
			continue
		}
		if d := math.Hypot(s.X-bx, s.Y-by); d < bestD {
			best, bestD = s, d
		}
	}
	return best
}
