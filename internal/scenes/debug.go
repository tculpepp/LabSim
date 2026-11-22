package scenes

import "fmt"

// DebugScene jumps straight to a named screen. Used by tools/snapshot for
// visual checks; not reachable from normal play.
func (a *App) DebugScene(name string) error {
	play := func() *Play {
		a.PlayerName = "Scientist"
		p := NewPlay(a)
		a.Replace(p)
		return p
	}
	switch name {
	case "title":
	case "creator":
		a.Replace(NewCreator(a))
	case "play":
		play()
	case "welcome":
		a.StartNewGame("Scientist", a.Level)
	case "map":
		a.Push(&MapView{play: play()})
	case "bag":
		a.Push(NewInventory(play().backpack))
	case "pause":
		play()
		a.Push(NewPause(a))
	case "locked":
		p := play()
		p.px, p.py = 1740, 630 // hallway beside the Robot Workshop door
		p.interact(a)
	case "night":
		p := play()
		p.clock.Advance(dayLength * 9 / 10)
	default:
		return fmt.Errorf("unknown scene %q", name)
	}
	return nil
}
