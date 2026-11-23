package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/geom"
	"labsim/internal/ui"
)

// TriviaLab: quick rounds of space questions (E-SPACE-2). Each first right
// answer is a discovery with its own fun fact.
type TriviaLab struct {
	expBase
	round  []experiment.TriviaQ
	idx    int
	score  int
	picked int // answer position picked for this question, -1 = none yet
	done   bool
	played int // rounds played this visit, so each round is different
}

func NewTriviaLab() *TriviaLab {
	return &TriviaLab{expBase: expBase{titleKey: "trivia.title", darkBG: true, bg: ui.Hex("#2b2d5a"), expID: "trivia"}, picked: -1}
}

func (tv *TriviaLab) Opaque() bool { return true }

func (tv *TriviaLab) start(a *App) {
	seed := uint64(a.State.Clock.Day())*1009 + uint64(tv.played)*31
	tv.played++
	tv.round = experiment.TriviaRound(seed, a.Level() == content.Expert, experiment.RoundSize)
	tv.idx, tv.score, tv.picked, tv.done, tv.note = 0, 0, -1, false, ""
}

func (tv *TriviaLab) answerRects(a *App) []geom.Rect {
	m := tv.main(a)
	return grid(geom.Rect{X: m.X + 10, Y: m.Y + 110, W: m.W - 20}, 4, 2, 64)
}

func (tv *TriviaLab) nextRect(a *App) geom.Rect {
	m := tv.main(a)
	return geom.Rect{X: m.X + m.W/2 - 80, Y: m.Y + 260, W: 160, H: 44}
}

func (tv *TriviaLab) Update(a *App) error {
	if tv.tick(a) || tv.busy() {
		return nil
	}
	if tv.round == nil {
		tv.start(a)
	}
	if tv.done {
		if a.Pointer.TapIn(tv.nextRect(a)) {
			tv.start(a)
		}
		return nil
	}
	for i, r := range tv.answerRects(a)[:tv.round[tv.idx].N] {
		if a.Pointer.TapIn(r) {
			tv.pick(a, i)
		}
	}
	if tv.picked >= 0 && a.Pointer.TapIn(tv.nextRect(a)) {
		tv.next(a)
	}
	return nil
}

func answerKey(q experiment.TriviaQ, pos int) string {
	return fmt.Sprintf("trivia.%s.a%d", q.ID, q.Order[pos])
}

func (tv *TriviaLab) pick(a *App, pos int) {
	if tv.done || tv.picked >= 0 {
		return
	}
	q := tv.round[tv.idx]
	tv.picked = pos
	if q.IsRight(pos) {
		tv.score++
		tv.note, tv.noteGood = a.T("trivia.right"), true
		a.Sound("coin")
		tv.discoverQuietly(a, "trivia."+q.ID)
		return
	}
	a.Sound("fizzle")
	tv.note, tv.noteGood = a.Tf("trivia.wrong", a.T(fmt.Sprintf("trivia.%s.a0", q.ID))), false
}

func (tv *TriviaLab) next(a *App) {
	if tv.picked < 0 {
		return
	}
	tv.idx++
	tv.picked, tv.note = -1, ""
	if tv.idx < len(tv.round) {
		return
	}
	tv.done = true
	if experiment.TriviaPassed(tv.score, len(tv.round)) {
		tv.note, tv.noteGood = a.Tf("trivia.pass", tv.score, len(tv.round)), true
		tv.succeed(a, "")
	} else {
		tv.note, tv.noteGood = a.Tf("trivia.fail", tv.score, len(tv.round)), false
	}
}

func (tv *TriviaLab) Draw(a *App, screen *ebiten.Image) {
	tv.drawFrame(a, screen)
	m := tv.main(a)
	if tv.round == nil {
		return
	}
	ui.Text(screen, a.Tf("trivia.score", tv.score), ui.Bold(ui.SizeSmall), m.X+m.W-14, m.Y+10, ui.ColAccent, text.AlignEnd)
	if tv.done {
		ui.Button{R: tv.nextRect(a), Label: a.T("trivia.again")}.Draw(screen, true)
		for i := 0; i < len(tv.round); i++ {
			c := ui.ColPanelLine
			if i < tv.score {
				c = ui.ColCrystal
			}
			ui.Diamond(screen, m.X+m.W/2-float64(len(tv.round)-1)*20+float64(i)*40, m.Y+150, 12, c)
		}
		tv.drawOverlay(a, screen)
		return
	}
	q := tv.round[tv.idx]
	ui.Text(screen, a.Tf("trivia.qnum", tv.idx+1, len(tv.round)), ui.Bold(ui.SizeSmall), m.X+14, m.Y+10, ui.ColText, text.AlignStart)
	lines := ui.Wrap(a.T(triviaQKey(q.ID)), ui.Bold(ui.SizeBody), m.W-28)
	ui.Lines(screen, lines, ui.Bold(ui.SizeBody), ui.SizeBody, m.X+m.W/2, m.Y+40, ui.ColText, text.AlignCenter)
	for i, r := range tv.answerRects(a)[:q.N] {
		fill := ui.ColPanel
		switch {
		case tv.picked >= 0 && q.IsRight(i):
			fill = ui.Hex("#b8f0c0")
		case i == tv.picked:
			fill = ui.Hex("#ffc8c0")
		}
		ui.RoundRect(screen, r, 12, fill, 3, ui.ColPanelLine)
		al := ui.Wrap(a.T(answerKey(q, i)), ui.Bold(ui.SizeSmall), r.W-16)
		lh := ui.LineHeight(ui.SizeSmall)
		ui.Lines(screen, al, ui.Bold(ui.SizeSmall), ui.SizeSmall, r.X+r.W/2, r.Y+r.H/2-lh*float64(len(al))/2, ui.ColText, text.AlignCenter)
	}
	if tv.picked >= 0 {
		ui.Button{R: tv.nextRect(a), Label: a.T("trivia.next")}.Draw(screen, true)
	}
	tv.drawOverlay(a, screen)
}
