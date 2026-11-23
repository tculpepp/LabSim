package scenes

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

func (p *Play) drawHUD(a *App, screen *ebiten.Image) {
	small, body := ui.Bold(ui.SizeSmall), ui.Bold(ui.SizeBody)
	// Status pill: day, crystals, backpack.
	ui.RoundRect(screen, geom.Rect{X: 8, Y: 8, W: 300, H: 40}, 20, ui.ColPanel, 2, ui.ColPanelLine)
	if p.s.Clock.IsNight() {
		ui.Circle(screen, 30, 28, 9, color.RGBA{0xf0, 0xe6, 0x8c, 0xff})
		ui.Circle(screen, 34, 25, 8, ui.ColPanel) // crescent moon
	} else {
		ui.Circle(screen, 30, 28, 9, color.RGBA{0xff, 0xc8, 0x3c, 0xff})
	}
	ui.Text(screen, a.Tf("hud.day", p.s.Clock.Day()), body, 46, 17, ui.ColText, text.AlignStart)
	ui.Diamond(screen, 134, 28, 10, ui.ColCrystal)
	ui.Text(screen, fmt.Sprint(p.s.Wallet.Crystals()), body, 148, 17, ui.ColText, text.AlignStart)
	ui.Text(screen, a.Tf("hud.bag", p.s.Backpack.UsedSlots(), p.s.Backpack.Slots()), body, 210, 17, ui.ColText, text.AlignStart)

	for _, b := range []ui.Button{
		{R: p.notesButton(a), Label: a.T("hud.notes")},
		{R: p.mapButton(a), Label: a.T("hud.map")},
		{R: p.bagButton(a), Label: a.T("hud.bagButton")},
		{R: p.pauseButton(a), Label: "II"},
	} {
		b.Draw(screen, false)
	}

	p.drawJobs(a, screen)

	// Current room label, bottom-left, plus a warning if saving is broken.
	ui.Text(screen, p.roomName(a), small, 14, float64(a.H)-28, ui.ColTextLight, text.AlignStart)
	if a.SaveFailing() {
		ui.Text(screen, a.T("save.failed"), small, 14, 56, ui.ColFocus, text.AlignStart)
	}

	if p.bannerT > 0 {
		face := ui.Bold(ui.SizeHeading)
		w := ui.Measure(p.banner, face) + 48
		cx, _ := a.center()
		r := geom.Centered(cx, 84, w, 44)
		ui.Panel(screen, r)
		ui.TextCentered(screen, p.banner, face, cx, 84, ui.ColText)
	}

	if a.ShowTouch() {
		if p.joy.Active() {
			ox, oy := p.joy.Origin()
			kx, ky := p.joy.Knob()
			ui.Ring(screen, ox, oy, p.joy.Radius, 4, ui.ColPanel)
			ui.Circle(screen, kx, ky, 22, ui.ColPanel)
		}
		u := p.useButton(a)
		ux, uy := u.Center()
		ui.Circle(screen, ux, uy, u.W/2, ui.ColPanel)
		ui.Ring(screen, ux, uy, u.W/2, 3, ui.ColPanelLine)
		ui.TextCentered(screen, a.T("hud.use"), body, ux, uy, ui.ColText)
	}
}

// trackerRect is the job list under the status pill (tap to collapse).
func (p *Play) trackerRect(a *App) geom.Rect {
	h := 30.0
	if !p.hideJobs {
		if b := a.Board(); b != nil {
			h += float64(len(b.All())) * 22
		} else {
			h += 22
		}
	}
	return geom.Rect{X: 8, Y: 54, W: 310, H: h}
}

func (p *Play) drawJobs(a *App, screen *ebiten.Image) {
	r := p.trackerRect(a)
	ui.RoundRect(screen, r, 12, color.RGBA{0xff, 0xfb, 0xf2, 0xe6}, 2, ui.ColPanelLine)
	arrow := "-"
	if p.hideJobs {
		arrow = "+"
	}
	ui.Text(screen, arrow+" "+a.T("jobs.title"), ui.Bold(ui.SizeSmall), r.X+12, r.Y+7, ui.ColAccent, text.AlignStart)
	if p.hideJobs {
		return
	}
	small := ui.Regular(ui.SizeSmall)
	b := a.Board()
	if b == nil {
		ui.Text(screen, a.T("jobs.none"), small, r.X+12, r.Y+29, ui.ColText, text.AlignStart)
		return
	}
	for i, j := range b.All() {
		y := r.Y + 29 + float64(i)*22
		box := geom.Rect{X: r.X + 12, Y: y + 2, W: 14, H: 14}
		ui.RoundRect(screen, box, 3, ui.ColPanel, 1.5, ui.ColPanelLine)
		label := j.Title
		if b.IsDone(j.ID) {
			ui.Line(screen, box.X+3, box.Y+7, box.X+6, box.Y+11, 2.5, ui.Hex("#3aa858"))
			ui.Line(screen, box.X+6, box.Y+11, box.X+12, box.Y+2, 2.5, ui.Hex("#3aa858"))
		} else if j.ItemID != "" {
			label = fmt.Sprintf("%s (%d/%d)", j.Title, min(p.s.Backpack.Count(j.ItemID), j.Qty), j.Qty)
		}
		for ui.Measure(label, small) > r.W-44 && len(label) > 4 {
			label = label[:len(label)-4] + "..."
		}
		c := ui.ColText
		if b.IsDone(j.ID) {
			c = color.RGBA{0x88, 0x88, 0x88, 0xff}
		}
		ui.Text(screen, label, small, r.X+32, y, c, text.AlignStart)
	}
}
