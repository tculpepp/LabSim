package npc

import (
	"testing"

	"labsim/internal/geom"
)

func free(float64, float64) bool { return true }

func TestWandererStaysInAreaAndMoves(t *testing.T) {
	area := geom.Rect{X: 100, Y: 100, W: 300, H: 200}
	w := NewWanderer(area, 7, 1.2)
	sx, sy := w.X, w.Y
	moved := false
	for i := 0; i < 5000; i++ {
		w.Update(free)
		if !area.Contains(w.X, w.Y) {
			t.Fatalf("left area at step %d: %v,%v", i, w.X, w.Y)
		}
		if w.X != sx || w.Y != sy {
			moved = true
		}
	}
	if !moved {
		t.Fatal("never moved")
	}
}

func TestWandererDeterministic(t *testing.T) {
	area := geom.Rect{X: 0, Y: 0, W: 200, H: 200}
	a, b := NewWanderer(area, 3, 1), NewWanderer(area, 3, 1)
	for i := 0; i < 500; i++ {
		a.Update(free)
		b.Update(free)
	}
	if a.X != b.X || a.Y != b.Y {
		t.Fatal("same seed diverged")
	}
}

func TestWandererAvoidsBlocked(t *testing.T) {
	area := geom.Rect{X: 0, Y: 0, W: 200, H: 200}
	w := NewWanderer(area, 5, 1.5)
	w.X, w.Y = 50, 100
	blocked := func(x, _ float64) bool { return x < 100 } // right half is blocked
	for i := 0; i < 3000; i++ {
		w.Update(blocked)
		if w.X >= 100 {
			t.Fatalf("walked into blocked area: %v", w.X)
		}
	}
}

func TestWandererPausesSometimes(t *testing.T) {
	w := NewWanderer(geom.Rect{W: 300, H: 300}, 11, 1)
	still := 0
	for i := 0; i < 3000; i++ {
		px, py := w.X, w.Y
		w.Update(free)
		if px == w.X && py == w.Y {
			still++
		}
	}
	if still == 0 {
		t.Fatal("animals should stop and sniff around sometimes")
	}
}
