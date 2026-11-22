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
	const bw, bh = 16.0, 21.0 // "tall" body
	hy := y - (4+bh)*u - 7*u
	hair := Hex("#7a4a2a")
	// Lopsided hair: an extra tuft sticking up on one side.
	Circle(dst, x+5*u, hy-9*u, 3.4*u, hair)
	// Short beard around the chin, smile still showing.
	Arc(dst, x, hy+1.5*u, 7.2*u, 0.1*math.Pi, 0.9*math.Pi, 2.4*u, hair)
	// Blue eyes.
	Circle(dst, x-3.3*u, hy+0.5*u, 1.3*u, Hex("#3a7bd5"))
	Circle(dst, x+3.3*u, hy+0.5*u, 1.3*u, Hex("#3a7bd5"))
	Circle(dst, x-3.3*u, hy+0.5*u, 0.6*u, colEye)
	Circle(dst, x+3.3*u, hy+0.5*u, 0.6*u, colEye)
	// Right-handed: holds a crystal in his right hand (our left, he faces us).
	hx, hyHand := x-(bw/2+1.5)*u, y-(4+bh)*u+bh*0.6*u
	Diamond(dst, hx, hyHand-4*u-math.Sin(t*2)*u, 3.5*u, ColCrystal)
	Circle(dst, hx, hyHand, 2.8*u, Hex("#fbe0cc"))
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
