package save

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func sample() Data {
	return Data{
		Name: "Doctor Sparkle", Level: "junior", X: 840, Y: 630, Room: "main_hall",
		Crystals: 42, BackpackSlots: 6,
		Items: []Item{{ID: "star_dust", Qty: 3}},
		Day:   4, DayElapsedMs: 12345,
		Unlocked: []string{"robot_workshop"}, Flags: []string{"welcomed"},
	}
}

func TestRoundTrip(t *testing.T) {
	c := DefaultCodec()
	b, err := c.Encode(sample())
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	want := sample()
	want.Version = CurrentVersion
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip:\n got %+v\nwant %+v", got, want)
	}
}

func TestDecodeCorrupt(t *testing.T) {
	c := DefaultCodec()
	for name, src := range map[string]string{
		"not json":       `{{{`,
		"no version":     `{"name":"x","day":1,"backpackSlots":6}`,
		"negative coins": `{"version":1,"name":"x","day":1,"backpackSlots":6,"crystals":-5}`,
		"day zero":       `{"version":1,"name":"x","day":0,"backpackSlots":6}`,
		"no slots":       `{"version":1,"name":"x","day":1,"backpackSlots":0}`,
		"bad item qty":   `{"version":1,"name":"x","day":1,"backpackSlots":6,"items":[{"id":"a","qty":0}]}`,
		"too many items": `{"version":1,"name":"x","day":1,"backpackSlots":1,"items":[{"id":"a","qty":1},{"id":"b","qty":1}]}`,
		"wrong type":     `{"version":"one"}`,
	} {
		if _, err := c.Decode([]byte(src)); !errors.Is(err, ErrCorrupt) {
			t.Errorf("%s: err = %v, want ErrCorrupt", name, err)
		}
	}
}

func TestDecodeFutureVersion(t *testing.T) {
	c := DefaultCodec()
	_, err := c.Decode([]byte(`{"version":99,"name":"x","day":1,"backpackSlots":6}`))
	if !errors.Is(err, ErrFutureVersion) {
		t.Fatalf("err = %v, want ErrFutureVersion", err)
	}
}

// SV-3: an old save is migrated step by step to the current version.
func TestDecodeMigratesOldVersion(t *testing.T) {
	c := Codec{
		Version: 3,
		Migrations: map[int]Migration{
			1: func(raw map[string]any) error { raw["name"] = raw["playerName"]; delete(raw, "playerName"); return nil },
			2: func(raw map[string]any) error { raw["crystals"] = raw["coins"]; delete(raw, "coins"); return nil },
		},
	}
	old := `{"version":1,"playerName":"Old Timer","coins":7,"day":2,"backpackSlots":6}`
	d, err := c.Decode([]byte(old))
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "Old Timer" || d.Crystals != 7 || d.Version != 3 {
		t.Fatalf("migrated = %+v", d)
	}
}

func TestDecodeMissingMigration(t *testing.T) {
	c := Codec{Version: 2}
	_, err := c.Decode([]byte(`{"version":1,"name":"x","day":1,"backpackSlots":6}`))
	if !errors.Is(err, ErrCorrupt) {
		t.Fatalf("err = %v, want ErrCorrupt", err)
	}
}

func TestDecodeIgnoresUnknownFields(t *testing.T) {
	c := DefaultCodec()
	if _, err := c.Decode([]byte(`{"version":1,"name":"x","day":1,"backpackSlots":6,"futureThing":true}`)); err != nil {
		t.Fatalf("unknown field should be ignored: %v", err)
	}
}

func testStore(t *testing.T, s Store) {
	t.Helper()
	if _, err := s.Load(); !errors.Is(err, ErrNoSave) {
		t.Fatalf("empty Load err = %v, want ErrNoSave", err)
	}
	if err := s.Save([]byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save([]byte("two")); err != nil {
		t.Fatal(err)
	}
	if b, err := s.Load(); err != nil || string(b) != "two" {
		t.Fatalf("Load = %q, %v", b, err)
	}
	if err := s.Backup([]byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); !errors.Is(err, ErrNoSave) {
		t.Fatalf("after Clear err = %v, want ErrNoSave", err)
	}
	if err := s.Clear(); err != nil {
		t.Fatalf("Clear twice: %v", err)
	}
}

func TestMemStore(t *testing.T) {
	m := &MemStore{}
	testStore(t, m)
	if string(m.BackupData) != "bad" {
		t.Fatalf("backup = %q", m.BackupData)
	}
}

func TestFileStore(t *testing.T) {
	dir := t.TempDir()
	fs := FileStore{Path: filepath.Join(dir, "nested", "save.json")}
	testStore(t, fs)
	if b, err := os.ReadFile(fs.Path + ".corrupt"); err != nil || string(b) != "bad" {
		t.Fatalf("backup file = %q, %v", b, err)
	}
	left, _ := filepath.Glob(filepath.Join(dir, "nested", "*.tmp*"))
	if len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}

func TestAutosaverInterval(t *testing.T) {
	n := 0
	a := NewAutosaver(30*time.Second, func() error { n++; return nil })
	a.Tick(29 * time.Second)
	if n != 0 {
		t.Fatal("saved too early")
	}
	if !a.Tick(time.Second) || n != 1 {
		t.Fatalf("expected save at 30s, n=%d", n)
	}
	a.Tick(29 * time.Second)
	if n != 1 {
		t.Fatal("timer did not reset")
	}
}

func TestAutosaverNowResetsTimer(t *testing.T) {
	n := 0
	a := NewAutosaver(30*time.Second, func() error { n++; return nil })
	a.Tick(25 * time.Second)
	a.Now()
	a.Tick(25 * time.Second)
	if n != 1 {
		t.Fatalf("n = %d, want 1 (Now resets interval)", n)
	}
}

func TestAutosaverRecordsError(t *testing.T) {
	boom := errors.New("disk full")
	a := NewAutosaver(time.Second, func() error { return boom })
	a.Now()
	if !errors.Is(a.LastErr(), boom) {
		t.Fatalf("LastErr = %v", a.LastErr())
	}
}
