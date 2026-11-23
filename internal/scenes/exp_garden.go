package scenes

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

const toolCan = "can"

var flowerColors = map[string][2]string{
	"sunflower": {"#ffd23f", "#8d6040"},
	"tulip":     {"#ff5a7a", "#ff5a7a"},
	"rose":      {"#e8312f", "#b01f2a"},
	"daisy":     {"#ffffff", "#ffd23f"},
}

// GardenLab: grow one flower or a whole garden (E-BIO-2). Plants need
// water and light, but not too much water.
type GardenLab struct {
	expBase
	full bool
	pots []experiment.Pot
	tool string // a seed id, toolCan, or ""
	sun  bool
	day  int
}

func NewGardenLab() *GardenLab {
	g := &GardenLab{expBase: expBase{titleKey: "garden.title", bg: ui.Hex("#a8e6b0"), expID: "garden"}, sun: true}
	g.setMode(false)
	return g
}

func (g *GardenLab) Opaque() bool { return true }

func (g *GardenLab) setMode(full bool) {
	n := 1
	if full {
		n = experiment.GardenSize
	}
	g.full, g.pots, g.day, g.note, g.tool = full, make([]experiment.Pot, n), 1, "", ""
}

func (g *GardenLab) tools() []string {
	return append(append([]string{}, experiment.Flowers...), toolCan)
}

func (g *GardenLab) modeRects(a *App) []geom.Rect {
	m := g.main(a)
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 8, W: min(280, m.W-100)}, 2, 2, 32)
}

func (g *GardenLab) toolRects(a *App) []geom.Rect {
	m := g.main(a)
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 48, W: m.W - 20}, len(g.tools()), len(g.tools()), 34)
}

func (g *GardenLab) potArea(a *App) geom.Rect {
	m := g.main(a)
	return geom.Rect{X: m.X + 10, Y: m.Y + 90, W: m.W - 20, H: m.H - 190}
}

func (g *GardenLab) potRects(a *App) []geom.Rect {
	area := g.potArea(a)
	if !g.full {
		return []geom.Rect{geom.Centered(area.X+area.W/2, area.Y+area.H/2+10, 140, area.H-20)}
	}
	h := area.H/2 - 4
	return grid(geom.Rect{X: area.X, Y: area.Y + 2, W: area.W}, len(g.pots), 3, h)
}

func (g *GardenLab) controlRects(a *App) (lamp, next geom.Rect) {
	m := g.main(a)
	y := m.Y + m.H - 92
	lamp = geom.Rect{X: m.X + 12, Y: y, W: m.W/2 - 18, H: 38}
	next = geom.Rect{X: m.X + m.W/2 + 6, Y: y, W: m.W/2 - 18, H: 38}
	return
}

func (g *GardenLab) Update(a *App) error {
	if g.tick(a) || g.busy() {
		return nil
	}
	for i, r := range g.modeRects(a) {
		if a.Pointer.TapIn(r) && (i == 1) != g.full {
			g.setMode(i == 1)
		}
	}
	for i, r := range g.toolRects(a) {
		if a.Pointer.TapIn(r) {
			g.tool = g.tools()[i]
		}
	}
	for i, r := range g.potRects(a) {
		if a.Pointer.TapIn(r) {
			g.usePot(a, i)
		}
	}
	lamp, next := g.controlRects(a)
	if a.Pointer.TapIn(lamp) {
		g.sun = !g.sun
		a.Sound("click")
	}
	if a.Pointer.TapIn(next) {
		g.nextDay(a)
	}
	return nil
}

// usePot plants the chosen seed or waters the pot.
func (g *GardenLab) usePot(a *App, i int) {
	p := &g.pots[i]
	switch {
	case g.tool == toolCan && p.Planted():
		p.AddWater()
		a.Sound("pickup")
	case g.tool == toolCan || g.tool == "":
		g.fizzle(a, "garden.hint.empty")
	case !p.Planted():
		p.Seed = g.tool
		a.Sound("place")
	}
}

// nextDay grows every pot, celebrates new flowers, and hints at problems.
func (g *GardenLab) nextDay(a *App) {
	g.day++
	var bloomed []string
	problem, grew := "", false
	for i := range g.pots {
		p := &g.pots[i]
		was := p.Bloomed()
		switch p.NextDay(g.sun) {
		case experiment.GrowOK:
			grew = true
		case experiment.GrowSoggy:
			problem = "garden.hint.soggy"
		case experiment.GrowDry:
			if problem != "garden.hint.soggy" {
				problem = "garden.hint.dry"
			}
		case experiment.GrowDark:
			if problem == "" {
				problem = "garden.hint.dark"
			}
		}
		if !was && p.Bloomed() {
			bloomed = append(bloomed, p.Seed)
		}
	}
	if problem == "" && !grew && len(bloomed) == 0 && !g.anyPlanted() {
		problem = "garden.hint.empty"
	}
	g.note, g.noteGood = "", true
	if grew {
		g.note = a.T("garden.grew")
	}
	if len(bloomed) > 0 {
		for _, f := range bloomed[1:] {
			g.discoverQuietly(a, "garden."+f)
		}
		g.note = a.Tf("garden.bloom", a.T(flowerKey(bloomed[0])))
		disc := "garden." + bloomed[0]
		if g.full && experiment.GardenBloomed(g.pots) {
			g.discoverQuietly(a, disc)
			disc, g.note = "garden.full", a.T("garden.fullDone")
		}
		g.succeed(a, disc)
	}
	if problem != "" {
		g.fizzle(a, problem)
	}
}

func (g *GardenLab) anyPlanted() bool {
	for _, p := range g.pots {
		if p.Planted() {
			return true
		}
	}
	return false
}

func (g *GardenLab) Draw(a *App, screen *ebiten.Image) {
	g.drawFrame(a, screen)
	m := g.main(a)
	mr := g.modeRects(a)
	smallButton(screen, mr[0], a.T("garden.mode.one"), !g.full)
	smallButton(screen, mr[1], a.T("garden.mode.all"), g.full)
	ui.Text(screen, a.Tf("garden.day", g.day), ui.Bold(ui.SizeBody), m.X+m.W-14, m.Y+12, ui.ColText, text.AlignEnd)
	for i, r := range g.toolRects(a) {
		id := g.tools()[i]
		label := a.T("garden.can")
		if id != toolCan {
			label = a.T(flowerKey(id))
		}
		smallButton(screen, r, label, id == g.tool)
	}

	area := g.potArea(a)
	if !g.sun {
		ui.RoundRect(screen, area, 10, color.RGBA{0x20, 0x20, 0x50, 0x50}, 0, nil)
	} else {
		glow := 0.5 + 0.5*math.Sin(g.t*2)
		ui.RoundRect(screen, geom.Rect{X: area.X + area.W/2 - 70, Y: area.Y - 4, W: 140, H: 10}, 4, ui.Hex("#fff3a0"), 2, ui.ColWall)
		ui.RoundRect(screen, geom.Rect{X: area.X + area.W/2 - 90, Y: area.Y + 6, W: 180, H: 18}, 9, color.RGBA{0xff, 0xf3, 0xa0, uint8(40 + 40*glow)}, 0, nil)
	}
	scale := 1.4
	if g.full {
		scale = 0.75
	}
	for i, r := range g.potRects(a) {
		drawPot(screen, g.pots[i], r, scale, g.t)
	}

	lamp, next := g.controlRects(a)
	lk := "garden.lamp.off"
	if g.sun {
		lk = "garden.lamp.on"
	}
	smallButton(screen, lamp, a.T(lk), g.sun)
	ui.Button{R: next, Label: a.T("garden.next")}.Draw(screen, false)
	g.drawOverlay(a, screen)
}

// drawPot draws a flowerpot with its plant at the right stage.
func drawPot(screen *ebiten.Image, p experiment.Pot, r geom.Rect, u, t float64) {
	cx, base := r.X+r.W/2, r.Y+r.H-4
	w, h := 44*u, 34*u
	top := base - h
	ui.Poly(screen, ui.Hex("#c8643a"), cx-w/2, top, cx+w/2, top, cx+w*0.38, base, cx-w*0.38, base)
	ui.RoundRect(screen, geom.Rect{X: cx - w/2 - 3, Y: top - 6*u, W: w + 6, H: 8 * u}, 3, ui.Hex("#d8784a"), 1.5, ui.ColWall)
	soil := ui.Hex("#6a4a32")
	if p.Soggy() {
		soil = ui.Hex("#3a4a6a")
	}
	ui.RoundRect(screen, geom.Rect{X: cx - w/2 + 3, Y: top - 4*u, W: w - 6, H: 4 * u}, 2, soil, 0, nil)
	for i := 0; i < min(p.Water, experiment.MaxWater+1); i++ {
		dc := ui.Hex("#4f8df0")
		if i >= experiment.MaxWater {
			dc = ui.ColLocked
		}
		ui.Circle(screen, cx+w/2+8+float64(i%2)*6, base-6-float64(i)*7*u, 3*u, dc)
	}
	if !p.Planted() {
		return
	}
	soilY := top - 4*u
	sway := math.Sin(t*1.5) * 2 * u
	col := flowerColors[p.Seed]
	stemTop := soilY - float64(p.Stage)*14*u
	switch {
	case p.Stage == 0:
		ui.Circle(screen, cx, soilY-1, 3*u, ui.Hex("#c8a060"))
		return
	default:
		ui.Line(screen, cx, soilY, cx+sway, stemTop, 3*u, ui.Hex("#3aa858"))
	}
	if p.Stage >= 2 {
		ly := soilY - 14*u
		ui.Poly(screen, ui.Hex("#5ad85a"), cx, ly, cx-14*u, ly-8*u, cx-4*u, ly+2*u)
		ui.Poly(screen, ui.Hex("#5ad85a"), cx, ly-6*u, cx+14*u, ly-14*u, cx+4*u, ly-4*u)
	}
	fx := cx + sway
	switch {
	case p.Stage == 1:
		ui.Circle(screen, fx, stemTop, 4*u, ui.Hex("#5ad85a"))
	case p.Stage == experiment.BloomStage-1:
		ui.Circle(screen, fx, stemTop, 6*u, ui.Hex(col[0]))
		ui.Poly(screen, ui.Hex("#3aa858"), fx-6*u, stemTop+2*u, fx+6*u, stemTop+2*u, fx, stemTop+8*u)
	case p.Bloomed():
		drawFlower(screen, p.Seed, fx, stemTop, u)
	}
}

func drawFlower(screen *ebiten.Image, kind string, x, y, u float64) {
	col := flowerColors[kind]
	switch kind {
	case "tulip":
		ui.Poly(screen, ui.Hex(col[0]), x-10*u, y-14*u, x-4*u, y-6*u, x, y-16*u, x+4*u, y-6*u, x+10*u, y-14*u, x+9*u, y+2*u, x-9*u, y+2*u)
	case "rose":
		ui.Circle(screen, x, y, 10*u, ui.Hex(col[0]))
		ui.Arc(screen, x, y, 6*u, 0, math.Pi*1.5, 2*u, ui.Hex(col[1]))
		ui.Arc(screen, x+1*u, y, 3*u, math.Pi, math.Pi*2.4, 2*u, ui.Hex(col[1]))
	default: // sunflower, daisy: petals around a center
		n, pr, cr := 10, 7*u, 6*u
		if kind == "daisy" {
			n, pr, cr = 12, 5*u, 4*u
		}
		for i := 0; i < n; i++ {
			ang := float64(i) * 2 * math.Pi / float64(n)
			ui.Circle(screen, x+math.Cos(ang)*(cr+pr*0.7), y+math.Sin(ang)*(cr+pr*0.7), pr*0.7, ui.Hex(col[0]))
		}
		ui.Circle(screen, x, y, cr, ui.Hex(col[1]))
	}
}
