package experiment

import (
	"math"
	"testing"
)

// ---- Notebook ----

func TestNotebook(t *testing.T) {
	n := NewNotebook([]string{"slime.glow"})
	if !n.Has("slime.glow") || n.Has("slime.bouncy") {
		t.Fatal("restore")
	}
	if !n.Add("slime.bouncy") {
		t.Fatal("first add should report new")
	}
	if n.Add("slime.bouncy") {
		t.Fatal("second add should not be new")
	}
	if got := n.Found(); len(got) != 2 || got[0] != "slime.bouncy" {
		t.Fatalf("Found = %v (want sorted)", got)
	}
}

// ---- Slime (E-CHEM-1) ----

func TestSlimeRecipes(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"slime_base", "glow_powder"}, "glow"},
		{[]string{"bouncy_beads", "slime_base"}, "bouncy"},
		{[]string{"slime_base", "rainbow_drops"}, "rainbow"},
		{[]string{"slime_base", "stretch_gel"}, "stretchy"},
	}
	for _, c := range cases {
		r := MixSlime(c.in)
		if !r.OK || r.Slime != c.want || r.Discovery != "slime."+c.want {
			t.Errorf("MixSlime(%v) = %+v, want %s", c.in, r, c.want)
		}
	}
}

func TestSlimeFailuresGiveHints(t *testing.T) {
	cases := map[string][]string{
		"slime.hint.base":    {"glow_powder"},
		"slime.hint.addin":   {"slime_base"},
		"slime.hint.tooMany": {"slime_base", "glow_powder", "bouncy_beads"},
		"slime.hint.empty":   {},
		"slime.hint.odd":     {"slime_base", "moon_rock"},
	}
	for hint, in := range cases {
		r := MixSlime(in)
		if r.OK || r.Hint != hint {
			t.Errorf("MixSlime(%v) = %+v, want hint %s", in, r, hint)
		}
	}
	if !MixSlime([]string{"slime_base", "glow_powder", "bouncy_beads"}).Boom {
		t.Error("too many add-ins should go BOOM")
	}
}

func TestIsSlimeIngredient(t *testing.T) {
	if !IsSlimeIngredient("slime_base") || !IsSlimeIngredient("stretch_gel") || IsSlimeIngredient("battery") {
		t.Fatal("IsSlimeIngredient")
	}
}

// ---- Circuit (E-PHYS-1) ----

func TestCircuitConductors(t *testing.T) {
	for _, id := range []string{"paperclip", "coin", "foil", "pencil", "key"} {
		m, ok := Material(id)
		if !ok || !m.Conducts {
			t.Errorf("%s should conduct", id)
		}
	}
	for _, id := range []string{"rubber_band", "plastic_spoon", "wood_stick", "eraser", "marble"} {
		m, ok := Material(id)
		if !ok || m.Conducts {
			t.Errorf("%s should not conduct", id)
		}
	}
}

func TestCircuitLights(t *testing.T) {
	if !Lights("foil", true) || Lights("foil", false) || Lights("rubber_band", true) {
		t.Fatal("series circuit: bulb lights only with a conductor AND a closed switch")
	}
}

func TestCircuitMaterialsForLevel(t *testing.T) {
	basic := Materials(false)
	kit := Materials(true)
	if len(kit) <= len(basic) {
		t.Fatal("Spark Kit should add materials")
	}
	for _, m := range basic {
		if m.Kit {
			t.Fatalf("%s needs the kit", m.ID)
		}
	}
}

func TestCircuitTest(t *testing.T) {
	r := TestCircuit("coin", true, true)
	if !r.Lit || !r.Correct || r.Discovery != "circuit.coin" {
		t.Fatalf("coin = %+v", r)
	}
	r = TestCircuit("rubber_band", true, true)
	if r.Lit || r.Correct || r.Hint == "" {
		t.Fatalf("rubber band predicted light = %+v", r)
	}
	r = TestCircuit("coin", false, false) // switch open: predicted dark, right!
	if r.Lit || !r.Correct {
		t.Fatalf("open switch = %+v", r)
	}
}

// ---- Microscope (E-BIO-1) ----

func TestSlides(t *testing.T) {
	basic := Slides(false)
	all := Slides(true)
	if len(basic) != 4 || len(all) <= len(basic) {
		t.Fatalf("slides: basic %d, all %d", len(basic), len(all))
	}
	for _, s := range all {
		if s.Answer == "" {
			t.Fatalf("%s has no answer", s.ID)
		}
	}
}

func TestSlideChoices(t *testing.T) {
	s, _ := Slide("pond")
	for _, n := range []int{3, 4} {
		c := Choices(s, n, 42)
		if len(c) != n {
			t.Fatalf("got %d choices, want %d", len(c), n)
		}
		found, seen := false, map[string]bool{}
		for _, x := range c {
			if x == s.Answer {
				found = true
			}
			if seen[x] {
				t.Fatalf("duplicate choice %q", x)
			}
			seen[x] = true
		}
		if !found {
			t.Fatal("answer missing from choices")
		}
	}
	a, b := Choices(s, 3, 7), Choices(s, 3, 7)
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("same seed, different order")
		}
	}
}

func TestLookAt(t *testing.T) {
	s, _ := Slide("onion")
	if r := LookAt(s, s.Answer); !r.Correct || r.Discovery != "micro.onion" {
		t.Fatalf("right = %+v", r)
	}
	if r := LookAt(s, "creatures"); r.Correct || r.Hint != "micro.hint.onion" {
		t.Fatalf("wrong = %+v", r)
	}
}

// ---- Gravity (E-SPACE-1) ----

func TestGravityRealValues(t *testing.T) {
	want := map[string]float64{"moon": 1.62, "mars": 3.71, "earth": 9.81, "jupiter": 24.79, "pluto": 0.62}
	for id, g := range want {
		p, ok := Planet(id)
		if !ok || p.G != g {
			t.Errorf("%s g = %v, want %v", id, p.G, g)
		}
	}
}

func TestJumpPhysics(t *testing.T) {
	// h = v²/2g, with a kid's 2.5 m/s jump: about 32 cm on Earth.
	if h := JumpHeight(9.81); math.Abs(h-0.3186) > 0.001 {
		t.Fatalf("Earth jump = %v", h)
	}
	moon, earth := JumpHeight(1.62), JumpHeight(9.81)
	if r := moon / earth; math.Abs(r-9.81/1.62) > 1e-9 {
		t.Fatalf("ratio = %v", r)
	}
	// t = sqrt(2h/g): a 2 m drop takes about 0.64 s on Earth.
	if ft := FallTime(2, 9.81); math.Abs(ft-0.6386) > 0.001 {
		t.Fatalf("fall time = %v", ft)
	}
}

func TestPlanetsByTelescope(t *testing.T) {
	if len(Planets(false)) != 3 || len(Planets(true)) != 5 {
		t.Fatal("telescope adds Jupiter and Pluto")
	}
}

func TestGravityQuestions(t *testing.T) {
	// Junior: where do you jump highest? (lowest gravity of the planets shown)
	if got := HighestJump(Planets(false)); got != "moon" {
		t.Fatalf("highest = %s", got)
	}
	if got := HighestJump(Planets(true)); got != "pluto" {
		t.Fatalf("highest with telescope = %s", got)
	}
	// Expert: about how many times higher on the Moon than Earth? 6.
	if got := MoonEarthRatio(); got != 6 {
		t.Fatalf("ratio = %d", got)
	}
}
