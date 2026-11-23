package scenes

import (
	"testing"

	"labsim/internal/content"
	"labsim/internal/save"
	"labsim/internal/state"
)

func TestTouchKeys(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	for _, c := range touchChoices {
		if !a.Text.Has(touchKey(c)) {
			t.Errorf("missing %s", touchKey(c))
		}
	}
}

func TestSettingsChangesAreSaved(t *testing.T) {
	store := &save.MemStore{}
	a, _ := playing(t, store)
	st := &Settings{}
	a.Push(st)
	st.layout(a)
	// Buttons come in </> pairs per row: volume, sound, level, day, touch, fps.
	st.acts[0](a)              // volume down
	st.acts[5](a)              // level flip
	st.acts[7](a)              // day longer
	st.acts[9](a)              // touch next
	st.acts[len(st.acts)-1](a) // done
	s := a.State.Settings
	if s.Volume != 60 || a.State.Level != content.Expert || s.DayMinutes != 15 || s.Touch != state.TouchOn {
		t.Fatalf("settings = %+v, level %v", s, a.State.Level)
	}
	b, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	d, err := save.DefaultCodec().Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Settings == nil || d.Settings.Volume != 60 || d.Settings.DayMinutes != 15 {
		t.Fatalf("saved settings = %+v", d.Settings)
	}
}

func TestShowTouchSetting(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Settings.Touch = state.TouchOn
	if !a.ShowTouch() {
		t.Fatal("on")
	}
	a.State.Settings.Touch = state.TouchOff
	if a.ShowTouch() {
		t.Fatal("off")
	}
}
