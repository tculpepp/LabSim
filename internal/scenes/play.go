package scenes

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"labsim/internal/clock"
	"labsim/internal/economy"
	"labsim/internal/geom"
	"labsim/internal/input"
	"labsim/internal/inventory"
	"labsim/internal/ui"
	"labsim/internal/world"
)

const (
	playerSize    = 24.0
	playerSpeed   = 3.0 // logical px per tick at full stick
	interactReach = 20.0
	dayLength     = 10 * time.Minute
	bannerTicks   = 120
)

// Play is the main walking-around-the-lab scene.
type Play struct {
	px, py   float64
	unlocked map[string]bool
	clock    *clock.Clock
	wallet   *economy.Wallet
	backpack *inventory.Backpack
	joy      *input.Joystick
	showMap  bool

	roomID   string
	banner   string
	bannerT  int
	message  string
	messageT int
}

func NewPlay(a *App) *Play {
	return &Play{
		px:       a.Map.Spawn.X,
		py:       a.Map.Spawn.Y,
		unlocked: map[string]bool{},
		clock:    clock.New(dayLength),
		wallet:   economy.NewWallet(0),
		backpack: inventory.NewBackpack(6),
		joy:      input.NewJoystick(50, 8),
		roomID:   "main_hall",
	}
}

func (p *Play) isUnlocked(id string) bool { return p.unlocked[id] }

func (p *Play) box() geom.Rect { return geom.Centered(p.px, p.py, playerSize, playerSize) }

// HUD button rects (screen space).
func (p *Play) mapButton(a *App) geom.Rect {
	return geom.Rect{X: float64(a.W) - 132, Y: 8, W: 60, H: 32}
}
func (p *Play) bagButton(a *App) geom.Rect {
	return geom.Rect{X: float64(a.W) - 68, Y: 8, W: 60, H: 32}
}
func (p *Play) useButton(a *App) geom.Rect {
	return geom.Centered(float64(a.W)-70, float64(a.H)-70, 84, 84)
}

func (p *Play) Update(a *App) error {
	ticks := ebiten.TPS()
	p.clock.Advance(time.Second / time.Duration(ticks))
	if p.bannerT > 0 {
		p.bannerT--
	}
	if p.messageT > 0 {
		p.messageT--
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyM) || a.Pointer.TapIn(p.mapButton(a)) {
		p.showMap = !p.showMap
	}
	if p.showMap {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || a.Pointer.AnyTap() {
			p.showMap = false
		}
		return nil
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
	p.px, p.py = nb.Center()

	p.updateRoom(a)

	interact := inpututil.IsKeyJustPressed(ebiten.KeyE) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(a.Pointer.TouchDevice() && a.Pointer.TapIn(p.useButton(a)))
	if interact {
		p.interact(a)
	}
	return nil
}

func (p *Play) updateJoystick(a *App) {
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		fx, fy := float64(x), float64(y)
		onButton := p.useButton(a).Contains(fx, fy) || p.mapButton(a).Contains(fx, fy) || p.bagButton(a).Contains(fx, fy)
		if !onButton && fx < float64(a.W)*0.6 {
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

func (p *Play) updateRoom(a *App) {
	id := ""
	name := "Hallway"
	if r := a.Map.RoomAt(p.px, p.py); r != nil {
		id, name = r.ID, r.Name
	}
	if id != p.roomID {
		p.roomID = id
		if id != "" {
			p.banner, p.bannerT = name, bannerTicks
		}
	}
}

func (p *Play) interact(a *App) {
	d := a.Map.DoorNear(p.box(), interactReach)
	if d == nil || !a.Map.DoorLocked(*d, p.isUnlocked) {
		return
	}
	r := a.Map.Room(d.RoomID)
	p.message = fmt.Sprintf("The %s is locked! It costs %d crystals to open.", r.Name, r.UnlockPrice)
	p.messageT = bannerTicks * 2
}

func (p *Play) Draw(a *App, screen *ebiten.Image) {
	m := a.Map
	cx, cy := world.CameraFor(p.px, p.py, float64(a.W), float64(a.H), m.Width, m.Height)

	screen.Fill(ui.ColHallway)
	for _, r := range m.Rooms {
		c, _ := world.ParseHexColor(r.Color)
		ui.Fill(screen, r.Bounds, cx, cy, ui.RGB(c))
		ebitenutil.DebugPrintAt(screen, r.Name, int(r.Bounds.X-cx)+14, int(r.Bounds.Y-cy)+12)
	}
	for _, w := range m.Walls() {
		ui.Fill(screen, w, cx, cy, ui.ColWall)
	}
	for _, d := range m.Doors() {
		if !m.DoorLocked(d, p.isUnlocked) {
			continue
		}
		ui.Fill(screen, d.Rect, cx, cy, ui.ColLocked)
		dx, dy := d.Rect.Center()
		ui.Padlock(screen, dx-cx-24, dy-cy, 16, ui.ColLocked)
		r := m.Room(d.RoomID)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", r.UnlockPrice), int(dx-cx)-60, int(dy-cy)+14)
		ui.Diamond(screen, dx-cx-66, dy-cy+22, 6, ui.ColCrystal)
	}

	// Player placeholder: round scientist with goggles.
	sx, sy := p.px-cx, p.py-cy
	ui.Circle(screen, sx, sy, playerSize/2, ui.ColTextLight)
	ui.Ring(screen, sx, sy, playerSize/2, 2, ui.ColAccent)
	ui.Circle(screen, sx-4, sy-3, 3, ui.ColAccent)
	ui.Circle(screen, sx+4, sy-3, 3, ui.ColAccent)

	if p.clock.IsNight() {
		ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColNight)
	}
	p.drawHUD(a, screen)
	if p.showMap {
		p.drawMiniMap(a, screen)
	}
}

func (p *Play) drawHUD(a *App, screen *ebiten.Image) {
	ui.Fill(screen, geom.Rect{X: 8, Y: 8, W: 260, H: 32}, 0, 0, ui.ColPanel)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Day %d", p.clock.Day()), 16, 16)
	ui.Diamond(screen, 86, 24, 8, ui.ColCrystal)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", p.wallet.Crystals()), 98, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Bag %d/%d", p.backpack.UsedSlots(), p.backpack.Slots()), 170, 16)

	for _, b := range []struct {
		r     geom.Rect
		label string
	}{{p.mapButton(a), "Map"}, {p.bagButton(a), "Bag"}} {
		ui.Fill(screen, b.r, 0, 0, ui.ColPanel)
		ui.Stroke(screen, b.r, 0, 0, 2, ui.ColPanelLine)
		ebitenutil.DebugPrintAt(screen, b.label, int(b.r.X)+18, int(b.r.Y)+9)
	}

	if p.bannerT > 0 {
		w := float64(len(p.banner)*6 + 32)
		r := geom.Rect{X: float64(a.W)/2 - w/2, Y: 50, W: w, H: 28}
		ui.Fill(screen, r, 0, 0, ui.ColPanel)
		ebitenutil.DebugPrintAt(screen, p.banner, int(r.X)+16, int(r.Y)+7)
	}
	if p.messageT > 0 {
		r := geom.Rect{X: 40, Y: float64(a.H) - 80, W: float64(a.W) - 200, H: 40}
		ui.Fill(screen, r, 0, 0, ui.ColPanel)
		ui.Stroke(screen, r, 0, 0, 2, ui.ColPanelLine)
		ebitenutil.DebugPrintAt(screen, p.message, int(r.X)+12, int(r.Y)+13)
	}

	if a.Pointer.TouchDevice() {
		if p.joy.Active() {
			ox, oy := p.joy.Origin()
			kx, ky := p.joy.Knob()
			ui.Ring(screen, ox, oy, p.joy.Radius, 3, ui.ColPanel)
			ui.Circle(screen, kx, ky, 20, ui.ColPanel)
		}
		u := p.useButton(a)
		ux, uy := u.Center()
		ui.Circle(screen, ux, uy, u.W/2, ui.ColPanel)
		ui.Ring(screen, ux, uy, u.W/2, 3, ui.ColPanelLine)
		ebitenutil.DebugPrintAt(screen, "Use", int(ux)-9, int(uy)-8)
	}
}

func (p *Play) drawMiniMap(a *App, screen *ebiten.Image) {
	m := a.Map
	ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColShade)
	scale := min(float64(a.W)*0.85/m.Width, float64(a.H)*0.8/m.Height)
	ox := (float64(a.W) - m.Width*scale) / 2
	oy := (float64(a.H) - m.Height*scale) / 2
	ui.Fill(screen, geom.Rect{X: ox, Y: oy, W: m.Width * scale, H: m.Height * scale}, 0, 0, ui.ColHallway)
	for _, r := range m.Rooms {
		rr := geom.Rect{X: ox + r.Bounds.X*scale, Y: oy + r.Bounds.Y*scale, W: r.Bounds.W * scale, H: r.Bounds.H * scale}
		c, _ := world.ParseHexColor(r.Color)
		ui.Fill(screen, rr, 0, 0, ui.RGB(c))
		ui.Stroke(screen, rr, 0, 0, 2, ui.ColWall)
		ebitenutil.DebugPrintAt(screen, r.Name, int(rr.X)+4, int(rr.Y)+4)
		if r.Locked && !p.isUnlocked(r.ID) {
			rx, ry := rr.Center()
			ui.Padlock(screen, rx, ry, 18, ui.ColLocked)
		}
	}
	ui.Circle(screen, ox+p.px*scale, oy+p.py*scale, 6, ui.ColAccent)
	ui.Ring(screen, ox+p.px*scale, oy+p.py*scale, 6, 2, ui.ColTextLight)
	ebitenutil.DebugPrintAt(screen, "Lab Map  (tap or press M to close)", int(ox), int(oy)-18)
}
