package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/geom"
	"labsim/internal/state"
	"labsim/internal/ui"
)

var dayChoices = []int{5, 10, 15, 20}
var touchChoices = []string{state.TouchAuto, state.TouchOn, state.TouchOff}

// Settings is the grown-up options screen (ST-1, AU-2).
type Settings struct {
	menu ui.Menu
	acts []func(a *App)
	rows []settingRow
}

type settingRow struct {
	label, value string
	y            float64
}

func (st *Settings) Opaque() bool { return true }

func (st *Settings) add(r geom.Rect, label string, fn func(a *App)) {
	st.menu.Buttons = append(st.menu.Buttons, ui.Button{R: r, Label: label})
	st.acts = append(st.acts, fn)
}

func next[T comparable](list []T, cur T, dir int) T {
	for i, v := range list {
		if v == cur {
			return list[(i+dir+len(list))%len(list)]
		}
	}
	return list[0]
}

func (st *Settings) layout(a *App) {
	st.menu.Buttons, st.acts, st.rows = st.menu.Buttons[:0], st.acts[:0], st.rows[:0]
	s := &a.State.Settings
	W := float64(a.W)
	left, right := 40.0, W-40
	row := func(i int, label, value string, prev, nxt func(*App)) {
		y := 74 + float64(i)*58
		st.rows = append(st.rows, settingRow{label, value, y})
		st.add(geom.Rect{X: right - 260, Y: y, W: 48, H: 44}, "<", prev)
		st.add(geom.Rect{X: right - 48, Y: y, W: 48, H: 44}, ">", nxt)
	}
	_ = left
	vol := fmt.Sprintf("%d%%", s.Volume)
	if s.Muted {
		vol = a.T("settings.muted")
	}
	row(0, a.T("settings.volume"), vol,
		func(*App) { s.Muted = false; s.Volume = max(0, s.Volume-10) },
		func(a *App) { s.Muted = false; s.Volume = min(100, s.Volume+10); a.Sound("coin") })
	onOff := func(b bool) string {
		if b {
			return a.T("settings.on")
		}
		return a.T("settings.off")
	}
	row(1, a.T("settings.sound"), onOff(!s.Muted), func(*App) { s.Muted = !s.Muted }, func(*App) { s.Muted = !s.Muted })
	lvl := a.T("creator.junior")
	if a.State.Level == content.Expert {
		lvl = a.T("creator.expert")
	}
	flip := func(*App) {
		if a.State.Level == content.Expert {
			a.State.Level = content.Junior
		} else {
			a.State.Level = content.Expert
		}
	}
	row(2, a.T("settings.level"), lvl, flip, flip)
	row(3, a.T("settings.day"), a.Tf("settings.minutes", s.DayMinutes),
		func(*App) { a.State.SetDayMinutes(next(dayChoices, s.DayMinutes, -1)) },
		func(*App) { a.State.SetDayMinutes(next(dayChoices, s.DayMinutes, 1)) })
	row(4, a.T("settings.touch"), a.T(touchKey(s.Touch)),
		func(*App) { s.Touch = next(touchChoices, s.Touch, -1) },
		func(*App) { s.Touch = next(touchChoices, s.Touch, 1) })
	row(5, a.T("settings.fps"), onOff(s.ShowFPS), func(*App) { s.ShowFPS = !s.ShowFPS }, func(*App) { s.ShowFPS = !s.ShowFPS })
	st.add(geom.Rect{X: W - 140, Y: 10, W: 120, H: 44}, a.T("settings.done"), func(a *App) { st.close(a) })
}

// touchKey is the text key for a touch-controls choice (checked in tests).
func touchKey(t string) string { return "settings.touch." + t }

func (st *Settings) close(a *App) {
	a.SaveNow()
	a.Pop()
}

func (st *Settings) Update(a *App) error {
	st.layout(a)
	if closePressed() {
		st.close(a)
		return nil
	}
	if i := st.menu.Update(&a.Pointer); i >= 0 {
		st.acts[i](a)
	}
	return nil
}

func (st *Settings) Draw(a *App, screen *ebiten.Image) {
	st.layout(a)
	screen.Fill(colTitleBG)
	ui.Text(screen, a.T("settings.title"), ui.Bold(ui.SizeHeading+4), 20, 16, ui.ColTextLight, text.AlignStart)
	right := float64(a.W) - 40
	for _, r := range st.rows {
		ui.Text(screen, r.label, ui.Bold(ui.SizeBody), 40, r.y+10, ui.ColTextLight, text.AlignStart)
		ui.TextCentered(screen, r.value, ui.Bold(ui.SizeBody), right-130, r.y+22, ui.ColTextLight)
	}
	ui.TextCentered(screen, a.T("settings.note"), ui.Regular(ui.SizeSmall), float64(a.W)/2, float64(a.H)-24, ui.ColTextLight)
	st.menu.Draw(screen)
}
