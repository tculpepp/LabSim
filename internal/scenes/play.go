package scenes

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/input"
	"labsim/internal/state"
	"labsim/internal/ui"
	"labsim/internal/world"
)

const (
	playerSize    = 24.0
	playerSpeed   = 3.0 // logical px per tick at full stick
	interactReach = 20.0
	bannerTicks   = 120
)

// Play is the main walking-around-the-lab scene.
type Play struct {
	s   *state.State
	joy *input.Joystick

	roomID  string
	banner  string
	bannerT int
}

func NewPlay(a *App) *Play {
	p := &Play{s: a.State, joy: input.NewJoystick(50, 8)}
	if r := a.Map.RoomAt(p.s.X, p.s.Y); r != nil {
		p.roomID = r.ID
	}
	p.s.Room = p.roomID
	return p
}

func (p *Play) Opaque() bool { return true }

func (p *Play) isUnlocked(id string) bool { return p.s.IsUnlocked(id) }

func (p *Play) box() geom.Rect { return geom.Centered(p.s.X, p.s.Y, playerSize, playerSize) }

// HUD buttons (screen space), right to left: pause, bag, map.
func (p *Play) hudButton(a *App, i int) geom.Rect {
	return geom.Rect{X: float64(a.W) - 8 - float64(i+1)*66, Y: 8, W: 60, H: 40}
}
func (p *Play) pauseButton(a *App) geom.Rect { return p.hudButton(a, 0) }
func (p *Play) bagButton(a *App) geom.Rect   { return p.hudButton(a, 1) }
func (p *Play) mapButton(a *App) geom.Rect   { return p.hudButton(a, 2) }
func (p *Play) useButton(a *App) geom.Rect {
	return geom.Centered(float64(a.W)-70, float64(a.H)-70, 88, 88)
}

func (p *Play) onButton(a *App, x, y float64) bool {
	for _, r := range []geom.Rect{p.useButton(a), p.mapButton(a), p.bagButton(a), p.pauseButton(a)} {
		if r.Contains(x, y) {
			return true
		}
	}
	return false
}

func (p *Play) Update(a *App) error {
	dt := time.Second / time.Duration(ebiten.TPS())
	if p.s.Clock.Advance(dt) > 0 {
		p.banner, p.bannerT = a.Tf("day.new", p.s.Clock.Day()), bannerTicks*2
	} else if p.bannerT > 0 {
		p.bannerT--
	}
	a.autosave.Tick(dt)

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyM) || a.Pointer.TapIn(p.mapButton(a)):
		p.joy.End(p.joy.ID())
		a.Push(&MapView{play: p})
		return nil
	case inpututil.IsKeyJustPressed(ebiten.KeyI) || inpututil.IsKeyJustPressed(ebiten.KeyB) || a.Pointer.TapIn(p.bagButton(a)):
		p.joy.End(p.joy.ID())
		a.Push(NewInventory(p.s.Backpack))
		return nil
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) || a.Pointer.TapIn(p.pauseButton(a)):
		p.joy.End(p.joy.ID())
		a.Push(NewPause(a))
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
	p.s.X, p.s.Y = nb.Center()
	p.updateRoom(a)

	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		(a.Pointer.TouchDevice() && a.Pointer.TapIn(p.useButton(a))) {
		p.interact(a)
	}
	return nil
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

func (p *Play) interact(a *App) {
	d := a.Map.DoorNear(p.box(), interactReach)
	if d == nil || !a.Map.DoorLocked(*d, p.isUnlocked) {
		return
	}
	r := a.Map.Room(d.RoomID)
	a.Say("gizmo", "gizmo.name", a.Tf("door.locked", r.Name, r.UnlockPrice))
}

func (p *Play) Draw(a *App, screen *ebiten.Image) {
	m := a.Map
	cx, cy := world.CameraFor(p.s.X, p.s.Y, float64(a.W), float64(a.H), m.Width, m.Height)

	screen.Fill(ui.ColHallway)
	for _, r := range m.Rooms {
		c, _ := world.ParseHexColor(r.Color)
		ui.Fill(screen, r.Bounds, cx, cy, ui.RGB(c))
	}
	for _, w := range m.Walls() {
		ui.Fill(screen, w, cx, cy, ui.ColWall)
	}
	for _, r := range m.Rooms {
		ui.Text(screen, r.Name, ui.Bold(ui.SizeBody), r.Bounds.X-cx+18, r.Bounds.Y-cy+16, color.RGBA{0, 0, 0, 0x70}, text.AlignStart)
	}
	for _, d := range m.Doors() {
		if !m.DoorLocked(d, p.isUnlocked) {
			continue
		}
		ui.Fill(screen, d.Rect, cx, cy, ui.ColLocked)
		dx, dy := d.Rect.Center()
		sx, sy := dx-cx-60, dy-cy-80 // badge up and away from the doorway
		r := m.Room(d.RoomID)
		ui.RoundRect(screen, geom.Centered(sx, sy, 76, 54), 10, ui.ColPanel, 2, ui.ColLocked)
		ui.Padlock(screen, sx, sy-8, 14, ui.ColLocked)
		ui.Diamond(screen, sx-14, sy+14, 6, ui.ColCrystal)
		ui.Text(screen, fmt.Sprint(r.UnlockPrice), ui.Bold(ui.SizeSmall), sx-4, sy+5, ui.ColText, text.AlignStart)
	}

	// Player placeholder: round scientist with goggles (real avatar in M4).
	sx, sy := p.s.X-cx, p.s.Y-cy
	ui.Circle(screen, sx, sy, playerSize/2, ui.ColTextLight)
	ui.Ring(screen, sx, sy, playerSize/2, 2, ui.ColAccent)
	ui.Circle(screen, sx-4, sy-3, 3, ui.ColAccent)
	ui.Circle(screen, sx+4, sy-3, 3, ui.ColAccent)

	if p.s.Clock.IsNight() {
		ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColNight)
	}
	p.drawHUD(a, screen)
}

func (p *Play) drawHUD(a *App, screen *ebiten.Image) {
	small, body := ui.Bold(ui.SizeSmall), ui.Bold(ui.SizeBody)
	// Status pill: day, crystals, backpack.
	ui.RoundRect(screen, geom.Rect{X: 8, Y: 8, W: 300, H: 40}, 20, ui.ColPanel, 2, ui.ColPanelLine)
	if p.s.Clock.IsNight() {
		ui.Circle(screen, 30, 28, 9, color.RGBA{0xf0, 0xe6, 0x8c, 0xff})
		ui.Circle(screen, 34, 25, 8, ui.ColPanel) // crescent moon
	} else {
		ui.Circle(screen, 30, 28, 9, color.RGBA{0xff, 0xc8, 0x3c, 0xff})
	}
	ui.Text(screen, a.Tf("hud.day", p.s.Clock.Day()), body, 46, 17, ui.ColText, text.AlignStart)
	ui.Diamond(screen, 134, 28, 10, ui.ColCrystal)
	ui.Text(screen, fmt.Sprint(p.s.Wallet.Crystals()), body, 148, 17, ui.ColText, text.AlignStart)
	ui.Text(screen, a.Tf("hud.bag", p.s.Backpack.UsedSlots(), p.s.Backpack.Slots()), body, 210, 17, ui.ColText, text.AlignStart)

	for _, b := range []ui.Button{
		{R: p.mapButton(a), Label: a.T("hud.map")},
		{R: p.bagButton(a), Label: a.T("hud.bagButton")},
		{R: p.pauseButton(a), Label: "II"},
	} {
		b.Draw(screen, false)
	}

	// Current room label, bottom-left, plus a warning if saving is broken.
	ui.Text(screen, p.roomName(a), small, 14, float64(a.H)-28, ui.ColTextLight, text.AlignStart)
	if a.SaveFailing() {
		ui.Text(screen, a.T("save.failed"), small, 14, 56, ui.ColFocus, text.AlignStart)
	}

	if p.bannerT > 0 {
		face := ui.Bold(ui.SizeHeading)
		w := ui.Measure(p.banner, face) + 48
		cx, _ := a.center()
		r := geom.Centered(cx, 84, w, 44)
		ui.Panel(screen, r)
		ui.TextCentered(screen, p.banner, face, cx, 84, ui.ColText)
	}

	if a.Pointer.TouchDevice() {
		if p.joy.Active() {
			ox, oy := p.joy.Origin()
			kx, ky := p.joy.Knob()
			ui.Ring(screen, ox, oy, p.joy.Radius, 4, ui.ColPanel)
			ui.Circle(screen, kx, ky, 22, ui.ColPanel)
		}
		u := p.useButton(a)
		ux, uy := u.Center()
		ui.Circle(screen, ux, uy, u.W/2, ui.ColPanel)
		ui.Ring(screen, ux, uy, u.W/2, 3, ui.ColPanelLine)
		ui.TextCentered(screen, a.T("hud.use"), body, ux, uy, ui.ColText)
	}
}
