package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"labsim/internal/geom"
)

// Tap is a completed click or touch (released this frame) in logical coordinates.
type Tap struct {
	X, Y     float64
	consumed bool
}

// Pointer collects mouse clicks and touch taps once per frame.
type Pointer struct {
	taps       []Tap
	touchSeen  bool
	touchStart map[ebiten.TouchID][2]float64
}

// maxTapTravel: a touch that moves farther than this is a drag, not a tap.
const maxTapTravel = 20

func (p *Pointer) Update() {
	p.taps = p.taps[:0]
	if p.touchStart == nil {
		p.touchStart = map[ebiten.TouchID][2]float64{}
	}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		p.touchSeen = true
		x, y := ebiten.TouchPosition(id)
		p.touchStart[id] = [2]float64{float64(x), float64(y)}
	}
	touched := false
	for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
		touched = true
		x, y := inpututil.TouchPositionFInPreviousTick(id)
		if s, ok := p.touchStart[id]; ok {
			dx, dy := x-s[0], y-s[1]
			if dx*dx+dy*dy <= maxTapTravel*maxTapTravel {
				p.taps = append(p.taps, Tap{X: x, Y: y})
			}
			delete(p.touchStart, id)
		}
	}
	// Browsers may also synthesize mouse events from touch; ignore them then.
	if !touched && !p.touchSeen && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		p.taps = append(p.taps, Tap{X: float64(x), Y: float64(y)})
	}
}

// TouchDevice reports whether the player has used touch (T-6: show touch controls).
func (p *Pointer) TouchDevice() bool { return p.touchSeen }

// TapIn consumes and reports a tap inside r.
func (p *Pointer) TapIn(r geom.Rect) bool {
	for i := range p.taps {
		t := &p.taps[i]
		if !t.consumed && r.Contains(t.X, t.Y) {
			t.consumed = true
			return true
		}
	}
	return false
}

// AnyTap consumes and reports any unconsumed tap.
func (p *Pointer) AnyTap() bool {
	for i := range p.taps {
		if !p.taps[i].consumed {
			p.taps[i].consumed = true
			return true
		}
	}
	return false
}
