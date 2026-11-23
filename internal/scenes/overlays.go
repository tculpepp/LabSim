package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"labsim/internal/geom"
	"labsim/internal/inventory"
	"labsim/internal/textwrap"
	"labsim/internal/ui"
	"labsim/internal/world"
)

func closePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}

func advancePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyE)
}

// ---- Dialog ----

const dialogLines = 3

// Dialog shows a character speaking, page by page.
type Dialog struct {
	who, name string
	text      string
	pages     [][]string
	page      int
	wrappedW  float64
}

func NewDialog(a *App, who, name, text string) *Dialog {
	return &Dialog{who: who, name: name, text: text}
}

func (d *Dialog) Opaque() bool { return false }

func (d *Dialog) box(a *App) geom.Rect {
	return geom.Rect{X: 16, Y: float64(a.H) - 150, W: float64(a.W) - 32, H: 134}
}

func (d *Dialog) textWidth(a *App) float64 { return d.box(a).W - 150 }

func (d *Dialog) wrap(a *App) {
	if w := d.textWidth(a); w != d.wrappedW {
		d.wrappedW = w
		d.pages = textwrap.Pages(ui.Wrap(d.text, ui.Regular(ui.SizeBody), w), dialogLines)
		d.page = min(d.page, len(d.pages)-1)
	}
}

func (d *Dialog) Update(a *App) error {
	d.wrap(a)
	if advancePressed() || a.Pointer.AnyTap() || closePressed() {
		d.page++
		if d.page >= len(d.pages) || closePressed() {
			a.Pop()
		}
	}
	return nil
}

func (d *Dialog) Draw(a *App, screen *ebiten.Image) {
	d.wrap(a)
	b := d.box(a)
	ui.Panel(screen, b)
	px, py := b.X+64, b.Y+b.H/2+6
	ui.Portrait(screen, d.who, px, py, 40)
	switch d.who {
	case "nova":
		ui.DrawNova(screen, a.Avatars, px, py+50, 2)
	case "tom":
		ui.DrawTom(screen, a.Avatars, px, py+58, 1.8, 0)
	}
	tx := b.X + 128
	ui.Text(screen, d.name, ui.Bold(ui.SizeBody), tx, b.Y+12, ui.ColAccent, text.AlignStart)
	ui.Lines(screen, d.pages[d.page], ui.Regular(ui.SizeBody), ui.SizeBody, tx, b.Y+38, ui.ColText, text.AlignStart)
	// "More" arrow.
	ax, ay := float32(b.X+b.W-26), float32(b.Y+b.H-22)
	var p vector.Path
	p.MoveTo(ax-8, ay-5)
	p.LineTo(ax+8, ay-5)
	p.LineTo(ax, ay+5)
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(ui.ColAccent)
	vector.FillPath(screen, &p, nil, op)
}

// ---- Inventory ----

// Inventory shows the backpack contents (I-5).
type Inventory struct{ bag *inventory.Backpack }

func NewInventory(bag *inventory.Backpack) *Inventory { return &Inventory{bag: bag} }

func (v *Inventory) Opaque() bool { return false }

func (v *Inventory) Update(a *App) error {
	if closePressed() || inpututil.IsKeyJustPressed(ebiten.KeyI) || inpututil.IsKeyJustPressed(ebiten.KeyB) || a.Pointer.AnyTap() {
		a.Pop()
	}
	return nil
}

func (v *Inventory) Draw(a *App, screen *ebiten.Image) {
	ui.Shade(screen, a.W, a.H)
	cx, _ := a.center()
	b := geom.Centered(cx, float64(a.H)/2, min(float64(a.W)-40, 520), 360)
	ui.Panel(screen, b)
	ui.TextCentered(screen, a.T("inventory.title"), ui.Bold(ui.SizeHeading), cx, b.Y+30, ui.ColText)
	ui.TextCentered(screen, a.Tf("inventory.slots", v.bag.UsedSlots(), v.bag.Slots()), ui.Regular(ui.SizeSmall), cx, b.Y+58, ui.ColText)
	items := v.bag.Items()
	if len(items) == 0 {
		lines := ui.Wrap(a.T("inventory.empty"), ui.Regular(ui.SizeBody), b.W-60)
		ui.Lines(screen, lines, ui.Regular(ui.SizeBody), ui.SizeBody, cx, b.Y+150, ui.ColText, text.AlignCenter)
	} else {
		cols := 2
		cellW := (b.W - 60) / float64(cols)
		for i, it := range items {
			x := b.X + 30 + float64(i%cols)*cellW
			y := b.Y + 84 + float64(i/cols)*44
			a.DrawItem(screen, it.ID, x+16, y+12, 26)
			ui.Text(screen, a.ItemName(it.ID), ui.Bold(ui.SizeBody), x+38, y+1, ui.ColText, text.AlignStart)
			ui.Text(screen, fmt.Sprintf("x%d", it.Qty), ui.Bold(ui.SizeBody), x+cellW-14, y+1, ui.ColAccent, text.AlignEnd)
		}
	}
	ui.TextCentered(screen, a.T("overlay.close"), ui.Regular(ui.SizeSmall), cx, b.Y+b.H-22, ui.ColAccent)
}

// ---- Map ----

// MapView is the full lab map overlay (M-5).
type MapView struct{ play *Play }

func (mv *MapView) Opaque() bool { return false }

func (mv *MapView) Update(a *App) error {
	if closePressed() || inpututil.IsKeyJustPressed(ebiten.KeyM) || a.Pointer.AnyTap() {
		a.Pop()
	}
	return nil
}

func (mv *MapView) Draw(a *App, screen *ebiten.Image) {
	m, p := a.Map, mv.play
	ui.Shade(screen, a.W, a.H)
	scale := min(float64(a.W)*0.9/m.Width, float64(a.H)*0.78/m.Height)
	ox := (float64(a.W) - m.Width*scale) / 2
	oy := (float64(a.H)-m.Height*scale)/2 + 14
	cx, _ := a.center()
	ui.TextCentered(screen, a.T("map.title"), ui.Bold(ui.SizeHeading), cx, oy-22, ui.ColTextLight)
	ui.RoundRect(screen, geom.Rect{X: ox, Y: oy, W: m.Width * scale, H: m.Height * scale}, 8, ui.ColHallway, 3, ui.ColTextLight)
	for _, r := range m.Rooms {
		rr := geom.Rect{X: ox + r.Bounds.X*scale, Y: oy + r.Bounds.Y*scale, W: r.Bounds.W * scale, H: r.Bounds.H * scale}
		c, _ := world.ParseHexColor(r.Color)
		ui.RoundRect(screen, rr, 6, ui.RGB(c), 2, ui.ColWall)
		rx, _ := rr.Center()
		lines := ui.Wrap(r.Name, ui.Bold(ui.SizeSmall), rr.W-8)
		ui.Lines(screen, lines, ui.Bold(ui.SizeSmall), ui.SizeSmall, rx, rr.Y+8, ui.ColText, text.AlignCenter)
		if r.Locked && !p.isUnlocked(r.ID) {
			ui.Padlock(screen, rr.X+rr.W-14, rr.Y+16, 14, ui.ColLocked)
		}
	}
	px, py := ox+p.s.X*scale, oy+p.s.Y*scale
	ui.Circle(screen, px, py, 7, ui.ColAccent)
	ui.Ring(screen, px, py, 7, 2, ui.ColTextLight)
	ui.Text(screen, a.T("map.you"), ui.Bold(ui.SizeSmall), px+10, py-8, ui.ColTextLight, text.AlignStart)
}

// ---- Pause ----

// Pause is the pause menu.
type Pause struct {
	menu    ui.Menu
	actions []func(a *App)
}

func NewPause(a *App) *Pause {
	p := &Pause{}
	p.add(a.T("pause.resume"), func(a *App) { a.Pop() })
	p.add(a.T("pause.settings"), func(a *App) { a.Push(&Settings{}) })
	p.add(a.T("pause.startover"), func(a *App) {
		a.Confirm("confirm.startOver", func(a *App) { a.StartOver() })
	})
	return p
}

func (p *Pause) add(label string, fn func(a *App)) {
	p.menu.Buttons = append(p.menu.Buttons, ui.Button{Label: label})
	p.actions = append(p.actions, fn)
}

func (p *Pause) Opaque() bool { return false }

func (p *Pause) panel(a *App) geom.Rect {
	cx, cy := a.center()
	return geom.Centered(cx, cy, 340, 110+float64(len(p.menu.Buttons))*62)
}

func (p *Pause) layout(a *App) {
	r := p.panel(a)
	cx, _ := r.Center()
	for i := range p.menu.Buttons {
		p.menu.Buttons[i].R = geom.Centered(cx, r.Y+105+float64(i)*62, 260, 50)
	}
}

func (p *Pause) Update(a *App) error {
	p.layout(a)
	if closePressed() {
		a.Pop()
		return nil
	}
	if i := p.menu.Update(&a.Pointer); i >= 0 {
		p.actions[i](a)
	}
	return nil
}

func (p *Pause) Draw(a *App, screen *ebiten.Image) {
	p.layout(a)
	ui.Shade(screen, a.W, a.H)
	r := p.panel(a)
	ui.Panel(screen, r)
	cx, _ := r.Center()
	ui.TextCentered(screen, a.T("pause.title"), ui.Bold(ui.SizeHeading), cx, r.Y+40, ui.ColText)
	p.menu.Draw(screen)
}

// ---- Confirm ----

// Confirm asks a yes/no question. "No" is focused by default so a stray
// Enter press never deletes anything.
type Confirm struct {
	question string
	onYes    func(a *App)
	menu     ui.Menu
}

func NewConfirm(a *App, questionKey string, onYes func(a *App)) *Confirm {
	return NewConfirmText(a, a.T(questionKey), onYes)
}

// NewConfirmText asks an already-formatted question.
func NewConfirmText(a *App, question string, onYes func(a *App)) *Confirm {
	c := &Confirm{question: question, onYes: onYes}
	c.menu.Buttons = []ui.Button{{Label: a.T("confirm.no")}, {Label: a.T("confirm.yes")}}
	return c
}

func (c *Confirm) Opaque() bool { return false }

func (c *Confirm) panel(a *App) geom.Rect {
	cx, cy := a.center()
	return geom.Centered(cx, cy, min(float64(a.W)-40, 520), 200)
}

func (c *Confirm) layout(a *App) {
	r := c.panel(a)
	cx, _ := r.Center()
	c.menu.Buttons[0].R = geom.Centered(cx-120, r.Y+r.H-50, 210, 50)
	c.menu.Buttons[1].R = geom.Centered(cx+120, r.Y+r.H-50, 210, 50)
}

func (c *Confirm) Update(a *App) error {
	c.layout(a)
	if closePressed() {
		a.Pop()
		return nil
	}
	switch c.menu.Update(&a.Pointer) {
	case 0:
		a.Pop()
	case 1:
		a.Pop()
		c.onYes(a)
	}
	return nil
}

func (c *Confirm) Draw(a *App, screen *ebiten.Image) {
	c.layout(a)
	ui.Shade(screen, a.W, a.H)
	r := c.panel(a)
	ui.Panel(screen, r)
	cx, _ := r.Center()
	lines := ui.Wrap(c.question, ui.Bold(ui.SizeBody), r.W-60)
	ui.Lines(screen, lines, ui.Bold(ui.SizeBody), ui.SizeBody, cx, r.Y+36, ui.ColText, text.AlignCenter)
	c.menu.Draw(screen)
}
