package scenes

import (
	"errors"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/decor"
	"labsim/internal/geom"
	"labsim/internal/shop"
	"labsim/internal/ui"
)

// Decorate is the My Lab decorating table (P-6): pick a decoration, then
// tap a square to put it there. Tap a placed one to pick it back up.
type Decorate struct {
	layout *decor.Layout
	sel    string
	t      float64
	msg    string
	menu   ui.Menu
	ids    []string // decoration per list button (after Done/Put away)
}

func NewDecorate(a *App) *Decorate { return &Decorate{layout: a.LabLayout(), msg: "decorate.help"} }

func (d *Decorate) Opaque() bool { return true }

func (d *Decorate) owned(a *App) []shop.Entry {
	var out []shop.Entry
	for _, e := range a.ShopCat.OfKind(shop.KindDecor) {
		if a.State.Owns(e.Key()) {
			out = append(out, e)
		}
	}
	return out
}

// view maps the grid onto the screen.
func (d *Decorate) view(a *App) (ox, oy, cell float64) {
	area := geom.Rect{X: 20, Y: 76, W: float64(a.W) - 280, H: float64(a.H) - 92}
	cell = min(area.W/float64(d.layout.W), area.H/float64(d.layout.H))
	ox = area.X + (area.W-cell*float64(d.layout.W))/2
	oy = area.Y + (area.H-cell*float64(d.layout.H))/2
	return
}

func (d *Decorate) cellRect(a *App, c decor.Cell) geom.Rect {
	ox, oy, cell := d.view(a)
	return geom.Rect{X: ox + float64(c.X)*cell, Y: oy + float64(c.Y)*cell, W: cell, H: cell}
}

func (d *Decorate) layoutMenu(a *App) {
	d.menu.Buttons, d.ids = d.menu.Buttons[:0], d.ids[:0]
	W := float64(a.W)
	d.menu.Buttons = append(d.menu.Buttons,
		ui.Button{R: geom.Rect{X: W - 130, Y: 10, W: 110, H: 40}, Label: a.T("decorate.done")},
		ui.Button{R: geom.Rect{X: W - 250, Y: float64(a.H) - 56, W: 230, H: 42}, Label: a.T("decorate.putAway")})
	for i, e := range d.owned(a) {
		d.menu.Buttons = append(d.menu.Buttons, ui.Button{R: geom.Rect{X: W - 250, Y: 60 + float64(i)*39, W: 230, H: 34}, Label: e.Name})
		d.ids = append(d.ids, e.ID)
		if i == 8 { // room for nine; the catalog has nine
			break
		}
	}
}

func (d *Decorate) close(a *App) {
	a.State.Decor = d.layout.Placements()
	a.SaveNow()
	a.Pop()
}

func (d *Decorate) Update(a *App) error {
	d.t += 1.0 / 60
	d.layoutMenu(a)
	if closePressed() {
		d.close(a)
		return nil
	}
	switch i := d.menu.Update(&a.Pointer); {
	case i == 0:
		d.close(a)
		return nil
	case i == 1:
		if d.sel != "" {
			d.layout.Remove(d.sel)
			d.sel, d.msg = "", "decorate.help"
		}
	case i >= 2:
		d.sel, d.msg = d.ids[i-2], "decorate.where"
	}
	for y := 0; y < d.layout.H; y++ {
		for x := 0; x < d.layout.W; x++ {
			c := decor.Cell{X: x, Y: y}
			if !a.Pointer.TapIn(d.cellRect(a, c)) {
				continue
			}
			if id, ok := d.layout.At(c); ok && id != d.sel {
				d.sel, d.msg = id, "decorate.where"
				continue
			}
			if d.sel == "" {
				continue
			}
			switch err := d.layout.Place(d.sel, c); {
			case errors.Is(err, decor.ErrBlocked):
				d.msg = "decorate.blocked"
			case err == nil:
				a.Sound("place")
				d.sel, d.msg = "", "decorate.help"
			}
		}
	}
	return nil
}

func (d *Decorate) Draw(a *App, screen *ebiten.Image) {
	d.layoutMenu(a)
	screen.Fill(ui.Hex("#ffd6b0"))
	ui.Text(screen, a.T("decorate.title"), ui.Bold(ui.SizeHeading+4), 20, 14, ui.ColText, text.AlignStart)
	_, _, cell := d.view(a)
	for y := 0; y < d.layout.H; y++ {
		for x := 0; x < d.layout.W; x++ {
			c := decor.Cell{X: x, Y: y}
			r := d.cellRect(a, c)
			fill := ui.Hex("#fff0e0")
			if d.layout.Blocked(c) {
				fill = ui.Hex("#d8c0a8")
			}
			ui.RoundRect(screen, r.Expand(-2), 6, fill, 1, ui.Hex("#e0b890"))
		}
	}
	// Draw placed decor (flat first), highlighting the selection.
	for pass := 0; pass < 2; pass++ {
		for _, p := range d.layout.Placements() {
			if flatDecor(p.ID) != (pass == 0) {
				continue
			}
			r := d.cellRect(a, p.Cell)
			if p.ID == d.sel {
				ui.RoundRect(screen, r.Expand(-1), 6, nil, 3, ui.ColFocus)
			}
			drawDecor(screen, p.ID, r.X+r.W/2, r.Y+r.H-4, cell/labCell, d.t)
		}
	}
	d.menu.Draw(screen)
	for i, id := range d.ids {
		r := d.menu.Buttons[2+i].R
		if id == d.sel {
			ui.RoundRect(screen, r.Expand(3), 12, nil, 3, ui.ColCrystal)
		}
		if _, placed := d.layout.Where(id); placed {
			ui.Diamond(screen, r.X+r.W-14, r.Y+r.H/2, 5, ui.Hex("#3aa858"))
		}
	}
	if len(d.ids) == 0 {
		lines := ui.Wrap(a.T("decorate.none"), ui.Regular(ui.SizeBody), 220)
		ui.Lines(screen, lines, ui.Regular(ui.SizeBody), ui.SizeBody, float64(a.W)-245, 70, ui.ColText, text.AlignStart)
	}
	ui.Text(screen, a.T(d.msg), ui.Bold(ui.SizeSmall), 22, 50, ui.ColAccent, text.AlignStart)
}
