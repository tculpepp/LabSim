package scenes

import (
	"fmt"

	"labsim/internal/content"
	"labsim/internal/state"
)

// DebugScene jumps straight to a named screen. Used by tools/snapshot for
// visual checks; not reachable from normal play.
func (a *App) DebugScene(name string) error {
	play := func() *Play {
		a.State = state.New("Scientist", content.Junior, a.Map.Spawn.X, a.Map.Spawn.Y)
		p := NewPlay(a)
		a.Replace(p)
		return p
	}
	switch name {
	case "title":
	case "title-continue":
		a.State = state.New("Scientist", content.Junior, a.Map.Spawn.X, a.Map.Spawn.Y)
		a.SaveNow()
		a.State = nil
		a.checkSave()
		a.Replace(NewTitle(a))
	case "creator":
		a.Replace(NewCreator(a))
	case "play":
		play()
	case "welcome":
		a.StartNewGame("Scientist", content.Junior)
	case "map":
		a.Push(&MapView{play: play()})
	case "bag":
		a.Push(NewInventory(play().s.Backpack))
	case "pause":
		play()
		a.Push(NewPause(a))
	case "confirm":
		play()
		a.Push(NewPause(a))
		a.Confirm("confirm.startOver", func(*App) {})
	case "locked":
		p := play()
		p.s.X, p.s.Y = 1740, 630 // hallway beside the Robot Workshop door
		p.interact(a)
	case "night":
		p := play()
		p.s.Clock.Advance(state.DayLength * 9 / 10)
	default:
		return fmt.Errorf("unknown scene %q", name)
	}
	return nil
}
