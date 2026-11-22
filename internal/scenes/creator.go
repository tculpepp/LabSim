package scenes

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/avatar"
	"labsim/internal/content"
	"labsim/internal/geom"
	"labsim/internal/names"
	"labsim/internal/ui"
)

// Avatar creator tabs (A-1..A-8). The last tab is name + level.
var creatorTabs = []struct {
	key   string
	parts []string
}{
	{"creator.tab.look", []string{"body", "skin", "hair", "hairColor"}},
	{"creator.tab.clothes", []string{"coat", "shirt", "shoes", "pattern"}},
	{"creator.tab.extras", []string{"eyewear", "hat", "gear", "pet"}},
	{"creator.tab.name", nil},
}

// Creator makes a new scientist, or edits the current one (edit = true, A-6).
type Creator struct {
	edit   bool
	av     avatar.Avatar
	level  content.Level
	pick   names.Pick
	custom string // typed name; overrides the preset builder when set
	tab    int

	menu    ui.Menu
	actions []func(a *App)
	labels  []labelAt
	t       float64

	focusSet bool
}

type labelAt struct {
	text   string
	x, y   float64
	center bool
}

func NewCreator(a *App) *Creator {
	seed := uint64(time.Now().UnixNano())
	return &Creator{
		av:    a.Avatars.Normalize(a.Avatars.Default(), nil),
		level: content.Junior,
		pick:  a.Names.Random(seed),
	}
}

// NewEditCreator opens the creator on the current scientist (wardrobe/mirror).
func NewEditCreator(a *App) *Creator {
	s := a.State
	c := &Creator{edit: true, av: a.Avatars.Normalize(s.Avatar, s.Owns), level: s.Level, custom: s.Name}
	return c
}

func (c *Creator) Opaque() bool { return true }

func (c *Creator) owned(a *App) func(string) bool {
	if a.State != nil {
		return a.State.Owns
	}
	return nil
}

func (c *Creator) name(a *App) string {
	if c.custom != "" {
		return c.custom
	}
	return a.Names.Build(c.pick)
}

func (c *Creator) add(r geom.Rect, label string, fn func(a *App)) {
	c.menu.Buttons = append(c.menu.Buttons, ui.Button{R: r, Label: label})
	c.actions = append(c.actions, fn)
}

// layout rebuilds the buttons for the current tab. Order is stable per tab so focus survives.
func (c *Creator) layout(a *App) {
	c.menu.Buttons, c.actions, c.labels = c.menu.Buttons[:0], c.actions[:0], c.labels[:0]
	W := float64(a.W)
	tabW := (W - 40) / float64(len(creatorTabs))
	for i, tb := range creatorTabs {
		i := i
		c.add(geom.Rect{X: 20 + float64(i)*tabW + 3, Y: 58, W: tabW - 6, H: 38}, a.T(tb.key), func(*App) {
			c.tab = i
		})
	}
	left := 270.0
	rowX, rowW := left, W-left-20
	row := func(i int) float64 { return 112 + float64(i)*62 }
	arrows := func(i int, label, value string, prev, next func(*App)) {
		y := row(i)
		c.labels = append(c.labels, labelAt{label, rowX, y - 2, false}, labelAt{value, rowX + rowW/2, y + 32, true})
		c.add(geom.Rect{X: rowX, Y: y + 14, W: 48, H: 36}, "<", prev)
		c.add(geom.Rect{X: rowX + rowW - 48, Y: y + 14, W: 48, H: 36}, ">", next)
	}
	if parts := creatorTabs[c.tab].parts; parts != nil {
		for i, part := range parts {
			part := part
			name := ""
			if o, ok := a.Avatars.Option(part, c.av[part]); ok {
				name = o.Name
			}
			label := part
			for _, p := range a.Avatars.Parts {
				if p.ID == part {
					label = p.Name
				}
			}
			arrows(i, label, name,
				func(a *App) { c.av = a.Avatars.Cycle(c.av, part, -1, c.owned(a)) },
				func(a *App) { c.av = a.Avatars.Cycle(c.av, part, 1, c.owned(a)) })
		}
	} else {
		words := []struct {
			key  string
			list []string
			idx  *int
		}{
			{"creator.word.title", a.Names.Titles, &c.pick.Title},
			{"creator.word.adjective", a.Names.Adjectives, &c.pick.Adjective},
			{"creator.word.noun", a.Names.Nouns, &c.pick.Noun},
		}
		for i, w := range words {
			w := w
			c.pick = a.Names.Normalize(c.pick)
			arrows(i, a.T(w.key), w.list[*w.idx],
				func(*App) { *w.idx--; c.custom = "" },
				func(*App) { *w.idx++; c.custom = "" })
		}
		y := row(3) + 6
		half := (rowW - 8) / 2
		c.add(geom.Rect{X: rowX, Y: y, W: half, H: 40}, a.T("creator.surprise"), func(a *App) {
			c.pick, c.custom = a.Names.Random(uint64(time.Now().UnixNano())), ""
		})
		c.add(geom.Rect{X: rowX + half + 8, Y: y, W: half, H: 40}, a.T("creator.type"), func(a *App) {
			a.Push(NewKeyboard(c.custom, func(_ *App, name string) { c.custom = name }))
		})
		c.labels = append(c.labels, labelAt{a.T("creator.level"), rowX, y + 48, false})
		c.add(geom.Rect{X: rowX, Y: y + 70, W: half, H: 40}, a.T("creator.junior"), func(*App) { c.level = content.Junior })
		c.add(geom.Rect{X: rowX + half + 8, Y: y + 70, W: half, H: 40}, a.T("creator.expert"), func(*App) { c.level = content.Expert })
	}
	label := a.T("creator.start")
	if c.edit {
		label = a.T("creator.done")
	}
	c.add(geom.Rect{X: 20, Y: float64(a.H) - 60, W: 230, H: 48}, label, func(a *App) { c.finish(a) })
	if !c.focusSet {
		c.menu.Focus, c.focusSet = len(c.menu.Buttons)-1, true
	}
}

func (c *Creator) finish(a *App) {
	name := c.name(a)
	if c.edit {
		s := a.State
		s.Name, s.Level, s.Avatar = name, c.level, c.av
		a.SaveNow()
		a.Pop()
		return
	}
	a.StartNewGame(name, c.level, c.av)
}

func (c *Creator) Update(a *App) error {
	c.t += 1.0 / 60
	c.layout(a)
	if c.edit && closePressed() {
		a.Pop()
		return nil
	}
	if i := c.menu.Update(&a.Pointer); i >= 0 {
		c.actions[i](a)
	}
	return nil
}

func (c *Creator) Draw(a *App, screen *ebiten.Image) {
	c.layout(a)
	screen.Fill(colTitleBG)
	cx, _ := a.center()
	title := a.T("creator.title")
	if c.edit {
		title = a.T("creator.editTitle")
	}
	ui.TextCentered(screen, title, ui.Bold(ui.SizeHeading), cx, 28, ui.ColTextLight)

	// Preview panel: big avatar, pet, and name.
	prev := geom.Rect{X: 20, Y: 106, W: 236, H: float64(a.H) - 180}
	ui.Panel(screen, prev)
	px, _ := prev.Center()
	ui.DrawAvatar(screen, a.Avatars, c.av, px-14, prev.Y+prev.H-54, 3.6, 1)
	ui.DrawPet(screen, c.av["pet"], px+70, prev.Y+prev.H-54, 2.4, c.t)
	nameLines := ui.Wrap(c.name(a), ui.Bold(ui.SizeBody), prev.W-20)
	ui.Lines(screen, nameLines, ui.Bold(ui.SizeBody), ui.SizeBody, px, prev.Y+prev.H-40, ui.ColText, text.AlignCenter)

	c.menu.Draw(screen)
	ui.RoundRect(screen, c.menu.Buttons[c.tab].R.Expand(3), 12, nil, 3, ui.ColCrystal)
	for _, l := range c.labels {
		if l.center {
			ui.TextCentered(screen, l.text, ui.Bold(ui.SizeBody), l.x, l.y, ui.ColTextLight)
		} else {
			ui.Text(screen, l.text, ui.Bold(ui.SizeSmall), l.x, l.y, ui.ColTextLight, text.AlignStart)
		}
	}
	if creatorTabs[c.tab].parts == nil {
		// Mark the chosen level.
		n := len(c.menu.Buttons)
		sel := c.menu.Buttons[n-3].R
		if c.level == content.Expert {
			sel = c.menu.Buttons[n-2].R
		}
		ui.RoundRect(screen, sel.Expand(4), 14, nil, 4, ui.ColCrystal)
	}
}
