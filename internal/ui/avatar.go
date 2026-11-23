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

// figure is the layout of a minifig-style scientist: where the head and the
// hands end up, so NPCs can add details (a beard, a held crystal) on top.
type figure struct {
	hx, hy     float64 // head center
	hw, hh     float64 // head size
	top        float64 // top of the head
	torsoTop   float64
	shoulderW  float64 // torso width at the shoulders
	hipW       float64 // torso width at the hips
	torsoH     float64
	legH       float64
	hipH       float64
	handL      [2]float64
	handR      [2]float64
	shoulderLY float64
}

// layoutFigure places a minifig with feet at (x, y). The body type changes
// torso width and the length of torso and legs.
func layoutFigure(body string, x, y, u float64) figure {
	tw, th, lh := 17.0, 10.5, 7.0
	switch body {
	case "tall":
		tw, th, lh = 15, 12, 9.5
	case "small":
		tw, th, lh = 15, 9, 5
	case "strong":
		tw, th, lh = 21, 11, 7
	}
	f := figure{hw: 13 * u, hh: 11 * u, hipW: tw * u, shoulderW: (tw - 4) * u, torsoH: th * u, legH: lh * u, hipH: 2.5 * u}
	f.torsoTop = y - f.legH - f.hipH - f.torsoH
	headBottom := f.torsoTop - 1.5*u
	f.hx, f.hy = x, headBottom-f.hh/2
	f.top = headBottom - f.hh
	f.shoulderLY = f.torsoTop + 2.5*u
	handY := f.torsoTop + f.torsoH*0.95
	f.handL = [2]float64{x - f.hipW/2 - 2.5*u, handY}
	f.handR = [2]float64{x + f.hipW/2 + 2.5*u, handY}
	return f
}

var colPants = Hex("#4a5878")

// cHand draws a minifig "C" hand: a ring with a gap facing forward-down.
func cHand(dst *ebiten.Image, x, y, u float64, c color.RGBA) {
	from, to := 0.5*math.Pi+0.7, 2.5*math.Pi-0.7
	Arc(dst, x, y, 1.7*u, from, to, 2.6*u, colOutline)
	Arc(dst, x, y, 1.7*u, from, to, 1.7*u, c)
}

// DrawAvatar draws a minifig-style scientist with feet at (x, y). u is the
// size unit (1 in the world, ~4 in the creator). facing is -1 (left), 1
// (right), or 0 (facing the camera).
func DrawAvatar(dst *ebiten.Image, cat *avatar.Catalog, av avatar.Avatar, x, y, u, facing float64) {
	col := func(part string, def string) color.RGBA {
		if o, ok := cat.Option(part, av[part]); ok && o.Color != "" {
			return Hex(o.Color)
		}
		return Hex(def)
	}
	skin, hairC := col("skin", "#f1c7a0"), col("hairColor", "#7a4a2a")
	coat, shirt, shoes := col("coat", "#ffffff"), col("shirt", "#3fb8af"), col("shoes", "#333333")

	f := layoutFigure(av["body"], x, y, u)
	hx, hy, T := f.hx, f.hy, f.top
	Shadow(dst, x, y, f.hipW+8*u)

	// Gear worn on the back goes behind everything.
	switch av["gear"] {
	case "backpack":
		RoundRect(dst, geom.Rect{X: x - facing*f.hipW/2 - 5*u, Y: f.torsoTop + u, W: 10 * u, H: f.torsoH * 0.9}, 3*u, Hex("#a0673a"), u, colOutline)
	case "jetpack":
		for _, dx := range []float64{-4, 4} {
			cx := x - facing*(f.hipW/2-2*u) + dx*u
			RoundRect(dst, geom.Rect{X: cx - 3*u, Y: f.torsoTop - u, W: 6 * u, H: f.torsoH}, 3*u, Hex("#9aa6b8"), u, colOutline)
			Poly(dst, Hex("#ff9a3c"), cx-2*u, f.torsoTop+f.torsoH-u, cx+2*u, f.torsoTop+f.torsoH-u, cx, f.torsoTop+f.torsoH+4*u)
		}
	}
	// Hair that sits behind the head.
	switch av["hair"] {
	case "long":
		RoundRect(dst, geom.Rect{X: hx - 8*u, Y: T - 1.5*u, W: 16 * u, H: f.hh + 6*u}, 3*u, hairC, 0.8*u, colOutline)
	case "curly":
		for a := 150.0; a <= 390; a += 30 {
			rad := a * math.Pi / 180
			Circle(dst, hx+math.Cos(rad)*7*u, hy+math.Sin(rad)*6.5*u, 3.4*u, hairC)
		}
	case "ponytail":
		Circle(dst, hx-facing*9*u, hy-1*u, 3.5*u, hairC)
		Circle(dst, hx-facing*10.5*u, hy+3.5*u, 2.8*u, hairC)
	case "bun":
		Circle(dst, hx, T-2.8*u, 3.8*u, hairC)
	}

	// Legs and hips: two blocky legs joined at the hip, shoes at the bottom.
	hipY := y - f.legH - f.hipH
	RoundRect(dst, geom.Rect{X: x - f.hipW/2, Y: hipY, W: f.hipW, H: f.hipH + u}, 0.8*u, colPants, 0.8*u, colOutline)
	legW := f.hipW/2 - 0.6*u
	for _, lx := range []float64{x - f.hipW/2, x + 0.6*u} {
		RoundRect(dst, geom.Rect{X: lx, Y: hipY + f.hipH, W: legW, H: f.legH}, 1.2*u, colPants, 0.8*u, colOutline)
		RoundRect(dst, geom.Rect{X: lx, Y: y - 2.8*u, W: legW, H: 2.8 * u}, 1.2*u, shoes, 0.8*u, colOutline)
	}

	// Arms join at the shoulders and angle out to C-shaped hands.
	hand := skin
	if av["gear"] == "gloves" {
		hand = Hex("#7a5cff")
	}
	for _, side := range []float64{-1, 1} {
		sx := x + side*(f.shoulderW/2-0.5*u)
		h := f.handR
		if side < 0 {
			h = f.handL
		}
		ex, ey := h[0]-side*0.3*u, h[1]-2*u
		Line(dst, sx, f.shoulderLY, ex, ey, 5*u, colOutline)
		Line(dst, sx, f.shoulderLY, ex, ey, 3.6*u, coat)
		Circle(dst, sx, f.shoulderLY, 2.5*u, colOutline)
		Circle(dst, sx, f.shoulderLY, 1.8*u, coat)
		cHand(dst, h[0], h[1], u, hand)
	}

	// Torso: a trapezoid, wider at the hips. Outline first, then the coat.
	tt, tb := f.torsoTop, f.torsoTop+f.torsoH
	o := 0.8 * u
	Poly(dst, colOutline, x-f.shoulderW/2-o, tt-o, x+f.shoulderW/2+o, tt-o, x+f.hipW/2+o, tb+o, x-f.hipW/2-o, tb+o)
	Poly(dst, coat, x-f.shoulderW/2, tt, x+f.shoulderW/2, tt, x+f.hipW/2, tb, x-f.hipW/2, tb)
	Poly(dst, shirt, x-3*u, tt, x+3*u, tt, x, tt+5.5*u)
	Line(dst, x, tt+5.5*u, x, tb, 0.8*u, colOutline)
	switch av["pattern"] {
	case "stripes":
		for _, fr := range []float64{0.55, 0.8} {
			w := f.shoulderW + (f.hipW-f.shoulderW)*fr - 2*u
			Fill(dst, geom.Rect{X: x - w/2, Y: tt + f.torsoH*fr, W: w, H: 1.4 * u}, 0, 0, shirt)
		}
	case "dots":
		for _, d := range [][2]float64{{-4, 0.4}, {4, 0.35}, {-4.5, 0.75}, {4.5, 0.72}} {
			Circle(dst, x+d[0]*u, tt+f.torsoH*d[1], 1.3*u, shirt)
		}
	case "stars":
		Diamond(dst, x-4*u, tt+f.torsoH*0.5, 2.2*u, colGold)
		Diamond(dst, x+4.5*u, tt+f.torsoH*0.72, 2.2*u, colGold)
	}

	// Neck, then the cylinder head with a stud on top.
	RoundRect(dst, geom.Rect{X: hx - 3*u, Y: T + f.hh - u, W: 6 * u, H: 2.8 * u}, 0.6*u, darker(skin, 0.85), 0, nil)
	head := geom.Rect{X: hx - f.hw/2, Y: T, W: f.hw, H: f.hh}
	RoundRect(dst, head, 3.2*u, skin, 0.8*u, colOutline)
	if (av["hair"] == "bald" || av["hair"] == "") && (av["hat"] == "" || av["hat"] == "none") {
		RoundRect(dst, geom.Rect{X: hx - 3*u, Y: T - 2.2*u, W: 6 * u, H: 2.8 * u}, 0.8*u, skin, 0.8*u, colOutline)
	}
	ex := facing * 0.9 * u
	Circle(dst, hx-2.6*u+ex, hy+0.3*u, 1.1*u, colEye)
	Circle(dst, hx+2.6*u+ex, hy+0.3*u, 1.1*u, colEye)
	Circle(dst, hx-4.3*u+ex, hy+2.6*u, 1.2*u, colCheek)
	Circle(dst, hx+4.3*u+ex, hy+2.6*u, 1.2*u, colCheek)
	Arc(dst, hx+ex, hy+1.6*u, 2*u, 0.15*math.Pi, 0.85*math.Pi, 0.8*u, colEye)

	// Hair piece on top of the head.
	switch av["hair"] {
	case "", "bald":
	case "spiky":
		for _, dx := range []float64{-5, -1.7, 1.7, 5} {
			Poly(dst, hairC, hx+(dx-2.2)*u, T+u, hx+(dx+2.2)*u, T+u, hx+dx*u, T-6*u)
		}
		RoundRect(dst, geom.Rect{X: hx - 7*u, Y: T - 1.5*u, W: 14 * u, H: 4 * u}, 2*u, hairC, 0, nil)
	case "curly":
		for _, dx := range []float64{-4.5, 0, 4.5} {
			Circle(dst, hx+dx*u, T, 3*u, hairC)
		}
		RoundRect(dst, geom.Rect{X: hx - 7*u, Y: T - 0.5*u, W: 14 * u, H: 3 * u}, 1.5*u, hairC, 0, nil)
	default:
		RoundRect(dst, geom.Rect{X: hx - 7.3*u, Y: T - 2*u, W: 14.6 * u, H: 5.5 * u}, 2.8*u, hairC, 0.8*u, colOutline)
		Fill(dst, geom.Rect{X: hx - 7.3*u + 0.4*u, Y: T + 1.5*u, W: 1.8 * u, H: 3 * u}, 0, 0, hairC)
		Fill(dst, geom.Rect{X: hx + 7.3*u - 2.2*u, Y: T + 1.5*u, W: 1.8 * u, H: 3 * u}, 0, 0, hairC)
	}

	// Eyewear.
	switch av["eyewear"] {
	case "goggles":
		Fill(dst, geom.Rect{X: hx - f.hw/2, Y: hy - 0.6*u, W: f.hw, H: 1.6 * u}, 0, 0, Hex("#566078"))
		for _, dx := range []float64{-2.6, 2.6} {
			Circle(dst, hx+dx*u+ex, hy+0.3*u, 2.3*u, color.RGBA{0xa8, 0xe4, 0xff, 0xb0})
			Ring(dst, hx+dx*u+ex, hy+0.3*u, 2.3*u, 0.8*u, Hex("#566078"))
		}
	case "glasses":
		for _, dx := range []float64{-2.6, 2.6} {
			Ring(dst, hx+dx*u+ex, hy+0.3*u, 2*u, 0.55*u, colEye)
		}
		Line(dst, hx-0.6*u+ex, hy+0.1*u, hx+0.6*u+ex, hy+0.1*u, 0.55*u, colEye)
	case "sunglasses":
		for _, dx := range []float64{-2.6, 2.6} {
			Circle(dst, hx+dx*u+ex, hy+0.3*u, 2.1*u, colEye)
		}
		Line(dst, hx-0.7*u+ex, hy+0.1*u, hx+0.7*u+ex, hy+0.1*u, 0.7*u, colEye)
	}

	// Hats.
	switch av["hat"] {
	case "beanie":
		c := Hex("#e8514a")
		RoundRect(dst, geom.Rect{X: hx - 7.6*u, Y: T - 4*u, W: 15.2 * u, H: 8 * u}, 4*u, c, 0, nil)
		Fill(dst, geom.Rect{X: hx - 7.6*u, Y: T + 2*u, W: 15.2 * u, H: 2.2 * u}, 0, 0, darker(c, 0.8))
		Circle(dst, hx, T-4.5*u, 2.2*u, Hex("#ffffff"))
	case "cap":
		c := Hex("#3a6fd8")
		RoundRect(dst, geom.Rect{X: hx - 7*u, Y: T - 3.5*u, W: 14 * u, H: 6 * u}, 3*u, c, 0, nil)
		RoundRect(dst, geom.Rect{X: hx + facing*3.5*u - 6*u, Y: T + 1.8*u, W: 12 * u, H: 2 * u}, u, darker(c, 0.8), 0, nil)
	case "party":
		Poly(dst, Hex("#ff7ac8"), hx-5*u, T+0.5*u, hx+5*u, T+0.5*u, hx, T-12*u)
		Circle(dst, hx, T-12*u, 1.8*u, colGold)
	case "tophat":
		RoundRect(dst, geom.Rect{X: hx - 8.5*u, Y: T - 1.5*u, W: 17 * u, H: 2.6 * u}, u, colEye, 0, nil)
		RoundRect(dst, geom.Rect{X: hx - 5.5*u, Y: T - 13*u, W: 11 * u, H: 12 * u}, u, colEye, 0, nil)
		Fill(dst, geom.Rect{X: hx - 5.5*u, Y: T - 3.5*u, W: 11 * u, H: 1.8 * u}, 0, 0, Hex("#e8514a"))
	case "crown":
		Poly(dst, colGold, hx-6*u, T+u, hx+6*u, T+u, hx+6*u, T-6*u, hx+3*u, T-3*u,
			hx, T-7*u, hx-3*u, T-3*u, hx-6*u, T-6*u)
		Circle(dst, hx, T-0.8*u, 1.1*u, Hex("#e8514a"))
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

// DrawGizmo draws Gizmo, the round green owl helper who floats next to the
// player (C-3). Lab goggles pushed up on the forehead make Gizmo our own owl.
func DrawGizmo(dst *ebiten.Image, x, y, u, t float64) {
	float := math.Sin(t*2.5) * 2 * u
	Shadow(dst, x, y+14*u, 12*u)
	cy := y + float
	green, dark := Hex("#4fbf6a"), Hex("#36975a")
	orange := Hex("#ff9a3c")

	// Flapping wings, ear tufts, then the round body and belly.
	flap := math.Sin(t*9) * 1.5 * u
	for _, s := range []float64{-1, 1} {
		Poly(dst, dark, x+s*7*u, cy-3*u, x+s*12*u, cy+1*u+flap, x+s*7.5*u, cy+5*u)
		Poly(dst, green, x+s*4*u, cy-6*u, x+s*8*u, cy-12*u, x+s*8.5*u, cy-5*u)
	}
	Circle(dst, x, cy, 9*u, green)
	Ring(dst, x, cy, 9*u, u, colOutline)
	Circle(dst, x, cy+3.8*u, 4.6*u, Hex("#d6f5c4"))

	// Big friendly eyes that blink every few seconds.
	blink := math.Mod(t, 4) < 0.15
	for _, s := range []float64{-1, 1} {
		ex, ey := x+s*3.4*u, cy-2.3*u
		if blink {
			Line(dst, ex-2.2*u, ey, ex+2.2*u, ey, 0.9*u, colEye)
			continue
		}
		Circle(dst, ex, ey, 3.4*u, Hex("#ffffff"))
		Ring(dst, ex, ey, 3.4*u, 0.7*u, colOutline)
		Circle(dst, ex+0.4*u, ey+0.3*u, 1.7*u, colEye)
		Circle(dst, ex+0.9*u, ey-0.4*u, 0.6*u, Hex("#ffffff"))
	}
	Poly(dst, orange, x-1.6*u, cy+0.6*u, x+1.6*u, cy+0.6*u, x, cy+3.2*u)

	// Lab goggles pushed up on the forehead.
	Line(dst, x-6*u, cy-7*u, x+6*u, cy-7*u, 0.9*u, Hex("#566078"))
	for _, s := range []float64{-1, 1} {
		Circle(dst, x+s*2.3*u, cy-7*u, 1.9*u, color.RGBA{0xa8, 0xe4, 0xff, 0xff})
		Ring(dst, x+s*2.3*u, cy-7*u, 1.9*u, 0.7*u, Hex("#566078"))
	}
	// Little orange feet.
	Circle(dst, x-2.5*u, cy+9*u, 1.4*u, orange)
	Circle(dst, x+2.5*u, cy+9*u, 1.4*u, orange)
}
