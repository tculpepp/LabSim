package scenes

import (
	"testing"

	"labsim/internal/avatar"
	"labsim/internal/content"
	"labsim/internal/save"
)

func newApp(t *testing.T, store save.Store) *App {
	t.Helper()
	a, err := NewApp(store)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestFreshStartHasNoContinue(t *testing.T) {
	a := newApp(t, &save.MemStore{})
	if a.HasSave() || a.notice != "" {
		t.Fatalf("HasSave=%v notice=%q", a.HasSave(), a.notice)
	}
}

func TestNewGameSavesAndContinues(t *testing.T) {
	store := &save.MemStore{}
	a := newApp(t, store)
	a.StartNewGame("Dr. Pip", content.Expert, avatar.Avatar{"hat": "cap", "pet": "cat"})
	a.State.Wallet.Earn(12)
	a.State.Backpack.Add("star_dust", 2)
	a.State.X, a.State.Y = 260, 210 // inside the Chemistry Lab
	a.SaveNow()

	b := newApp(t, store) // "reload the page"
	if !b.HasSave() {
		t.Fatal("expected Continue after reload")
	}
	if err := b.ContinueGame(); err != nil {
		t.Fatal(err)
	}
	s := b.State
	if s.Name != "Dr. Pip" || s.Level != content.Expert || s.Wallet.Crystals() != 12 || s.Backpack.Count("star_dust") != 2 {
		t.Fatalf("restored %+v", s)
	}
	if s.Avatar["hat"] != "cap" || s.Avatar["pet"] != "cat" || s.Avatar["skin"] == "" {
		t.Fatalf("avatar not restored/normalized: %v", s.Avatar)
	}
	if s.X != 260 || s.Y != 210 || !s.Flag("welcomed") {
		t.Fatalf("position/flags not restored: %v,%v welcomed=%v", s.X, s.Y, s.Flag("welcomed"))
	}
	if b.Level() != content.Expert || b.T("door.locked") != b.Text.Get("door.locked", content.Expert) {
		t.Fatal("text should follow the saved difficulty")
	}
}

func TestCorruptSaveIsBackedUp(t *testing.T) {
	store := &save.MemStore{}
	store.Save([]byte(`{"version":1, garbage`))
	a := newApp(t, store)
	if a.HasSave() || a.notice != "title.corrupt" {
		t.Fatalf("HasSave=%v notice=%q", a.HasSave(), a.notice)
	}
	if string(store.BackupData) != `{"version":1, garbage` {
		t.Fatalf("backup = %q", store.BackupData)
	}
}

func TestFutureSaveIsBackedUp(t *testing.T) {
	store := &save.MemStore{}
	store.Save([]byte(`{"version":999,"name":"x","day":1,"backpackSlots":6}`))
	a := newApp(t, store)
	if a.HasSave() || a.notice != "title.future" || len(store.BackupData) == 0 {
		t.Fatalf("HasSave=%v notice=%q backup=%q", a.HasSave(), a.notice, store.BackupData)
	}
}

func TestContinueInsideWallResetsToSpawn(t *testing.T) {
	store := &save.MemStore{}
	a := newApp(t, store)
	a.StartNewGame("x", content.Junior, nil)
	a.State.X, a.State.Y = 84, 100 // on the Chemistry Lab's left wall
	a.SaveNow()

	b := newApp(t, store)
	if err := b.ContinueGame(); err != nil {
		t.Fatal(err)
	}
	if b.State.X != b.Map.Spawn.X || b.State.Y != b.Map.Spawn.Y {
		t.Fatalf("expected spawn, got %v,%v", b.State.X, b.State.Y)
	}
}

func TestStartOverClearsSave(t *testing.T) {
	store := &save.MemStore{}
	a := newApp(t, store)
	a.StartNewGame("x", content.Junior, nil)
	a.StartOver()
	if a.State != nil || a.HasSave() {
		t.Fatal("state or save survived Start Over")
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("store still has a save")
	}
}

func TestSaveNowWithoutGameIsNoop(t *testing.T) {
	store := &save.MemStore{}
	a := newApp(t, store)
	a.SaveNow()
	if _, err := store.Load(); err == nil {
		t.Fatal("title screen should not write a save")
	}
}

func TestPaidAvatarPartsNeedOwnership(t *testing.T) {
	a := newApp(t, &save.MemStore{})
	a.StartNewGame("x", content.Junior, avatar.Avatar{"hat": "crown"})
	if a.State.Avatar["hat"] == "crown" {
		t.Fatal("new scientist got a paid crown for free")
	}
}

func TestEditCreatorSavesLook(t *testing.T) {
	store := &save.MemStore{}
	a := newApp(t, store)
	a.StartNewGame("x", content.Junior, nil)
	c := NewEditCreator(a)
	c.av["hat"] = "beanie"
	c.custom = "Dr. New"
	c.level = content.Expert
	a.Push(c)
	c.finish(a)
	if a.State.Avatar["hat"] != "beanie" || a.State.Name != "Dr. New" || a.State.Level != content.Expert {
		t.Fatalf("edit not applied: %+v", a.State)
	}
	b := newApp(t, store)
	b.ContinueGame()
	if b.State.Name != "Dr. New" || b.State.Avatar["hat"] != "beanie" {
		t.Fatal("edit not saved")
	}
}
