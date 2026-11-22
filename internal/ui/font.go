package ui

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"labsim/internal/textwrap"
)

// Go fonts (BSD licence) for now; swap for a rounder, friendlier font in the art pass (M9).
var regularSrc, boldSrc *text.GoTextFaceSource

func init() {
	var err error
	if regularSrc, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF)); err != nil {
		panic(err)
	}
	if boldSrc, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF)); err != nil {
		panic(err)
	}
}

// Text sizes in logical pixels (screen is 480 tall).
const (
	SizeSmall   = 14
	SizeBody    = 18
	SizeHeading = 24
	SizeTitle   = 56
)

func Regular(size float64) text.Face { return &text.GoTextFace{Source: regularSrc, Size: size} }
func Bold(size float64) text.Face    { return &text.GoTextFace{Source: boldSrc, Size: size} }

// LineHeight is the spacing used between wrapped lines.
func LineHeight(size float64) float64 { return size * 1.35 }

// Text draws s with its top-left (or top-center/right per align) at (x, y).
func Text(dst *ebiten.Image, s string, face text.Face, x, y float64, c color.Color, align text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.PrimaryAlign = align
	text.Draw(dst, s, face, op)
}

// TextCentered draws s centered on (cx, cy).
func TextCentered(dst *ebiten.Image, s string, face text.Face, cx, cy float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx, cy)
	op.ColorScale.ScaleWithColor(c)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(dst, s, face, op)
}

// Measure returns the width of s in face.
func Measure(s string, face text.Face) float64 {
	w, _ := text.Measure(s, face, 0)
	return w
}

// Wrap word-wraps s to maxW using face.
func Wrap(s string, face text.Face, maxW float64) []string {
	return textwrap.Wrap(s, maxW, func(t string) float64 { return Measure(t, face) })
}

// Lines draws pre-wrapped lines starting at (x, y).
func Lines(dst *ebiten.Image, lines []string, face text.Face, size, x, y float64, c color.Color, align text.Align) {
	for i, l := range lines {
		Text(dst, l, face, x, y+float64(i)*LineHeight(size), c, align)
	}
}

// TextOutlined draws centered text with a dark outline so it reads on any floor.
func TextOutlined(dst *ebiten.Image, s string, face text.Face, cx, cy float64, c, outline color.Color) {
	for _, d := range [][2]float64{{-1.5, 0}, {1.5, 0}, {0, -1.5}, {0, 1.5}, {1.2, 1.2}, {-1.2, 1.2}, {1.2, -1.2}, {-1.2, -1.2}} {
		TextCentered(dst, s, face, cx+d[0], cy+d[1], outline)
	}
	TextCentered(dst, s, face, cx, cy, c)
}
