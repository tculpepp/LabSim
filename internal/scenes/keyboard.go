package scenes

import (
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"labsim/internal/filter"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

// Keyboard is an on-screen keyboard for typing a name. Tablets can't open the
// system keyboard for a canvas game, so we draw our own; physical keys work too.
type Keyboard struct {
	text   string
	errKey string
	onDone func(a *App, name string)
}

var keyRows = []string{"QWERTYUIOP", "ASDFGHJKL'", "ZXCVBNM-."}

const kbMax = 16 // typed names are shorter than preset ones

func NewKeyboard(initial string, onDone func(a *App, name string)) *Keyboard {
	return &Keyboard{text: initial, onDone: onDone}
}

func (k *Keyboard) Opaque() bool { return true }

func (k *Keyboard) keyRect(a *App, row, col int) geom.Rect {
	w := (float64(a.W) - 40) / 10
	return geom.Rect{X: 20 + float64(col)*w + 2, Y: 196 + float64(row)*54, W: w - 4, H: 48}
}

func (k *Keyboard) bottom(a *App) (space, del, done, cancel geom.Rect) {
	y := 196 + 3*54.0
	w := float64(a.W) - 40
	cancel = geom.Rect{X: 20, Y: y, W: w*0.2 - 4, H: 48}
	space = geom.Rect{X: 20 + w*0.2, Y: y, W: w*0.4 - 4, H: 48}
	del = geom.Rect{X: 20 + w*0.6, Y: y, W: w*0.2 - 4, H: 48}
	done = geom.Rect{X: 20 + w*0.8, Y: y, W: w*0.2 - 4, H: 48}
	return
}

// add types a character, capitalizing the start of each word.
func (k *Keyboard) add(r rune) {
	if len([]rune(k.text)) >= kbMax {
		return
	}
	if unicode.IsLetter(r) {
		if k.text == "" || k.text[len(k.text)-1] == ' ' || k.text[len(k.text)-1] == '-' {
			r = unicode.ToUpper(r)
		} else {
			r = unicode.ToLower(r)
		}
	}
	if r == ' ' && (k.text == "" || k.text[len(k.text)-1] == ' ') {
		return
	}
	k.text += string(r)
	k.errKey = ""
}

func (k *Keyboard) backspace() {
	if r := []rune(k.text); len(r) > 0 {
		k.text = string(r[:len(r)-1])
	}
	k.errKey = ""
}

func (k *Keyboard) finish(a *App) {
	name := filter.Clean(k.text)
	if err := filter.Check(name); err != nil {
		k.errKey = nameErrorKey(err)
		return
	}
	a.Pop()
	k.onDone(a, name)
}

func nameErrorKey(err error) string {
	switch err {
	case filter.ErrEmpty:
		return "name.empty"
	case filter.ErrTooLong:
		return "name.tooLong"
	case filter.ErrBadChars:
		return "name.badChars"
	default:
		return "name.blocked"
	}
}

func (k *Keyboard) Update(a *App) error {
	for _, r := range ebiten.AppendInputChars(nil) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '.' || r == '\'' {
			k.add(r)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		k.backspace()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		k.finish(a)
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.Pop()
		return nil
	}
	for row, keys := range keyRows {
		for col, r := range keys {
			if a.Pointer.TapIn(k.keyRect(a, row, col)) {
				k.add(r)
			}
		}
	}
	space, del, done, cancel := k.bottom(a)
	switch {
	case a.Pointer.TapIn(space):
		k.add(' ')
	case a.Pointer.TapIn(del):
		k.backspace()
	case a.Pointer.TapIn(done):
		k.finish(a)
	case a.Pointer.TapIn(cancel):
		a.Pop()
	}
	return nil
}

func (k *Keyboard) Draw(a *App, screen *ebiten.Image) {
	screen.Fill(colTitleBG)
	cx, _ := a.center()
	ui.TextCentered(screen, a.T("keyboard.title"), ui.Bold(ui.SizeHeading), cx, 40, ui.ColTextLight)
	field := geom.Centered(cx, 100, min(float64(a.W)-80, 460), 56)
	ui.RoundRect(screen, field, 14, ui.ColPanel, 3, ui.ColPanelLine)
	shown := k.text
	if (ebiten.Tick()/30)%2 == 0 {
		shown += "|"
	}
	ui.TextCentered(screen, shown, ui.Bold(ui.SizeHeading), cx, 100, ui.ColText)
	if k.errKey != "" {
		ui.TextCentered(screen, a.T(k.errKey), ui.Bold(ui.SizeBody), cx, 156, ui.ColFocus)
	}
	for row, keys := range keyRows {
		for col, r := range keys {
			ui.Button{R: k.keyRect(a, row, col), Label: string(r)}.Draw(screen, false)
		}
	}
	space, del, done, cancel := k.bottom(a)
	ui.Button{R: cancel, Label: a.T("keyboard.cancel")}.Draw(screen, false)
	ui.Button{R: space, Label: a.T("keyboard.space")}.Draw(screen, false)
	ui.Button{R: del, Label: a.T("keyboard.delete")}.Draw(screen, false)
	ui.Button{R: done, Label: a.T("keyboard.done")}.Draw(screen, true)
}
