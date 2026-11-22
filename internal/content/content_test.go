package content

import (
	"errors"
	"testing"
)

const sample = `{
  "greet": {"junior": "Hi %s!", "expert": "Welcome, %s."},
  "only.junior": {"junior": "Easy words"}
}`

func TestGetByLevel(t *testing.T) {
	c, err := Load([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Get("greet", Junior); got != "Hi %s!" {
		t.Errorf("junior = %q", got)
	}
	if got := c.Get("greet", Expert); got != "Welcome, %s." {
		t.Errorf("expert = %q", got)
	}
}

func TestExpertFallsBackToJunior(t *testing.T) {
	c, _ := Load([]byte(sample))
	if got := c.Get("only.junior", Expert); got != "Easy words" {
		t.Fatalf("fallback = %q", got)
	}
}

func TestMissingKeyIsVisible(t *testing.T) {
	c, _ := Load([]byte(sample))
	if got := c.Get("nope", Junior); got != "[nope]" {
		t.Fatalf("missing = %q", got)
	}
}

func TestLoadRequiresJunior(t *testing.T) {
	_, err := Load([]byte(`{"x": {"expert": "only expert"}}`))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if _, err := Load([]byte(`{`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bad json err = %v", err)
	}
}

func TestLoadRejectsMismatchedVerbs(t *testing.T) {
	_, err := Load([]byte(`{"x": {"junior": "%d crystals", "expert": "%s crystals"}}`))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}

func TestParseLevel(t *testing.T) {
	if l, ok := ParseLevel("expert"); !ok || l != Expert {
		t.Fatal("expert")
	}
	if l, ok := ParseLevel("bogus"); ok || l != Junior {
		t.Fatal("bogus should fall back to junior, not ok")
	}
}

func TestKeysSorted(t *testing.T) {
	c, _ := Load([]byte(sample))
	k := c.Keys()
	if len(k) != 2 || k[0] != "greet" || k[1] != "only.junior" {
		t.Fatalf("Keys = %v", k)
	}
}
