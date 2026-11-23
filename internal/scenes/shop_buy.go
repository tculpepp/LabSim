package scenes

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/geom"
	"labsim/internal/shop"
	"labsim/internal/ui"
)

// Buy side of the shop (P-5).

var shopCats = []struct {
	key  string
	kind shop.Kind
}{
	{"shop.cat.tools", shop.KindTool},
	{"shop.cat.outfits", shop.KindOutfit},
	{"shop.cat.decor", shop.KindDecor},
	{"shop.cat.bags", shop.KindBackpack},
	{"shop.cat.rooms", shop.KindRoom},
}

func (sh *Shop) listTop() float64 {
	if sh.tab == 0 {
		return 156
	}
	return 112
}

func (sh *Shop) catRect(a *App, i int) geom.Rect {
	x0, w := 248.0, float64(a.W)-264
	cw := w / float64(len(shopCats))
	return geom.Rect{X: x0 + float64(i)*cw + 2, Y: 108, W: cw - 4, H: 38}
}

func (sh *Shop) addCategoryTabs(a *App) {
	if sh.tab != 0 {
		return
	}
	for i, c := range shopCats {
		i := i
		sh.add(sh.catRect(a, i), a.T(c.key), func(*App) { sh.cat, sh.page = i, 0 })
	}
}

func (sh *Shop) drawCategoryMarker(a *App, screen *ebiten.Image) {
	if sh.tab == 0 {
		ui.RoundRect(screen, sh.catRect(a, sh.cat).Expand(3), 12, nil, 3, ui.ColCrystal)
	}
}

func (sh *Shop) entries(a *App) []shop.Entry {
	if shopCats[sh.cat].kind == shop.KindOutfit {
		return a.Outfits()
	}
	return a.ShopCat.OfKind(shopCats[sh.cat].kind)
}

func (sh *Shop) buyRows(a *App) []shopRow {
	var out []shopRow
	for _, e := range sh.entries(a) {
		e := e
		row := shopRow{
			icon:   func(screen *ebiten.Image, x, y float64) { drawShopIcon(a, screen, e, x, y) },
			name:   e.Name,
			detail: e.Desc,
			price:  e.Price,
			button: a.T("shop.buyOne"),
			ok:     true,
			act:    func(a *App) { sh.buy(a, e) },
		}
		switch err := shop.Check(e, 1<<30, a.State.Owns); {
		case errors.Is(err, shop.ErrOwned):
			row.button, row.ok = a.T("shop.owned"), false
		case errors.Is(err, shop.ErrNeedsFirst):
			row.button, row.ok = a.T("shop.locked"), false
		}
		out = append(out, row)
	}
	return out
}

func (sh *Shop) buy(a *App, e shop.Entry) {
	switch err := a.Buy(e); {
	case errors.Is(err, shop.ErrTooPoor):
		sh.tipKey = "tom.poor"
		a.Sound("fizzle")
	case err != nil:
		sh.tipKey = "tom.hello"
	default:
		sh.tipKey = boughtKey(e.Kind)
	}
}

// drawShopIcon draws a small picture for a shop entry.
func drawShopIcon(a *App, screen *ebiten.Image, e shop.Entry, x, y float64) {
	switch e.Kind {
	case shop.KindOutfit:
		if e.Part == "pet" {
			ui.DrawPet(screen, e.Option, x, y+10, 0.9, 0)
			return
		}
		c := color.Color(ui.ColAccent)
		if o, ok := a.Avatars.Option(e.Part, e.Option); ok && o.Color != "" {
			c = ui.Hex(o.Color)
		}
		ui.Circle(screen, x, y, 12, c)
		ui.Ring(screen, x, y, 12, 2, ui.ColWall)
	case shop.KindDecor:
		drawDecor(screen, e.ID, x, y+12, 0.55, 0)
	case shop.KindBackpack:
		ui.RoundRect(screen, geom.Centered(x, y+2, 22, 24), 7, ui.Hex("#e8514a"), 2, ui.ColWall)
		ui.RoundRect(screen, geom.Centered(x, y+6, 14, 9), 3, ui.Hex("#ff8a80"), 1, ui.ColWall)
		ui.Arc(screen, x, y-10, 6, 3.14, 6.28, 2.5, ui.ColWall)
	case shop.KindRoom:
		ui.Padlock(screen, x, y+6, 26, ui.ColLocked)
	case shop.KindTool:
		switch e.ID {
		case "spark_kit":
			ui.Poly(screen, ui.ColFocus, x+3, y-13, x-8, y+2, x-1, y+2, x-4, y+13, x+8, y-3, x+1, y-3)
		case "super_microscope":
			ui.RoundRect(screen, geom.Centered(x+2, y-4, 8, 20), 3, ui.Hex("#566078"), 0, nil)
			ui.RoundRect(screen, geom.Centered(x, y+11, 22, 5), 2, ui.Hex("#2e3648"), 0, nil)
			ui.Arc(screen, x-2, y+2, 8, 1.6, 4.2, 3, ui.Hex("#2e3648"))
		case "telescope":
			ui.Line(screen, x-10, y+6, x+10, y-8, 7, ui.Hex("#3a6fd8"))
			ui.Line(screen, x-2, y+2, x-8, y+13, 2, ui.ColWall)
			ui.Line(screen, x-2, y+2, x+4, y+13, 2, ui.ColWall)
		}
	}
}

// boughtKey is Tom's thank-you line for a kind of purchase (checked in tests).
func boughtKey(k shop.Kind) string { return "tom.bought." + string(k) }
