// Package input turns keyboard and touch state into movement intents.
// It has no Ebitengine dependency; the UI layer feeds it raw positions.
package input

import "math"

// KeyVector converts direction keys into a unit (or zero) movement vector.
func KeyVector(left, right, up, down bool) (float64, float64) {
	x, y := 0.0, 0.0
	if left {
		x--
	}
	if right {
		x++
	}
	if up {
		y--
	}
	if down {
		y++
	}
	if l := math.Hypot(x, y); l > 0 {
		return x / l, y / l
	}
	return 0, 0
}

// Joystick is a floating on-screen joystick: it appears where the finger
// first touches and tracks that one finger until it lifts.
type Joystick struct {
	Radius   float64
	DeadZone float64

	active bool
	id     int
	ox, oy float64
	x, y   float64
}

func NewJoystick(radius, deadZone float64) *Joystick {
	return &Joystick{Radius: radius, DeadZone: deadZone}
}

// Begin starts tracking touch id at (x, y). Returns false if already tracking another touch.
func (j *Joystick) Begin(id int, x, y float64) bool {
	if j.active {
		return false
	}
	j.active, j.id = true, id
	j.ox, j.oy, j.x, j.y = x, y, x, y
	return true
}

func (j *Joystick) Move(id int, x, y float64) {
	if j.active && id == j.id {
		j.x, j.y = x, y
	}
}

func (j *Joystick) End(id int) {
	if j.active && id == j.id {
		j.active = false
	}
}

func (j *Joystick) Active() bool               { return j.active }
func (j *Joystick) ID() int                    { return j.id }
func (j *Joystick) Origin() (float64, float64) { return j.ox, j.oy }

// Knob returns the knob position, clamped to the joystick radius.
func (j *Joystick) Knob() (float64, float64) {
	dx, dy := j.x-j.ox, j.y-j.oy
	if d := math.Hypot(dx, dy); d > j.Radius {
		dx, dy = dx/d*j.Radius, dy/d*j.Radius
	}
	return j.ox + dx, j.oy + dy
}

// Vector returns a movement vector with magnitude 0..1, rescaled past the dead zone.
func (j *Joystick) Vector() (float64, float64) {
	if !j.active {
		return 0, 0
	}
	dx, dy := j.x-j.ox, j.y-j.oy
	d := math.Hypot(dx, dy)
	if d <= j.DeadZone {
		return 0, 0
	}
	mag := (math.Min(d, j.Radius) - j.DeadZone) / (j.Radius - j.DeadZone)
	return dx / d * mag, dy / d * mag
}
