package scenes

import (
	"testing"

	"labsim/internal/save"
	"labsim/internal/state"
)

func TestJobsFromNovaDeliverAndPay(t *testing.T) {
	store := &save.MemStore{}
	a, p := playing(t, store)
	if a.Board() != nil {
		t.Fatal("jobs before visiting Dr. Nova")
	}
	if st := p.jobTarget(a); st == nil || st.ID != "nova" {
		t.Fatalf("Gizmo should point to Dr. Nova first, got %v", st)
	}
	a.TalkToNova()
	a.Pop() // dialog
	b := a.Board()
	if b == nil || len(b.All()) != len(a.JobPool.FirstDay) {
		t.Fatalf("first-day jobs not given: %v", b)
	}
	// First-day job: 3 star dust to Dr. Nova.
	a.State.Backpack.Add("star_dust", 3)
	p.useStation(a, a.Map.Station("nova"))
	if a.State.Wallet.Crystals() != 15 || a.State.Backpack.Count("star_dust") != 0 {
		t.Fatalf("not paid: crystals=%d dust=%d", a.State.Wallet.Crystals(), a.State.Backpack.Count("star_dust"))
	}
	if !a.State.Flag("tut.paid") {
		t.Fatal("tutorial pay tip not shown")
	}
	// Using the desk again with no items doesn't pay twice.
	a.State.Backpack.Add("star_dust", 3)
	for a.scenes.Len() > 1 {
		a.Pop()
	}
	p.useStation(a, a.Map.Station("nova"))
	if a.State.Wallet.Crystals() != 15 {
		t.Fatal("paid twice")
	}

	// Reload keeps today's progress.
	c := newApp(t, store)
	c.ContinueGame()
	if cb := c.Board(); cb == nil || !cb.IsDone("dust_nova") {
		t.Fatal("job progress lost on reload")
	}

	// Next day: jobs expire, Dr. Nova deals a fresh random set.
	a.State.Clock.Advance(state.DayLength)
	if a.Board() != nil {
		t.Fatal("yesterday's jobs still active")
	}
	got := a.GiveJobs()
	if len(got) != jobsPerDay {
		t.Fatalf("dealt %d jobs, want %d", len(got), jobsPerDay)
	}
}

func TestExperimentDonePaysExperimentJobs(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.GiveJobs([]string{"dust_nova"})
	if a.ExperimentDone("slime") != 0 {
		t.Fatal("paid for a job we don't have")
	}
}
