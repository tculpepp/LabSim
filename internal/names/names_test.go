package names

import (
	"errors"
	"testing"
)

const sample = `{"titles":["Doctor","Captain"],"adjectives":["Sparkle"],"nouns":["Beaker","Comet"]}`

func TestBuild(t *testing.T) {
	b, err := Load([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Build(Pick{0, 0, 1}); got != "Doctor Sparkle Comet" {
		t.Fatalf("Build = %q", got)
	}
	// indexes wrap both ways
	if got := b.Build(Pick{3, -1, 2}); got != "Captain Sparkle Beaker" {
		t.Fatalf("wrapped Build = %q", got)
	}
}

func TestRandomDeterministic(t *testing.T) {
	b, _ := Load([]byte(sample))
	if b.Random(7) != b.Random(7) {
		t.Fatal("same seed gave different names")
	}
}

func TestAllCombos(t *testing.T) {
	b, _ := Load([]byte(sample))
	if n := len(b.All()); n != 4 {
		t.Fatalf("All = %d names, want 4", n)
	}
}

func TestLoadRejectsEmptyLists(t *testing.T) {
	if _, err := Load([]byte(`{"titles":[],"adjectives":["a"],"nouns":["b"]}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
}
