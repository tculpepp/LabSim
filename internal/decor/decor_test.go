package decor

import (
	"errors"
	"testing"
)

func TestPlaceMoveRemove(t *testing.T) {
	l := NewLayout(4, 3, []Cell{{0, 0}}, nil)
	if err := l.Place("plant", Cell{1, 1}); err != nil {
		t.Fatal(err)
	}
	if id, ok := l.At(Cell{1, 1}); !ok || id != "plant" {
		t.Fatal("At")
	}
	// Placing again moves it (each decoration exists once).
	if err := l.Place("plant", Cell{2, 2}); err != nil {
		t.Fatal(err)
	}
	if _, ok := l.At(Cell{1, 1}); ok {
		t.Fatal("old cell still filled")
	}
	if c, ok := l.Where("plant"); !ok || c != (Cell{2, 2}) {
		t.Fatal("Where")
	}
	l.Remove("plant")
	if _, ok := l.Where("plant"); ok || len(l.Placements()) != 0 {
		t.Fatal("Remove")
	}
}

func TestPlaceErrors(t *testing.T) {
	l := NewLayout(4, 3, []Cell{{0, 0}}, nil)
	l.Place("rug", Cell{3, 2})
	cases := map[error]Cell{ErrOutside: {4, 0}, ErrBlocked: {0, 0}, ErrTaken: {3, 2}}
	for want, c := range cases {
		if err := l.Place("lamp", c); !errors.Is(err, want) {
			t.Errorf("Place at %v = %v, want %v", c, err, want)
		}
	}
	if err := l.Place("lamp", Cell{-1, 0}); !errors.Is(err, ErrOutside) {
		t.Error("negative cell")
	}
	// Moving onto its own cell is fine.
	if err := l.Place("rug", Cell{3, 2}); err != nil {
		t.Errorf("re-place on own cell: %v", err)
	}
}

func TestRestoreDropsInvalid(t *testing.T) {
	l := NewLayout(4, 3, []Cell{{0, 0}}, []Placement{
		{"plant", Cell{1, 1}}, {"lamp", Cell{0, 0}}, {"rug", Cell{9, 9}}, {"globe", Cell{1, 1}},
	})
	p := l.Placements()
	if len(p) != 1 || p[0].ID != "plant" {
		t.Fatalf("Placements = %v", p)
	}
}

func TestPlacementsSorted(t *testing.T) {
	l := NewLayout(5, 5, nil, nil)
	l.Place("b", Cell{2, 2})
	l.Place("a", Cell{3, 1})
	p := l.Placements()
	if p[0].ID != "a" || p[1].ID != "b" {
		t.Fatalf("not sorted: %v", p)
	}
}
