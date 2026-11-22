package world

import (
	"errors"
	"testing"

	"labsim/internal/geom"
)

func TestStationsBlockAndAreFound(t *testing.T) {
	m := load(t, testMap)
	st, err := LoadStations([]byte(`{"stations":[{"id":"desk","kind":"wardrobe","room":"hall","x":100,"y":60}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := m.AddStations(st); err != nil {
		t.Fatal(err)
	}
	if got := st[0].Rect(); got != geom.Centered(100, 60, DefaultStationW, DefaultStationH) {
		t.Fatalf("Rect = %+v", got)
	}
	// Walking up into the desk stops below it.
	box := geom.Centered(100, 120, 24, 24)
	moved := m.Move(box, 0, -60, nil)
	if moved.Y < 60+DefaultStationH/2 {
		t.Fatalf("walked through station: %+v", moved)
	}
	if s := m.StationNear(moved, 10); s == nil || s.ID != "desk" {
		t.Fatalf("StationNear = %v", s)
	}
	if m.StationNear(geom.Centered(100, 170, 24, 24), 10) != nil {
		t.Fatal("station found from too far")
	}
	if m.Station("desk") == nil || m.Station("nope") != nil {
		t.Fatal("Station lookup")
	}
}

func TestStationNearPicksClosest(t *testing.T) {
	m := load(t, testMap)
	st, _ := LoadStations([]byte(`{"stations":[
	  {"id":"a","kind":"x","room":"hall","x":60,"y":100,"w":20,"h":20},
	  {"id":"b","kind":"x","room":"hall","x":120,"y":100,"w":20,"h":20}]}`))
	if err := m.AddStations(st); err != nil {
		t.Fatal(err)
	}
	if s := m.StationNear(geom.Centered(100, 100, 10, 10), 30); s == nil || s.ID != "b" {
		t.Fatalf("closest = %v", s)
	}
}

func TestAddStationsValidation(t *testing.T) {
	cases := map[string]string{
		"unknown room": `{"stations":[{"id":"a","kind":"x","room":"nope","x":100,"y":100}]}`,
		"outside room": `{"stations":[{"id":"a","kind":"x","room":"hall","x":250,"y":100}]}`,
		"blocks door":  `{"stations":[{"id":"a","kind":"x","room":"hall","x":100,"y":180}]}`,
		"duplicate":    `{"stations":[{"id":"a","kind":"x","room":"hall","x":60,"y":60},{"id":"a","kind":"x","room":"hall","x":140,"y":60}]}`,
		"overlap":      `{"stations":[{"id":"a","kind":"x","room":"hall","x":60,"y":60},{"id":"b","kind":"x","room":"hall","x":70,"y":60}]}`,
		"missing kind": `{"stations":[{"id":"a","room":"hall","x":60,"y":60}]}`,
	}
	for name, src := range cases {
		m := load(t, testMap)
		st, err := LoadStations([]byte(src))
		if err == nil {
			err = m.AddStations(st)
		}
		if !errors.Is(err, ErrInvalidMap) {
			t.Errorf("%s: err = %v, want ErrInvalidMap", name, err)
		}
	}
}
