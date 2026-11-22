package state

import (
	"errors"
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
