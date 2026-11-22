// Package npc moves the lab animals around (C-4).
package npc

import (
	"math"
	"math/rand/v2"

	"labsim/internal/geom"
)

// Wanderer walks to random spots in its area, pausing in between.
type Wanderer struct {
	X, Y   float64
	Facing float64 // -1 left, 1 right
	Moving bool

	area   geom.Rect
	speed  float64
	tx, ty float64
	wait   int
	r      *rand.Rand
}

func NewWanderer(area geom.Rect, seed uint64, speed float64) *Wanderer {
	w := &Wanderer{area: area, speed: speed, Facing: 1, r: rand.New(rand.NewPCG(seed, seed+99))}
	w.X, w.Y = w.randomPoint()
	w.tx, w.ty = w.X, w.Y
	return w
}

func (w *Wanderer) randomPoint() (float64, float64) {
	return w.area.X + w.r.Float64()*w.area.W, w.area.Y + w.r.Float64()*w.area.H
}

// Update moves one tick. ok reports whether a spot is walkable.
func (w *Wanderer) Update(ok func(x, y float64) bool) {
	w.Moving = false
	if w.wait > 0 {
		w.wait--
		return
	}
	dx, dy := w.tx-w.X, w.ty-w.Y
	d := math.Hypot(dx, dy)
	if d < w.speed {
		w.wait = 30 + w.r.IntN(150)
		for try := 0; try < 10; try++ {
			if x, y := w.randomPoint(); ok(x, y) {
				w.tx, w.ty = x, y
				break
			}
		}
		return
	}
	nx, ny := w.X+dx/d*w.speed, w.Y+dy/d*w.speed
	if !ok(nx, ny) || !w.area.Contains(nx, ny) {
		w.tx, w.ty = w.X, w.Y // give up on this target; pick another after a pause
		return
	}
	w.X, w.Y, w.Moving = nx, ny, true
	if math.Abs(dx) > 0.1 {
		w.Facing = math.Copysign(1, dx)
	}
}
