// Package clock implements the in-game day cycle.
package clock

import "time"

// NightStart is the fraction of the day after which it is night.
const NightStart = 0.8

// Clock tracks the current in-game day and time within it.
type Clock struct {
	dayLength time.Duration
	day       int
	elapsed   time.Duration
}

func New(dayLength time.Duration) *Clock {
	return &Clock{dayLength: dayLength, day: 1}
}

// Restore rebuilds a clock from saved state.
func Restore(dayLength time.Duration, day int, elapsed time.Duration) *Clock {
	c := New(dayLength)
	if day > 1 {
		c.day = day
	}
	c.Advance(elapsed)
	return c
}

func (c *Clock) Day() int                 { return c.day }
func (c *Clock) Elapsed() time.Duration   { return c.elapsed }
func (c *Clock) DayLength() time.Duration { return c.dayLength }

// Progress is 0..1 through the current day.
func (c *Clock) Progress() float64 {
	return float64(c.elapsed) / float64(c.dayLength)
}

func (c *Clock) IsNight() bool { return c.Progress() >= NightStart }

// Advance moves time forward and returns how many new days began.
func (c *Clock) Advance(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	c.elapsed += d
	n := int(c.elapsed / c.dayLength)
	c.elapsed %= c.dayLength
	c.day += n
	return n
}
