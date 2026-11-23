package scenes

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const (
	pxPerMetre = 34.0
	jumpSecs   = 1.4
)

var planetColors = map[string][2]string{
	"moon":    {"#c8c8d0", "#9a9aa6"},
	"mars":    {"#e0703a", "#a8482a"},
	"earth":   {"#4fa86a", "#3a6fd8"},
	"jupiter": {"#e8b07a", "#b8784a"},
	"pluto":   {"#e8d0c0", "#b89a8a"},
}

// GravityLab: predict, then jump on different worlds (E-SPACE-1).
type GravityLab struct {
	expBase
	planet   string
	jumpT    float64 // seconds into the current jump, <0 when standing
	answered bool
}

func NewGravityLab() *GravityLab {
	return &GravityLab{expBase: expBase{titleKey: "gravity.title", darkBG: true, bg: ui.Hex("#2b2d5a"), expID: "gravity"}, planet: "earth", jumpT: -1}
}

func (g *GravityLab) Opaque() bool { return true }

func (g *GravityLab) planets(a *App) []experiment.PlanetInfo {
	return experiment.Planets(a.State.Owns("tool:telescope"))
}

// question returns the answer choices and the right one for this level.
func (g *GravityLab) question(a *App) (labels, ids []string, answer string) {
	if a.Level() == content.Expert {
		for _, n := range []int{2, 6, 10, 20} {
			labels, ids = append(labels, a.Tf("gravity.times", n)), append(ids, strconv.Itoa(n))
		}
		return labels, ids, strconv.Itoa(experiment.MoonEarthRatio())
	}
	ps := g.planets(a)
	for _, p := range ps {
		labels, ids = append(labels, a.T(planetKey(p.ID))), append(ids, p.ID)
	}
	return labels, ids, experiment.HighestJump(ps)
}

func (g *GravityLab) answerRects(a *App, n int) []geom.Rect {
	m := g.main(a)
	return grid(geom.Rect{X: m.X + 12, Y: m.Y + 46, W: m.W - 24, H: 44}, n, n, 40)
}

func (g *GravityLab) planetRects(a *App) []geom.Rect {
	m := g.main(a)
	return grid(geom.Rect{X: m.X + 12, Y: m.Y + m.H - 58, W: m.W - 24, H: 44}, len(g.planets(a)), len(g.planets(a)), 44)
}

func (g *GravityLab) Update(a *App) error {
	if g.tick(a) || g.busy() {
		return nil
	}
	if g.jumpT >= 0 {
		g.jumpT += 1.0 / 60
		if g.jumpT > jumpSecs {
			g.jumpT = -1
		}
	}
	if !g.answered {
		_, ids, answer := g.question(a)
		for i, r := range g.answerRects(a, len(ids)) {
			if a.Pointer.TapIn(r) {
				if ids[i] == answer {
					g.answered = true
					g.note, g.noteGood = a.T("gravity.right"), true
					g.succeed(a, "")
				} else {
					g.boom(a, "gravity.hint")
				}
			}
		}
	}
	ps := g.planets(a)
	for i, r := range g.planetRects(a) {
		if a.Pointer.TapIn(r) {
			g.planet, g.jumpT = ps[i].ID, 0
			a.Sound("jump")
			g.discoverQuietly(a, "gravity."+ps[i].ID)
		}
	}
	return nil
}

func heightWord(h float64) string {
	switch {
	case h < 0.2:
		return "gravity.word.tiny"
	case h < 0.5:
		return "gravity.word.small"
	case h < 2.5:
		return "gravity.word.high"
	}
	return "gravity.word.super"
}

func (g *GravityLab) Draw(a *App, screen *ebiten.Image) {
	g.drawFrame(a, screen)
	m := g.main(a)
	p, _ := experiment.Planet(g.planet)
	h := experiment.JumpHeight(p.G)

	// Space view, clipped to the main panel.
	view := screen.SubImage(image.Rect(int(m.X)+3, int(m.Y)+3, int(m.X+m.W)-3, int(m.Y+m.H)-3)).(*ebiten.Image)
	view.Fill(ui.Hex("#141633"))
	for i := 0; i < 40; i++ {
		sx := m.X + math.Mod(float64(i)*97.3, m.W)
		sy := m.Y + math.Mod(float64(i)*53.7, m.H*0.7)
		ui.Circle(view, sx, sy, 1+math.Mod(float64(i), 3)*0.5, ui.ColTextLight)
	}
	// Ground: a big curved planet surface.
	col := planetColors[g.planet]
	groundY := m.Y + m.H - 90
	ui.Circle(view, m.X+m.W/2, groundY+620, 640, ui.Hex(col[0]))
	ui.Circle(view, m.X+m.W*0.3, groundY+18, 10, ui.Hex(col[1]))
	ui.Circle(view, m.X+m.W*0.75, groundY+30, 14, ui.Hex(col[1]))

	// The jump: a parabola that takes the same time on every world, but
	// reaches the real height for that gravity.
	lift := 0.0
	if g.jumpT >= 0 {
		k := g.jumpT / jumpSecs
		lift = 4 * k * (1 - k) * math.Min(h*pxPerMetre, 230)
	}
	ax := m.X + m.W*0.5
	ui.DrawAvatar(view, a.Avatars, a.State.Avatar, ax, groundY+6-lift, 1.6, 1)
	if lift > 4 {
		ui.Line(view, ax+40, groundY, ax+40, groundY-lift, 2, ui.ColFocus)
	}

	// Info.
	info := a.T(heightWord(h))
	if a.Level() == content.Expert {
		info = a.Tf("gravity.height", fmt.Sprintf("%.2f m", h)) + "   " +
			a.Tf("gravity.fall", fmt.Sprintf("%.2f s", experiment.FallTime(2, p.G)))
	} else {
		info += "  (" + fmt.Sprintf("%.1f m", h) + ")"
	}
	ui.Text(screen, a.T(planetKey(g.planet))+": "+info, ui.Bold(ui.SizeBody), m.X+16, m.Y+100, ui.ColTextLight, text.AlignStart)

	if !g.answered {
		q := "gravity.q"
		if a.Level() == content.Expert {
			q = "gravity.q.expert"
		}
		lines := ui.Wrap(a.T(q), ui.Bold(ui.SizeBody), m.W-24)
		ui.Lines(screen, lines, ui.Bold(ui.SizeBody), ui.SizeBody, m.X+m.W/2, m.Y+8, ui.ColTextLight, text.AlignCenter)
		labels, _, _ := g.question(a)
		for i, r := range g.answerRects(a, len(labels)) {
			ui.Button{R: r, Label: labels[i]}.Draw(screen, false)
		}
	} else {
		ui.TextCentered(screen, a.T("gravity.tap"), ui.Bold(ui.SizeBody), m.X+m.W/2, m.Y+22, ui.ColTextLight)
	}
	ps := g.planets(a)
	for i, r := range g.planetRects(a) {
		ui.Button{R: r, Label: a.T(planetKey(ps[i].ID))}.Draw(screen, ps[i].ID == g.planet)
	}
	g.drawOverlay(a, screen)
}
