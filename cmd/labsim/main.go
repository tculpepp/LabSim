// Command labsim is the LabSim game entry point (desktop and WASM).
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/scenes"
)

func main() {
	app, err := scenes.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1080, 720)
	ebiten.SetWindowTitle("LabSim")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
