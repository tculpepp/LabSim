package jobs

import (
	"errors"
	"testing"
)

func deliveryJob() Job {
	return Job{
		ID:     "deliver-dust",
		Title:  "Bring 3 jars of star dust to the Space Lab",
		Kind:   KindDeliver,
		ItemID: "star_dust",
		Qty:    3,
		Room:   "space_lab",
		Reward: 15,
	}
}

func TestNewBoardHasOpenJobs(t *testing.T) {
	b := NewBoard([]Job{deliveryJob()})
	open := b.Open()
	if len(open) != 1 || open[0].ID != "deliver-dust" {
		t.Fatalf("Open = %+v", open)
	}
}

func TestCompleteReturnsRewardOnce(t *testing.T) {
	b := NewBoard([]Job{deliveryJob()})
	reward, err := b.Complete("deliver-dust")
	if err != nil || reward != 15 {
		t.Fatalf("Complete = %d, %v", reward, err)
	}
	if _, err := b.Complete("deliver-dust"); !errors.Is(err, ErrAlreadyDone) {
		t.Fatalf("second Complete err = %v, want ErrAlreadyDone", err)
	}
	if len(b.Open()) != 0 {
		t.Fatal("completed job still open")
	}
}

func TestCompleteUnknownJob(t *testing.T) {
	b := NewBoard(nil)
	if _, err := b.Complete("nope"); !errors.Is(err, ErrUnknownJob) {
		t.Fatalf("err = %v, want ErrUnknownJob", err)
	}
}

func TestAllDone(t *testing.T) {
	j2 := deliveryJob()
	j2.ID = "other"
	b := NewBoard([]Job{deliveryJob(), j2})
	if b.AllDone() {
		t.Fatal("AllDone true with open jobs")
	}
	b.Complete("deliver-dust")
	b.Complete("other")
	if !b.AllDone() {
		t.Fatal("AllDone false after completing all")
	}
}

func TestDealPicksNDistinctJobsDeterministically(t *testing.T) {
	pool := make([]Job, 6)
	for i := range pool {
		pool[i] = deliveryJob()
		pool[i].ID = string(rune('a' + i))
	}
	a := Deal(pool, 3, 42)
	b := Deal(pool, 3, 42)
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
	if got := Deal([]Job{deliveryJob()}, 3, 1); len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}
