package ui

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/geom"
)

// Heart draws a little heart centered at (x, y).
func Heart(dst *ebiten.Image, x, y, s float64) {
	c := Hex("#ff5a8a")
	Circle(dst, x-s*0.25, y-s*0.1, s*0.3, c)
	Circle(dst, x+s*0.25, y-s*0.1, s*0.3, c)
	Poly(dst, c, x-s*0.53, y, x+s*0.53, y, x, y+s*0.55)
}

// DrawAnimal draws a lab animal standing at (x, y) (C-4). Animals are
// always shown happy and cared for (S-6).
func DrawAnimal(dst *ebiten.Image, kind string, x, y, u, t, facing float64, moving bool) {
	hop := 0.0
	if moving {
		hop = math.Abs(math.Sin(t*8)) * 2 * u
	}
	switch kind {
	case "bunny":
		Shadow(dst, x, y, 14*u)
		y -= hop
		c := Hex("#f8f4ee")
		Circle(dst, x-facing*6*u, y-5*u, 2.5*u, Hex("#ffffff"))
		RoundRect(dst, geom.Rect{X: x - 6*u, Y: y - 10*u, W: 12 * u, H: 9 * u}, 4.5*u, c, 0.8*u, colOutline)
		hx := x + facing*4*u
		for _, dx := range []float64{-1.8, 1.8} {
			RoundRect(dst, geom.Rect{X: hx + dx*u - 1.4*u, Y: y - 20*u, W: 2.8 * u, H: 9 * u}, 1.4*u, c, 0.6*u, colOutline)
			RoundRect(dst, geom.Rect{X: hx + dx*u - 0.6*u, Y: y - 18.5*u, W: 1.2 * u, H: 6 * u}, 0.6*u, Hex("#ffb3c8"), 0, nil)
		}
		Circle(dst, hx, y-11*u, 4.5*u, c)
		Circle(dst, hx+facing*1.6*u, y-11.5*u, 0.8*u, colEye)
		Circle(dst, hx+facing*4*u, y-10.5*u, 0.7*u, Hex("#ff8aa8"))
	case "hamster":
		Shadow(dst, x, y, 12*u)
		y -= hop * 0.5
		c := Hex("#e6a65d")
		Circle(dst, x-3.5*u, y-9*u, 1.8*u, c)
		Circle(dst, x+3.5*u, y-9*u, 1.8*u, c)
		RoundRect(dst, geom.Rect{X: x - 6.5*u, Y: y - 9*u, W: 13 * u, H: 9 * u}, 4.5*u, c, 0.8*u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 3.5*u, Y: y - 5*u, W: 7 * u, H: 4.5 * u}, 2*u, Hex("#fff3e0"), 0, nil)
		Circle(dst, x-2*u+facing*u, y-6.5*u, 0.8*u, colEye)
		Circle(dst, x+2*u+facing*u, y-6.5*u, 0.8*u, colEye)
	case "butterfly":
		fy := y - 18*u + math.Sin(t*3+x)*3*u
		flap := 0.35 + 0.65*math.Abs(math.Sin(t*10+x))
		for _, side := range []float64{-1, 1} {
			wx := x + side*4*u*flap
			Circle(dst, wx, fy-2*u, 3.6*u*flap+0.5, Hex("#ff9a3c"))
			Circle(dst, wx, fy+2.5*u, 2.6*u*flap+0.5, Hex("#c89bff"))
		}
		RoundRect(dst, geom.Rect{X: x - 0.8*u, Y: fy - 4*u, W: 1.6 * u, H: 8 * u}, 0.8*u, colEye, 0, nil)
	case "axolotl":
		c := Hex("#ffb3cf")
		wig := math.Sin(t*4) * 1.5 * u
		RoundRect(dst, geom.Rect{X: x - 8*u, Y: y - 5*u, W: 14 * u, H: 5 * u}, 2.5*u, c, 0.6*u, colOutline)
		Poly(dst, c, x-facing*7*u, y-3*u, x-facing*13*u, y-5*u+wig, x-facing*13*u, y-1*u+wig)
		hx := x + facing*7*u
		Circle(dst, hx, y-4*u, 3.6*u, c)
		for _, a := range []float64{-0.9, -0.3, 0.3} {
			for _, side := range []float64{-1, 1} {
				Line(dst, hx+side*2.5*u, y-4*u, hx+side*(4.5*u)+math.Sin(a)*u, y-4*u+a*4*u-2*u, 0.9*u, Hex("#ff6fa0"))
			}
		}
		Circle(dst, hx-1.2*u, y-4.5*u, 0.6*u, colEye)
		Circle(dst, hx+1.2*u, y-4.5*u, 0.6*u, colEye)
	}
}

// Tank draws the axolotl's fish tank occupying r.
func Tank(dst *ebiten.Image, r geom.Rect, t float64) {
	RoundRect(dst, r, 6, Hex("#bfe8ff"), 2, colOutline)
	Fill(dst, geom.Rect{X: r.X + 3, Y: r.Y + r.H - 10, W: r.W - 6, H: 7}, 0, 0, Hex("#f2d9a0"))
	Line(dst, r.X+14, r.Y+r.H-8, r.X+12+math.Sin(t*2)*2, r.Y+12, 3, Hex("#3aa858"))
	Line(dst, r.X+20, r.Y+r.H-8, r.X+22+math.Sin(t*2+1)*2, r.Y+20, 3, Hex("#3aa858"))
	Circle(dst, r.X+r.W-16, r.Y+14-math.Mod(t*8, 10), 2, Hex("#ffffff"))
}
