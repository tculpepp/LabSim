package scenes

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

// CircuitLab: put a material in the gap, predict, then test (E-PHYS-1).
// Expert players also get a switch to think about (D-2).
type CircuitLab struct {
	expBase
	chosen   string
	switchOn bool
	phase    int // 0 pick, 1 predict, 2 result
	lit      bool
	result   string
}

func NewCircuitLab() *CircuitLab {
	return &CircuitLab{expBase: expBase{titleKey: "circuit.title", bg: ui.Hex("#ffe8a8"), expID: "circuit"}, switchOn: true}
}

func (c *CircuitLab) Opaque() bool { return true }

func (c *CircuitLab) hasSwitch(a *App) bool { return a.Level() == content.Expert }

func (c *CircuitLab) loop(a *App) geom.Rect {
	m := c.main(a)
	return geom.Rect{X: m.X + 50, Y: m.Y + 46, W: m.W - 100, H: 150}
}

func (c *CircuitLab) gap(a *App) (float64, float64) {
	l := c.loop(a)
	return l.X + l.W/2, l.Y + l.H
}

func (c *CircuitLab) switchRect(a *App) geom.Rect {
	l := c.loop(a)
	return geom.Centered(l.X+l.W, l.Y+l.H/2, 44, 56)
}

func (c *CircuitLab) tray(a *App) geom.Rect {
	m := c.main(a)
	return geom.Rect{X: m.X + 12, Y: m.Y + 236, W: m.W - 24, H: 140}
}

func (c *CircuitLab) answers(a *App) (yes, no geom.Rect) {
	m := c.main(a)
	y := m.Y + 260
	return geom.Rect{X: m.X + 20, Y: y, W: m.W/2 - 30, H: 50}, geom.Rect{X: m.X + m.W/2 + 10, Y: y, W: m.W/2 - 30, H: 50}
}

func (c *CircuitLab) mats(a *App) []experiment.Mat {
	return experiment.Materials(a.State.Owns("tool:spark_kit"))
}

func (c *CircuitLab) Update(a *App) error {
	if c.tick(a) || c.busy() {
		return nil
	}
	if !c.hasSwitch(a) {
		c.switchOn = true
	} else if c.phase != 1 && a.Pointer.TapIn(c.switchRect(a)) {
		c.switchOn = !c.switchOn
		c.phase, c.result = min(c.phase, 1), ""
		if c.chosen != "" {
			c.phase = 1
		}
	}
	if c.phase == 1 {
		yes, no := c.answers(a)
		switch {
		case a.Pointer.TapIn(yes):
			c.test(a, true)
		case a.Pointer.TapIn(no):
			c.test(a, false)
		}
		return nil
	}
	ms := c.mats(a)
	for i, r := range grid(c.tray(a), len(ms), 5, 62) {
		if a.Pointer.TapIn(r) {
			c.chosen, c.phase, c.lit, c.result, c.note = ms[i].ID, 1, false, "", ""
		}
	}
	return nil
}

func (c *CircuitLab) test(a *App, predictLit bool) {
	r := experiment.TestCircuit(c.chosen, c.switchOn, predictLit)
	c.phase, c.lit = 2, r.Lit
	name := a.T(matKey(c.chosen))
	switch {
	case r.Lit:
		c.result = a.Tf("circuit.lit", name)
	case !c.switchOn:
		c.result = a.T("circuit.darkSwitch")
	default:
		c.result = a.Tf("circuit.dark", name)
	}
	if r.Correct {
		c.note, c.noteGood = a.T("circuit.right"), true
	} else {
		c.note, c.noteGood = a.T("circuit.surprise"), false
		c.fizzle(a, r.Hint)
	}
	if r.Lit {
		c.succeed(a, r.Discovery)
	} else if c.switchOn {
		c.discoverQuietly(a, r.Discovery)
	}
}

func (c *CircuitLab) Draw(a *App, screen *ebiten.Image) {
	c.drawFrame(a, screen)
	l := c.loop(a)
	wire := ui.Hex("#d8423a")
	gx, gy := c.gap(a)
	// Wires around the loop, with the gap at the bottom.
	ui.Line(screen, l.X, l.Y, l.X+l.W, l.Y, 4, wire)
	ui.Line(screen, l.X, l.Y, l.X, l.Y+l.H, 4, wire)
	ui.Line(screen, l.X+l.W, l.Y, l.X+l.W, l.Y+l.H, 4, wire)
	ui.Line(screen, l.X, l.Y+l.H, gx-34, gy, 4, wire)
	ui.Line(screen, gx+34, gy, l.X+l.W, l.Y+l.H, 4, wire)
	// Battery on the left.
	ui.RoundRect(screen, geom.Centered(l.X, l.Y+l.H/2, 34, 70), 6, ui.Hex("#ffb347"), 2, ui.ColWall)
	ui.TextCentered(screen, "+", ui.Bold(ui.SizeHeading), l.X, l.Y+l.H/2-18, ui.ColText)
	ui.TextCentered(screen, "-", ui.Bold(ui.SizeHeading), l.X, l.Y+l.H/2+18, ui.ColText)
	// Bulb at the top.
	bx, by := l.X+l.W/2, l.Y
	if c.lit {
		g := 0.6 + 0.4*math.Sin(c.t*6)
		ui.Circle(screen, bx, by-6, 44+6*g, color.RGBA{0xff, 0xf0, 0x80, uint8(90 * g)})
		for i := 0; i < 8; i++ {
			ang := float64(i) * math.Pi / 4
			ui.Line(screen, bx+math.Cos(ang)*30, by-6+math.Sin(ang)*30, bx+math.Cos(ang)*42, by-6+math.Sin(ang)*42, 3, ui.ColFocus)
		}
		ui.Circle(screen, bx, by-6, 22, ui.Hex("#fff36a"))
	} else {
		ui.Circle(screen, bx, by-6, 22, ui.Hex("#f4f4ee"))
	}
	ui.Ring(screen, bx, by-6, 22, 2, ui.ColWall)
	ui.RoundRect(screen, geom.Centered(bx, by+18, 22, 14), 3, ui.Hex("#9aa6b8"), 0, nil)
	// Switch (Expert).
	if c.hasSwitch(a) {
		sr := c.switchRect(a)
		ui.RoundRect(screen, sr, 8, ui.ColPanel, 2, ui.ColWall)
		sx, sy := sr.Center()
		if c.switchOn {
			ui.Line(screen, sx, sy-18, sx, sy+18, 5, ui.ColWall)
		} else {
			ui.Line(screen, sx, sy+18, sx+16, sy-12, 5, ui.ColWall)
		}
		label := a.T("circuit.off")
		if c.switchOn {
			label = a.T("circuit.on")
		}
		ui.TextCentered(screen, a.Tf("circuit.switch", label), ui.Bold(ui.SizeSmall), sx, sr.Y+sr.H+12, ui.ColText)
	}
	// The gap and whatever is in it.
	ui.RoundRect(screen, geom.Centered(gx, gy, 68, 44), 10, ui.ColPanel, 2, color.RGBA{0x30, 0x60, 0xd0, 0x80})
	if c.chosen != "" {
		drawMaterial(screen, c.chosen, gx, gy, 1.2)
	}

	switch c.phase {
	case 0:
		ui.Text(screen, a.T("circuit.pick"), ui.Bold(ui.SizeBody), c.tray(a).X, c.tray(a).Y-26, ui.ColText, text.AlignStart)
		ms := c.mats(a)
		for i, r := range grid(c.tray(a), len(ms), 5, 62) {
			id := ms[i].ID
			tile{r: r, label: a.T(matKey(id)), icon: func(sc *ebiten.Image, x, y float64) { drawMaterial(sc, id, x, y, 0.9) }}.draw(screen, false)
		}
	case 1:
		m := c.main(a)
		ui.TextCentered(screen, a.T("circuit.ask"), ui.Bold(ui.SizeHeading), m.X+m.W/2, m.Y+232, ui.ColText)
		yes, no := c.answers(a)
		ui.Button{R: yes, Label: a.T("circuit.yes")}.Draw(screen, false)
		ui.Button{R: no, Label: a.T("circuit.no")}.Draw(screen, false)
	case 2:
		m := c.main(a)
		lines := ui.Wrap(c.result, ui.Bold(ui.SizeBody), m.W-40)
		ui.Lines(screen, lines, ui.Bold(ui.SizeBody), ui.SizeBody, m.X+m.W/2, m.Y+226, ui.ColText, text.AlignCenter)
		ms := c.mats(a)
		tray := c.tray(a)
		tray.Y += 40
		for i, r := range grid(tray, len(ms), 5, 44) {
			id := ms[i].ID
			tile{r: r, label: a.T(matKey(id)), on: a.State.Notebook.Has("circuit." + id)}.draw(screen, false)
		}
	}
	c.drawOverlay(a, screen)
}

// drawMaterial draws a circuit test material.
func drawMaterial(screen *ebiten.Image, id string, x, y, s float64) {
	silver, dark := ui.Hex("#c8ced8"), ui.Hex("#2e3648")
	switch id {
	case "paperclip":
		ui.RoundRect(screen, geom.Centered(x, y, 30*s, 12*s), 6*s, nil, 2.5*s, silver)
		ui.RoundRect(screen, geom.Centered(x+2*s, y, 20*s, 6*s), 3*s, nil, 2*s, silver)
	case "coin":
		ui.Circle(screen, x, y, 12*s, ui.Hex("#e0a83a"))
		ui.Ring(screen, x, y, 9*s, 1.5*s, ui.Hex("#b07a1a"))
	case "foil":
		ui.Poly(screen, silver, x-14*s, y-8*s, x+10*s, y-11*s, x+15*s, y+6*s, x-8*s, y+10*s)
		ui.Line(screen, x-6*s, y-6*s, x+4*s, y+4*s, 1.2*s, ui.Hex("#ffffff"))
	case "rubber_band":
		ui.Ring(screen, x, y, 11*s, 3*s, ui.Hex("#e8514a"))
	case "plastic_spoon":
		ui.Circle(screen, x-8*s, y, 7*s, ui.Hex("#7ad0ff"))
		ui.Line(screen, x-2*s, y, x+16*s, y, 4*s, ui.Hex("#7ad0ff"))
	case "wood_stick":
		ui.RoundRect(screen, geom.Centered(x, y, 32*s, 7*s), 3*s, ui.Hex("#c08a5a"), 1, ui.Hex("#8d6040"))
	case "pencil":
		ui.RoundRect(screen, geom.Centered(x, y, 26*s, 8*s), 2*s, ui.Hex("#ffd23f"), 1, dark)
		ui.Poly(screen, ui.Hex("#f1c7a0"), x+13*s, y-4*s, x+13*s, y+4*s, x+20*s, y)
		ui.Circle(screen, x+19*s, y, 1.5*s, dark)
	case "key":
		ui.Ring(screen, x-9*s, y, 6*s, 3*s, ui.Hex("#e0a83a"))
		ui.Line(screen, x-3*s, y, x+16*s, y, 3.5*s, ui.Hex("#e0a83a"))
		ui.Line(screen, x+12*s, y, x+12*s, y+6*s, 3*s, ui.Hex("#e0a83a"))
	case "eraser":
		ui.RoundRect(screen, geom.Centered(x, y, 26*s, 14*s), 3*s, ui.Hex("#ffb3d9"), 1, dark)
	case "marble":
		ui.Circle(screen, x, y, 10*s, ui.Hex("#8fc1ff"))
		ui.Circle(screen, x-3*s, y-3*s, 3*s, ui.Hex("#ffffff"))
	}
}
