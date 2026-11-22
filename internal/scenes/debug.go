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
		a.State = state.New("Doctor Sparkle Beaker", content.Junior, a.Map.Spawn.X, a.Map.Spawn.Y)
		a.State.Avatar = a.Avatars.Default()
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
	case "creator-name":
		c := NewCreator(a)
		c.tab = 3
		a.Replace(c)
	case "keyboard":
		a.Replace(NewCreator(a))
		k := NewKeyboard("Shit", func(*App, string) {})
		k.text = "Doctor Poo"
		k.errKey = "name.blocked"
		a.Push(k)
	case "office":
		p := play()
		p.s.X, p.s.Y = 840, 230
		p.updateFollowers()
	case "shop":
		p := play()
		p.s.X, p.s.Y = 1400, 220
	case "mylab":
		p := play()
		p.s.X, p.s.Y = 760, 1000
		p.s.Avatar = map[string]string{"body": "strong", "skin": "s5", "hair": "curly", "hairColor": "pink",
			"coat": "green", "shirt": "yellow", "shoes": "red", "pattern": "dots", "eyewear": "goggles",
			"hat": "beanie", "gear": "backpack", "pet": "slime"}
	case "play":
		play()
	case "welcome":
		a.StartNewGame("Scientist", content.Junior, a.Avatars.Default())
	case "map":
		a.Push(&MapView{play: play()})
	case "bag":
		p := play()
		for _, id := range []string{"star_dust", "battery", "glow_powder", "moon_rock", "pond_sample"} {
			p.s.Backpack.Add(id, 2)
		}
		a.Push(NewInventory(p.s.Backpack))
	case "storage":
		p := play()
		p.s.X, p.s.Y = 1400, 1150
		p.refreshPickups(a)
	case "shop-sell":
		p := play()
		for _, id := range []string{"star_dust", "battery", "moon_rock"} {
			p.s.Backpack.Add(id, 3)
		}
		sh := NewShop(a)
		sh.tab = 1
		a.Push(sh)
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
