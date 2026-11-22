package scenes

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
)

var colTitleBG = color.RGBA{0x30, 0x60, 0xd0, 0xff}

// Title is the start screen.
type Title struct {
	menu    ui.Menu
	actions []func(a *App)
}

func NewTitle(a *App) *Title {
	t := &Title{}
	if a.HasSave() {
		t.addButton(a.T("title.continue"), func(a *App) {
			if err := a.ContinueGame(); err != nil {
				log.Printf("continue: %v", err)
				a.saved, a.notice = nil, "title.corrupt"
				a.Replace(NewTitle(a))
			}
		})
		t.addButton(a.T("title.new"), func(a *App) {
			a.Confirm("confirm.newOverwrite", func(a *App) { a.Replace(NewCreator(a)) })
		})
	} else {
		t.addButton(a.T("title.new"), func(a *App) { a.Replace(NewCreator(a)) })
	}
	return t
}

func (t *Title) addButton(label string, fn func(a *App)) {
	t.menu.Buttons = append(t.menu.Buttons, ui.Button{Label: label})
	t.actions = append(t.actions, fn)
}

func (t *Title) Opaque() bool { return true }

func (t *Title) layout(a *App) {
	cx, _ := a.center()
	for i := range t.menu.Buttons {
		t.menu.Buttons[i].R = geom.Centered(cx, 300+float64(i)*62, 260, 50)
	}
}

func (t *Title) Update(a *App) error {
	t.layout(a)
	if i := t.menu.Update(&a.Pointer); i >= 0 {
		t.actions[i](a)
	}
	return nil
}

func (t *Title) Draw(a *App, screen *ebiten.Image) {
	t.layout(a)
	screen.Fill(colTitleBG)
	cx, _ := a.center()
	// Bubbling beaker.
	ui.Circle(screen, cx, 110, 46, ui.ColTextLight)
	ui.Circle(screen, cx, 122, 32, color.RGBA{0x7c, 0xf0, 0x8c, 0xff})
	ui.Circle(screen, cx-12, 70, 7, ui.ColCrystal)
	ui.Circle(screen, cx+14, 56, 5, ui.ColCrystal)
	ui.TextCentered(screen, "LabSim", ui.Bold(ui.SizeTitle), cx, 190, ui.ColTextLight)
	tag := ui.Wrap(a.T("title.tagline"), ui.Regular(ui.SizeBody), float64(a.W)-80)
	ui.Lines(screen, tag, ui.Regular(ui.SizeBody), ui.SizeBody, cx, 222, ui.ColTextLight, text.AlignCenter)
	t.menu.Draw(screen)
	if a.notice != "" {
		lines := ui.Wrap(a.T(a.notice), ui.Regular(ui.SizeSmall), float64(a.W)-80)
		ui.Lines(screen, lines, ui.Regular(ui.SizeSmall), ui.SizeSmall, cx, float64(a.H)-60, ui.ColFocus, text.AlignCenter)
	}
}
