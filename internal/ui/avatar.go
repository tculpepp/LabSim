package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"labsim/internal/avatar"
	"labsim/internal/geom"
	"labsim/internal/world"
)

// Hex parses "#rrggbb", falling back to magenta so mistakes are obvious.
func Hex(s string) color.RGBA {
	c, err := world.ParseHexColor(s)
	if err != nil {
		return color.RGBA{0xff, 0x00, 0xff, 0xff}
	}
	return RGB(c)
}

var (
	colOutline = color.RGBA{0x2e, 0x36, 0x48, 0xcc}
	colEye     = color.RGBA{0x22, 0x22, 0x2a, 0xff}
	colShadow  = color.RGBA{0, 0, 0, 0x40}
	colCheek   = color.RGBA{0xff, 0x8a, 0x9a, 0x70}
	colGold    = color.RGBA{0xff, 0xcf, 0x3a, 0xff}
)

func darker(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), c.A}
}

// Poly fills a polygon.
func Poly(dst *ebiten.Image, c color.Color, pts ...float64) {
	var p vector.Path
	p.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i+1 < len(pts); i += 2 {
		p.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, &p, nil, op)
}

// Arc strokes a circular arc (angles in radians, clockwise from +x).
func Arc(dst *ebiten.Image, x, y, r, from, to, width float64, c color.Color) {
	var p vector.Path
	p.Arc(float32(x), float32(y), float32(r), float32(from), float32(to), vector.Clockwise)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: float32(width), LineCap: vector.LineCapRound}, op)
}

// Line strokes a straight line.
func Line(dst *ebiten.Image, x0, y0, x1, y1, width float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(width), c, true)
}

// Shadow draws a soft oval shadow on the ground at (x, y).
func Shadow(dst *ebiten.Image, x, y, w float64) {
	RoundRect(dst, geom.Centered(x, y, w, w*0.25), w*0.12, colShadow, 0, nil)
}

// DrawAvatar draws a scientist with feet at (x, y). u is the size unit
// (1 in the world, ~4 in the creator). facing is -1 (left) or 1 (right).
func DrawAvatar(dst *ebiten.Image, cat *avatar.Catalog, av avatar.Avatar, x, y, u, facing float64) {
	col := func(part string, def string) color.RGBA {
		if o, ok := cat.Option(part, av[part]); ok && o.Color != "" {
			return Hex(o.Color)
		}
		return Hex(def)
	}
	skin, hairC := col("skin", "#f1c7a0"), col("hairColor", "#7a4a2a")
	coat, shirt, shoes := col("coat", "#ffffff"), col("shirt", "#3fb8af"), col("shoes", "#333333")

	bw, bh := 18.0, 16.0
	switch av["body"] {
	case "tall":
		bw, bh = 16, 21
	case "small":
		bw, bh = 15, 13
	case "strong":
		bw, bh = 22, 17
	}
	Shadow(dst, x, y, 22*u)
	bodyTop := y - (4+bh)*u
	hx, hy, r := x, bodyTop-7*u, 9*u

	// Gear worn on the back goes behind everything.
	switch av["gear"] {
	case "backpack":
		RoundRect(dst, geom.Rect{X: x - facing*bw/2*u - 5*u, Y: bodyTop + 2*u, W: 10 * u, H: bh * 0.75 * u}, 3*u, Hex("#a0673a"), u, colOutline)
	case "jetpack":
		for _, dx := range []float64{-4, 4} {
			cx := x - facing*(bw/2-1)*u + dx*u
			RoundRect(dst, geom.Rect{X: cx - 3*u, Y: bodyTop, W: 6 * u, H: bh * 0.8 * u}, 3*u, Hex("#9aa6b8"), u, colOutline)
			Poly(dst, Hex("#ff9a3c"), cx-2*u, bodyTop+bh*0.8*u, cx+2*u, bodyTop+bh*0.8*u, cx, bodyTop+(bh*0.8+5)*u)
		}
	}
	// Hair that sits behind the head.
	switch av["hair"] {
	case "long":
		RoundRect(dst, geom.Rect{X: hx - 10.5*u, Y: hy - 10*u, W: 21 * u, H: 22 * u}, 8*u, hairC, 0, nil)
	case "curly":
		for a := 150.0; a <= 390; a += 30 {
			rad := a * math.Pi / 180
			Circle(dst, hx+math.Cos(rad)*8.5*u, hy+math.Sin(rad)*8.5*u, 4*u, hairC)
		}
	case "ponytail":
		Circle(dst, hx-facing*10*u, hy-1*u, 4*u, hairC)
		Circle(dst, hx-facing*12*u, hy+4*u, 3.2*u, hairC)
	case "bun":
		Circle(dst, hx, hy-11*u, 4.5*u, hairC)
	}
	if av["hair"] != "bald" && av["hair"] != "" {
		Circle(dst, hx, hy-1.5*u, r*1.04, hairC)
	}

	// Shoes, body (lab coat), shirt, pattern, hands.
	Circle(dst, x-4.5*u, y-2*u, 3.2*u, shoes)
	Circle(dst, x+4.5*u, y-2*u, 3.2*u, shoes)
	body := geom.Rect{X: x - bw/2*u, Y: bodyTop, W: bw * u, H: bh * u}
	RoundRect(dst, body, 6*u, coat, u, colOutline)
	RoundRect(dst, geom.Rect{X: x - 3*u, Y: bodyTop, W: 6 * u, H: 7 * u}, 2*u, shirt, 0, nil)
	Line(dst, x, bodyTop+7*u, x, bodyTop+bh*u-u, 0.8*u, colOutline)
	switch av["pattern"] {
	case "stripes":
		for _, f := range []float64{0.55, 0.78} {
			Fill(dst, geom.Rect{X: body.X + u, Y: bodyTop + bh*f*u, W: body.W - 2*u, H: 1.6 * u}, 0, 0, shirt)
		}
	case "dots":
		for _, d := range [][2]float64{{-5, 0.5}, {5, 0.45}, {-4, 0.8}, {5, 0.78}} {
			Circle(dst, x+d[0]*u, bodyTop+bh*d[1]*u, 1.4*u, shirt)
		}
	case "stars":
		Diamond(dst, x-5*u, bodyTop+bh*0.55*u, 2.4*u, colGold)
		Diamond(dst, x+5*u, bodyTop+bh*0.7*u, 2.4*u, colGold)
	}
	hand := skin
	if av["gear"] == "gloves" {
		hand = Hex("#7a5cff")
	}
	Circle(dst, x-(bw/2+1.5)*u, bodyTop+bh*0.6*u, 2.8*u, hand)
	Circle(dst, x+(bw/2+1.5)*u, bodyTop+bh*0.6*u, 2.8*u, hand)

	// Head and face.
	Circle(dst, hx, hy, r, skin)
	Ring(dst, hx, hy, r, 0.8*u, colOutline)
	ex := facing * 1.2 * u
	Circle(dst, hx-3.3*u+ex, hy+0.5*u, 1.3*u, colEye)
	Circle(dst, hx+3.3*u+ex, hy+0.5*u, 1.3*u, colEye)
	Circle(dst, hx-5.5*u+ex, hy+3*u, 1.4*u, colCheek)
	Circle(dst, hx+5.5*u+ex, hy+3*u, 1.4*u, colCheek)
	Arc(dst, hx+ex, hy+2.5*u, 2.3*u, 0.15*math.Pi, 0.85*math.Pi, 0.9*u, colEye)

	// Hair on top of the head (fringe).
	switch av["hair"] {
	case "", "bald":
	case "spiky":
		for _, dx := range []float64{-6, -2, 2, 6} {
			Poly(dst, hairC, hx+(dx-2.5)*u, hy-6*u, hx+(dx+2.5)*u, hy-6*u, hx+dx*u, hy-13*u)
		}
		RoundRect(dst, geom.Rect{X: hx - 8*u, Y: hy - 9.5*u, W: 16 * u, H: 4.5 * u}, 2*u, hairC, 0, nil)
	case "curly":
		for _, dx := range []float64{-5, 0, 5} {
			Circle(dst, hx+dx*u, hy-7*u, 3.2*u, hairC)
		}
	default:
		RoundRect(dst, geom.Rect{X: hx - 8*u, Y: hy - 9.5*u, W: 16 * u, H: 5 * u}, 2.5*u, hairC, 0, nil)
	}

	// Eyewear.
	switch av["eyewear"] {
	case "goggles":
		Fill(dst, geom.Rect{X: hx - r, Y: hy - 0.8*u, W: 2 * r, H: 1.8 * u}, 0, 0, Hex("#566078"))
		for _, dx := range []float64{-3.3, 3.3} {
			Circle(dst, hx+dx*u+ex, hy+0.5*u, 2.9*u, color.RGBA{0xa8, 0xe4, 0xff, 0xb0})
			Ring(dst, hx+dx*u+ex, hy+0.5*u, 2.9*u, 0.9*u, Hex("#566078"))
		}
	case "glasses":
		for _, dx := range []float64{-3.3, 3.3} {
			Ring(dst, hx+dx*u+ex, hy+0.5*u, 2.4*u, 0.6*u, colEye)
		}
		Line(dst, hx-0.9*u+ex, hy+0.3*u, hx+0.9*u+ex, hy+0.3*u, 0.6*u, colEye)
	case "sunglasses":
		for _, dx := range []float64{-3.3, 3.3} {
			Circle(dst, hx+dx*u+ex, hy+0.5*u, 2.6*u, colEye)
		}
		Line(dst, hx-1*u+ex, hy+0.3*u, hx+1*u+ex, hy+0.3*u, 0.7*u, colEye)
	}

	// Hats.
	switch av["hat"] {
	case "beanie":
		c := Hex("#e8514a")
		RoundRect(dst, geom.Rect{X: hx - 9.5*u, Y: hy - 13*u, W: 19 * u, H: 9 * u}, 5*u, c, 0, nil)
		Fill(dst, geom.Rect{X: hx - 9.5*u, Y: hy - 6.5*u, W: 19 * u, H: 2.5 * u}, 0, 0, darker(c, 0.8))
		Circle(dst, hx, hy-13.5*u, 2.5*u, Hex("#ffffff"))
	case "cap":
		c := Hex("#3a6fd8")
		RoundRect(dst, geom.Rect{X: hx - 8.5*u, Y: hy - 13*u, W: 17 * u, H: 7 * u}, 4*u, c, 0, nil)
		RoundRect(dst, geom.Rect{X: hx + facing*4*u - 7*u, Y: hy - 7*u, W: 14 * u, H: 2.2 * u}, u, darker(c, 0.8), 0, nil)
	case "party":
		Poly(dst, Hex("#ff7ac8"), hx-6*u, hy-8*u, hx+6*u, hy-8*u, hx, hy-22*u)
		Circle(dst, hx, hy-22*u, 2*u, colGold)
	case "tophat":
		RoundRect(dst, geom.Rect{X: hx - 10*u, Y: hy - 10.5*u, W: 20 * u, H: 2.8 * u}, u, colEye, 0, nil)
		RoundRect(dst, geom.Rect{X: hx - 6*u, Y: hy - 23*u, W: 12 * u, H: 13 * u}, u, colEye, 0, nil)
		Fill(dst, geom.Rect{X: hx - 6*u, Y: hy - 13.5*u, W: 12 * u, H: 2 * u}, 0, 0, Hex("#e8514a"))
	case "crown":
		Poly(dst, colGold, hx-7*u, hy-8*u, hx+7*u, hy-8*u, hx+7*u, hy-16*u, hx+3.5*u, hy-12*u,
			hx, hy-17*u, hx-3.5*u, hy-12*u, hx-7*u, hy-16*u)
		Circle(dst, hx, hy-10*u, 1.2*u, Hex("#e8514a"))
	}
}

// DrawPet draws a pet sitting at (x, y). t is time in seconds (for idle bounce).
func DrawPet(dst *ebiten.Image, id string, x, y, u, t float64) {
	Shadow(dst, x, y, 14*u)
	y -= math.Abs(math.Sin(t*5)) * 1.5 * u
	eyes := func(cx, cy, gap, size float64) {
		Circle(dst, cx-gap*u, cy, size*u, colEye)
		Circle(dst, cx+gap*u, cy, size*u, colEye)
	}
	switch id {
	case "robot":
		RoundRect(dst, geom.Rect{X: x - 6*u, Y: y - 13*u, W: 12 * u, H: 11 * u}, 3*u, Hex("#b8c4d6"), 0.8*u, colOutline)
		RoundRect(dst, geom.Rect{X: x - 4.5*u, Y: y - 11.5*u, W: 9 * u, H: 5 * u}, 1.5*u, Hex("#2e3648"), 0, nil)
		Circle(dst, x-2*u, y-9*u, 1*u, ColCrystal)
		Circle(dst, x+2*u, y-9*u, 1*u, ColCrystal)
		Line(dst, x, y-13*u, x, y-16*u, 0.8*u, colOutline)
		Circle(dst, x, y-16.5*u, 1.4*u, Hex("#e05a5a"))
		Circle(dst, x-4*u, y-1.5*u, 1.8*u, colOutline)
		Circle(dst, x+4*u, y-1.5*u, 1.8*u, colOutline)
	case "cat":
		c := Hex("#f0a04b")
		Poly(dst, c, x-6*u, y-9*u, x-2*u, y-12*u, x-5.5*u, y-15*u)
		Poly(dst, c, x+6*u, y-9*u, x+2*u, y-12*u, x+5.5*u, y-15*u)
		Circle(dst, x+7*u, y-3*u, 2*u, c)
		Circle(dst, x, y-7*u, 6.5*u, c)
		eyes(x, y-8*u, 2.4, 1)
		Circle(dst, x, y-5.8*u, 0.8*u, Hex("#e05a8a"))
	case "slime":
		c := Hex("#7cf08c")
		RoundRect(dst, geom.Rect{X: x - 7.5*u, Y: y - 8*u, W: 15 * u, H: 8 * u}, 4*u, c, 0, nil)
		Circle(dst, x, y-8*u, 5.5*u, c)
		Circle(dst, x-2.5*u, y-11*u, 1.4*u, color.RGBA{0xff, 0xff, 0xff, 0xb0})
		eyes(x, y-6*u, 2.4, 1.1)
	case "puppy":
		c := Hex("#c89a6a")
		Circle(dst, x, y-7*u, 6.5*u, c)
		Circle(dst, x-6*u, y-6*u, 2.6*u, darker(c, 0.7))
		Circle(dst, x+6*u, y-6*u, 2.6*u, darker(c, 0.7))
		eyes(x, y-8*u, 2.4, 1)
		Circle(dst, x, y-5.5*u, 1.2*u, colEye)
	case "owl":
		c := Hex("#a8784a")
		Poly(dst, c, x-6*u, y-11*u, x-3*u, y-12*u, x-6*u, y-16*u)
		Poly(dst, c, x+6*u, y-11*u, x+3*u, y-12*u, x+6*u, y-16*u)
		Circle(dst, x, y-7.5*u, 7*u, c)
		for _, dx := range []float64{-2.8, 2.8} {
			Circle(dst, x+dx*u, y-9*u, 2.6*u, Hex("#ffffff"))
			Circle(dst, x+dx*u, y-9*u, 1.3*u, colEye)
		}
		Poly(dst, Hex("#ff9a3c"), x-1.2*u, y-6.5*u, x+1.2*u, y-6.5*u, x, y-4.5*u)
	case "dragon":
		c := Hex("#5cc878")
		Poly(dst, Hex("#b38cff"), x-5*u, y-8*u, x-12*u, y-14*u, x-10*u, y-5*u)
		Poly(dst, Hex("#b38cff"), x+5*u, y-8*u, x+12*u, y-14*u, x+10*u, y-5*u)
		Poly(dst, colGold, x-4*u, y-11*u, x-2*u, y-12*u, x-4*u, y-16*u)
		Poly(dst, colGold, x+4*u, y-11*u, x+2*u, y-12*u, x+4*u, y-16*u)
		Circle(dst, x, y-7*u, 6.5*u, c)
		Circle(dst, x, y-4.5*u, 3*u, Hex("#c8f0a0"))
		eyes(x, y-8.5*u, 2.4, 1.1)
	}
}

// DrawGizmo draws Gizmo, the round floating helper robot with a screen face (C-3).
func DrawGizmo(dst *ebiten.Image, x, y, u, t float64) {
	float := math.Sin(t*2.5) * 2 * u
	Shadow(dst, x, y+14*u, 12*u)
	cy := y + float
	Circle(dst, x, cy, 9*u, Hex("#9ad8ff"))
	Ring(dst, x, cy, 9*u, u, colOutline)
	RoundRect(dst, geom.Rect{X: x - 6*u, Y: cy - 4*u, W: 12 * u, H: 7.5 * u}, 2*u, Hex("#2e3648"), 0, nil)
	blink := math.Mod(t, 4) < 0.15
	if blink {
		Line(dst, x-3.5*u, cy, x-1.5*u, cy, 0.8*u, ColCrystal)
		Line(dst, x+1.5*u, cy, x+3.5*u, cy, 0.8*u, ColCrystal)
	} else {
		Circle(dst, x-2.5*u, cy-0.3*u, 1.2*u, ColCrystal)
		Circle(dst, x+2.5*u, cy-0.3*u, 1.2*u, ColCrystal)
	}
	Line(dst, x, cy-9*u, x, cy-12*u, 0.8*u, colOutline)
	Circle(dst, x, cy-12.5*u, 1.6*u, Hex("#e05a5a"))
}
