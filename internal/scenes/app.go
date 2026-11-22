// Package scenes wires game state, logic packages, and UI into Ebitengine scenes.
package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/data"
	"labsim/internal/content"
	"labsim/internal/geom"
	"labsim/internal/stack"
	"labsim/internal/ui"
	"labsim/internal/world"
)

// Scene is one screen or overlay. Only the top scene updates; Visible scenes draw.
type Scene interface {
	Update(a *App) error
	Draw(a *App, screen *ebiten.Image)
	Opaque() bool
}

// App is the ebiten.Game implementation and owns shared resources.
type App struct {
	W, H       int // logical screen size
	outW, outH int // outside (window/CSS) size
	Map        *world.Map
	Text       *content.Catalog
	Pointer    ui.Pointer
	Level      content.Level
	PlayerName string

	scenes stack.Stack[Scene]
}

func NewApp() (*App, error) {
	m, err := world.Load(data.Map)
	if err != nil {
		return nil, fmt.Errorf("load map: %w", err)
	}
	txt, err := content.Load(data.Text)
	if err != nil {
		return nil, fmt.Errorf("load text: %w", err)
	}
	a := &App{W: 720, H: geom.LogicalH, Map: m, Text: txt, Level: content.Junior}
	a.scenes.Push(NewTitle(a))
	return a, nil
}

// T returns player-facing text for the current difficulty.
func (a *App) T(key string) string { return a.Text.Get(key, a.Level) }

// Tf formats player-facing text.
func (a *App) Tf(key string, args ...any) string { return fmt.Sprintf(a.T(key), args...) }

func (a *App) Push(s Scene)    { a.scenes.Push(s) }
func (a *App) Pop()            { a.scenes.Pop() }
func (a *App) Replace(s Scene) { a.scenes.Replace(s) }

// Say opens a dialog from a character.
func (a *App) Say(who, nameKey, text string) {
	a.Push(NewDialog(a, who, a.T(nameKey), text))
}

func (a *App) Update() error {
	a.Pointer.Update()
	if top, ok := a.scenes.Top(); ok {
		return top.Update(a)
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	for _, s := range a.scenes.Visible() {
		s.Draw(a, screen)
	}
	if a.Pointer.TouchDevice() && a.outW < a.outH {
		ui.Fill(screen, geom.Rect{W: float64(a.W), H: float64(a.H)}, 0, 0, ui.ColWall)
		ui.TextCentered(screen, a.T("rotate.hint"), ui.Bold(ui.SizeHeading), float64(a.W)/2, float64(a.H)/2, ui.ColTextLight)
	}
}

func (a *App) Layout(outW, outH int) (int, int) {
	a.outW, a.outH = outW, outH
	a.W, a.H = geom.LogicalSize(outW, outH)
	return a.W, a.H
}

// center returns the screen center.
func (a *App) center() (float64, float64) { return float64(a.W) / 2, float64(a.H) / 2 }
