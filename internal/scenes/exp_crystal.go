package scenes

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const crystalGrowTicks = 90

// CrystalLab: dissolve powder in water, let it cool, grow crystals (E-CHEM-2).
type CrystalLab struct {
	expBase
	powder string
	hot    bool
	spoons int
	grown  string // powder whose crystals are on display
	growT  int
}

func NewCrystalLab() *CrystalLab {
	return &CrystalLab{expBase: expBase{titleKey: "crystal.title", bg: ui.Hex("#c8f0c8"), expID: "crystal"}, hot: true, spoons: 1}
}

func (c *CrystalLab) Opaque() bool { return true }

func (c *CrystalLab) powderRects(a *App) []geom.Rect {
	m := c.main(a)
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 30, W: m.W - 20}, len(experiment.Crystals()), 4, 34)
}

func (c *CrystalLab) waterRects(a *App) []geom.Rect {
	m := c.main(a)
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 94, W: m.W - 20}, 2, 2, 34)
}

func (c *CrystalLab) spoonRects(a *App) (minus, plus, grow geom.Rect) {
	m := c.main(a)
	y := m.Y + 158
	minus = geom.Rect{X: m.X + 12, Y: y, W: 40, H: 36}
	plus = geom.Rect{X: m.X + 132, Y: y, W: 40, H: 36}
	grow = geom.Rect{X: m.X + m.W - 132, Y: y, W: 120, H: 36}
	return
}

// reset clears the jar whenever a setting changes.
func (c *CrystalLab) reset() { c.grown, c.note = "", "" }

func (c *CrystalLab) Update(a *App) error {
	if c.tick(a) || c.busy() {
		return nil
	}
	if c.grown != "" && c.growT < crystalGrowTicks {
		c.growT++
	}
	for i, r := range c.powderRects(a) {
		if a.Pointer.TapIn(r) {
			c.powder = experiment.Crystals()[i].ID
			c.reset()
		}
	}
	for i, r := range c.waterRects(a) {
		if a.Pointer.TapIn(r) {
			c.hot = i == 0
			c.reset()
		}
	}
	minus, plus, grow := c.spoonRects(a)
	switch {
	case a.Pointer.TapIn(minus) && c.spoons > 1:
		c.spoons--
		c.reset()
	case a.Pointer.TapIn(plus) && c.spoons < experiment.MaxSpoons:
		c.spoons++
		c.reset()
	case a.Pointer.TapIn(grow):
		c.grow(a)
	}
	return nil
}

func (c *CrystalLab) grow(a *App) {
	r := experiment.GrowCrystal(c.powder, c.hot, c.spoons)
	if !r.OK {
		c.grown = ""
		c.fizzle(a, r.Hint)
		return
	}
	d, _ := a.Disc.Get(r.Discovery)
	c.grown, c.growT = c.powder, 0
	c.note, c.noteGood = a.Tf("crystal.grew", d.Name), true
	c.succeed(a, r.Discovery)
}

func (c *CrystalLab) Draw(a *App, screen *ebiten.Image) {
	c.drawFrame(a, screen)
	m := c.main(a)
	label := func(key string, y float64) {
		ui.Text(screen, a.T(key), ui.Bold(ui.SizeSmall), m.X+12, y, ui.ColText, text.AlignStart)
	}
	label("crystal.pick", m.Y+8)
	for i, r := range c.powderRects(a) {
		id := experiment.Crystals()[i].ID
		smallButton(screen, r, a.T(powderKey(id)), id == c.powder)
		if a.State.Notebook.Has("crystal." + id) {
			ui.Diamond(screen, r.X+r.W-8, r.Y+8, 5, ui.ColCrystal)
		}
	}
	label("crystal.water", m.Y+72)
	wr := c.waterRects(a)
	smallButton(screen, wr[0], a.T("crystal.hot"), c.hot)
	smallButton(screen, wr[1], a.T("crystal.cold"), !c.hot)
	label("crystal.spoons", m.Y+136)
	minus, plus, grow := c.spoonRects(a)
	ui.Button{R: minus, Label: "-"}.Draw(screen, false)
	ui.Button{R: plus, Label: "+"}.Draw(screen, false)
	ui.TextCentered(screen, fmt.Sprint(c.spoons), ui.Bold(ui.SizeHeading), (minus.X+minus.W+plus.X)/2, minus.Y+minus.H/2, ui.ColText)
	ui.Button{R: grow, Label: a.T("crystal.grow")}.Draw(screen, c.powder != "")

	c.drawJar(a, screen, geom.Rect{X: m.X + m.W/2 - 70, Y: m.Y + 210, W: 140, H: m.H - 250})
	c.drawOverlay(a, screen)
}

func (c *CrystalLab) drawJar(a *App, screen *ebiten.Image, j geom.Rect) {
	water := ui.Hex("#bfe4ff")
	if c.hot {
		water = ui.Hex("#ffd9c4")
	}
	ui.RoundRect(screen, geom.Rect{X: j.X + 8, Y: j.Y - 8, W: j.W - 16, H: 12}, 4, ui.Hex("#dff2ff"), 2, ui.ColWall)
	ui.RoundRect(screen, j, 16, ui.Hex("#eef8ff"), 3, ui.ColWall)
	wtop := j.Y + 20
	ui.RoundRect(screen, geom.Rect{X: j.X + 4, Y: wtop, W: j.W - 8, H: j.H - 24}, 12, water, 0, nil)
	cx := j.X + j.W/2
	if c.hot && c.grown == "" {
		for i := -1; i <= 1; i++ {
			x := cx + float64(i)*24
			ui.Arc(screen, x+4*math.Sin(c.t*3+float64(i)), j.Y-22, 6, math.Pi*0.6, math.Pi*1.4, 2, color.RGBA{0x99, 0x99, 0xaa, 0x90})
		}
	}
	// A pencil across the top with a string hanging down: crystals grow on it.
	ui.Line(screen, j.X-12, j.Y-4, j.X+j.W+12, j.Y-4, 6, ui.Hex("#ffd23f"))
	ui.Line(screen, cx, j.Y-4, cx, j.Y+j.H-30, 1.5, ui.Hex("#8d6040"))
	if c.grown == "" {
		if c.powder != "" {
			// Undissolved powder settles at the bottom.
			n := min(c.spoons, 6)
			for i := 0; i < n; i++ {
				ui.Circle(screen, cx-20+float64(i)*8, j.Y+j.H-12, 6, color.RGBA{0xff, 0xff, 0xff, 0xc0})
			}
		}
		return
	}
	k := float64(c.growT) / crystalGrowTicks
	drawCrystals(screen, c.grown, cx, j.Y+j.H*0.35, j.H*0.55, k)
	if c.growT < crystalGrowTicks {
		ui.TextCentered(screen, a.T("crystal.waiting"), ui.Regular(ui.SizeSmall), cx, j.Y+j.H+14, ui.ColText)
	}
}

// drawCrystals grows each kind's real shape along the string (k = 0..1).
func drawCrystals(screen *ebiten.Image, kind string, x, y, h, k float64) {
	for i := 0; i < 7; i++ {
		cy := y + float64(i)*h/7
		side := float64(i%2*2 - 1)
		s := (6 + float64((i*5)%7)) * k
		cx := x + side*s*0.6
		switch kind {
		case "salt":
			ui.RoundRect(screen, geom.Centered(cx, cy, s*1.6, s*1.6), 1, color.RGBA{0xf8, 0xf8, 0xff, 0xf0}, 1.5, ui.Hex("#9aa6b8"))
		case "sugar":
			ui.Circle(screen, cx, cy, s*1.1, ui.Hex("#ffe3a8"))
			ui.Diamond(screen, cx+s*0.4, cy-s*0.3, s*0.6, ui.Hex("#fff4d8"))
		case "alum":
			ui.Diamond(screen, cx, cy, s*1.4, ui.Hex("#c9ecff"))
			ui.Line(screen, cx-s*1.4, cy, cx+s*1.4, cy, 1, ui.Hex("#7ab8e0"))
		case "epsom":
			for n := 0; n < 3; n++ {
				ang := float64(n)*0.5 - 0.5 + side*0.8
				ui.Line(screen, cx, cy, cx+math.Cos(ang)*s*2.4*side, cy+math.Sin(ang)*s*2.4, 2, ui.Hex("#f2f6ff"))
			}
		}
	}
}
