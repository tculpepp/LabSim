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
		{R: p.mapButton(a), Label: a.T("hud.map")},
		{R: p.bagButton(a), Label: a.T("hud.bagButton")},
		{R: p.pauseButton(a), Label: "II"},
	} {
		b.Draw(screen, false)
	}

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

	if a.Pointer.TouchDevice() {
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
