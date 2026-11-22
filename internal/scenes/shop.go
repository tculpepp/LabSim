package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

const shopRows = 5

// Shop is Tom's shop. Tom loves to help people become better scientists (C-2).
type Shop struct {
	tab    int // 0 = buy, 1 = sell
	cat    int // buy category
	page   int
	tip    string
	tipKey string
	menu   ui.Menu
	acts   []func(a *App)
}

func NewShop(a *App) *Shop {
	sh := &Shop{tipKey: "tom.hello"}
	return sh
}

func (sh *Shop) Opaque() bool { return true }

func (sh *Shop) add(r geom.Rect, label string, fn func(a *App)) {
	sh.menu.Buttons = append(sh.menu.Buttons, ui.Button{R: r, Label: label})
	sh.acts = append(sh.acts, fn)
}

// shopRow is one line in the list.
type shopRow struct {
	icon   func(screen *ebiten.Image, x, y float64)
	name   string
	detail string
	price  int
	button string
	ok     bool // button enabled
	act    func(a *App)
}

func (sh *Shop) rows(a *App) []shopRow {
	if sh.tab == 1 {
		return sh.sellRows(a)
	}
	return sh.buyRows(a)
}

func (sh *Shop) sellRows(a *App) []shopRow {
	var out []shopRow
	for _, st := range a.State.Backpack.Items() {
		id := st.ID
		it, _ := a.Items.Get(id)
		out = append(out, shopRow{
			icon:   func(screen *ebiten.Image, x, y float64) { a.DrawItem(screen, id, x, y, 26) },
			name:   it.Name,
			detail: fmt.Sprintf("x%d", st.Qty),
			price:  it.Sell,
			button: a.T("shop.sell"),
			ok:     it.Sell > 0,
			act: func(a *App) {
				if a.State.Backpack.Remove(id, 1) == nil {
					a.State.Wallet.Earn(it.Sell)
					a.Sound("coin")
					sh.tipKey = "tom.sold"
					a.SaveNow()
				}
			},
		})
	}
	return out
}

func (sh *Shop) layout(a *App) []shopRow {
	sh.menu.Buttons, sh.acts = sh.menu.Buttons[:0], sh.acts[:0]
	W := float64(a.W)
	sh.add(geom.Rect{X: W - 130, Y: 10, W: 110, H: 40}, a.T("shop.leave"), func(a *App) { a.Pop() })
	sh.add(geom.Rect{X: 250, Y: 62, W: 130, H: 38}, a.T("shop.buy"), func(*App) { sh.tab, sh.page = 0, 0 })
	sh.add(geom.Rect{X: 390, Y: 62, W: 130, H: 38}, a.T("shop.sellTab"), func(*App) { sh.tab, sh.page = 1, 0 })
	sh.addCategoryTabs(a)
	rows := sh.rows(a)
	pages := max(1, (len(rows)+shopRows-1)/shopRows)
	sh.page = min(sh.page, pages-1)
	listTop := sh.listTop()
	for i := sh.page * shopRows; i < min(len(rows), (sh.page+1)*shopRows); i++ {
		r := rows[i]
		y := listTop + float64(i-sh.page*shopRows)*50
		if r.ok {
			sh.add(geom.Rect{X: W - 130, Y: y + 6, W: 110, H: 38}, r.button, r.act)
		}
	}
	if pages > 1 {
		sh.add(geom.Rect{X: 250, Y: float64(a.H) - 50, W: 60, H: 38}, "<", func(*App) { sh.page = (sh.page + pages - 1) % pages })
		sh.add(geom.Rect{X: W - 80, Y: float64(a.H) - 50, W: 60, H: 38}, ">", func(*App) { sh.page = (sh.page + 1) % pages })
	}
	return rows
}

func (sh *Shop) Update(a *App) error {
	sh.layout(a)
	if closePressed() {
		a.Pop()
		return nil
	}
	if i := sh.menu.Update(&a.Pointer); i >= 0 {
		sh.acts[i](a)
	}
	return nil
}

func (sh *Shop) Draw(a *App, screen *ebiten.Image) {
	rows := sh.layout(a)
	W := float64(a.W)
	screen.Fill(ui.Hex("#f9d2e0"))
	ui.Text(screen, a.T("shop.title"), ui.Bold(ui.SizeHeading+4), 20, 16, ui.ColText, text.AlignStart)
	// Wallet.
	ui.RoundRect(screen, geom.Rect{X: W - 270, Y: 12, W: 126, H: 36}, 18, ui.ColPanel, 2, ui.ColPanelLine)
	ui.Diamond(screen, W-250, 30, 9, ui.ColCrystal)
	ui.Text(screen, fmt.Sprint(a.State.Wallet.Crystals()), ui.Bold(ui.SizeBody), W-234, 19, ui.ColText, text.AlignStart)

	// Tom and his tip.
	tom := geom.Rect{X: 16, Y: 62, W: 220, H: float64(a.H) - 78}
	ui.Panel(screen, tom)
	ui.DrawTom(screen, a.Avatars, tom.X+tom.W/2, tom.Y+150, 3, float64(ebiten.Tick())/60)
	ui.Text(screen, a.T("tom.name"), ui.Bold(ui.SizeBody), tom.X+14, tom.Y+166, ui.ColAccent, text.AlignStart)
	tip := ui.Wrap(a.T(sh.tipKey), ui.Regular(ui.SizeSmall+1), tom.W-28)
	ui.Lines(screen, tip, ui.Regular(ui.SizeSmall+1), ui.SizeSmall+1, tom.X+14, tom.Y+192, ui.ColText, text.AlignStart)

	listTop := sh.listTop()
	if len(rows) == 0 {
		msg := ui.Wrap(a.T(map[int]string{0: "shop.nothing", 1: "shop.nothingToSell"}[sh.tab]), ui.Regular(ui.SizeBody), W-290)
		ui.Lines(screen, msg, ui.Regular(ui.SizeBody), ui.SizeBody, 254, listTop+20, ui.ColText, text.AlignStart)
	}
	for i := sh.page * shopRows; i < min(len(rows), (sh.page+1)*shopRows); i++ {
		r := rows[i]
		y := listTop + float64(i-sh.page*shopRows)*50
		ui.RoundRect(screen, geom.Rect{X: 248, Y: y, W: W - 264, H: 46}, 10, ui.ColPanel, 0, nil)
		r.icon(screen, 272, y+23)
		ui.Text(screen, r.name, ui.Bold(ui.SizeBody), 294, y+4, ui.ColText, text.AlignStart)
		ui.Text(screen, r.detail, ui.Regular(ui.SizeSmall), 294, y+26, ui.ColText, text.AlignStart)
		if r.price > 0 {
			ui.Diamond(screen, W-170, y+23, 7, ui.ColCrystal)
			ui.Text(screen, fmt.Sprint(r.price), ui.Bold(ui.SizeBody), W-160, y+12, ui.ColText, text.AlignStart)
		}
		if !r.ok && r.button != "" {
			ui.TextCentered(screen, r.button, ui.Bold(ui.SizeSmall), W-75, y+25, ui.ColText)
		}
	}
	sh.menu.Draw(screen)
	selTab := sh.menu.Buttons[1+sh.tab].R
	ui.RoundRect(screen, selTab.Expand(3), 14, nil, 3, ui.ColAccent)
	sh.drawCategoryMarker(screen)
}
