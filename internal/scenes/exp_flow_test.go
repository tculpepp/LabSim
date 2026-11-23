package scenes

import (
	"testing"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/save"
)

func TestSlimeSuccessDiscoversAndPaysJob(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_slime"})
	a.State.Backpack.Add("slime_base", 1)
	a.State.Backpack.Add("glow_powder", 1)
	s := NewSlimeLab()
	a.Push(s)
	s.bowl = []string{"slime_base", "glow_powder"}
	s.mix(a)
	if s.result != "glow" || !a.State.Notebook.Has("slime.glow") {
		t.Fatalf("result=%q notebook=%v", s.result, a.State.Notebook.Found())
	}
	if a.State.Backpack.Count("slime_base") != 0 || a.State.Backpack.Count("glow_powder") != 0 {
		t.Fatal("ingredients not used up")
	}
	want := a.Disc.Bonus + 15
	if a.State.Wallet.Crystals() != want {
		t.Fatalf("crystals = %d, want %d (bonus + job)", a.State.Wallet.Crystals(), want)
	}
	// Second glow slime: no new bonus.
	a.State.Backpack.Add("slime_base", 1)
	a.State.Backpack.Add("glow_powder", 1)
	s.sparkleT, s.pendingFC = 0, ""
	s.bowl = []string{"slime_base", "glow_powder"}
	s.mix(a)
	if a.State.Wallet.Crystals() != want {
		t.Fatal("discovery bonus paid twice")
	}
}

func TestSlimeBoomUsesIngredientsAndGivesHint(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Backpack.Add("glow_powder", 1)
	s := NewSlimeLab()
	a.Push(s)
	s.bowl = []string{"glow_powder"}
	s.mix(a)
	if s.boomT == 0 || s.hintKey != "slime.hint.base" {
		t.Fatalf("boom=%d hint=%q", s.boomT, s.hintKey)
	}
	if a.State.Backpack.Count("glow_powder") != 0 {
		t.Fatal("boom should use up the ingredient (E-4)")
	}
}

func TestSlimeGentleFailureKeepsIngredients(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Backpack.Add("slime_base", 1)
	s := NewSlimeLab()
	a.Push(s)
	s.bowl = []string{"slime_base"}
	s.mix(a)
	if a.State.Backpack.Count("slime_base") != 1 {
		t.Fatal("plain base alone shouldn't waste it")
	}
}

func TestCircuitLightsAndDiscovers(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_circuit"})
	c := NewCircuitLab()
	a.Push(c)
	c.chosen, c.phase = "foil", 1
	c.test(a, true)
	if !c.lit || !a.State.Notebook.Has("circuit.foil") || a.Board().IsDone("exp_circuit") == false {
		t.Fatalf("lit=%v notebook=%v", c.lit, a.State.Notebook.Found())
	}
	c.chosen, c.phase = "rubber_band", 1
	c.test(a, true) // wrong prediction
	if c.lit || !a.State.Notebook.Has("circuit.rubber_band") {
		t.Fatal("insulator test")
	}
}

func TestMicroscopeNeedsPondSample(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	ml := NewMicroLab()
	a.Push(ml)
	ml.pick(a, mustSlide("pond"))
	if ml.slide != "" {
		t.Fatal("pond slide opened without a sample")
	}
	a.State.Backpack.Add("pond_sample", 1)
	ml.pick(a, mustSlide("pond"))
	if ml.slide != "pond" || a.State.Backpack.Count("pond_sample") != 0 || len(ml.choices) != 3 {
		t.Fatalf("slide=%q choices=%d", ml.slide, len(ml.choices))
	}
}

func TestMicroscopeExpertHasFourChoices(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Level = content.Expert
	ml := NewMicroLab()
	ml.pick(a, mustSlide("leaf"))
	if len(ml.choices) != 4 {
		t.Fatalf("expert choices = %d", len(ml.choices))
	}
}

func TestToolsUnlockMore(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	ml, g, c := NewMicroLab(), NewGravityLab(), NewCircuitLab()
	n1, n2, n3 := len(ml.slides(a)), len(g.planets(a)), len(c.mats(a))
	a.State.Own("tool:super_microscope")
	a.State.Own("tool:telescope")
	a.State.Own("tool:spark_kit")
	if len(ml.slides(a)) <= n1 || len(g.planets(a)) <= n2 || len(c.mats(a)) <= n3 {
		t.Fatal("tools should unlock more slides, planets, and materials (E-7)")
	}
}

func TestGravityQuestion(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	g := NewGravityLab()
	_, ids, answer := g.question(a)
	if answer != experiment.HighestJump(g.planets(a)) || len(ids) != 3 {
		t.Fatalf("junior answer = %s", answer)
	}
	a.State.Level = content.Expert
	_, _, answer = g.question(a)
	if answer != "6" {
		t.Fatalf("expert answer = %s", answer)
	}
}
