// Package ui holds Ebitengine drawing and input helpers shared by scenes.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"labsim/internal/geom"
)

// Palette: bright, friendly colors (AR-1).
var (
	ColHallway   = color.RGBA{0x5a, 0x6a, 0x7a, 0xff}
	ColWall      = color.RGBA{0x2e, 0x36, 0x48, 0xff}
	ColLocked    = color.RGBA{0xe0, 0x5a, 0x5a, 0xff}
	ColPanel     = color.RGBA{0xff, 0xfb, 0xf2, 0xf0}
	ColPanelLine = color.RGBA{0x30, 0x60, 0xd0, 0xff}
	ColText      = color.RGBA{0x22, 0x26, 0x33, 0xff}
	ColTextLight = color.RGBA{0xff, 0xff, 0xff, 0xff}
	ColAccent    = color.RGBA{0x30, 0x60, 0xd0, 0xff}
	ColCrystal   = color.RGBA{0x5c, 0xd6, 0xf0, 0xff}
	ColShade     = color.RGBA{0x00, 0x00, 0x00, 0x88}
	ColNight     = color.RGBA{0x10, 0x10, 0x40, 0x70}
)

// RGB converts a parsed hex color to an opaque color.
func RGB(c [3]uint8) color.RGBA { return color.RGBA{c[0], c[1], c[2], 0xff} }

// Fill draws r offset by (-ox, -oy) (camera space).
func Fill(dst *ebiten.Image, r geom.Rect, ox, oy float64, c color.Color) {
	vector.FillRect(dst, float32(r.X-ox), float32(r.Y-oy), float32(r.W), float32(r.H), c, false)
}

// Stroke outlines r offset by (-ox, -oy).
func Stroke(dst *ebiten.Image, r geom.Rect, ox, oy, width float64, c color.Color) {
	vector.StrokeRect(dst, float32(r.X-ox), float32(r.Y-oy), float32(r.W), float32(r.H), float32(width), c, true)
}

func Circle(dst *ebiten.Image, x, y, r float64, c color.Color) {
	vector.FillCircle(dst, float32(x), float32(y), float32(r), c, true)
}

func Ring(dst *ebiten.Image, x, y, r, width float64, c color.Color) {
	vector.StrokeCircle(dst, float32(x), float32(y), float32(r), float32(width), c, true)
}

// Diamond draws a small crystal icon centered at (x, y).
func Diamond(dst *ebiten.Image, x, y, size float64, c color.Color) {
	var p vector.Path
	p.MoveTo(float32(x), float32(y-size))
	p.LineTo(float32(x+size*0.7), float32(y))
	p.LineTo(float32(x), float32(y+size))
	p.LineTo(float32(x-size*0.7), float32(y))
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, &p, nil, op)
}

// Padlock draws a little lock icon centered at (x, y).
func Padlock(dst *ebiten.Image, x, y, size float64, c color.Color) {
	Ring(dst, x, y-size*0.35, size*0.35, size*0.18, c)
	vector.FillRect(dst, float32(x-size*0.5), float32(y-size*0.2), float32(size), float32(size*0.75), c, true)
}
