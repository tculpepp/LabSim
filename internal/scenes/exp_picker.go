package scenes

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

// expEntry is one experiment you can run at a room's bench.
type expEntry struct {
	id       string // job experiment id; picker text is picker.desc.<id>
	titleKey string
	disc     string // discovery id prefix, for the progress count
	open     func() Scene
}

var experimentsByRoom = map[string][]expEntry{
	"chemistry": {
		{"slime", "slime.title", "slime.", func() Scene { return NewSlimeLab() }},
		{"crystal", "crystal.title", "crystal.", func() Scene { return NewCrystalLab() }},
	},
	"physics": {
		{"circuit", "circuit.title", "circuit.", func() Scene { return NewCircuitLab() }},
		{"gforce", "gforce.title", "gforce.", func() Scene { return NewGForceLab() }},
	},
	"greenhouse": {
		{"microscope", "micro.title", "micro.", func() Scene { return NewMicroLab() }},
		{"garden", "garden.title", "garden.", func() Scene { return NewGardenLab() }},
	},
	"space_lab": {
		{"gravity", "gravity.title", "gravity.", func() Scene { return NewGravityLab() }},
		{"trivia", "trivia.title", "trivia.", func() Scene { return NewTriviaLab() }},
	},
	"robot_workshop": {
		{"robot", "robot.title", "robot.", func() Scene { return NewRobotLab() }},
	},
}

// roomExperiments lists the experiments on offer in a room, in menu order.
func roomExperiments(room string) []expEntry { return experimentsByRoom[room] }

// ExpPicker lets the player choose which of the room's experiments to run.
type ExpPicker struct {
	entries []expEntry
	menu    ui.Menu
}

func NewExpPicker(room string) *ExpPicker { return &ExpPicker{entries: roomExperiments(room)} }

func (p *ExpPicker) Opaque() bool { return false }

func (p *ExpPicker) panel(a *App) geom.Rect {
	w := min(float64(a.W)-40, 560)
	h := 110 + float64(len(p.entries))*92
	return geom.Rect{X: (float64(a.W) - w) / 2, Y: (float64(a.H) - h) / 2, W: w, H: h}
}

func (p *ExpPicker) layout(a *App) {
	r := p.panel(a)
	p.menu.Buttons = p.menu.Buttons[:0]
	for i := range p.entries {
		p.menu.Buttons = append(p.menu.Buttons, ui.Button{R: geom.Rect{X: r.X + 20, Y: r.Y + 56 + float64(i)*92, W: r.W - 40, H: 82}})
	}
	p.menu.Buttons = append(p.menu.Buttons, ui.Button{R: geom.Rect{X: r.X + r.W/2 - 60, Y: r.Y + r.H - 46, W: 120, H: 38}, Label: a.T("exp.back")})
}

// choose swaps the picker for the chosen experiment, so Back from the
// experiment returns straight to the lab.
func (p *ExpPicker) choose(a *App, i int) {
	a.Pop()
	a.Push(p.entries[i].open())
}

func (p *ExpPicker) Update(a *App) error {
	p.layout(a)
	if closePressed() {
		a.Pop()
		return nil
	}
	switch i := p.menu.Update(&a.Pointer); {
	case i < 0:
	case i < len(p.entries):
		p.choose(a, i)
	default:
		a.Pop()
	}
	return nil
}

// progress counts discoveries found and total for an experiment.
func (a *App) progress(prefix string) (found, total int) {
	for _, d := range a.Disc.List {
		if strings.HasPrefix(d.ID, prefix) {
			total++
			if a.State.Notebook.Has(d.ID) {
				found++
			}
		}
	}
	return
}

func (p *ExpPicker) Draw(a *App, screen *ebiten.Image) {
	p.layout(a)
	ui.Shade(screen, a.W, a.H)
	r := p.panel(a)
	ui.Panel(screen, r)
	ui.TextCentered(screen, a.T("picker.title"), ui.Bold(ui.SizeHeading), r.X+r.W/2, r.Y+28, ui.ColText)
	for i, e := range p.entries {
		b := p.menu.Buttons[i].R
		fill := ui.ColPanel
		if i == p.menu.Focus && !a.Pointer.TouchDevice() {
			fill = ui.ColFocus
		}
		ui.RoundRect(screen, b, 12, fill, 3, ui.ColPanelLine)
		ui.Text(screen, a.T(e.titleKey), ui.Bold(ui.SizeBody+2), b.X+16, b.Y+12, ui.ColText, text.AlignStart)
		ui.Text(screen, a.T(pickerDescKey(e.id)), ui.Regular(ui.SizeSmall), b.X+16, b.Y+44, ui.ColText, text.AlignStart)
		found, total := a.progress(e.disc)
		ui.Diamond(screen, b.X+b.W-110, b.Y+22, 7, ui.ColCrystal)
		ui.Text(screen, a.Tf("picker.found", found, total), ui.Bold(ui.SizeSmall), b.X+b.W-98, b.Y+13, ui.ColAccent, text.AlignStart)
	}
	back := p.menu.Buttons[len(p.entries)]
	back.Draw(screen, p.menu.Focus == len(p.entries) && !a.Pointer.TouchDevice())
}
