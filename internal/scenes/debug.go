package scenes

import (
	"fmt"
	"labsim/internal/decor"

	"labsim/internal/content"
	"labsim/internal/experiment"
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
	case "shop-buy", "shop-outfits":
		p := play()
		p.s.Wallet.Earn(75)
		p.s.Own("tool:spark_kit")
		sh := NewShop(a)
		if name == "shop-outfits" {
			sh.cat = 1
		}
		a.Push(sh)
	case "settings":
		p := play()
		p.s.Settings.ShowFPS = true
		a.Push(&Settings{})
	case "decorate":
		p := play()
		for _, id := range []string{"plant", "rug", "lamp", "bookshelf", "rocket", "trophy", "globe", "fishbowl", "beanbag"} {
			p.s.Own("decor:" + id)
		}
		p.s.Decor = []decor.Placement{{ID: "plant", Cell: decor.Cell{X: 0, Y: 5}}, {ID: "rug", Cell: decor.Cell{X: 4, Y: 3}},
			{ID: "bookshelf", Cell: decor.Cell{X: 6, Y: 1}}, {ID: "lamp", Cell: decor.Cell{X: 8, Y: 5}}, {ID: "rocket", Cell: decor.Cell{X: 2, Y: 2}}}
		d := NewDecorate(a)
		d.sel = "lamp"
		a.Push(d)
	case "mylab":
		p := play()
		for _, id := range []string{"plant", "rug", "lamp", "bookshelf", "rocket", "trophy", "globe", "fishbowl", "beanbag"} {
			p.s.Own("decor:" + id)
		}
		p.s.Decor = []decor.Placement{{ID: "plant", Cell: decor.Cell{X: 0, Y: 5}}, {ID: "rug", Cell: decor.Cell{X: 4, Y: 3}},
			{ID: "bookshelf", Cell: decor.Cell{X: 6, Y: 1}}, {ID: "lamp", Cell: decor.Cell{X: 8, Y: 5}}, {ID: "rocket", Cell: decor.Cell{X: 2, Y: 2}},
			{ID: "globe", Cell: decor.Cell{X: 5, Y: 5}}, {ID: "fishbowl", Cell: decor.Cell{X: 3, Y: 5}}, {ID: "trophy", Cell: decor.Cell{X: 8, Y: 2}}, {ID: "beanbag", Cell: decor.Cell{X: 1, Y: 4}}}
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
	case "nova-jobs":
		p := play()
		p.s.X, p.s.Y = 840, 210
		a.TalkToNova()
	case "hall":
		p := play()
		for i := 0; i < 200; i++ {
			p.updateAnimals(a)
		}
		p.s.X, p.s.Y = 760, 640
		a.State.GiveJobs([]string{"dust_nova", "rock_gravity", "battery_circuit"})
		a.State.MarkJobDone("dust_nova")
		a.State.Backpack.Add("battery", 1)
	case "greenhouse":
		p := play()
		p.s.X, p.s.Y = 300, 1130
		for i := 0; i < 100; i++ {
			p.updateAnimals(a)
		}
		p.pet(a, "axolotl", 380, 1130)
		for i := range p.hearts {
			p.hearts[i].t = 12
		}
	case "slime":
		p := play()
		for _, id := range []string{"slime_base", "glow_powder", "bouncy_beads", "rainbow_drops"} {
			p.s.Backpack.Add(id, 2)
		}
		s := NewSlimeLab()
		s.bowl = []string{"slime_base", "rainbow_drops"}
		a.Push(s)
	case "slime-made":
		play()
		s := NewSlimeLab()
		s.result, s.note, s.noteGood, s.t = "glow", "You made Glow-in-the-Dark Slime!", true, 1
		a.Push(s)
	case "slime-boom":
		play()
		s := NewSlimeLab()
		a.Push(s)
		s.boom(a, "slime.hint.base")
		s.boomT = boomTicks / 2
	case "circuit":
		play()
		c := NewCircuitLab()
		a.Push(c)
	case "circuit-lit":
		play()
		a.State.Level = content.Expert
		c := NewCircuitLab()
		c.chosen, c.phase, c.lit, c.result, c.note, c.noteGood = "pencil", 2, true, "The bulb lights up! Pencil Lead is a conductor.", "Great prediction!", true
		a.Push(c)
	case "micro":
		play()
		a.State.Own("tool:super_microscope")
		m := NewMicroLab()
		m.pick(a, mustSlide("onion"))
		a.Push(m)
	case "gravity":
		play()
		g := NewGravityLab()
		g.planet, g.jumpT = "moon", jumpSecs/2
		a.Push(g)
	case "picker":
		play()
		a.State.Notebook.Add("slime.glow")
		a.Push(NewExpPicker("chemistry"))
	case "crystal":
		play()
		c := NewCrystalLab()
		c.powder, c.spoons, c.grown, c.growT, c.t = "alum", 3, "alum", crystalGrowTicks, 1
		a.Push(c)
	case "crystal-start":
		play()
		c := NewCrystalLab()
		c.powder, c.spoons, c.t = "sugar", 4, 1
		a.Push(c)
	case "gforce":
		play()
		g := NewGForceLab()
		g.current, g.rideT = "coaster", 1
		a.Push(g)
	case "gforce-expert":
		play()
		a.State.Level = content.Expert
		g := NewGForceLab()
		g.current, g.rideT = "jet", 1
		a.Push(g)
	case "trivia":
		play()
		tv := NewTriviaLab()
		a.Push(tv)
		tv.start(a)
		tv.round[0] = experiment.TriviaRound(5, false, 20)[0]
		for i, q := range tv.round {
			if q.ID == "moonlight" {
				tv.round[0], tv.round[i] = tv.round[i], tv.round[0]
			}
		}
		tv.pick(a, 0)
		tv.sparkleT, tv.pendingFC = 0, ""
	case "garden":
		play()
		g := NewGardenLab()
		g.pots[0] = experiment.Pot{Seed: "sunflower", Stage: experiment.BloomStage, Water: 1}
		a.Push(g)
	case "garden-full":
		play()
		g := NewGardenLab()
		g.setMode(true)
		for i := range g.pots {
			g.pots[i] = experiment.Pot{Seed: experiment.Flowers[i%4], Stage: i % 5, Water: i % 3}
		}
		g.tool, g.day = toolCan, 3
		a.Push(g)
	case "robot":
		play()
		r := NewRobotLab()
		r.bot = experiment.Robot{"head": "sonar", "arms": "magnet", "movers": "tracks"}
		r.challenge, r.passed, r.runT = "sand", true, robotRunSecs/2
		a.Push(r)
	case "robot-stairs":
		play()
		a.State.Level = content.Expert
		r := NewRobotLab()
		r.bot = experiment.Robot{"head": "camera", "arms": "claw", "movers": "legs"}
		r.challenge, r.passed, r.runT = "stairs", true, robotRunSecs
		a.Push(r)
	case "notebook-space":
		play()
		for _, id := range []string{"gravity.moon", "trivia.sun", "trivia.hottest"} {
			a.State.Notebook.Add(id)
		}
		a.Push(&NotebookView{tab: 3})
	case "notebook":
		play()
		for _, id := range []string{"slime.glow", "slime.rainbow", "circuit.foil"} {
			a.State.Notebook.Add(id)
		}
		a.Push(&NotebookView{})
	case "fact":
		play()
		a.Push(NewFactCard("gravity.moon", true))
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
