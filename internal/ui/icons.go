package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/geom"
)

var colGlass = color.RGBA{0xe8, 0xf4, 0xff, 0xff}

// ItemIcon draws an item icon of the given kind and color centered at (x, y).
// s is roughly the icon height.
func ItemIcon(dst *ebiten.Image, icon string, c color.Color, x, y, s float64) {
	u := s / 20
	switch icon {
	case "jar":
		RoundRect(dst, geom.Rect{X: x - 6*u, Y: y - 6*u, W: 12 * u, H: 14 * u}, 3*u, colGlass, u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 5*u, Y: y - 1*u, W: 10 * u, H: 8 * u}, 2*u, c, 0, nil)
		RoundRect(dst, geom.Rect{X: x - 6.5*u, Y: y - 9*u, W: 13 * u, H: 4 * u}, u, Hex("#8d6040"), 0, nil)
	case "bottle":
		RoundRect(dst, geom.Rect{X: x - 6*u, Y: y - 3*u, W: 12 * u, H: 12 * u}, 4*u, c, u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 2.5*u, Y: y - 9*u, W: 5 * u, H: 7 * u}, u, colGlass, u, colOutline)
	case "beads":
		for i, d := range [][2]float64{{-4, 3}, {4, 3}, {0, -3}, {0, 6}} {
			cc := c
			if i%2 == 1 {
				cc = Hex("#ff7ac8")
			}
			Circle(dst, x+d[0]*u, y+d[1]*u, 3.5*u, cc)
		}
	case "tube":
		RoundRect(dst, geom.Rect{X: x - 3.5*u, Y: y - 9*u, W: 7 * u, H: 18 * u}, 3.5*u, colGlass, u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 2.5*u, Y: y - 1*u, W: 5 * u, H: 9 * u}, 2.5*u, c, 0, nil)
	case "battery":
		RoundRect(dst, geom.Rect{X: x - 9*u, Y: y - 5*u, W: 16 * u, H: 10 * u}, 2*u, c, u, colOutline)
		Fill(dst, geom.Rect{X: x + 7*u, Y: y - 2*u, W: 2.5 * u, H: 4 * u}, 0, 0, colOutline)
		Fill(dst, geom.Rect{X: x - 9*u, Y: y - 5*u, W: 5 * u, H: 10 * u}, 0, 0, colOutline)
	case "bulb":
		Circle(dst, x, y-2*u, 7*u, c)
		Ring(dst, x, y-2*u, 7*u, u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 3.5*u, Y: y + 4*u, W: 7 * u, H: 5 * u}, u, Hex("#9aa6b8"), 0, nil)
	case "rock":
		Poly(dst, c, x-8*u, y+3*u, x-5*u, y-5*u, x+2*u, y-7*u, x+8*u, y-2*u, x+7*u, y+6*u, x-3*u, y+7*u)
		Circle(dst, x-2*u, y-1*u, 1.5*u, darker(Hex("#b9b9c8"), 0.75))
		Circle(dst, x+3*u, y+3*u, 1.2*u, darker(Hex("#b9b9c8"), 0.75))
	default:
		Diamond(dst, x, y, 8*u, c)
	}
}
