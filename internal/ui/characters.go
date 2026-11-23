package ui

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"labsim/internal/avatar"
	"labsim/internal/geom"
)

// Dr. Nova: loves space and always smiles (C-1).
var novaLook = avatar.Avatar{
	"body": "round", "skin": "s4", "hair": "bun", "hairColor": "black", "coat": "white",
	"shirt": "navy", "shoes": "black", "pattern": "stars", "eyewear": "glasses", "hat": "none", "gear": "none",
}

// Tom: tall, light skin, blue eyes, brown hair a little lopsided, short beard, right-handed (C-2).
var tomLook = avatar.Avatar{
	"body": "tall", "skin": "s1", "hair": "short", "hairColor": "brown", "coat": "blue",
	"shirt": "orange", "shoes": "black", "pattern": "none", "eyewear": "none", "hat": "none", "gear": "none",
}

func DrawNova(dst *ebiten.Image, cat *avatar.Catalog, x, y, u float64) {
	DrawAvatar(dst, cat, novaLook, x, y, u, 0)
}

func DrawTom(dst *ebiten.Image, cat *avatar.Catalog, x, y, u, t float64) {
	DrawAvatar(dst, cat, tomLook, x, y, u, 0)
	f := layoutFigure(tomLook["body"], x, y, u)
	hx, hy := f.hx, f.hy
	hair := Hex("#7a4a2a")
	// Lopsided hair: an extra tuft sticking up on one side.
	Circle(dst, hx+4*u, f.top-2*u, 2.8*u, hair)
	// Short beard around the chin, smile still showing.
	Arc(dst, hx, hy+0.5*u, 4.6*u, 0.1*math.Pi, 0.9*math.Pi, 2*u, hair)
	// Blue eyes.
	for _, dx := range []float64{-2.6, 2.6} {
		Circle(dst, hx+dx*u, hy+0.3*u, 1.1*u, Hex("#3a7bd5"))
		Circle(dst, hx+dx*u, hy+0.3*u, 0.5*u, colEye)
	}
	// Right-handed: holds a crystal in his right hand (our left, he faces us).
	h := f.handL
	Diamond(dst, h[0], h[1]-4.5*u-math.Sin(t*2)*u, 3.5*u, ColCrystal)
}

// Furniture -------------------------------------------------------------

var (
	colWood     = Hex("#c08a5a")
	colWoodDark = Hex("#8d6040")
	colBench    = Hex("#dfe6ee")
	colBenchTop = Hex("#f6f9fc")
)

// Desk draws a wooden desk or counter occupying r.
func Desk(dst *ebiten.Image, r geom.Rect) {
	RoundRect(dst, r, 5, colWood, 2, colWoodDark)
	Fill(dst, geom.Rect{X: r.X + 4, Y: r.Y + r.H*0.45, W: r.W - 8, H: 2}, 0, 0, colWoodDark)
}

// Bench draws a white lab bench occupying r.
func Bench(dst *ebiten.Image, r geom.Rect) {
	RoundRect(dst, r, 5, colBench, 2, colOutline)
	RoundRect(dst, geom.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H * 0.45}, 5, colBenchTop, 0, nil)
}
