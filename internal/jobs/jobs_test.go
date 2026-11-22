package jobs

import (
	"errors"
	"strings"
	"testing"
)

func deliveryJob() Job {
	return Job{
		ID:      "deliver-dust",
		Title:   "Bring 3 jars of star dust to the Space Lab",
		Kind:    KindDeliver,
		ItemID:  "star_dust",
		Qty:     3,
		Station: "exp_gravity",
		Reward:  15,
	}
}

func expJob() Job {
	return Job{ID: "make-slime", Title: "Make a slime", Kind: KindExperiment, Experiment: "slime", Station: "exp_slime", Reward: 15}
}

func TestNewBoardHasOpenJobs(t *testing.T) {
	b := NewBoard([]Job{deliveryJob()}, nil)
	open := b.Open()
	if len(open) != 1 || open[0].ID != "deliver-dust" {
		t.Fatalf("Open = %+v", open)
	}
}

func TestBoardRestoresDone(t *testing.T) {
	b := NewBoard([]Job{deliveryJob(), expJob()}, []string{"make-slime"})
	if len(b.Open()) != 1 || !b.IsDone("make-slime") || b.IsDone("deliver-dust") {
		t.Fatal("done list not restored")
	}
}

func TestCompleteReturnsRewardOnce(t *testing.T) {
	b := NewBoard([]Job{deliveryJob()}, nil)
	reward, err := b.Complete("deliver-dust")
	if err != nil || reward != 15 {
		t.Fatalf("Complete = %d, %v", reward, err)
	}
	if _, err := b.Complete("deliver-dust"); !errors.Is(err, ErrAlreadyDone) {
		t.Fatalf("second Complete err = %v, want ErrAlreadyDone", err)
	}
	if len(b.Open()) != 0 || len(b.Done()) != 1 {
		t.Fatal("completed job still open")
	}
}

func TestCompleteUnknownJob(t *testing.T) {
	b := NewBoard(nil, nil)
	if _, err := b.Complete("nope"); !errors.Is(err, ErrUnknownJob) {
		t.Fatalf("err = %v, want ErrUnknownJob", err)
	}
}

func TestAllDone(t *testing.T) {
	b := NewBoard([]Job{deliveryJob(), expJob()}, nil)
	if b.AllDone() {
		t.Fatal("AllDone true with open jobs")
	}
	b.Complete("deliver-dust")
	b.Complete("make-slime")
	if !b.AllDone() {
		t.Fatal("AllDone false after completing all")
	}
}

func TestDeliveryFor(t *testing.T) {
	b := NewBoard([]Job{deliveryJob(), expJob()}, nil)
	have := map[string]int{"star_dust": 2}
	count := func(id string) int { return have[id] }
	if j := b.DeliveryFor("exp_gravity", count); j != nil {
		t.Fatal("not enough star dust yet")
	}
	have["star_dust"] = 3
	if j := b.DeliveryFor("exp_gravity", count); j == nil || j.ID != "deliver-dust" {
		t.Fatalf("DeliveryFor = %v", j)
	}
	if j := b.DeliveryFor("nova", count); j != nil {
		t.Fatal("wrong station matched")
	}
	b.Complete("deliver-dust")
	if b.DeliveryFor("exp_gravity", count) != nil {
		t.Fatal("done job matched")
	}
}

func TestExperimentJobs(t *testing.T) {
	b := NewBoard([]Job{deliveryJob(), expJob()}, nil)
	if got := b.ExperimentJobs("gravity"); len(got) != 0 {
		t.Fatal("wrong experiment matched")
	}
	got := b.ExperimentJobs("slime")
	if len(got) != 1 || got[0].ID != "make-slime" {
		t.Fatalf("ExperimentJobs = %v", got)
	}
}

func pool(n int) []Job {
	p := make([]Job, n)
	for i := range p {
		p[i] = deliveryJob()
		p[i].ID = string(rune('a' + i))
		p[i].Qty = 1
	}
	return p
}

func TestDealPicksNDistinctJobsDeterministically(t *testing.T) {
	supply := map[string]int{"star_dust": 10}
	a := Deal(pool(6), 3, 42, supply)
	b := Deal(pool(6), 3, 42, supply)
	if len(a) != 3 {
		t.Fatalf("len = %d, want 3", len(a))
	}
	seen := map[string]bool{}
	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatal("same seed gave different deal")
		}
		if seen[a[i].ID] {
			t.Fatalf("duplicate job %q", a[i].ID)
		}
		seen[a[i].ID] = true
	}
}

func TestDealSmallPool(t *testing.T) {
	if got := Deal(pool(1), 3, 1, map[string]int{"star_dust": 9}); len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

// Jobs dealt together must not ask for more items than spawn in a day.
func TestDealRespectsSupply(t *testing.T) {
	big := pool(6)
	for i := range big {
		big[i].Qty = 3
	}
	for seed := uint64(0); seed < 50; seed++ {
		got := Deal(big, 3, seed, map[string]int{"star_dust": 7})
		need := 0
		for _, j := range got {
			need += j.Qty
		}
		if need > 7 {
			t.Fatalf("seed %d: deal needs %d star dust, only 7 spawn", seed, need)
		}
		if len(got) == 0 {
			t.Fatalf("seed %d: dealt nothing", seed)
		}
	}
}

func TestLoadPool(t *testing.T) {
	src := `{"firstDay":["a"],"pool":[
	  {"id":"a","title":"A","kind":"deliver","item":"star_dust","qty":1,"station":"nova","reward":5},
	  {"id":"b","title":"B","kind":"experiment","experiment":"slime","station":"exp_slime","reward":5}]}`
	p, err := LoadPool([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Jobs) != 2 || len(p.FirstDay) != 1 {
		t.Fatalf("pool = %+v", p)
	}
	if j, ok := p.Get("b"); !ok || j.Experiment != "slime" {
		t.Fatal("Get")
	}
	bad := map[string]string{
		"json":           `{`,
		"dup":            strings.Replace(src, `"id":"b"`, `"id":"a"`, 1),
		"bad kind":       strings.Replace(src, `"kind":"deliver"`, `"kind":"dance"`, 1),
		"no reward":      strings.Replace(src, `"reward":5}]`, `"reward":0}]`, 1),
		"deliver no qty": strings.Replace(src, `"qty":1`, `"qty":0`, 1),
		"unknown first":  strings.Replace(src, `"firstDay":["a"]`, `"firstDay":["z"]`, 1),
		"no station":     strings.Replace(src, `"station":"nova",`, ``, 1),
	}
	for name, s := range bad {
		if _, err := LoadPool([]byte(s)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}
