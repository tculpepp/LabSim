package state

import (
	"time"

	"labsim/internal/clock"
	"labsim/internal/save"
)

// Touch controls (ST-1): auto shows them on touch devices only.
const (
	TouchAuto = "auto"
	TouchOn   = "on"
	TouchOff  = "off"
)

// Settings are the grown-up options (ST-1, AU-2).
type Settings struct {
	Volume     int // 0..100
	Muted      bool
	DayMinutes int // 5..20
	Touch      string
	ShowFPS    bool
}

func DefaultSettings() Settings {
	return Settings{Volume: 70, DayMinutes: int(DayLength / time.Minute), Touch: TouchAuto}
}

// Normalize clamps values so a hand-edited save can't break the game.
func (s Settings) Normalize() Settings {
	s.Volume = max(0, min(100, s.Volume))
	if s.DayMinutes == 0 {
		s.DayMinutes = DefaultSettings().DayMinutes
	}
	s.DayMinutes = max(5, min(20, s.DayMinutes))
	if s.Touch != TouchOn && s.Touch != TouchOff {
		s.Touch = TouchAuto
	}
	return s
}

// SetDayMinutes changes the day length, keeping how far through today we are.
func (s *State) SetDayMinutes(m int) {
	s.Settings.DayMinutes = m
	s.Settings = s.Settings.Normalize()
	newLen := time.Duration(s.Settings.DayMinutes) * time.Minute
	elapsed := time.Duration(s.Clock.Progress() * float64(newLen))
	s.Clock = clock.Restore(newLen, s.Clock.Day(), elapsed)
}

func (s Settings) toSave() *save.Settings {
	return &save.Settings{Volume: s.Volume, Muted: s.Muted, DayMinutes: s.DayMinutes, Touch: s.Touch, ShowFPS: s.ShowFPS}
}

func settingsFromSave(d *save.Settings) Settings {
	if d == nil {
		return DefaultSettings()
	}
	return Settings{Volume: d.Volume, Muted: d.Muted, DayMinutes: d.DayMinutes, Touch: d.Touch, ShowFPS: d.ShowFPS}.Normalize()
}
