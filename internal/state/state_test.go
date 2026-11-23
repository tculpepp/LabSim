package state

import (
	"errors"
	"labsim/internal/decor"
	"testing"
	"time"

	"labsim/internal/content"
	"labsim/internal/save"
)

func TestNewState(t *testing.T) {
	s := New("Dr. Pip", content.Expert, 10, 20)
	if s.Name != "Dr. Pip" || s.Level != content.Expert || s.X != 10 || s.Y != 20 {
		t.Fatalf("New = %+v", s)
	}
	if s.Wallet.Crystals() != 0 || s.Backpack.Slots() != StartBackpackSlots || s.Clock.Day() != 1 {
		t.Fatal("bad starting values")
	}
	if s.IsUnlocked("robot_workshop") {
		t.Fatal("nothing should start unlocked")
	}
}

func TestUnlockAndFlags(t *testing.T) {
	s := New("x", content.Junior, 0, 0)
	s.Unlock("robot_workshop")
	if !s.IsUnlocked("robot_workshop") {
		t.Fatal("Unlock failed")
	}
	s.SetFlag("welcomed")
	if !s.Flag("welcomed") || s.Flag("other") {
		t.Fatal("flags wrong")
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	s := New("Dr. Pip", content.Expert, 100, 200)
	s.Room = "space_lab"
	s.Wallet.Earn(55)
	s.Backpack.Add("star_dust", 2)
	s.Backpack.Add("battery", 1)
	s.Backpack.Upgrade(2)
	s.Clock.Advance(DayLength + 90*time.Second)
	s.Unlock("robot_workshop")
	s.SetFlag("welcomed")
	s.Avatar = map[string]string{"hat": "cap", "pet": "cat"}
	s.Own("avatar:hat:crown")
	s.MarkPicked("d2-r0-1")
	s.Notebook.Add("slime.glow")
	s.Decor = []decor.Placement{{ID: "plant", Cell: decor.Cell{X: 2, Y: 1}}}

	codec := save.DefaultCodec()
	b, err := codec.Encode(s.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	d, err := codec.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	r, err := FromSave(d)
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != s.Name || r.Level != s.Level || r.X != 100 || r.Y != 200 || r.Room != "space_lab" {
		t.Fatalf("player = %+v", r)
	}
	if r.Wallet.Crystals() != 55 || r.Backpack.Count("star_dust") != 2 || r.Backpack.Count("battery") != 1 || r.Backpack.Slots() != StartBackpackSlots+2 {
		t.Fatal("wallet/backpack not restored")
	}
	if r.Clock.Day() != 2 || r.Clock.Elapsed() != 90*time.Second {
		t.Fatalf("clock day=%d elapsed=%v", r.Clock.Day(), r.Clock.Elapsed())
	}
	if !r.IsUnlocked("robot_workshop") || !r.Flag("welcomed") {
		t.Fatal("unlocks/flags not restored")
	}
	if r.Avatar["hat"] != "cap" || r.Avatar["pet"] != "cat" || !r.Owns("avatar:hat:crown") || r.Owns("avatar:hat:tophat") {
		t.Fatalf("avatar/owned not restored: %v", r.Avatar)
	}
	if len(r.Decor) != 1 || r.Decor[0].Cell.X != 2 {
		t.Fatalf("decor not restored: %v", r.Decor)
	}
	if !r.Notebook.Has("slime.glow") || r.Notebook.Has("slime.bouncy") {
		t.Fatal("notebook not restored")
	}
	if !r.Picked("d2-r0-1") || r.Picked("d2-r0-2") {
		t.Fatal("picked not restored")
	}
}

func TestSnapshotDeterministic(t *testing.T) {
	s := New("x", content.Junior, 0, 0)
	s.Unlock("b")
	s.Unlock("a")
	s.SetFlag("z")
	s.SetFlag("y")
	d := s.Snapshot()
	if d.Unlocked[0] != "a" || d.Flags[0] != "y" {
		t.Fatalf("not sorted: %v %v", d.Unlocked, d.Flags)
	}
}

func TestFromSaveUnknownLevelFallsBackToJunior(t *testing.T) {
	d := New("x", content.Junior, 0, 0).Snapshot()
	d.Level = "wizard"
	r, err := FromSave(d)
	if err != nil || r.Level != content.Junior {
		t.Fatalf("level = %v, err = %v", r.Level, err)
	}
}

func TestFromSaveBadBackpack(t *testing.T) {
	d := New("x", content.Junior, 0, 0).Snapshot()
	d.BackpackSlots = 1
	d.Items = []save.Item{{ID: "a", Qty: 1}, {ID: "b", Qty: 1}}
	if _, err := FromSave(d); !errors.Is(err, save.ErrCorrupt) {
		t.Fatalf("err = %v, want ErrCorrupt", err)
	}
}

func TestPickedResetsOnNewDay(t *testing.T) {
	s := New("x", content.Junior, 0, 0)
	s.MarkPicked("d1-r0-0")
	if !s.Picked("d1-r0-0") {
		t.Fatal("not picked")
	}
	s.Clock.Advance(DayLength)
	if s.Picked("d1-r0-0") {
		t.Fatal("picked list should reset on a new day")
	}
}

func TestJobsForToday(t *testing.T) {
	s := New("x", content.Junior, 0, 0)
	if _, ok := s.TodaysJobs(); ok {
		t.Fatal("no jobs before visiting Dr. Nova")
	}
	s.GiveJobs([]string{"a", "b"})
	s.MarkJobDone("a")
	ids, ok := s.TodaysJobs()
	if !ok || len(ids) != 2 || !s.JobDone("a") || s.JobDone("b") {
		t.Fatalf("jobs = %v ok=%v", ids, ok)
	}
	d := s.Snapshot()
	r, _ := FromSave(d)
	if ids, ok := r.TodaysJobs(); !ok || len(ids) != 2 || !r.JobDone("a") {
		t.Fatal("jobs not saved")
	}
	s.Clock.Advance(DayLength)
	if _, ok := s.TodaysJobs(); ok {
		t.Fatal("yesterday's jobs should expire")
	}
	if s.JobDone("a") {
		t.Fatal("done marks should expire with the jobs")
	}
}

func TestSettingsDefaultsAndRoundTrip(t *testing.T) {
	s := New("Ada", content.Junior, 0, 0)
	if s.Settings != DefaultSettings() || s.Settings.Volume != 70 || s.Settings.DayMinutes != 10 {
		t.Fatalf("defaults = %+v", s.Settings)
	}
	s.Settings.Volume, s.Settings.Muted, s.Settings.Touch = 30, true, TouchOn
	r, err := FromSave(s.Snapshot())
	if err != nil || r.Settings != s.Settings {
		t.Fatalf("restored %+v, %v", r.Settings, err)
	}
	// Old saves without settings get defaults.
	d := s.Snapshot()
	d.Settings = nil
	if r, _ := FromSave(d); r.Settings != DefaultSettings() {
		t.Fatal("old save should get default settings")
	}
}

func TestSettingsNormalize(t *testing.T) {
	got := Settings{Volume: 150, DayMinutes: 2, Touch: "sideways"}.Normalize()
	if got.Volume != 100 || got.DayMinutes != 5 || got.Touch != TouchAuto {
		t.Fatalf("Normalize = %+v", got)
	}
	if (Settings{Volume: -5, DayMinutes: 99}).Normalize().Volume != 0 {
		t.Fatal("negative volume")
	}
	if (Settings{DayMinutes: 99}).Normalize().DayMinutes != 20 {
		t.Fatal("day too long")
	}
}

func TestSetDayLengthKeepsDay(t *testing.T) {
	s := New("Ada", content.Junior, 0, 0)
	s.Clock.Advance(DayLength * 3 / 2) // day 2, halfway
	s.SetDayMinutes(20)
	if s.Clock.Day() != 2 || s.Settings.DayMinutes != 20 {
		t.Fatalf("day %d, %+v", s.Clock.Day(), s.Settings)
	}
}
