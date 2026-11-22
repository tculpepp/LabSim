package save

import "time"

// Autosaver saves on a fixed interval and on demand (SV-1).
type Autosaver struct {
	interval time.Duration
	since    time.Duration
	save     func() error
	lastErr  error
}

func NewAutosaver(interval time.Duration, save func() error) *Autosaver {
	return &Autosaver{interval: interval, save: save}
}

// Tick advances the timer; returns true if it saved.
func (a *Autosaver) Tick(dt time.Duration) bool {
	a.since += dt
	if a.since < a.interval {
		return false
	}
	a.Now()
	return true
}

// Now saves immediately and restarts the interval.
func (a *Autosaver) Now() {
	a.since = 0
	a.lastErr = a.save()
}

// LastErr is the result of the most recent save attempt.
func (a *Autosaver) LastErr() error { return a.lastErr }
