package input

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestKeyVector(t *testing.T) {
	cases := []struct {
		name                  string
		left, right, up, down bool
		wantX, wantY          float64
	}{
		{"none", false, false, false, false, 0, 0},
		{"right", false, true, false, false, 1, 0},
		{"up", false, false, true, false, 0, -1},
		{"opposites cancel", true, true, false, false, 0, 0},
		{"diagonal normalized", false, true, false, true, math.Sqrt2 / 2, math.Sqrt2 / 2},
	}
	for _, c := range cases {
		x, y := KeyVector(c.left, c.right, c.up, c.down)
		if !near(x, c.wantX) || !near(y, c.wantY) {
			t.Errorf("%s: got %v,%v want %v,%v", c.name, x, c.wantX, y, c.wantY)
		}
	}
}

func TestJoystickInactiveIsZero(t *testing.T) {
	j := NewJoystick(50, 10)
	if x, y := j.Vector(); x != 0 || y != 0 || j.Active() {
		t.Fatal("inactive joystick should be zero")
	}
}

func TestJoystickDeadZone(t *testing.T) {
	j := NewJoystick(50, 10)
	j.Begin(1, 100, 100)
	j.Move(1, 105, 100)
	if x, y := j.Vector(); x != 0 || y != 0 {
		t.Fatalf("dead zone vector = %v,%v", x, y)
	}
}

func TestJoystickFullAndClamped(t *testing.T) {
	j := NewJoystick(50, 10)
	j.Begin(1, 100, 100)
	j.Move(1, 150, 100)
	if x, y := j.Vector(); !near(x, 1) || !near(y, 0) {
		t.Fatalf("full right = %v,%v", x, y)
	}
	j.Move(1, 100, 300)
	if x, y := j.Vector(); !near(x, 0) || !near(y, 1) {
		t.Fatalf("clamped down = %v,%v", x, y)
	}
	if kx, ky := j.Knob(); !near(kx, 100) || !near(ky, 150) {
		t.Fatalf("knob = %v,%v want 100,150", kx, ky)
	}
}

func TestJoystickHalfway(t *testing.T) {
	j := NewJoystick(50, 10)
	j.Begin(1, 0, 0)
	j.Move(1, 30, 0) // (30-10)/(50-10) = 0.5
	if x, _ := j.Vector(); !near(x, 0.5) {
		t.Fatalf("x = %v, want 0.5", x)
	}
}

func TestJoystickIgnoresOtherTouches(t *testing.T) {
	j := NewJoystick(50, 10)
	if !j.Begin(1, 0, 0) {
		t.Fatal("first Begin should succeed")
	}
	if j.Begin(2, 500, 500) {
		t.Fatal("second touch should not take over")
	}
	j.Move(2, 999, 999)
	if x, y := j.Vector(); x != 0 || y != 0 {
		t.Fatal("other touch moved joystick")
	}
	j.End(2)
	if !j.Active() {
		t.Fatal("other touch ended joystick")
	}
	j.End(1)
	if j.Active() {
		t.Fatal("End did not release")
	}
}
