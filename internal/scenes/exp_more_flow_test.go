package scenes

import (
	"testing"

	"labsim/internal/content"
	"labsim/internal/experiment"
	"labsim/internal/save"
)

func top(a *App) Scene {
	s, _ := a.scenes.Top()
	return s
}

func TestEveryExperimentBenchOffersItsRoomsExperiments(t *testing.T) {
	a := newApp(t, &save.MemStore{})
	want := map[string][]string{
		"chemistry":      {"slime", "crystal"},
		"physics":        {"circuit", "gforce"},
		"greenhouse":     {"microscope", "garden"},
		"space_lab":      {"gravity", "trivia"},
		"robot_workshop": {"robot"},
	}
	seen := map[string]bool{}
	for _, st := range a.Map.Stations() {
		if st.Kind != "experiment" {
			continue
		}
		seen[st.Room] = true
		got := roomExperiments(st.Room)
		if len(got) != len(want[st.Room]) {
			t.Fatalf("%s: %d experiments, want %v", st.Room, len(got), want[st.Room])
		}
		for i, e := range got {
			if e.id != want[st.Room][i] {
				t.Errorf("%s[%d] = %s", st.Room, i, e.id)
			}
			for _, k := range []string{e.titleKey, pickerDescKey(e.id)} {
				if !a.Text.Has(k) {
					t.Errorf("missing text %s", k)
				}
			}
			if e.open() == nil {
				t.Errorf("%s opens nothing", e.id)
			}
		}
	}
	for room := range want {
		if !seen[room] {
			t.Errorf("no experiment bench in %s", room)
		}
	}
}

func TestBenchWithTwoExperimentsShowsPicker(t *testing.T) {
	a, p := playing(t, &save.MemStore{})
	p.useStation(a, a.Map.Station("exp_slime"))
	pk, ok := top(a).(*ExpPicker)
	if !ok {
		t.Fatalf("top = %T, want picker", top(a))
	}
	pk.choose(a, 1)
	if _, ok := top(a).(*CrystalLab); !ok {
		t.Fatalf("top = %T, want crystal lab", top(a))
	}
	a.Pop()
	if _, ok := top(a).(*Play); !ok {
		t.Fatalf("back from the experiment should return to the lab, got %T", top(a))
	}
}

func TestBenchWithOneExperimentOpensIt(t *testing.T) {
	a, p := playing(t, &save.MemStore{})
	p.useStation(a, a.Map.Station("workshop"))
	if _, ok := top(a).(*RobotLab); !ok {
		t.Fatalf("top = %T, want robot lab", top(a))
	}
}

func TestCrystalGrowsAndPaysJob(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_crystal"})
	c := NewCrystalLab()
	a.Push(c)
	c.powder, c.hot, c.spoons = "sugar", false, 4
	c.grow(a)
	if a.State.Notebook.Has("crystal.sugar") || c.grown != "" {
		t.Fatal("cold sugar shouldn't grow")
	}
	c.hot = true
	c.grow(a)
	if !a.State.Notebook.Has("crystal.sugar") || c.grown != "sugar" || !a.Board().IsDone("exp_crystal") {
		t.Fatalf("grown=%q notebook=%v", c.grown, a.State.Notebook.Found())
	}
}

func TestGForceRidesAndQuestion(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_gforce"})
	g := NewGForceLab()
	a.Push(g)
	g.ride(a, "coaster")
	if g.current != "coaster" || !a.State.Notebook.Has("gforce.coaster") {
		t.Fatal("ride not recorded")
	}
	_, ids, answer := g.question(a)
	if answer != "jet" || len(ids) != len(experiment.Rides()) {
		t.Fatalf("junior answer %s", answer)
	}
	g.answer(a, answer)
	if !g.answered || !a.Board().IsDone("exp_gforce") {
		t.Fatal("right answer should finish the job")
	}
	a.State.Level = content.Expert
	if _, _, ans := g.question(a); ans != "120" {
		t.Fatalf("expert answer %s", ans)
	}
}

func TestTriviaRoundScoresAndPays(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_trivia"})
	tv := NewTriviaLab()
	a.Push(tv)
	tv.start(a)
	if len(tv.round) != experiment.RoundSize {
		t.Fatalf("round len %d", len(tv.round))
	}
	for range tv.round {
		q := tv.round[tv.idx]
		right := 0
		for i, o := range q.Order {
			if o == 0 {
				right = i
			}
		}
		tv.pick(a, right)
		if !a.State.Notebook.Has("trivia." + q.ID) {
			t.Fatalf("no discovery for %s", q.ID)
		}
		tv.sparkleT, tv.pendingFC = 0, ""
		tv.next(a)
	}
	if tv.score != experiment.RoundSize || !tv.done || !a.Board().IsDone("exp_trivia") {
		t.Fatalf("score=%d done=%v", tv.score, tv.done)
	}
}

func TestTriviaWrongAnswerNoDiscovery(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	tv := NewTriviaLab()
	a.Push(tv)
	tv.start(a)
	q := tv.round[0]
	wrong := 0
	for i, o := range q.Order {
		if o != 0 {
			wrong = i
		}
	}
	tv.pick(a, wrong)
	if tv.score != 0 || a.State.Notebook.Has("trivia."+q.ID) || tv.picked != wrong {
		t.Fatal("wrong answer counted")
	}
	tv.pick(a, 0) // no changing your answer
	if tv.picked != wrong {
		t.Fatal("answer changed")
	}
}

func TestGardenOneFlowerBlooms(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"exp_garden"})
	g := NewGardenLab()
	a.Push(g)
	g.setMode(false)
	if len(g.pots) != 1 {
		t.Fatalf("pots %d", len(g.pots))
	}
	g.tool = "daisy"
	g.usePot(a, 0)
	g.nextDay(a) // no water yet
	if g.pots[0].Stage != 0 {
		t.Fatal("grew without water")
	}
	for i := 0; i < experiment.BloomStage; i++ {
		g.tool = toolCan
		g.usePot(a, 0)
		g.nextDay(a)
	}
	if !g.pots[0].Bloomed() || !a.State.Notebook.Has("garden.daisy") || !a.Board().IsDone("exp_garden") {
		t.Fatalf("pot %+v", g.pots[0])
	}
}

func TestFullGardenDiscovery(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	g := NewGardenLab()
	a.Push(g)
	g.setMode(true)
	if len(g.pots) != experiment.GardenSize {
		t.Fatalf("pots %d", len(g.pots))
	}
	for i := range g.pots {
		g.tool = experiment.Flowers[i%len(experiment.Flowers)]
		g.usePot(a, i)
	}
	g.sun = false
	for i := range g.pots {
		g.tool = toolCan
		g.usePot(a, i)
	}
	g.nextDay(a)
	if g.pots[0].Stage != 0 {
		t.Fatal("grew in the dark")
	}
	g.sun = true
	for d := 0; d < experiment.BloomStage; d++ {
		for i := range g.pots {
			if g.pots[i].Water == 0 {
				g.tool = toolCan
				g.usePot(a, i)
			}
		}
		g.nextDay(a)
	}
	if !a.State.Notebook.Has("garden.full") {
		t.Fatalf("pots %+v", g.pots)
	}
}

func TestRobotChallenge(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	r := NewRobotLab()
	a.Push(r)
	r.challenge = "sand"
	r.test(a)
	if r.boomT != 0 || a.State.Notebook.Has("robot.sand") {
		t.Fatal("unfinished robot should just get a hint")
	}
	r.bot = experiment.Robot{"head": "camera", "arms": "claw", "movers": "wheels"}
	r.test(a)
	if r.boomT == 0 || r.hintKey != "robot.hint.sand" {
		t.Fatalf("wheels in sand: boom=%d hint=%q", r.boomT, r.hintKey)
	}
	r.boomT = 0
	r.bot["movers"] = "tracks"
	r.test(a)
	if !r.passed || !a.State.Notebook.Has("robot.sand") {
		t.Fatal("tracks should cross the sand")
	}
	if len(r.challenges(a)) != len(experiment.RobotChallenges(false)) {
		t.Fatal("junior challenges")
	}
}
