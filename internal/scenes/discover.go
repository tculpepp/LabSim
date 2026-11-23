package scenes

import (
	"encoding/json"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/data"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

type science struct {
	ID, Name, Color string
}

type discovery struct {
	ID, Science, Name string
}

// Discoveries is the notebook catalog (data/discoveries.json).
type Discoveries struct {
	Bonus    int         `json:"bonus"`
	Sciences []science   `json:"sciences"`
	List     []discovery `json:"discoveries"`
}

func (d *Discoveries) Get(id string) (discovery, bool) {
	for _, x := range d.List {
		if x.ID == id {
			return x, true
		}
	}
	return discovery{}, false
}

func (a *App) loadDiscoveries() error {
	return json.Unmarshal(data.Discoveries, &a.Disc)
}

// Discover records a discovery. The first time, it pays the bonus (P-2, E-6)
// and returns true.
func (a *App) Discover(id string) bool {
	if !a.State.Notebook.Add(id) {
		return false
	}
	a.State.Wallet.Earn(a.Disc.Bonus)
	a.Sound("discover")
	a.SaveNow()
	return true
}

// factKey is the text key for a discovery's fun fact (checked in data tests).
func factKey(id string) string { return "fact." + id }

// ---- Fun Fact card (E-5) ----

type FactCard struct {
	id    string
	isNew bool
	btn   ui.Menu
}

func NewFactCard(id string, isNew bool) *FactCard { return &FactCard{id: id, isNew: isNew} }

func (f *FactCard) Opaque() bool { return false }

func (f *FactCard) panel(a *App) geom.Rect {
	cx, cy := a.center()
	return geom.Centered(cx, cy, min(float64(a.W)-40, 540), 320)
}

func (f *FactCard) Update(a *App) error {
	r := f.panel(a)
	f.btn.Buttons = []ui.Button{{R: geom.Centered(r.X+r.W/2, r.Y+r.H-36, 180, 46), Label: a.T("exp.cool")}}
	if f.btn.Update(&a.Pointer) == 0 || closePressed() || inpututil.IsKeyJustPressed(ebiten.KeyE) {
		a.Pop()
	}
	return nil
}

func (f *FactCard) Draw(a *App, screen *ebiten.Image) {
	ui.Shade(screen, a.W, a.H)
	r := f.panel(a)
	ui.RoundRect(screen, r, 18, ui.ColPanel, 4, ui.ColFocus)
	cx := r.X + r.W/2
	d, _ := a.Disc.Get(f.id)
	title := a.T("fact.title")
	if f.isNew {
		title = a.Tf("exp.discovery", a.Disc.Bonus)
	}
	ui.Diamond(screen, cx-ui.Measure(title, ui.Bold(ui.SizeBody))/2-18, r.Y+30, 9, ui.ColCrystal)
	ui.TextCentered(screen, title, ui.Bold(ui.SizeBody), cx, r.Y+30, ui.ColAccent)
	ui.TextCentered(screen, d.Name, ui.Bold(ui.SizeHeading), cx, r.Y+66, ui.ColText)
	lines := ui.Wrap(a.T(factKey(f.id)), ui.Regular(ui.SizeBody), r.W-60)
	ui.Lines(screen, lines, ui.Regular(ui.SizeBody), ui.SizeBody, cx, r.Y+96, ui.ColText, text.AlignCenter)
	f.btn.Draw(screen)
}

// ---- Science Notebook ----

type NotebookView struct {
	tab  int
	menu ui.Menu
	ids  []string
}

func (n *NotebookView) Opaque() bool { return true }

func (n *NotebookView) layout(a *App) {
	n.menu.Buttons, n.ids = n.menu.Buttons[:0], n.ids[:0]
	W := float64(a.W)
	tabW := (W - 40) / float64(len(a.Disc.Sciences))
	for i, s := range a.Disc.Sciences {
		found, total := a.scienceCount(s.ID)
		n.menu.Buttons = append(n.menu.Buttons, ui.Button{
			R: geom.Rect{X: 20 + float64(i)*tabW + 3, Y: 60, W: tabW - 6, H: 40}, Label: fmt.Sprintf("%s %d/%d", s.Name, found, total)})
		n.ids = append(n.ids, "")
	}
	n.menu.Buttons = append(n.menu.Buttons, ui.Button{R: geom.Rect{X: W - 120, Y: 10, W: 104, H: 40}, Label: a.T("exp.back")})
	n.ids = append(n.ids, "")
	sci := a.Disc.Sciences[n.tab].ID
	_, total := a.scienceCount(sci)
	cols := 2
	if total > 12 {
		cols = 3 // keep long lists on one screen (480 tall)
	}
	col := 0
	row := 0
	colW := (W - 40) / float64(cols)
	for _, d := range a.Disc.List {
		if d.Science != sci {
			continue
		}
		label := a.T("notebook.unknown")
		if a.State.Notebook.Has(d.ID) {
			label = d.Name
		}
		n.menu.Buttons = append(n.menu.Buttons, ui.Button{R: geom.Rect{X: 20 + float64(col)*colW + 3, Y: 116 + float64(row)*56, W: colW - 6, H: 48}, Label: label})
		n.ids = append(n.ids, d.ID)
		if col++; col == cols {
			col, row = 0, row+1
		}
	}
}

func (a *App) scienceCount(sci string) (found, total int) {
	for _, d := range a.Disc.List {
		if d.Science == sci {
			total++
			if a.State.Notebook.Has(d.ID) {
				found++
			}
		}
	}
	return
}

func (n *NotebookView) Update(a *App) error {
	n.layout(a)
	if closePressed() || inpututil.IsKeyJustPressed(ebiten.KeyN) {
		a.Pop()
		return nil
	}
	i := n.menu.Update(&a.Pointer)
	switch {
	case i < 0:
	case i < len(a.Disc.Sciences):
		n.tab = i
	case i == len(a.Disc.Sciences):
		a.Pop()
	default:
		if id := n.ids[i]; a.State.Notebook.Has(id) {
			a.Push(NewFactCard(id, false))
		}
	}
	return nil
}

func (n *NotebookView) Draw(a *App, screen *ebiten.Image) {
	n.layout(a)
	screen.Fill(ui.Hex("#f6e7c8"))
	ui.Text(screen, a.T("notebook.title"), ui.Bold(ui.SizeHeading+4), 20, 14, ui.ColText, text.AlignStart)
	n.menu.Draw(screen)
	s := a.Disc.Sciences[n.tab]
	ui.RoundRect(screen, n.menu.Buttons[n.tab].R.Expand(3), 14, nil, 4, ui.Hex(s.Color))
	ui.TextCentered(screen, a.T("notebook.tap"), ui.Regular(ui.SizeSmall), float64(a.W)/2, float64(a.H)-20, ui.ColText)
}

// Text keys built from experiment IDs (checked by data/discoveries_test.go).
func matKey(id string) string        { return "mat." + id }
func slideKey(id string) string      { return "micro.slide." + id }
func optKey(id string) string        { return "micro.opt." + id }
func planetKey(id string) string     { return "planet." + id }
func powderKey(id string) string     { return "powder." + id }
func flowerKey(id string) string     { return "flower." + id }
func rideKey(id string) string       { return "ride." + id }
func pickerDescKey(id string) string { return "picker.desc." + id }
func robotSlotKey(id string) string  { return "robot.slot." + id }
func robotPartKey(id string) string  { return "robot.part." + id }
func robotChKey(id string) string    { return "robot.ch." + id }
func robotGoalKey(id string) string  { return "robot.goal." + id }
func triviaQKey(id string) string    { return "trivia." + id + ".q" }
