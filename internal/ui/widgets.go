package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"labsim/internal/geom"
)

func roundPath(r geom.Rect, radius float64) *vector.Path {
	x, y, w, h := float32(r.X), float32(r.Y), float32(r.W), float32(r.H)
	rad := float32(min(radius, r.W/2, r.H/2))
	var p vector.Path
	p.MoveTo(x+rad, y)
	p.LineTo(x+w-rad, y)
	p.ArcTo(x+w, y, x+w, y+rad, rad)
	p.LineTo(x+w, y+h-rad)
	p.ArcTo(x+w, y+h, x+w-rad, y+h, rad)
	p.LineTo(x+rad, y+h)
	p.ArcTo(x, y+h, x, y+h-rad, rad)
	p.LineTo(x, y+rad)
	p.ArcTo(x, y, x+rad, y, rad)
	p.Close()
	return &p
}

// RoundRect fills a rounded rectangle (fill may be nil), with an optional outline (lineW > 0).
func RoundRect(dst *ebiten.Image, r geom.Rect, radius float64, fill color.Color, lineW float64, line color.Color) {
	p := roundPath(r, radius)
	if fill != nil {
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(fill)
		vector.FillPath(dst, p, nil, op)
	}
	if lineW > 0 {
		sop := &vector.DrawPathOptions{AntiAlias: true}
		sop.ColorScale.ScaleWithColor(line)
		vector.StrokePath(dst, p, &vector.StrokeOptions{Width: float32(lineW)}, sop)
	}
}

// Panel is the standard friendly rounded box.
func Panel(dst *ebiten.Image, r geom.Rect) {
	RoundRect(dst, r, 14, ColPanel, 3, ColPanelLine)
}

// Shade dims the whole screen behind an overlay.
func Shade(dst *ebiten.Image, w, h int) {
	Fill(dst, geom.Rect{W: float64(w), H: float64(h)}, 0, 0, ColShade)
}

// Button is a tappable rounded button.
type Button struct {
	R     geom.Rect
	Label string
}

func (b Button) Draw(dst *ebiten.Image, focused bool) {
	fill, txt := ColPanel, ColText
	if focused {
		fill = ColFocus
	}
	RoundRect(dst, b.R, 12, fill, 3, ColPanelLine)
	cx, cy := b.R.Center()
	TextCentered(dst, b.Label, Bold(SizeBody), cx, cy, txt)
}

// Menu is a list of buttons usable by tap, mouse, or keyboard.
type Menu struct {
	Buttons []Button
	Focus   int

	hideFocus bool // touch players have no "focus"; only keyboard/mouse do
}

// OnActivate, if set, runs whenever any menu button is activated (click sound).
var OnActivate func()

// Update returns the index of the activated button, or -1.
func (m *Menu) Update(p *Pointer) int {
	i := m.update(p)
	if i >= 0 && OnActivate != nil {
		OnActivate()
	}
	return i
}

func (m *Menu) update(p *Pointer) int {
	n := len(m.Buttons)
	if n == 0 {
		return -1
	}
	m.hideFocus = p.TouchDevice()
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
		m.hideFocus = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		m.Focus = (m.Focus + 1) % n
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		m.Focus = (m.Focus + n - 1) % n
	}
	if !p.TouchDevice() {
		x, y := ebiten.CursorPosition()
		for i, b := range m.Buttons {
			if b.R.Contains(float64(x), float64(y)) {
				m.Focus = i
			}
		}
	}
	for i, b := range m.Buttons {
		if p.TapIn(b.R) {
			m.Focus = i
			return i
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return m.Focus
	}
	return -1
}

func (m *Menu) Draw(dst *ebiten.Image) {
	for i, b := range m.Buttons {
		b.Draw(dst, i == m.Focus && !m.hideFocus)
	}
}

// Speaker portraits (placeholders until M9 art).
var portraitColors = map[string]color.RGBA{
	"gizmo": {0x9a, 0xd8, 0xff, 0xff},
	"nova":  {0xd2, 0xc4, 0xf5, 0xff},
	"tom":   {0xc8, 0x9a, 0x6a, 0xff},
}

// Portrait draws a round character portrait centered at (x, y).
func Portrait(dst *ebiten.Image, who string, x, y, r float64) {
	c, ok := portraitColors[who]
	if !ok {
		c = color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
	}
	Circle(dst, x, y, r, c)
	Ring(dst, x, y, r, 3, ColPanelLine)
	if who == "gizmo" {
		// Round floating robot with a screen face.
		screen := geom.Rect{X: x - r*0.6, Y: y - r*0.4, W: r * 1.2, H: r * 0.75}
		RoundRect(dst, screen, 6, ColWall, 0, nil)
		Circle(dst, x-r*0.25, y-r*0.05, r*0.1, ColCrystal)
		Circle(dst, x+r*0.25, y-r*0.05, r*0.1, ColCrystal)
		Fill(dst, geom.Rect{X: x - 1.5, Y: y - r*1.15, W: 3, H: r * 0.25}, 0, 0, ColWall)
		Circle(dst, x, y-r*1.15, 4, ColLocked)
	}
}
