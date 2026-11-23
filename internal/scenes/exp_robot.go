package scenes

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const robotRunSecs = 2.5

// RobotLab: snap together a robot and test it on a challenge (RW-2).
type RobotLab struct {
	expBase
	bot       experiment.Robot
	challenge string
	passed    bool
	runT      float64
}

func NewRobotLab() *RobotLab {
	return &RobotLab{expBase: expBase{titleKey: "robot.title", bg: ui.Hex("#b8b8b8"), expID: "robot"},
		bot: experiment.Robot{}, challenge: "floor", runT: robotRunSecs}
}

func (r *RobotLab) Opaque() bool { return true }

func (r *RobotLab) challenges(a *App) []experiment.RobotChallenge {
	return experiment.RobotChallenges(a.Level() == content.Expert)
}

func (r *RobotLab) partRects(a *App, row int) []geom.Rect {
	m := r.main(a)
	return grid(geom.Rect{X: m.X + 84, Y: m.Y + 8 + float64(row)*40, W: m.W - 94}, 3, 3, 34)
}

func (r *RobotLab) challengeRects(a *App) []geom.Rect {
	m := r.main(a)
	n := len(r.challenges(a))
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 152, W: m.W - 20}, n, n, 32)
}

func (r *RobotLab) testRect(a *App) geom.Rect {
	m := r.main(a)
	return geom.Rect{X: m.X + m.W - 130, Y: m.Y + 192, W: 118, H: 40}
}

func (r *RobotLab) stage(a *App) geom.Rect {
	m := r.main(a)
	return geom.Rect{X: m.X + 10, Y: m.Y + 240, W: m.W - 20, H: m.H - 276}
}

func (r *RobotLab) Update(a *App) error {
	if r.tick(a) || r.busy() {
		return nil
	}
	if r.runT < robotRunSecs {
		r.runT += 1.0 / 60
	}
	for row, slot := range experiment.RobotSlots {
		for i, rr := range r.partRects(a, row) {
			if a.Pointer.TapIn(rr) {
				r.bot[slot] = experiment.RobotParts[slot][i]
				r.passed, r.note = false, ""
				a.Sound("place")
			}
		}
	}
	cs := r.challenges(a)
	for i, rr := range r.challengeRects(a) {
		if a.Pointer.TapIn(rr) {
			r.challenge, r.passed, r.note = cs[i].ID, false, ""
		}
	}
	if a.Pointer.TapIn(r.testRect(a)) {
		r.test(a)
	}
	return nil
}

func (r *RobotLab) test(a *App) {
	var ch experiment.RobotChallenge
	for _, c := range r.challenges(a) {
		if c.ID == r.challenge {
			ch = c
		}
	}
	res := experiment.TryRobot(r.bot, ch)
	switch {
	case res.OK:
		r.passed, r.runT = true, 0
		r.note, r.noteGood = a.T("robot.passed"), true
		r.succeed(a, res.Discovery)
	case res.Boom:
		r.passed, r.runT = false, 0
		r.boom(a, res.Hint)
	default:
		r.fizzle(a, res.Hint)
	}
}

func (r *RobotLab) Draw(a *App, screen *ebiten.Image) {
	r.drawFrame(a, screen)
	m := r.main(a)
	for row, slot := range experiment.RobotSlots {
		rects := r.partRects(a, row)
		ui.Text(screen, a.T(robotSlotKey(slot)), ui.Bold(ui.SizeSmall), m.X+12, rects[0].Y+9, ui.ColText, text.AlignStart)
		for i, rr := range rects {
			p := experiment.RobotParts[slot][i]
			smallButton(screen, rr, a.T(robotPartKey(p)), r.bot[slot] == p)
		}
	}
	ui.Text(screen, a.T("robot.test"), ui.Bold(ui.SizeSmall), m.X+12, m.Y+130, ui.ColText, text.AlignStart)
	cs := r.challenges(a)
	for i, rr := range r.challengeRects(a) {
		smallButton(screen, rr, a.T(robotChKey(cs[i].ID)), cs[i].ID == r.challenge)
		if a.State.Notebook.Has("robot." + cs[i].ID) {
			ui.Diamond(screen, rr.X+rr.W-8, rr.Y+8, 5, ui.ColCrystal)
		}
	}
	goal := ui.Wrap(a.T(robotGoalKey(r.challenge)), ui.Regular(ui.SizeSmall), m.W-150)
	ui.Lines(screen, goal, ui.Regular(ui.SizeSmall), ui.SizeSmall, m.X+12, m.Y+196, ui.ColText, text.AlignStart)
	ui.Button{R: r.testRect(a), Label: a.T("robot.go")}.Draw(screen, r.bot.Complete())

	st := r.stage(a)
	view := screen.SubImage(image.Rect(int(st.X), int(st.Y), int(st.X+st.W), int(st.Y+st.H))).(*ebiten.Image)
	drawTerrain(view, r.challenge, st)
	k := math.Min(r.runT/robotRunSecs, 1)
	x := st.X + 50
	if r.passed {
		x += k * (st.W - 100)
	}
	y := st.Y + st.H - 14
	if r.challenge == "stairs" && r.passed {
		y -= k * stairRise(st)
	}
	if r.boomT > 0 {
		x += math.Sin(r.t*50) * 4
	}
	drawRobot(view, r.bot, x, y, 0.8, r.t)
	r.drawOverlay(a, screen)
}

func drawTerrain(screen *ebiten.Image, id string, st geom.Rect) {
	ground := st.Y + st.H - 14
	switch id {
	case "floor":
		ui.RoundRect(screen, st, 8, ui.Hex("#e8eef6"), 0, nil)
		ui.Fill(screen, geom.Rect{X: st.X, Y: ground, W: st.W, H: 14}, 0, 0, ui.Hex("#9aa6b8"))
	case "sand":
		ui.RoundRect(screen, st, 8, ui.Hex("#ffe9b8"), 0, nil)
		ui.Fill(screen, geom.Rect{X: st.X, Y: ground, W: st.W, H: 14}, 0, 0, ui.Hex("#f0c870"))
		for i := 0; i < 5; i++ {
			ui.Circle(screen, st.X+float64(i)*st.W/4, ground+18, 24, ui.Hex("#f0c870"))
		}
	case "stairs":
		ui.RoundRect(screen, st, 8, ui.Hex("#e8eef6"), 0, nil)
		n := 5
		for i := 0; i < n; i++ {
			h := 14 + float64(i+1)*stairRise(st)/float64(n)
			sx := st.X + 90 + float64(i)*(st.W-110)/float64(n)
			ui.Fill(screen, geom.Rect{X: sx, Y: st.Y + st.H - h, W: st.X + st.W - sx, H: h}, 0, 0, ui.Hex("#b89a7a"))
		}
	case "cave":
		ui.RoundRect(screen, st, 8, ui.Hex("#1a1a2a"), 0, nil)
		for i := 0; i < 6; i++ {
			ui.Circle(screen, st.X+40+float64(i)*st.W/6, ground+6, 12+float64(i%3)*5, ui.Hex("#4a4a5a"))
		}
	case "junkyard":
		ui.RoundRect(screen, st, 8, ui.Hex("#e8d8c0"), 0, nil)
		ui.Fill(screen, geom.Rect{X: st.X, Y: ground, W: st.W, H: 14}, 0, 0, ui.Hex("#8d6040"))
		for i := 0; i < 6; i++ {
			bx := st.X + 70 + float64(i)*st.W/7
			ui.RoundRect(screen, geom.Centered(bx, ground-3, 12, 6), 2, ui.Hex("#9aa6b8"), 1, ui.ColWall)
		}
	}
}

// drawRobot draws a robot from its parts; (x, y) is where it touches the ground.
func drawRobot(screen *ebiten.Image, bot experiment.Robot, x, y, u, t float64) {
	metal, dark := ui.Hex("#c8d0dc"), ui.ColWall
	ghost := color.RGBA{0x30, 0x60, 0xd0, 0x50}
	// Movers.
	switch bot["movers"] {
	case "wheels":
		for _, dx := range []float64{-14, 14} {
			ui.Circle(screen, x+dx*u, y-8*u, 8*u, dark)
			ui.Circle(screen, x+dx*u, y-8*u, 3*u, metal)
		}
	case "tracks":
		ui.RoundRect(screen, geom.Rect{X: x - 26*u, Y: y - 16*u, W: 52 * u, H: 16 * u}, 8*u, dark, 0, nil)
		for i := 0; i < 4; i++ {
			ui.Circle(screen, x+(-18+float64(i)*12)*u, y-8*u, 4*u, metal)
		}
	case "legs":
		step := math.Sin(t*6) * 4 * u
		ui.Line(screen, x-8*u, y-18*u, x-8*u+step, y, 4*u, dark)
		ui.Line(screen, x+8*u, y-18*u, x+8*u-step, y, 4*u, dark)
	default:
		ui.RoundRect(screen, geom.Rect{X: x - 22*u, Y: y - 16*u, W: 44 * u, H: 16 * u}, 6, nil, 2, ghost)
	}
	// Body.
	body := geom.Rect{X: x - 20*u, Y: y - 50*u, W: 40 * u, H: 34 * u}
	ui.RoundRect(screen, body, 6*u, metal, 2, dark)
	ui.Circle(screen, x, y-33*u, 5*u, ui.Hex("#ffd23f"))
	// Arms.
	for _, side := range []float64{-1, 1} {
		sx := x + side*20*u
		ex := sx + side*12*u
		switch bot["arms"] {
		case "claw":
			ui.Line(screen, sx, y-42*u, ex, y-30*u, 3*u, dark)
			ui.Line(screen, ex, y-30*u, ex+side*4*u, y-24*u, 2*u, dark)
			ui.Line(screen, ex, y-30*u, ex-side*3*u, y-23*u, 2*u, dark)
		case "magnet":
			ui.Line(screen, sx, y-42*u, ex, y-32*u, 3*u, dark)
			ui.Arc(screen, ex, y-28*u, 5*u, 0, math.Pi, 3*u, ui.ColLocked)
		case "scoop":
			ui.Line(screen, sx, y-42*u, ex, y-32*u, 3*u, dark)
			ui.Arc(screen, ex, y-30*u, 6*u, 0, math.Pi, 3*u, ui.Hex("#4f8df0"))
		default:
			ui.Line(screen, sx, y-42*u, ex, y-30*u, 2, ghost)
		}
	}
	// Head.
	head := geom.Rect{X: x - 15*u, Y: y - 74*u, W: 30 * u, H: 22 * u}
	ui.RoundRect(screen, head, 5*u, metal, 2, dark)
	switch bot["head"] {
	case "camera":
		for _, dx := range []float64{-7, 7} {
			ui.Circle(screen, x+dx*u, y-63*u, 5*u, dark)
			ui.Circle(screen, x+dx*u, y-63*u, 2*u, ui.ColCrystal)
		}
	case "antenna":
		ui.Line(screen, x, y-74*u, x, y-88*u, 2*u, dark)
		blink := uint8(0x80 + 0x7f*(0.5+0.5*math.Sin(t*5)))
		ui.Circle(screen, x, y-90*u, 4*u, color.RGBA{0xe0, 0x5a, 0x5a, blink})
		ui.Fill(screen, geom.Rect{X: x - 8*u, Y: y - 64*u, W: 16 * u, H: 3 * u}, 0, 0, dark)
	case "sonar":
		for _, side := range []float64{-1, 1} {
			ui.Arc(screen, x+side*15*u, y-63*u, 7*u, math.Pi/2+side*math.Pi/2-math.Pi/2, math.Pi/2+side*math.Pi/2+math.Pi/2, 3*u, dark)
		}
		ui.Fill(screen, geom.Rect{X: x - 8*u, Y: y - 66*u, W: 16 * u, H: 4 * u}, 0, 0, ui.ColCrystal)
	}
}

// stairRise is how high the stairs climb, leaving room for the robot on top.
func stairRise(st geom.Rect) float64 { return max(st.H-90, 10) }
