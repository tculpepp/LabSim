package scenes

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

// flatDecor lies on the floor and is drawn under everything else.
func flatDecor(id string) bool { return id == "rug" }

// drawDecor draws a decoration standing at bottom-centre (x, y), scale s
// (1 = one 40px grid cell in the world).
func drawDecor(screen *ebiten.Image, id string, x, y, s, t float64) {
	u := s * 2 // drawing units: a cell is 20u
	wall := ui.ColWall
	switch id {
	case "plant":
		ui.Shadow(screen, x, y, 14*u)
		ui.Poly(screen, ui.Hex("#c8643a"), x-7*u, y-10*u, x+7*u, y-10*u, x+5*u, y, x-5*u, y)
		for i, d := range [][2]float64{{-6, -18}, {0, -22}, {6, -17}, {-3, -14}, {4, -13}} {
			_ = i
			ui.Circle(screen, x+d[0]*u, y+d[1]*u, 4.5*u, ui.Hex("#3aa858"))
		}
	case "rug":
		ui.RoundRect(screen, geom.Centered(x, y-8*u, 36*u, 16*u), 8*u, ui.Hex("#9b6af0"), 1.5*s, wall)
		ui.RoundRect(screen, geom.Centered(x, y-8*u, 26*u, 9*u), 4*u, ui.Hex("#ffd23f"), 0, nil)
	case "beanbag":
		ui.Shadow(screen, x, y, 16*u)
		ui.Circle(screen, x, y-7*u, 9*u, ui.Hex("#ff8a3c"))
		ui.Circle(screen, x+2*u, y-12*u, 6*u, ui.Hex("#ffa866"))
	case "lamp":
		ui.Shadow(screen, x, y, 8*u)
		g := 0.5 + 0.5*math.Sin(t*2)
		ui.Circle(screen, x, y-14*u, 10*u, color.RGBA{0xff, 0x6a, 0xc8, uint8(40 + 40*g)})
		ui.Poly(screen, ui.Hex("#c0c8d8"), x-4*u, y-24*u, x+4*u, y-24*u, x+6*u, y-6*u, x-6*u, y-6*u)
		ui.Circle(screen, x, y-12*u-3*u*g, 2.2*u, ui.Hex("#ff6ac8"))
		ui.Circle(screen, x+1*u, y-18*u+2*u*g, 1.6*u, ui.Hex("#ff6ac8"))
		ui.RoundRect(screen, geom.Centered(x, y-3*u, 12*u, 6*u), 2*u, ui.Hex("#566078"), 0, nil)
	case "fishbowl":
		ui.RoundRect(screen, geom.Centered(x, y-3*u, 16*u, 6*u), 2*u, ui.Hex("#8d6040"), 0, nil)
		ui.Circle(screen, x, y-14*u, 9*u, ui.Hex("#bfe6ff"))
		ui.Ring(screen, x, y-14*u, 9*u, 1.2*s, wall)
		fx := x + math.Sin(t*1.5)*4*u
		ui.Poly(screen, ui.Hex("#ff8a3c"), fx-3*u, y-14*u, fx+3*u, y-16*u, fx+3*u, y-12*u)
		ui.Poly(screen, ui.Hex("#ff8a3c"), fx+3*u, y-14*u, fx+6*u, y-17*u, fx+6*u, y-11*u)
	case "bookshelf":
		ui.Shadow(screen, x, y, 16*u)
		r := geom.Rect{X: x - 9*u, Y: y - 30*u, W: 18 * u, H: 30 * u}
		ui.RoundRect(screen, r, 2*u, ui.Hex("#a8744a"), 1.5*s, wall)
		cols := []string{"#e8514a", "#3a6fd8", "#ffd23f", "#3aa858", "#9b6af0"}
		for row := 0; row < 3; row++ {
			for b := 0; b < 4; b++ {
				ui.RoundRect(screen, geom.Rect{X: r.X + 2*u + float64(b)*3.6*u, Y: r.Y + 2*u + float64(row)*9.5*u, W: 3 * u, H: 7.5 * u}, 0.5*u, ui.Hex(cols[(row+b)%len(cols)]), 0, nil)
			}
		}
	case "globe":
		ui.Shadow(screen, x, y, 10*u)
		ui.Line(screen, x, y, x, y-6*u, 2*u, ui.Hex("#8d6040"))
		ui.Circle(screen, x, y-15*u, 9*u, ui.Hex("#3a6fd8"))
		ui.Circle(screen, x-3*u+math.Sin(t)*2*u, y-17*u, 3.5*u, ui.Hex("#3aa858"))
		ui.Circle(screen, x+4*u+math.Sin(t)*2*u, y-12*u, 2.5*u, ui.Hex("#3aa858"))
		ui.Arc(screen, x, y-15*u, 11*u, 0.6*math.Pi, 1.9*math.Pi, 1.5*s, ui.Hex("#c8a04a"))
	case "rocket":
		ui.Shadow(screen, x, y, 12*u)
		ui.Poly(screen, ui.Hex("#e8514a"), x-8*u, y, x-4*u, y-8*u, x-4*u, y-2*u)
		ui.Poly(screen, ui.Hex("#e8514a"), x+8*u, y, x+4*u, y-8*u, x+4*u, y-2*u)
		ui.RoundRect(screen, geom.Centered(x, y-12*u, 9*u, 22*u), 4*u, ui.Hex("#f4f4ee"), 1.2*s, wall)
		ui.Poly(screen, ui.Hex("#e8514a"), x-4.5*u, y-21*u, x+4.5*u, y-21*u, x, y-30*u)
		ui.Circle(screen, x, y-14*u, 2.5*u, ui.Hex("#7ad0ff"))
	case "trophy":
		ui.Shadow(screen, x, y, 10*u)
		gold := ui.Hex("#ffc83c")
		ui.RoundRect(screen, geom.Centered(x, y-2*u, 12*u, 4*u), 1*u, ui.Hex("#8d6040"), 0, nil)
		ui.Line(screen, x, y-4*u, x, y-10*u, 2.5*u, gold)
		ui.Poly(screen, gold, x-7*u, y-22*u, x+7*u, y-22*u, x+4*u, y-11*u, x-4*u, y-11*u)
		ui.Ring(screen, x-7*u, y-18*u, 3*u, 1.2*u, gold)
		ui.Ring(screen, x+7*u, y-18*u, 3*u, 1.2*u, gold)
		ui.Diamond(screen, x, y-17*u, 2*u, ui.ColTextLight)
	}
}
