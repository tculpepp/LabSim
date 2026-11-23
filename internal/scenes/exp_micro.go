package scenes

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

// MicroLab is the Microscope Explorer (E-BIO-1): pick a slide, look, and
// say what you see.
type MicroLab struct {
	expBase
	slide   string
	choices []string
	solved  bool
}

func NewMicroLab() *MicroLab {
	return &MicroLab{expBase: expBase{titleKey: "micro.title", bg: ui.Hex("#a8e6b0"), expID: "microscope"}}
}

func (ml *MicroLab) Opaque() bool { return true }

func (ml *MicroLab) slides(a *App) []experiment.SlideInfo {
	return experiment.Slides(a.State.Owns("tool:super_microscope"))
}

func (ml *MicroLab) slideTray(a *App) geom.Rect {
	m := ml.main(a)
	return geom.Rect{X: m.X + 10, Y: m.Y + 34, W: m.W - 20, H: 80}
}

func (ml *MicroLab) lens(a *App) (float64, float64, float64) {
	m := ml.main(a)
	return m.X + m.W*0.3, m.Y + 238, min(92, m.W*0.25)
}

func (ml *MicroLab) choiceRects(a *App) []geom.Rect {
	m := ml.main(a)
	x := m.X + m.W*0.58
	out := make([]geom.Rect, len(ml.choices))
	for i := range out {
		out[i] = geom.Rect{X: x, Y: m.Y + 160 + float64(i)*50, W: m.W*0.42 - 14, H: 42}
	}
	return out
}

func (ml *MicroLab) Update(a *App) error {
	if ml.tick(a) || ml.busy() {
		return nil
	}
	ss := ml.slides(a)
	for i, r := range grid(ml.slideTray(a), len(ss), 4, 36) {
		if a.Pointer.TapIn(r) {
			ml.pick(a, ss[i])
		}
	}
	if ml.slide == "" || ml.solved {
		return nil
	}
	for i, r := range ml.choiceRects(a) {
		if a.Pointer.TapIn(r) {
			s, _ := experiment.Slide(ml.slide)
			res := experiment.LookAt(s, ml.choices[i])
			if res.Correct {
				ml.solved = true
				ml.note, ml.noteGood = a.T("micro.right"), true
				ml.succeed(a, res.Discovery)
			} else {
				ml.boom(a, res.Hint)
			}
		}
	}
	return nil
}

func (ml *MicroLab) pick(a *App, s experiment.SlideInfo) {
	if s.ID == ml.slide && !ml.solved {
		return
	}
	if s.Needs != "" {
		if a.State.Backpack.Remove(s.Needs, 1) != nil {
			ml.fizzle(a, "micro.needs")
			return
		}
	}
	n := 3
	if a.Level() == content.Expert {
		n = 4 // more look-alike choices for Experts (D-2)
	}
	ml.slide, ml.solved, ml.note = s.ID, false, ""
	ml.choices = experiment.Choices(s, n, uint64(a.State.Clock.Day())*131+uint64(len(s.ID)))
}

func (ml *MicroLab) Draw(a *App, screen *ebiten.Image) {
	ml.drawFrame(a, screen)
	m := ml.main(a)
	ui.Text(screen, a.T("micro.pick"), ui.Bold(ui.SizeSmall), m.X+12, m.Y+10, ui.ColText, text.AlignStart)
	ss := ml.slides(a)
	for i, r := range grid(ml.slideTray(a), len(ss), 4, 36) {
		ui.Button{R: r, Label: a.T(slideKey(ss[i].ID))}.Draw(screen, ss[i].ID == ml.slide)
		if a.State.Notebook.Has("micro." + ss[i].ID) {
			ui.Diamond(screen, r.X+r.W-8, r.Y+8, 5, ui.ColCrystal)
		}
	}
	x, y, r := ml.lens(a)
	ui.Circle(screen, x, y, r+10, ui.Hex("#2e3648"))
	ui.Circle(screen, x, y, r, ui.Hex("#fdfbe8"))
	if ml.slide != "" {
		drawSlideView(screen, ml.slide, x, y, r, ml.t)
	}
	ui.Ring(screen, x, y, r, 4, ui.Hex("#566078"))
	if ml.slide != "" {
		ui.TextCentered(screen, a.T("micro.ask"), ui.Bold(ui.SizeBody+2), m.X+m.W*0.79, m.Y+138, ui.ColText)
		for i, cr := range ml.choiceRects(a) {
			ui.Button{R: cr, Label: a.T(optKey(ml.choices[i]))}.Draw(screen, ml.solved && i < len(ml.choices) && ml.choices[i] == mustSlide(ml.slide).Answer)
		}
	}
	ml.drawOverlay(a, screen)
}

func mustSlide(id string) experiment.SlideInfo { s, _ := experiment.Slide(id); return s }

// drawSlideView draws a magnified sample inside the lens circle.
func drawSlideView(screen *ebiten.Image, id string, cx, cy, r, t float64) {
	rng := rand.New(rand.NewPCG(uint64(len(id))*7, 3))
	in := func(x, y, pad float64) bool { return math.Hypot(x-cx, y-cy) < r-pad }
	switch id {
	case "pond":
		ui.Circle(screen, cx, cy, r, ui.Hex("#dff3e8"))
		for i := 0; i < 6; i++ {
			bx := cx + (rng.Float64()-0.5)*r*1.2 + math.Sin(t*0.8+float64(i))*14
			by := cy + (rng.Float64()-0.5)*r*1.2 + math.Cos(t*0.6+float64(i)*2)*10
			if !in(bx, by, 24) {
				continue
			}
			ui.RoundRect(screen, geom.Centered(bx, by, 34, 16), 8, ui.Hex("#9fd8b8"), 1.5, ui.Hex("#3a8a5a"))
			ui.Circle(screen, bx-4, by, 3, ui.Hex("#3a8a5a"))
			for k := -2; k <= 2; k++ {
				ui.Line(screen, bx+float64(k)*6, by-8, bx+float64(k)*6+math.Sin(t*10+float64(k)), by-12, 1, ui.Hex("#3a8a5a"))
			}
		}
	case "leaf", "onion":
		fill, wall, dot := ui.Hex("#b6e88a"), ui.Hex("#4a9a3a"), ui.Hex("#2f8a2f")
		w, h := 30.0, 22.0
		if id == "onion" {
			fill, wall, dot, w, h = ui.Hex("#f8ecd8"), ui.Hex("#c89a7a"), ui.Hex("#9a5a5a"), 42, 18
		}
		for row := -6; row <= 6; row++ {
			off := 0.0
			if row%2 != 0 {
				off = w / 2
			}
			for col := -6; col <= 6; col++ {
				x := cx + float64(col)*w + off
				y := cy + float64(row)*h
				if !in(x, y, w*0.6) {
					continue
				}
				ui.RoundRect(screen, geom.Centered(x, y, w-3, h-3), 4, fill, 1.5, wall)
				if id == "leaf" {
					for k := 0; k < 3; k++ {
						ui.Circle(screen, x-8+float64(k)*8, y+float64(k%2)*4-2, 2.6, dot)
					}
				} else {
					ui.Circle(screen, x+6, y, 2.8, dot)
				}
			}
		}
	case "hair":
		ui.RoundRect(screen, geom.Rect{X: cx - r, Y: cy - 26, W: 2 * r, H: 52}, 20, ui.Hex("#a0673a"), 0, nil)
		for x := cx - r + 10; x < cx+r-10; x += 18 {
			ui.Arc(screen, x, cy, 24, -0.4*math.Pi, 0.4*math.Pi, 2, ui.Hex("#6a3f20"))
		}
	case "salt":
		for i := 0; i < 9; i++ {
			x, y := cx+(rng.Float64()-0.5)*r*1.3, cy+(rng.Float64()-0.5)*r*1.3
			s := 18 + rng.Float64()*18
			if in(x, y, s) {
				ui.RoundRect(screen, geom.Centered(x, y, s, s), 2, ui.Hex("#ffffff"), 2, ui.Hex("#9aa6b8"))
				ui.Line(screen, x-s/2+3, y-s/2+3, x+s/2-3, y-s/2+3, 1, ui.Hex("#d8e0ea"))
			}
		}
	case "sand":
		for i := 0; i < 12; i++ {
			x, y := cx+(rng.Float64()-0.5)*r*1.4, cy+(rng.Float64()-0.5)*r*1.4
			if !in(x, y, 16) {
				continue
			}
			c := []string{"#e0c890", "#c8a870", "#f0dca8", "#b89868"}[i%4]
			s := 9 + rng.Float64()*7
			ui.Poly(screen, ui.Hex(c), x-s, y, x-s*0.4, y-s*0.9, x+s*0.7, y-s*0.6, x+s, y+s*0.3, x+s*0.2, y+s, x-s*0.8, y+s*0.6)
		}
	case "wing":
		cols := []string{"#ff9a3c", "#ffd23f", "#2e3648", "#ff9a3c"}
		for row := -5; row <= 5; row++ {
			for col := -6; col <= 6; col++ {
				x := cx + float64(col)*20 + float64(row%2)*10
				y := cy + float64(row)*16
				if in(x, y, 12) {
					ui.RoundRect(screen, geom.Centered(x, y, 18, 22), 8, ui.Hex(cols[(row+col+20)%len(cols)]), 1, color.RGBA{0, 0, 0, 0x40})
				}
			}
		}
	}
}
