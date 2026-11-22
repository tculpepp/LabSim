package clock

import (
	"testing"
	"time"
)

func TestStartsOnDayOneMorning(t *testing.T) {
	c := New(10 * time.Minute)
	if c.Day() != 1 {
		t.Fatalf("Day = %d, want 1", c.Day())
	}
	if c.Progress() != 0 {
		t.Fatalf("Progress = %v, want 0", c.Progress())
	}
}

func TestAdvanceWithinDay(t *testing.T) {
	c := New(10 * time.Minute)
	if n := c.Advance(5 * time.Minute); n != 0 {
		t.Fatalf("new days = %d, want 0", n)
	}
	if c.Progress() != 0.5 {
		t.Fatalf("Progress = %v, want 0.5", c.Progress())
	}
}

func TestAdvanceRollsOverDay(t *testing.T) {
	c := New(10 * time.Minute)
	c.Advance(9 * time.Minute)
	if n := c.Advance(2 * time.Minute); n != 1 {
		t.Fatalf("new days = %d, want 1", n)
	}
	if c.Day() != 2 {
		t.Fatalf("Day = %d, want 2", c.Day())
	}
	if got := c.Elapsed(); got != time.Minute {
		t.Fatalf("Elapsed = %v, want 1m", got)
	}
}

func TestAdvanceMultipleDays(t *testing.T) {
	c := New(time.Minute)
	if n := c.Advance(3*time.Minute + 10*time.Second); n != 3 {
		t.Fatalf("new days = %d, want 3", n)
	}
	if c.Day() != 4 {
		t.Fatalf("Day = %d, want 4", c.Day())
	}
}

func TestAdvanceIgnoresNegative(t *testing.T) {
	c := New(time.Minute)
	c.Advance(-time.Second)
	if c.Elapsed() != 0 {
		t.Fatal("negative advance changed clock")
	}
}

func TestIsNight(t *testing.T) {
	c := New(100 * time.Second)
	c.Advance(70 * time.Second)
	if c.IsNight() {
		t.Fatal("70% should be day")
	}
	c.Advance(10 * time.Second)
	if !c.IsNight() {
		t.Fatal("80% should be night")
	}
}

func TestRestore(t *testing.T) {
	c := Restore(10*time.Minute, 5, 3*time.Minute)
	if c.Day() != 5 || c.Elapsed() != 3*time.Minute {
		t.Fatalf("restored day=%d elapsed=%v", c.Day(), c.Elapsed())
	}
}
