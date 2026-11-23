package scenes

import (
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const gforceKid = 30 // kg, for the Expert question

// GForceLab: ride things and feel how hard they push you (E-PHYS-2).
type GForceLab struct {
	expBase
	current  string
	rideT    float64
	answered bool
}

func NewGForceLab() *GForceLab {
	return &GForceLab{expBase: expBase{titleKey: "gforce.title", bg: ui.Hex("#ffe8a8"), expID: "gforce"}, current: "still"}
}

func (g *GForceLab) Opaque() bool { return true }

// question returns the answer choices and the right one for this level.
func (g *GForceLab) question(a *App) (labels, ids []string, answer string) {
	if a.Level() == content.Expert {
		for _, kg := range []int{30, 60, 120, 300} {
			labels, ids = append(labels, a.Tf("gforce.kg", kg)), append(ids, strconv.Itoa(kg))
		}
		coaster, _ := experiment.Ride("coaster")
		return labels, ids, strconv.Itoa(int(experiment.FeltWeight(gforceKid, coaster.G)))
	}
	for _, r := range experiment.Rides() {
		labels, ids = append(labels, a.T(rideKey(r.ID))), append(ids, r.ID)
	}
	return labels, ids, experiment.MostG(experiment.Rides())
}

func (g *GForceLab) questionKey(a *App) string {
	if a.Level() == content.Expert {
		return "gforce.q.expert"
	}
	return "gforce.q"
}

// answerRects sit just under the (possibly wrapped) question.
func (g *GForceLab) answerRects(a *App, n int) []geom.Rect {
	m := g.main(a)
	lines := len(ui.Wrap(a.T(g.questionKey(a)), ui.Bold(ui.SizeBody), m.W-24))
	y := m.Y + 14 + float64(lines)*ui.LineHeight(ui.SizeBody)
	return grid(geom.Rect{X: m.X + 10, Y: y, W: m.W - 20}, n, n, 34)
}

func (g *GForceLab) rideRects(a *App) []geom.Rect {
	m := g.main(a)
	n := len(experiment.Rides())
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + m.H - 52, W: m.W - 20}, n, n, 40)
}

func (g *GForceLab) Update(a *App) error {
	if g.tick(a) || g.busy() {
		return nil
	}
	g.rideT += 1.0 / 60
	if !g.answered {
		_, ids, _ := g.question(a)
		for i, r := range g.answerRects(a, len(ids)) {
			if a.Pointer.TapIn(r) {
				g.answer(a, ids[i])
			}
		}
	}
	for i, r := range g.rideRects(a) {
		if a.Pointer.TapIn(r) {
			g.ride(a, experiment.Rides()[i].ID)
		}
	}
	return nil
}

func (g *GForceLab) answer(a *App, id string) {
	if _, _, right := g.question(a); id != right {
		g.boom(a, "gforce.hint")
		return
	}
	g.answered = true
	g.note, g.noteGood = a.T("gforce.right"), true
	g.succeed(a, "")
}

func (g *GForceLab) ride(a *App, id string) {
	g.current, g.rideT = id, 0
	a.Sound("jump")
	g.discoverQuietly(a, "gforce."+id)
}

func gWord(gs float64) string {
	switch {
	case gs <= 1:
		return "gforce.word.normal"
	case gs < 2:
		return "gforce.word.light"
	case gs < 6:
		return "gforce.word.heavy"
	}
	return "gforce.word.mega"
}

func (g *GForceLab) Draw(a *App, screen *ebiten.Image) {
	g.drawFrame(a, screen)
	m := g.main(a)
	r, _ := experiment.Ride(g.current)

	if !g.answered {
		lines := ui.Wrap(a.T(g.questionKey(a)), ui.Bold(ui.SizeBody), m.W-24)
		ui.Lines(screen, lines, ui.Bold(ui.SizeBody), ui.SizeBody, m.X+m.W/2, m.Y+6, ui.ColText, text.AlignCenter)
		labels, _, _ := g.question(a)
		for i, rr := range g.answerRects(a, len(labels)) {
			smallButton(screen, rr, labels[i], false)
		}
	} else {
		ui.TextCentered(screen, a.T("gforce.tap"), ui.Bold(ui.SizeBody), m.X+m.W/2, m.Y+22, ui.ColText)
	}

	// The rider: one push arrow per g, and they sink into the seat.
	cx, seatY := m.X+m.W*0.36, m.Y+m.H-110
	shake := 0.0
	if r.G > 1 {
		shake = math.Sin(g.rideT*40) * math.Min(r.G, 9) * 0.3
	}
	sink := math.Min(r.G-1, 8) * 1.5
	ui.RoundRect(screen, geom.Rect{X: cx - 44, Y: seatY - 4, W: 88, H: 22}, 8, ui.Hex("#d8423a"), 2, ui.ColWall)
	ui.RoundRect(screen, geom.Rect{X: cx + 30, Y: seatY - 80, W: 16, H: 80}, 6, ui.Hex("#d8423a"), 2, ui.ColWall)
	ui.DrawAvatar(screen, a.Avatars, a.State.Avatar, cx+shake, seatY+sink, 1.5, 1)
	arrows := int(math.Round(r.G))
	for i := 0; i < arrows; i++ {
		ax := cx - 40 + float64(i%5)*20
		ay := seatY - 150 + float64(i/5)*24 + 4*math.Sin(g.rideT*6+float64(i))
		ui.Line(screen, ax, ay, ax, ay+14, 3, ui.ColAccent)
		ui.Poly(screen, ui.ColAccent, ax-6, ay+12, ax+6, ay+12, ax, ay+20)
	}

	// The g-meter dial.
	dx, dy, dr := m.X+m.W*0.72, m.Y+m.H-175, min(56, m.W*0.15)
	ui.Circle(screen, dx, dy, dr, ui.ColPanel)
	ui.Ring(screen, dx, dy, dr, 3, ui.ColWall)
	for gv := 0; gv <= 10; gv += 2 {
		ang := math.Pi*0.75 + float64(gv)/10*math.Pi*1.5
		ui.Line(screen, dx+math.Cos(ang)*dr*0.78, dy+math.Sin(ang)*dr*0.78, dx+math.Cos(ang)*dr*0.95, dy+math.Sin(ang)*dr*0.95, 2, ui.ColWall)
	}
	shown := r.G
	if g.rideT < 0.6 {
		shown = 1 + (r.G-1)*g.rideT/0.6
	}
	ang := math.Pi*0.75 + math.Min(shown, 10)/10*math.Pi*1.5
	ui.Line(screen, dx, dy, dx+math.Cos(ang)*dr*0.8, dy+math.Sin(ang)*dr*0.8, 3, ui.ColLocked)
	ui.Circle(screen, dx, dy, 4, ui.ColWall)
	gs := strconv.FormatFloat(r.G, 'f', -1, 64)
	ui.TextCentered(screen, a.Tf("gforce.meter", a.T(rideKey(r.ID)), gs), ui.Bold(ui.SizeBody), dx, dy+dr+16, ui.ColText)
	info := a.T(gWord(r.G))
	if a.Level() == content.Expert {
		info = a.Tf("gforce.feel", int(experiment.FeltWeight(gforceKid, r.G)))
	}
	ui.TextCentered(screen, info, ui.Bold(ui.SizeSmall), dx, dy+dr+38, ui.ColAccent)

	for i, rr := range g.rideRects(a) {
		id := experiment.Rides()[i].ID
		smallButton(screen, rr, a.T(rideKey(id)), id == g.current)
	}
	g.drawOverlay(a, screen)
}
