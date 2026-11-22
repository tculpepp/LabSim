// Command labsim is the LabSim game entry point (desktop and WASM).
package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"labsim/internal/clock"
	"labsim/internal/economy"
	"labsim/internal/inventory"
)

const (
	screenW     = 960
	screenH     = 640
	playerSpeed = 3.0
	playerSize  = 24
	dayLength   = 10 * time.Minute
	tick        = time.Second / 60
)

// room is a placeholder rectangle until real art/tilemaps arrive.
type room struct {
	name       string
	x, y, w, h float32
	col        color.RGBA
	locked     bool
}

var rooms = []room{
	{"Main Hall", 360, 240, 240, 160, color.RGBA{0xf4, 0xe1, 0xc1, 0xff}, false},
	{"Dr. Nova's Office", 360, 40, 240, 160, color.RGBA{0xc9, 0xd6, 0xf2, 0xff}, false},
	{"Shop", 640, 40, 200, 160, color.RGBA{0xf9, 0xd2, 0xe0, 0xff}, false},
	{"Chemistry Lab", 80, 40, 240, 160, color.RGBA{0xc8, 0xf0, 0xc8, 0xff}, false},
	{"Physics Lab", 80, 240, 240, 160, color.RGBA{0xff, 0xe8, 0xa8, 0xff}, false},
	{"Greenhouse", 80, 440, 240, 160, color.RGBA{0xa8, 0xe6, 0xb0, 0xff}, false},
	{"Space Lab", 640, 240, 200, 160, color.RGBA{0xd2, 0xc4, 0xf5, 0xff}, false},
	{"Storage Room", 640, 440, 200, 160, color.RGBA{0xe0, 0xd4, 0xc4, 0xff}, false},
	{"My Lab", 360, 440, 240, 160, color.RGBA{0xff, 0xd6, 0xb0, 0xff}, false},
	{"Robot Workshop", 860, 40, 80, 560, color.RGBA{0xb8, 0xb8, 0xb8, 0xff}, true},
}

type Game struct {
	px, py   float64
	clock    *clock.Clock
	wallet   *economy.Wallet
	backpack *inventory.Backpack
}

func newGame() *Game {
	return &Game{
		px:       480,
		py:       320,
		clock:    clock.New(dayLength),
		wallet:   economy.NewWallet(0),
		backpack: inventory.NewBackpack(6),
	}
}

func (g *Game) Update() error {
	g.clock.Advance(tick)

	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy++
	}
	// Touch: walk toward the finger. A proper on-screen joystick comes later (T-5).
	if ids := ebiten.AppendTouchIDs(nil); len(ids) > 0 {
		tx, ty := ebiten.TouchPosition(ids[0])
		dx, dy = float64(tx)-g.px, float64(ty)-g.py
		if math.Hypot(dx, dy) < playerSpeed {
			dx, dy = 0, 0
		}
	}
	if l := math.Hypot(dx, dy); l > 0 {
		g.px += dx / l * playerSpeed
		g.py += dy / l * playerSpeed
	}
	g.px = clamp(g.px, playerSize/2, screenW-playerSize/2)
	g.py = clamp(g.py, playerSize/2, screenH-playerSize/2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x5a, 0x6a, 0x7a, 0xff}) // hallway floor
	for _, r := range rooms {
		vector.FillRect(screen, r.x, r.y, r.w, r.h, r.col, false)
		label := r.name
		if r.locked {
			label += "\n(locked)"
		}
		ebitenutil.DebugPrintAt(screen, label, int(r.x)+6, int(r.y)+6)
	}

	// Player placeholder: round scientist.
	vector.FillCircle(screen, float32(g.px), float32(g.py), playerSize/2, color.RGBA{0xff, 0xff, 0xff, 0xff}, true)
	vector.StrokeCircle(screen, float32(g.px), float32(g.py), playerSize/2, 2, color.RGBA{0x30, 0x60, 0xd0, 0xff}, true)

	if g.clock.IsNight() {
		vector.FillRect(screen, 0, 0, screenW, screenH, color.RGBA{0x10, 0x10, 0x40, 0x70}, false)
	}

	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Day %d   Crystals: %d   Backpack: %d/%d",
			g.clock.Day(), g.wallet.Crystals(), g.backpack.UsedSlots(), g.backpack.Slots()),
		8, screenH-20)
}

func (g *Game) Layout(_, _ int) (int, int) { return screenW, screenH }

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("LabSim")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
