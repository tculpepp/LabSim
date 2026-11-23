package scenes

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const bowlSize = 3

// SlimeLab is the Glowing Slime sandbox (E-CHEM-1).
type SlimeLab struct {
	expBase
	bowl   []string
	result string // slime kind on display
}

func NewSlimeLab() *SlimeLab {
	return &SlimeLab{expBase: expBase{titleKey: "slime.title", bg: ui.Hex("#c8f0c8"), expID: "slime"}}
}

func (s *SlimeLab) Opaque() bool { return true }

// shelf lists slime ingredients in the backpack with how many are left.
func (s *SlimeLab) shelf(a *App) (ids []string, left []int) {
	for _, st := range a.State.Backpack.Items() {
		if !experiment.IsSlimeIngredient(st.ID) {
			continue
		}
		n := st.Qty
		for _, b := range s.bowl {
			if b == st.ID {
				n--
			}
		}
		ids, left = append(ids, st.ID), append(left, n)
	}
	return
}

func (s *SlimeLab) beaker(a *App) geom.Rect {
	m := s.main(a)
	return geom.Rect{X: m.X + 24, Y: m.Y + 40, W: m.W*0.42 - 24, H: 210}
}

func (s *SlimeLab) slot(a *App, i int) geom.Rect {
	b := s.beaker(a)
	return geom.Centered(b.X+b.W/2, b.Y+60+float64(i)*50, 60, 44)
}

func (s *SlimeLab) mixBtn(a *App) geom.Rect {
	b := s.beaker(a)
	return geom.Rect{X: b.X, Y: b.Y + b.H + 14, W: b.W*0.6 - 4, H: 44}
}

func (s *SlimeLab) emptyBtn(a *App) geom.Rect {
	b := s.beaker(a)
	return geom.Rect{X: b.X + b.W*0.6 + 4, Y: b.Y + b.H + 14, W: b.W*0.4 - 4, H: 44}
}

func (s *SlimeLab) shelfArea(a *App) geom.Rect {
	m := s.main(a)
	return geom.Rect{X: m.X + m.W*0.46, Y: m.Y + 44, W: m.W*0.54 - 14, H: m.H - 90}
}

func (s *SlimeLab) Update(a *App) error {
	if s.tick(a) || s.busy() {
		return nil
	}
	ids, left := s.shelf(a)
	for i, r := range grid(s.shelfArea(a), len(ids), 2, 70) {
		if (a.Pointer.TapIn(r) || inpututil.IsKeyJustPressed(ebiten.Key1+ebiten.Key(i))) && left[i] > 0 && len(s.bowl) < bowlSize {
			s.bowl, s.result, s.note = append(s.bowl, ids[i]), "", ""
		}
	}
	for i := len(s.bowl) - 1; i >= 0; i-- {
		if a.Pointer.TapIn(s.slot(a, i)) {
			s.bowl = append(s.bowl[:i], s.bowl[i+1:]...)
		}
	}
	if a.Pointer.TapIn(s.emptyBtn(a)) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		s.bowl, s.result = nil, ""
	}
	if a.Pointer.TapIn(s.mixBtn(a)) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.mix(a)
	}
	return nil
}

func (s *SlimeLab) mix(a *App) {
	r := experiment.MixSlime(s.bowl)
	if r.OK || r.Boom {
		// The only cost of a failed experiment is the ingredients used (E-4).
		for _, id := range s.bowl {
			a.State.Backpack.Remove(id, 1)
		}
		s.bowl = nil
	}
	switch {
	case r.OK:
		d, _ := a.Disc.Get(r.Discovery)
		s.result, s.note, s.noteGood = r.Slime, a.Tf("slime.made", d.Name), true
		s.succeed(a, r.Discovery)
	case r.Boom:
		s.result = ""
		s.boom(a, r.Hint)
	default:
		s.fizzle(a, r.Hint)
	}
}

func (s *SlimeLab) Draw(a *App, screen *ebiten.Image) {
	s.drawFrame(a, screen)
	b := s.beaker(a)
	// Beaker.
	ui.RoundRect(screen, geom.Rect{X: b.X + 10, Y: b.Y - 12, W: b.W - 20, H: 16}, 6, ui.Hex("#dff2ff"), 2, ui.ColWall)
	ui.RoundRect(screen, b, 22, ui.Hex("#eef8ff"), 3, ui.ColWall)
	if s.result != "" {
		drawSlime(screen, s.result, b.X+b.W/2, b.Y+b.H-60, s.t)
	} else {
		for i := 0; i < bowlSize; i++ {
			r := s.slot(a, i)
			if i < len(s.bowl) {
				ui.RoundRect(screen, r, 10, ui.ColPanel, 2, ui.ColPanelLine)
				a.DrawItem(screen, s.bowl[i], r.X+r.W/2, r.Y+r.H/2, 30)
			} else {
				ui.RoundRect(screen, r, 10, nil, 2, color.RGBA{0x30, 0x60, 0xd0, 0x40})
			}
		}
	}
	ui.Button{R: s.mixBtn(a), Label: a.T("slime.mix")}.Draw(screen, len(s.bowl) > 0)
	ui.Button{R: s.emptyBtn(a), Label: a.T("slime.empty")}.Draw(screen, false)

	area := s.shelfArea(a)
	ui.Text(screen, a.T("slime.shelf"), ui.Bold(ui.SizeSmall), area.X, area.Y-24, ui.ColText, text.AlignStart)
	ids, left := s.shelf(a)
	if len(ids) == 0 {
		lines := ui.Wrap(a.T("slime.none"), ui.Regular(ui.SizeBody), area.W)
		ui.Lines(screen, lines, ui.Regular(ui.SizeBody), ui.SizeBody, area.X, area.Y+10, ui.ColText, text.AlignStart)
	}
	for i, r := range grid(area, len(ids), 2, 70) {
		id := ids[i]
		tile{r: r, label: a.ItemName(id), sub: fmt.Sprintf("x%d", left[i]),
			icon: func(sc *ebiten.Image, x, y float64) { a.DrawItem(sc, id, x, y, 30) }}.draw(screen, false)
	}
	s.drawOverlay(a, screen)
}

// drawSlime draws a finished slime blob, each kind with its own animation.
func drawSlime(screen *ebiten.Image, kind string, x, y, t float64) {
	switch kind {
	case "glow":
		g := 0.5 + 0.5*math.Sin(t*3)
		ui.Circle(screen, x, y-10, 70+10*g, color.RGBA{0xb6, 0xff, 0x6a, uint8(40 + 50*g)})
		blob(screen, x, y, 1, ui.Hex("#9dff5a"))
	case "bouncy":
		h := math.Abs(math.Sin(t*4)) * 70
		squash := 1.0
		if h < 8 {
			squash = 0.75
		}
		ui.Shadow(screen, x, y+26, 60)
		blob(screen, x, y-h, squash, ui.Hex("#7ad0ff"))
	case "rainbow":
		cols := []string{"#ff5a5a", "#ff9a3c", "#ffd23f", "#5ad85a", "#4f8df0", "#9b6af0"}
		c := ui.Hex(cols[int(t*3)%len(cols)])
		blob(screen, x, y, 1, c)
		for i, cc := range cols {
			ui.Fill(screen, geom.Rect{X: x - 40, Y: y - 10 + float64(i)*4, W: 80, H: 3}, 0, 0, ui.Hex(cc))
		}
	case "stretchy":
		s := 1 + 0.35*math.Sin(t*2)
		ui.RoundRect(screen, geom.Centered(x, y, 120*s, 40/s), 20/s, ui.Hex("#c89bff"), 2, ui.ColWall)
		ui.Circle(screen, x-12, y-4, 4, color.Black)
		ui.Circle(screen, x+12, y-4, 4, color.Black)
		return
	}
}

// blob is a cute slime with a face.
func blob(screen *ebiten.Image, x, y, squash float64, c color.Color) {
	w, h := 110/squash, 70*squash
	ui.RoundRect(screen, geom.Centered(x, y+h*0.1, w, h*0.8), h*0.35, c, 2, ui.ColWall)
	ui.Circle(screen, x, y-h*0.15, h*0.45, c)
	ui.Circle(screen, x-14, y-6, 5, color.Black)
	ui.Circle(screen, x+14, y-6, 5, color.Black)
	ui.Circle(screen, x-20, y-h*0.35, 7, color.RGBA{0xff, 0xff, 0xff, 0x90})
	ui.Arc(screen, x, y+2, 8, 0.15*math.Pi, 0.85*math.Pi, 2.5, color.Black)
}
