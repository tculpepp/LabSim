// Package scenes wires game state, logic packages, and UI into Ebitengine scenes.
package scenes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/data"
	"labsim/internal/geom"
	"labsim/internal/ui"
	"labsim/internal/world"
)

// App is the ebiten.Game implementation.
type App struct {
	W, H    int // logical screen size
	Map     *world.Map
	Pointer ui.Pointer

	play *Play
}

func NewApp() (*App, error) {
	m, err := world.Load(data.Map)
	if err != nil {
		return nil, fmt.Errorf("load map: %w", err)
	}
	a := &App{W: 720, H: geom.LogicalH, Map: m}
	a.play = NewPlay(a)
	return a, nil
}

func (a *App) Update() error {
	a.Pointer.Update()
	return a.play.Update(a)
}

func (a *App) Draw(screen *ebiten.Image) { a.play.Draw(a, screen) }

func (a *App) Layout(outW, outH int) (int, int) {
	a.W, a.H = geom.LogicalSize(outW, outH)
	return a.W, a.H
}
