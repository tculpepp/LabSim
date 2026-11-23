package scenes

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"labsim/internal/geom"
	"labsim/internal/input"
	"labsim/internal/items"
	"labsim/internal/npc"
	"labsim/internal/state"
	"labsim/internal/ui"
	"labsim/internal/world"
)

const (
	playerSize    = 24.0
	playerSpeed   = 3.0 // logical px per tick at full stick
	interactReach = 18.0
	bannerTicks   = 120
)

// Play is the main walking-around-the-lab scene.
type Play struct {
	s   *state.State
	joy *input.Joystick

	t      float64 // seconds since scene start (animations)
	facing float64
	moving bool
	petX   float64
	petY   float64
	gizX   float64
	gizY   float64

	roomID  string
	banner  string
	bannerT int

	pickups   []items.Pickup
	pickDay   int
	floats    []floatText
	fullWarnT int

	animals  []*animal
	hideJobs bool // job tracker collapsed
	hearts   []heart
}

// animal is a wandering lab animal (C-4).
type animal struct {
	spec animalSpec
	w    *npc.Wanderer
}

// heart floats up when you pet an animal (S-6).
type heart struct {
	x, y float64
	t    int
}

// floatText is a little "+1 Star Dust" label that rises and fades.
type floatText struct {
	text string
	x, y float64
	t    int
	col  color.Color
}

const (
	grabReach  = 22.0
	floatTicks = 70
)

func NewPlay(a *App) *Play {
	s := a.State
	p := &Play{s: s, joy: input.NewJoystick(50, 8), facing: 1,
		petX: s.X - 22, petY: s.Y + 6, gizX: s.X + 26, gizY: s.Y - 34}
	if r := a.Map.RoomAt(s.X, s.Y); r != nil {
		p.roomID = r.ID
	}
	s.Room = p.roomID
	p.spawnAnimals(a)
	return p
}

func (p *Play) spawnAnimals(a *App) {
	for i, sp := range a.Animals {
		area := a.RoomRects[sp.Room].Expand(-40)
		if sp.Kind == "axolotl" {
			if st := a.Map.Station("tank"); st != nil {
				area = st.Rect().Expand(-12)
				area.Y += 6
			}
		}
		p.animals = append(p.animals, &animal{spec: sp, w: npc.NewWanderer(area, uint64(i+1)*977, sp.Speed)})
	}
}

func (p *Play) updateAnimals(a *App) {
	for _, an := range p.animals {
		switch an.spec.Kind {
		case "butterfly", "axolotl": // flying / swimming: no walls to bump into
			an.w.Update(func(float64, float64) bool { return true })
		default:
			an.w.Update(func(x, y float64) bool {
				return !a.Map.Collides(geom.Centered(x, y, 20, 14), nil) &&
					math.Hypot(x-p.s.X, y-p.s.Y) > 26 // don't walk into the player
			})
		}
	}
	kept := p.hearts[:0]
	for _, h := range p.hearts {
		if h.t++; h.t < floatTicks {
			kept = append(kept, h)
		}
	}
	p.hearts = kept
}

// nearAnimal is the closest pettable animal in reach.
func (p *Play) nearAnimal() *animal {
	var best *animal
	bestD := 34.0
	for _, an := range p.animals {
		if an.spec.Kind == "axolotl" {
			continue // in its tank; use the tank
		}
		if d := math.Hypot(an.w.X-p.s.X, an.w.Y-p.s.Y); d < bestD {
			best, bestD = an, d
		}
	}
	return best
}

func (p *Play) pet(a *App, kind string, x, y float64) {
	for i := 0; i < 3; i++ {
		p.hearts = append(p.hearts, heart{x: x - 10 + float64(i)*10, y: y - 20 - float64(i)*4, t: -i * 8})
	}
	p.float(a.T(animalKey(kind)), x, y-34, ui.ColTextLight)
	a.Sound("pet")
}

func (p *Play) Opaque() bool { return true }

func (p *Play) isUnlocked(id string) bool { return p.s.IsUnlocked(id) }

func (p *Play) box() geom.Rect { return geom.Centered(p.s.X, p.s.Y, playerSize, playerSize) }

// HUD buttons (screen space), right to left: pause, bag, map, notes.
func (p *Play) hudButton(a *App, i int) geom.Rect {
	return geom.Rect{X: float64(a.W) - 8 - float64(i+1)*66, Y: 8, W: 60, H: 40}
}
func (p *Play) pauseButton(a *App) geom.Rect { return p.hudButton(a, 0) }
func (p *Play) bagButton(a *App) geom.Rect   { return p.hudButton(a, 1) }
func (p *Play) mapButton(a *App) geom.Rect   { return p.hudButton(a, 2) }
func (p *Play) notesButton(a *App) geom.Rect { return p.hudButton(a, 3) }
func (p *Play) useButton(a *App) geom.Rect {
	return geom.Centered(float64(a.W)-70, float64(a.H)-70, 88, 88)
}

func (p *Play) onButton(a *App, x, y float64) bool {
	for _, r := range []geom.Rect{p.useButton(a), p.mapButton(a), p.bagButton(a), p.pauseButton(a), p.notesButton(a), p.trackerRect(a)} {
		if r.Contains(x, y) {
			return true
		}
	}
	return false
}

// open pushes an overlay, releasing the joystick first.
func (p *Play) open(a *App, s Scene) {
	p.joy.End(p.joy.ID())
	a.Push(s)
}

func (p *Play) Update(a *App) error {
	dt := time.Second / time.Duration(ebiten.TPS())
	p.t += dt.Seconds()
	if p.s.Clock.Advance(dt) > 0 {
		p.banner, p.bannerT = a.Tf("day.new", p.s.Clock.Day()), bannerTicks*2
		a.SaveNow()
	} else if p.bannerT > 0 {
		p.bannerT--
	}
	a.autosave.Tick(dt)

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyM) || a.Pointer.TapIn(p.mapButton(a)):
		p.open(a, &MapView{play: p})
		return nil
	case inpututil.IsKeyJustPressed(ebiten.KeyI) || inpututil.IsKeyJustPressed(ebiten.KeyB) || a.Pointer.TapIn(p.bagButton(a)):
		p.open(a, NewInventory(p.s.Backpack))
		return nil
	case inpututil.IsKeyJustPressed(ebiten.KeyN) || a.Pointer.TapIn(p.notesButton(a)):
		p.open(a, &NotebookView{})
		return nil
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) || a.Pointer.TapIn(p.pauseButton(a)):
		p.open(a, NewPause(a))
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyJ) || a.Pointer.TapIn(p.trackerRect(a)) {
		p.hideJobs = !p.hideJobs
	}
	p.updateJoystick(a)
	dx, dy := input.KeyVector(
		ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW),
		ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS),
	)
	if p.joy.Active() {
		dx, dy = p.joy.Vector()
	}
	nb := a.Map.Move(p.box(), dx*playerSpeed, dy*playerSpeed, p.isUnlocked)
	p.s.X, p.s.Y = nb.Center()
	p.moving = dx != 0 || dy != 0
	if math.Abs(dx) > 0.2 {
		p.facing = math.Copysign(1, dx)
	}
	p.updateFollowers()
	p.updateAnimals(a)
	p.updateRoom(a)
	p.refreshPickups(a)
	p.grab(a)
	p.tickFloats()

	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(a.Pointer.TouchDevice() && a.Pointer.TapIn(p.useButton(a))) {
		p.interact(a)
	}
	return nil
}

// updateFollowers eases the pet (behind) and Gizmo (floating beside) toward the player.
func (p *Play) updateFollowers() {
	ease := func(cur *float64, target, k float64) { *cur += (target - *cur) * k }
	ease(&p.petX, p.s.X-p.facing*24, 0.12)
	ease(&p.petY, p.s.Y+8, 0.12)
	ease(&p.gizX, p.s.X+p.facing*28, 0.06)
	ease(&p.gizY, p.s.Y-36, 0.06)
}

func (p *Play) updateJoystick(a *App) {
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		fx, fy := float64(x), float64(y)
		if !p.onButton(a, fx, fy) && fx < float64(a.W)*0.6 {
			p.joy.Begin(int(id), fx, fy)
		}
	}
	if p.joy.Active() {
		id := ebiten.TouchID(p.joy.ID())
		if inpututil.IsTouchJustReleased(id) {
			p.joy.End(int(id))
		} else {
			x, y := ebiten.TouchPosition(id)
			p.joy.Move(int(id), float64(x), float64(y))
		}
	}
}

func (p *Play) roomName(a *App) string {
	if r := a.Map.Room(p.roomID); r != nil {
		return r.Name
	}
	return a.T("room.hallway")
}

func (p *Play) updateRoom(a *App) {
	id := ""
	if r := a.Map.RoomAt(p.s.X, p.s.Y); r != nil {
		id = r.ID
	}
	if id != p.roomID {
		p.roomID = id
		p.s.Room = id
		a.SaveNow() // SV-1: save on room change
		if id != "" {
			p.banner, p.bannerT = p.roomName(a), bannerTicks
		}
	}
}

// nearStation is the station the player can use right now, if any.
func (p *Play) nearStation(a *App) *world.Station {
	return a.Map.StationNear(p.box(), interactReach)
}

func (p *Play) interact(a *App) {
	if st := p.nearStation(a); st != nil {
		p.useStation(a, st)
		return
	}
	if an := p.nearAnimal(); an != nil {
		p.pet(a, an.spec.Kind, an.w.X, an.w.Y)
		return
	}
	d := a.Map.DoorNear(p.box(), interactReach)
	if d == nil || !a.Map.DoorLocked(*d, p.isUnlocked) {
		return
	}
	r := a.Map.Room(d.RoomID)
	a.Say("gizmo", "gizmo.name", a.Tf("door.locked", r.Name, r.UnlockPrice))
}

func (p *Play) useStation(a *App, st *world.Station) {
	if j, reward, ok := a.TryDeliver(st.ID); ok {
		p.float(a.Tf("job.paid", reward), st.X, st.Y-40, ui.ColFocus)
		who, name := "gizmo", "gizmo.name"
		if st.ID == "nova" {
			who, name = "nova", "nova.name"
		}
		a.Say(who, name, a.Tf("job.delivered", j.Title, reward))
		p.afterPay(a)
		return
	}
	switch st.Kind {
	case "npc":
		a.TalkToNova()
	case "tank":
		p.pet(a, "axolotl", st.X, st.Y-10)
	case "wardrobe":
		p.open(a, NewEditCreator(a))
	case "shop":
		p.open(a, NewShop(a))
	case "decorate":
		p.open(a, NewDecorate(a))
	case "experiment":
		switch st.ID {
		case "exp_slime":
			p.open(a, NewSlimeLab())
		case "exp_circuit":
			p.open(a, NewCircuitLab())
		case "exp_microscope":
			p.open(a, NewMicroLab())
		case "exp_gravity":
			p.open(a, NewGravityLab())
		}
	default:
		a.Say("gizmo", "gizmo.name", a.T("station.soon"))
	}
}

// refreshPickups lays out today's items (I-4: they respawn every day).
func (p *Play) refreshPickups(a *App) {
	day := p.s.Clock.Day()
	if p.pickDay == day {
		return
	}
	p.pickDay = day
	free := func(x, y float64) bool { return !a.Map.Collides(geom.Centered(x, y, 28, 28), nil) }
	p.pickups = items.Generate(a.Spawns, a.RoomRects, day, free)
}

// grab picks up items the player walks over (I-1).
func (p *Play) grab(a *App) {
	if p.fullWarnT > 0 {
		p.fullWarnT--
	}
	for _, pk := range p.pickups {
		if p.s.Picked(pk.Key) || math.Hypot(pk.X-p.s.X, pk.Y-p.s.Y) > grabReach {
			continue
		}
		if err := p.s.Backpack.Add(pk.ItemID, 1); err != nil {
			if p.fullWarnT == 0 {
				p.banner, p.bannerT = a.T("bag.full"), bannerTicks*2
			}
			p.fullWarnT = bannerTicks * 3
			continue
		}
		p.s.MarkPicked(pk.Key)
		p.float(a.Tf("pickup.got", a.ItemName(pk.ItemID)), pk.X, pk.Y-20, ui.ColTextLight)
		a.Sound("pickup")
	}
}

func (p *Play) float(text string, x, y float64, c color.Color) {
	p.floats = append(p.floats, floatText{text: text, x: x, y: y, col: c})
}

func (p *Play) tickFloats() {
	kept := p.floats[:0]
	for _, f := range p.floats {
		f.t++
		if f.t < floatTicks {
			kept = append(kept, f)
		}
	}
	p.floats = kept
}

// afterPay shows Gizmo's tutorial tip the first time you get paid.
func (p *Play) afterPay(a *App) {
	if !a.State.Flag("tut.paid") {
		a.State.SetFlag("tut.paid")
		a.Say("gizmo", "gizmo.name", a.T("tut.paid"))
	}
}

// jobTarget is where Gizmo's arrow points: Dr. Nova if you have no jobs
// yet, otherwise the station for your next job (C-3).
func (p *Play) jobTarget(a *App) *world.Station {
	b := a.Board()
	if b == nil {
		return a.Map.Station("nova")
	}
	for _, j := range b.Open() {
		return a.Map.Station(j.Station)
	}
	return nil
}

// animalKey is the text key for petting an animal (checked by TestAnimalTexts).
func animalKey(kind string) string { return "animal." + kind }
