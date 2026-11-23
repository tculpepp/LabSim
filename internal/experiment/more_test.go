package experiment

import "testing"

// ---- Crystal Growing (E-CHEM-2) ----

func TestCrystalNeedsHotWaterAndEnoughSpoons(t *testing.T) {
	for _, c := range Crystals() {
		r := GrowCrystal(c.ID, true, c.Need)
		if !r.OK || r.Discovery != "crystal."+c.ID {
			t.Errorf("%s: hot water with %d spoons should grow: %+v", c.ID, c.Need, r)
		}
		if r := GrowCrystal(c.ID, true, c.MaxHot); !r.OK {
			t.Errorf("%s: %d spoons in hot water should still dissolve", c.ID, c.MaxHot)
		}
		if r := GrowCrystal(c.ID, true, c.Need-1); r.OK || r.Hint != "crystal.hint.weak" {
			t.Errorf("%s: too few spoons: %+v", c.ID, r)
		}
		if r := GrowCrystal(c.ID, true, c.MaxHot+1); r.OK || r.Hint != "crystal.hint.pile" {
			t.Errorf("%s: too many spoons: %+v", c.ID, r)
		}
	}
}

func TestCrystalColdWater(t *testing.T) {
	// Sugar barely dissolves enough in cold water; salt is about the same hot or cold.
	if r := GrowCrystal("sugar", false, 4); r.OK || r.Hint != "crystal.hint.cold" {
		t.Fatalf("cold sugar: %+v", r)
	}
	salt, _ := Crystal("salt")
	if r := GrowCrystal("salt", false, salt.Need); !r.OK {
		t.Fatalf("cold salt should work: %+v", r)
	}
	if r := GrowCrystal("nope", true, 3); r.OK {
		t.Fatal("unknown powder")
	}
}

func TestCrystalRangesAreSane(t *testing.T) {
	for _, c := range Crystals() {
		if c.Need < 1 || c.Need > c.MaxHot || c.MaxHot > MaxSpoons || c.MaxCold > c.MaxHot {
			t.Errorf("%s: bad ranges %+v", c.ID, c)
		}
	}
}

// ---- G-Force Explorer (E-PHYS-2) ----

func TestRides(t *testing.T) {
	if len(Rides()) < 4 {
		t.Fatal("need some rides")
	}
	if MostG(Rides()) != "jet" {
		t.Fatalf("most g = %s", MostG(Rides()))
	}
	still, _ := Ride("still")
	if still.G != 1 {
		t.Fatal("sitting still is 1 g")
	}
	if got := FeltWeight(30, 4); got != 120 {
		t.Fatalf("FeltWeight = %v", got)
	}
	if _, ok := Ride("bogus"); ok {
		t.Fatal("bogus ride")
	}
}

// ---- Space Trivia (E-SPACE-2) ----

func TestTriviaRound(t *testing.T) {
	j := TriviaRound(1, false, 5)
	if len(j) != 5 {
		t.Fatalf("round len %d", len(j))
	}
	seen := map[string]bool{}
	for _, q := range j {
		if q.Expert {
			t.Errorf("junior round got expert question %s", q.ID)
		}
		if seen[q.ID] {
			t.Errorf("duplicate %s", q.ID)
		}
		seen[q.ID] = true
		if len(q.Order) != q.N {
			t.Errorf("%s: order %v", q.ID, q.Order)
		}
	}
	if a, b := TriviaRound(7, true, 5), TriviaRound(7, true, 5); a[0].ID != b[0].ID {
		t.Fatal("same seed should give same round")
	}
	hasExpert := false
	for seed := uint64(0); seed < 20; seed++ {
		for _, q := range TriviaRound(seed, true, 5) {
			hasExpert = hasExpert || q.Expert
		}
	}
	if !hasExpert {
		t.Fatal("expert rounds never include expert questions")
	}
}

func TestTriviaAnswer(t *testing.T) {
	q := TriviaRound(3, false, 1)[0]
	right := -1
	for i, o := range q.Order {
		if o == 0 {
			right = i
		}
	}
	if right < 0 || !q.IsRight(right) {
		t.Fatalf("answer 0 must be somewhere in %v", q.Order)
	}
	if q.IsRight((right + 1) % q.N) {
		t.Fatal("wrong answer accepted")
	}
	if !TriviaPassed(3, 5) || TriviaPassed(2, 5) {
		t.Fatal("pass mark is 3 of 5")
	}
}

// ---- Greenhouse Garden (E-BIO-2) ----

func TestFlowerGrowsWithWaterAndSun(t *testing.T) {
	p := Pot{Seed: "tulip"}
	for day := 0; day < BloomStage; day++ {
		p.AddWater()
		if got := p.NextDay(true); got != GrowOK {
			t.Fatalf("day %d: %v", day, got)
		}
	}
	if !p.Bloomed() {
		t.Fatalf("stage %d", p.Stage)
	}
	if got := p.NextDay(true); got != GrowDone {
		t.Fatalf("bloomed pot: %v", got)
	}
}

func TestFlowerProblems(t *testing.T) {
	p := Pot{Seed: "rose"}
	if got := p.NextDay(true); got != GrowDry || p.Stage != 0 {
		t.Fatalf("no water: %v", got)
	}
	p.AddWater()
	if got := p.NextDay(false); got != GrowDark || p.Stage != 0 {
		t.Fatalf("no sun: %v", got)
	}
	for i := 0; i <= MaxWater; i++ {
		p.AddWater()
	}
	if !p.Soggy() {
		t.Fatal("overwatered pot should be soggy")
	}
	if got := p.NextDay(true); got != GrowSoggy || p.Stage != 0 {
		t.Fatalf("soggy: %v", got)
	}
	if p.Soggy() {
		t.Fatal("extra water should drain away overnight")
	}
	empty := Pot{}
	if got := empty.NextDay(true); got != GrowEmpty {
		t.Fatalf("empty: %v", got)
	}
}

func TestGardenFull(t *testing.T) {
	g := make([]Pot, GardenSize)
	if GardenBloomed(g) {
		t.Fatal("empty garden")
	}
	for i := range g {
		g[i] = Pot{Seed: Flowers[i%len(Flowers)], Stage: BloomStage}
	}
	if !GardenBloomed(g) {
		t.Fatal("all bloomed")
	}
	g[2].Stage--
	if GardenBloomed(g) {
		t.Fatal("one not bloomed")
	}
}

// ---- Build a Robot (RW-2) ----

func TestRobotChallenges(t *testing.T) {
	j, e := RobotChallenges(false), RobotChallenges(true)
	if len(j) < 3 || len(e) <= len(j) {
		t.Fatalf("junior %d expert %d", len(j), len(e))
	}
	for _, c := range e {
		bot := Robot{}
		for slot, part := range c.Need {
			bot[slot] = part
		}
		for _, slot := range RobotSlots {
			if bot[slot] == "" {
				bot[slot] = RobotParts[slot][0]
			}
		}
		r := TryRobot(bot, c)
		if !r.OK || r.Discovery != "robot."+c.ID {
			t.Errorf("%s: right parts failed: %+v", c.ID, r)
		}
	}
}

func TestRobotWrongOrMissingParts(t *testing.T) {
	sand := mustChallenge(t, "sand")
	if r := TryRobot(Robot{"head": "camera", "arms": "claw"}, sand); r.OK || r.Hint != "robot.hint.missing" || r.Boom {
		t.Fatalf("missing movers: %+v", r)
	}
	if r := TryRobot(Robot{"head": "camera", "arms": "claw", "movers": "wheels"}, sand); r.OK || r.Hint != "robot.hint.sand" || !r.Boom {
		t.Fatalf("wheels in sand: %+v", r)
	}
	for _, slot := range RobotSlots {
		if len(RobotParts[slot]) < 3 {
			t.Errorf("slot %s needs 3 parts", slot)
		}
	}
}

func mustChallenge(t *testing.T, id string) RobotChallenge {
	t.Helper()
	for _, c := range RobotChallenges(true) {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no challenge %s", id)
	return RobotChallenge{}
}
