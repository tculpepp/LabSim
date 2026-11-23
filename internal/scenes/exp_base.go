package scenes

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

const (
	boomTicks    = 70
	sparkleTicks = 60
)

// expBase is shared by the experiment screens: header, back button, the
// scientist side panel, funny booms (E-4), and success handling (E-5, E-6).
type expBase struct {
	titleKey  string
	darkBG    bool
	bg        color.Color
	expID     string // for job completion ("slime", "circuit", ...)
	t         float64
	boomT     int
	hintKey   string // shown by Gizmo after the boom settles
	sparkleT  int
	note      string // short status line under the main area
	noteGood  bool
	pendingFC string // discovery whose fact card opens after the sparkle
	pendingNew bool
}

func (e *expBase) backRect(a *App) geom.Rect {
	return geom.Rect{X: float64(a.W) - 120, Y: 10, W: 104, H: 40}
}

// side is the left panel with the scientist.
func (e *expBase) side(a *App) geom.Rect { return geom.Rect{X: 16, Y: 60, W: 200, H: float64(a.H) - 76} }

// main is the experiment area.
func (e *expBase) main(a *App) geom.Rect {
	return geom.Rect{X: 228, Y: 60, W: float64(a.W) - 244, H: float64(a.H) - 76}
}

// tick advances animations; returns true if the back button/Esc was pressed.
func (e *expBase) tick(a *App) bool {
	e.t += 1.0 / 60
	if e.boomT > 0 {
		e.boomT--
		if e.boomT == boomTicks/3 && e.hintKey != "" {
			a.Say("gizmo", "gizmo.name", a.T(e.hintKey))
			e.hintKey = ""
		}
	}
	if e.sparkleT > 0 {
		e.sparkleT--
		if e.sparkleT == 0 && e.pendingFC != "" {
			a.Push(NewFactCard(e.pendingFC, e.pendingNew))
			e.pendingFC = ""
		}
	}
	if closePressed() || a.Pointer.TapIn(e.backRect(a)) {
		a.Pop()
		return true
	}
	return false
}

// busy is true while an animation plays (ignore input).
func (e *expBase) busy() bool { return e.boomT > boomTicks/3 || e.sparkleT > 0 }

// boom: a cartoon explosion, soot on your face, then Gizmo's hint (E-4).
func (e *expBase) boom(a *App, hintKey string) {
	e.boomT, e.hintKey = boomTicks, hintKey
	e.note, e.noteGood = a.T("exp.boom"), false
	a.Sound("boom")
}

// fizzle is a gentle "not quite" with an immediate hint.
func (e *expBase) fizzle(a *App, hintKey string) {
	a.Sound("fizzle")
	a.Say("gizmo", "gizmo.name", a.T(hintKey))
}

// succeed records a discovery, pays any matching job, and queues the fun fact.
func (e *expBase) succeed(a *App, discovery string) {
	e.sparkleT = sparkleTicks
	a.Sound("success")
	isNew := discovery != "" && a.Discover(discovery)
	if paid := a.ExperimentDone(e.expID); paid > 0 {
		e.note, e.noteGood = a.Tf("exp.jobPaid", paid), true
	}
	if isNew {
		e.pendingFC, e.pendingNew = discovery, true
	}
}

// discoverQuietly records a discovery without a celebration (e.g. each
// material tested in the circuit). Returns true if new.
func (e *expBase) discoverQuietly(a *App, discovery string) bool {
	if a.Discover(discovery) {
		e.pendingFC, e.pendingNew = discovery, true
		e.sparkleT = max(e.sparkleT, sparkleTicks/2)
		return true
	}
	return false
}

func (e *expBase) drawFrame(a *App, screen *ebiten.Image) {
	screen.Fill(e.bg)
	tc := ui.ColText
	if e.darkBG {
		tc = ui.ColTextLight
	}
	ui.Text(screen, a.T(e.titleKey), ui.Bold(ui.SizeHeading+4), 20, 14, tc, text.AlignStart)
	ui.Button{R: e.backRect(a), Label: a.T("exp.back")}.Draw(screen, false)
	W := float64(a.W)
	ui.RoundRect(screen, geom.Rect{X: W - 250, Y: 12, W: 118, H: 36}, 18, ui.ColPanel, 2, ui.ColPanelLine)
	ui.Diamond(screen, W-230, 30, 9, ui.ColCrystal)
	ui.Text(screen, fmt.Sprint(a.State.Wallet.Crystals()), ui.Bold(ui.SizeBody), W-214, 19, ui.ColText, text.AlignStart)

	s := e.side(a)
	ui.Panel(screen, s)
	sx := s.X + s.W/2
	feet := s.Y + s.H - 30
	shake := 0.0
	if e.boomT > 0 {
		shake = math.Sin(e.t*60) * float64(e.boomT) / 12
	}
	ui.DrawAvatar(screen, a.Avatars, a.State.Avatar, sx+shake, feet, 2.6, 1)
	if e.boomT > 0 || e.boomT == 0 && e.note == a.T("exp.boom") {
		drawSoot(screen, sx+shake, feet, 2.6)
	}
	ui.DrawGizmo(screen, s.X+40, s.Y+40, 1.6, e.t)
	m := e.main(a)
	ui.Panel(screen, m)
}

// drawOverlay draws boom and sparkle effects plus the status note on top.
func (e *expBase) drawOverlay(a *App, screen *ebiten.Image) {
	m := e.main(a)
	cx, cy := m.Center()
	if e.boomT > 0 {
		k := 1 - float64(e.boomT)/boomTicks
		for i := 0; i < 9; i++ {
			ang := float64(i) * 2 * math.Pi / 9
			r := 30 + k*110
			ui.Circle(screen, cx+math.Cos(ang)*r*0.8, cy+math.Sin(ang)*r*0.6, 34*(1-k*0.6), color.RGBA{0x70, 0x70, 0x78, uint8(200 * (1 - k))})
		}
		ui.TextOutlined(screen, a.T("exp.boom"), ui.Bold(ui.SizeTitle), cx, cy-20*k, ui.ColFocus, ui.ColWall)
	}
	if e.sparkleT > 0 {
		k := 1 - float64(e.sparkleT)/sparkleTicks
		for i := 0; i < 12; i++ {
			ang := float64(i)*2*math.Pi/12 + k
			r := 20 + k*140
			ui.Diamond(screen, cx+math.Cos(ang)*r, cy+math.Sin(ang)*r*0.7, 8*(1-k)+2, ui.ColFocus)
		}
	}
	if e.note != "" {
		c := ui.ColLocked
		if e.noteGood {
			c = ui.Hex("#2f9e4f")
		}
		ui.TextCentered(screen, e.note, ui.Bold(ui.SizeBody), cx, m.Y+m.H-22, c)
	}
}

// drawSoot smudges the scientist's face after a boom (E-4).
func drawSoot(screen *ebiten.Image, x, feet, u float64) {
	hy := feet - (4+16)*u - 7*u
	soot := color.RGBA{0x30, 0x30, 0x30, 0x90}
	for _, d := range [][3]float64{{-5, -3, 2.2}, {4, -5, 1.8}, {6, 3, 2}, {-6, 4, 1.6}, {0, -7, 2.4}} {
		ui.Circle(screen, x+d[0]*u, hy+d[1]*u, d[2]*u, soot)
	}
	ui.Circle(screen, x-8*u, hy-12*u, 3*u, color.RGBA{0x90, 0x90, 0x98, 0x80})
	ui.Circle(screen, x+7*u, hy-14*u, 2.5*u, color.RGBA{0x90, 0x90, 0x98, 0x70})
}

// tile is a tappable card with an icon and a label.
type tile struct {
	r     geom.Rect
	label string
	sub   string
	icon  func(screen *ebiten.Image, x, y float64)
	on    bool
}

func (t tile) draw(screen *ebiten.Image, focus bool) {
	fill := ui.ColPanel
	if t.on {
		fill = ui.Hex("#d8ecff")
	}
	line := ui.ColPanelLine
	if focus {
		line = ui.ColFocus
	}
	ui.RoundRect(screen, t.r, 10, fill, 2, line)
	if t.icon != nil {
		t.icon(screen, t.r.X+t.r.W/2, t.r.Y+t.r.H/2-8)
	}
	ui.TextCentered(screen, t.label, ui.Regular(ui.SizeSmall), t.r.X+t.r.W/2, t.r.Y+t.r.H-14, ui.ColText)
	if t.sub != "" {
		ui.TextCentered(screen, t.sub, ui.Bold(ui.SizeSmall), t.r.X+t.r.W-12, t.r.Y+12, ui.ColAccent)
	}
}

// grid lays out n tiles in rows inside area.
func grid(area geom.Rect, n, cols int, h float64) []geom.Rect {
	w := area.W / float64(cols)
	out := make([]geom.Rect, n)
	for i := range out {
		out[i] = geom.Rect{X: area.X + float64(i%cols)*w + 3, Y: area.Y + float64(i/cols)*(h+6), W: w - 6, H: h}
	}
	return out
}
