package scenes

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"labsim/internal/geom"
	"labsim/internal/ui"
	"labsim/internal/world"
)

// sprite is something drawn in the world, sorted by its ground y so things
// in front overlap things behind.
type sprite struct {
	y    float64
	draw func()
}

func (p *Play) Draw(a *App, screen *ebiten.Image) {
	m := a.Map
	cx, cy := world.CameraFor(p.s.X, p.s.Y, float64(a.W), float64(a.H), m.Width, m.Height)

	screen.Fill(ui.ColHallway)
	for _, r := range m.Rooms {
		c, _ := world.ParseHexColor(r.Color)
		ui.Fill(screen, r.Bounds, cx, cy, ui.RGB(c))
		drawFloor(screen, r, cx, cy)
	}
	for _, w := range m.Walls() {
		ui.Fill(screen, w, cx, cy, ui.ColWall)
	}
	for _, r := range m.Rooms {
		ui.Text(screen, r.Name, ui.Bold(ui.SizeBody), r.Bounds.X-cx+18, r.Bounds.Y-cy+14, color.RGBA{0, 0, 0, 0x70}, text.AlignStart)
	}
	p.drawLockedDoors(a, screen, cx, cy)

	var sprites []sprite
	for _, st := range m.Stations() {
		st := st
		sprites = append(sprites, sprite{st.Y + st.H/2, func() { p.drawStation(a, screen, st, cx, cy) }})
	}
	feetY := p.s.Y + playerSize/2
	sprites = append(sprites,
		sprite{feetY, func() {
			bob := 0.0
			if p.moving {
				bob = -absSin(p.t*12) * 2
			}
			ui.DrawAvatar(screen, a.Avatars, p.s.Avatar, p.s.X-cx, feetY-cy+bob, 1, p.facing)
		}},
		sprite{p.petY + playerSize/2, func() {
			ui.DrawPet(screen, p.s.Avatar["pet"], p.petX-cx, p.petY+playerSize/2-cy, 1, p.t)
		}},
	)
	sort.SliceStable(sprites, func(i, j int) bool { return sprites[i].y < sprites[j].y })
	for _, s := range sprites {
		s.draw()
	}
	ui.DrawGizmo(screen, p.gizX-cx, p.gizY-cy, 0.9, p.t)

	if st := p.nearStation(a); st != nil {
		p.drawPrompt(a, screen, a.T(stationKey(st.ID)), st.X-cx, st.Y-st.H/2-cy-46)
	}

	if p.s.Clock.IsNight() {
		ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColNight)
	}
	p.drawHUD(a, screen)
}

func absSin(x float64) float64 {
	s := x - float64(int(x/3.14159265))*3.14159265
	if s < 0 {
		s = -s
	}
	// cheap |sin| via parabola, good enough for a bounce
	return 4 * s * (3.14159265 - s) / (3.14159265 * 3.14159265)
}

// drawFloor adds a subtle tile grid so rooms read as floors.
func drawFloor(screen *ebiten.Image, r world.Room, cx, cy float64) {
	line := color.RGBA{0, 0, 0, 0x0c}
	for x := r.Bounds.X + 40; x < r.Bounds.X+r.Bounds.W; x += 40 {
		ui.Fill(screen, geom.Rect{X: x, Y: r.Bounds.Y, W: 1, H: r.Bounds.H}, cx, cy, line)
	}
	for y := r.Bounds.Y + 40; y < r.Bounds.Y+r.Bounds.H; y += 40 {
		ui.Fill(screen, geom.Rect{X: r.Bounds.X, Y: y, W: r.Bounds.W, H: 1}, cx, cy, line)
	}
}

func (p *Play) drawLockedDoors(a *App, screen *ebiten.Image, cx, cy float64) {
	m := a.Map
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
}

// drawPrompt shows a speech-bubble label over something usable.
func (p *Play) drawPrompt(a *App, screen *ebiten.Image, label string, x, y float64) {
	face := ui.Bold(ui.SizeSmall)
	key := "E"
	if a.Pointer.TouchDevice() {
		key = a.T("hud.use")
	}
	s := key + ": " + label
	w := ui.Measure(s, face) + 20
	r := geom.Centered(x, y, w, 26)
	ui.RoundRect(screen, r, 12, ui.ColPanel, 2, ui.ColPanelLine)
	ui.Poly(screen, ui.ColPanelLine, x-6, r.Y+r.H, x+6, r.Y+r.H, x, r.Y+r.H+7)
	ui.TextCentered(screen, s, face, x, y, ui.ColText)
}

func (p *Play) drawStation(a *App, screen *ebiten.Image, st world.Station, cx, cy float64) {
	r := st.Rect()
	r.X -= cx
	r.Y -= cy
	x, _ := r.Center()
	switch st.Kind {
	case "npc":
		ui.DrawNova(screen, a.Avatars, x, r.Y+10, 1)
		ui.Desk(screen, r)
		ui.Circle(screen, x+30, r.Y+8, 7, ui.Hex("#4f8df0")) // little globe on the desk
		ui.Ring(screen, x+30, r.Y+8, 7, 1.5, ui.Hex("#2e3648"))
	case "shop":
		ui.DrawTom(screen, a.Avatars, x, r.Y+10, 1, p.t)
		ui.Desk(screen, r)
		for i, c := range []string{"#ffe27a", "#7ad0ff", "#ff7ac8"} {
			ui.RoundRect(screen, geom.Rect{X: r.X + 10 + float64(i)*12, Y: r.Y - 6, W: 9, H: 12}, 3, ui.Hex(c), 1, ui.ColWall)
		}
	case "wardrobe":
		ui.RoundRect(screen, r, 5, ui.Hex("#c08a5a"), 2, ui.Hex("#8d6040"))
		ui.RoundRect(screen, geom.Rect{X: r.X + 8, Y: r.Y + 6, W: r.W - 16, H: r.H - 14}, 10, ui.Hex("#cfeeff"), 2, ui.Hex("#8d6040"))
		ui.Line(screen, r.X+14, r.Y+14, r.X+22, r.Y+24, 2, color.RGBA{0xff, 0xff, 0xff, 0xc0})
	default:
		ui.Bench(screen, r)
		drawBenchProps(screen, st.ID, r, p.t)
	}
}

// drawBenchProps puts the right equipment on each experiment bench.
func drawBenchProps(screen *ebiten.Image, id string, r geom.Rect, t float64) {
	x, top := r.X+r.W/2, r.Y+4
	switch id {
	case "exp_slime":
		for i, c := range []string{"#7cf08c", "#ff7ac8", "#7ad0ff"} {
			bx := x - 24 + float64(i)*24
			ui.RoundRect(screen, geom.Rect{X: bx - 7, Y: top - 10, W: 14, H: 18}, 4, ui.Hex(c), 1.5, ui.ColWall)
		}
	case "exp_circuit":
		ui.RoundRect(screen, geom.Rect{X: x - 30, Y: top - 4, W: 20, H: 12}, 3, ui.Hex("#ffb347"), 1.5, ui.ColWall)
		ui.Line(screen, x-10, top+2, x+16, top+2, 2, ui.Hex("#d8423a"))
		ui.Circle(screen, x+22, top-2, 7, ui.Hex("#fff3a0"))
		ui.Ring(screen, x+22, top-2, 7, 1.5, ui.ColWall)
	case "exp_microscope":
		ui.RoundRect(screen, geom.Rect{X: x - 12, Y: top + 2, W: 24, H: 6}, 2, ui.Hex("#566078"), 0, nil)
		ui.Line(screen, x-2, top+2, x+4, top-18, 5, ui.Hex("#566078"))
		ui.Circle(screen, x+5, top-20, 4, ui.Hex("#2e3648"))
	case "exp_gravity":
		ui.Circle(screen, x-20, top-4, 10, ui.Hex("#d8842e"))
		ui.Circle(screen, x+8, top-2, 7, ui.Hex("#b9b9c8"))
		ui.Circle(screen, x+28, top, 5, ui.Hex("#4f8df0"))
	case "decorate":
		ui.Circle(screen, x-8, top, 9, ui.Hex("#f6e7c8"))
		for i, c := range []string{"#e8514a", "#3a6fd8", "#ffd23f"} {
			ui.Circle(screen, x-12+float64(i)*5, top-2+float64(i%2)*4, 2.4, ui.Hex(c))
		}
		ui.Line(screen, x+6, top+6, x+18, top-10, 3, ui.Hex("#8d6040"))
	case "workshop":
		for i := 0; i < 3; i++ {
			gx := x - 24 + float64(i)*24
			ui.Ring(screen, gx, top, 7, 4, ui.Hex("#9aa6b8"))
		}
	}
}

// stationKey is the text key for a station's label (checked by TestStationLabels).
func stationKey(id string) string { return "station." + id }
