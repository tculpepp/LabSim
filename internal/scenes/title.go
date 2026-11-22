package scenes

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
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

// Creator is the avatar creator. M2 stub: pick difficulty only; full creator is M4.
type Creator struct {
	level content.Level
	menu  ui.Menu
}

func NewCreator(a *App) *Creator {
	c := &Creator{level: content.Junior}
	c.menu.Buttons = []ui.Button{{Label: a.T("creator.junior")}, {Label: a.T("creator.expert")}, {Label: a.T("creator.start")}}
	c.menu.Focus = 2
	return c
}

func (c *Creator) Opaque() bool { return true }

func (c *Creator) layout(a *App) {
	cx, _ := a.center()
	c.menu.Buttons[0].R = geom.Centered(cx-130, 250, 220, 56)
	c.menu.Buttons[1].R = geom.Centered(cx+130, 250, 220, 56)
	c.menu.Buttons[2].R = geom.Centered(cx, 400, 220, 56)
}

func (c *Creator) Update(a *App) error {
	c.layout(a)
	switch c.menu.Update(&a.Pointer) {
	case 0:
		c.level = content.Junior
	case 1:
		c.level = content.Expert
	case 2:
		a.StartNewGame(a.T("creator.defaultName"), c.level)
	}
	return nil
}

func (c *Creator) Draw(a *App, screen *ebiten.Image) {
	c.layout(a)
	screen.Fill(colTitleBG)
	cx, _ := a.center()
	ui.TextCentered(screen, a.T("creator.title"), ui.Bold(ui.SizeHeading+8), cx, 50, ui.ColTextLight)
	ui.TextCentered(screen, a.T("creator.soon"), ui.Regular(ui.SizeBody), cx, 92, ui.ColTextLight)
	ui.TextCentered(screen, a.T("creator.level"), ui.Bold(ui.SizeHeading), cx, 180, ui.ColTextLight)
	c.menu.Draw(screen)
	// Mark the chosen level with a ring, separate from keyboard focus.
	sel := c.menu.Buttons[0].R
	if c.level == content.Expert {
		sel = c.menu.Buttons[1].R
	}
	ui.RoundRect(screen, sel.Expand(6), 16, color.Transparent, 4, ui.ColCrystal)
	descs := []string{a.T("creator.junior.desc"), a.T("creator.expert.desc")}
	for i, d := range descs {
		r := c.menu.Buttons[i].R
		lines := ui.Wrap(d, ui.Regular(ui.SizeSmall), r.W)
		ui.Lines(screen, lines, ui.Regular(ui.SizeSmall), ui.SizeSmall, r.X+r.W/2, r.Y+r.H+14, ui.ColTextLight, text.AlignCenter)
	}
}
