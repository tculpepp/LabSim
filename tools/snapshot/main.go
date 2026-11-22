// Command snapshot runs the game for a few frames and saves a PNG of the screen.
// Dev tool for checking rendering without playing: go run ./tools/snapshot out.png
package main

import (
	"errors"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/scenes"
)

const frames = 30

type snap struct {
	*scenes.App
	n    int
	out  string
	done bool
}

var errDone = errors.New("done")

func (s *snap) Update() error {
	if s.done {
		return errDone
	}
	s.n++
	return s.App.Update()
}

func (s *snap) Draw(screen *ebiten.Image) {
	s.App.Draw(screen)
	if s.n < frames || s.done {
		return
	}
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	f, err := os.Create(s.out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	s.done = true
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: snapshot <out.png>")
	}
	app, err := scenes.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1080, 720)
	if err := ebiten.RunGame(&snap{App: app, out: os.Args[1]}); err != nil && !errors.Is(err, errDone) {
		log.Fatal(err)
	}
}
