package scenes

import (
	"encoding/json"
	"fmt"
	"strings"

	"labsim/data"
	"labsim/internal/jobs"
)

const jobsPerDay = 3 // P-4

// loadJobs reads the job pool and works out daily item supply from spawns.
func (a *App) loadJobs() error {
	p, err := jobs.LoadPool(data.Jobs)
	if err != nil {
		return err
	}
	a.JobPool = p
	a.Supply = map[string]int{}
	for _, r := range a.Spawns {
		a.Supply[r.Item] += r.Count
	}
	return nil
}

// Board is today's job board, or nil if Professor Emily hasn't handed out jobs yet.
func (a *App) Board() *jobs.Board {
	ids, ok := a.State.TodaysJobs()
	if !ok {
		return nil
	}
	var list []jobs.Job
	for _, id := range ids {
		if j, ok := a.JobPool.Get(id); ok {
			list = append(list, j)
		}
	}
	return jobs.NewBoard(list, a.State.DoneJobs())
}

// GiveJobs deals today's jobs. Day one gets a gentle preset pair.
func (a *App) GiveJobs() []jobs.Job {
	var dealt []jobs.Job
	if !a.State.Flag("jobs.first") {
		for _, id := range a.JobPool.FirstDay {
			j, _ := a.JobPool.Get(id)
			dealt = append(dealt, j)
		}
		a.State.SetFlag("jobs.first")
	} else {
		dealt = jobs.Deal(a.JobPool.Jobs, jobsPerDay, uint64(a.State.Clock.Day())*7919, a.Supply)
	}
	ids := make([]string, len(dealt))
	for i, j := range dealt {
		ids[i] = j.ID
	}
	a.State.GiveJobs(ids)
	a.SaveNow()
	return dealt
}

// PayJob marks a job done and pays for it. Returns the reward (0 if already done).
func (a *App) PayJob(j jobs.Job) int {
	b := a.Board()
	if b == nil {
		return 0
	}
	reward, err := b.Complete(j.ID)
	if err != nil {
		return 0
	}
	a.State.MarkJobDone(j.ID)
	a.State.Wallet.Earn(reward)
	a.Sound("coin")
	a.SaveNow() // SV-1: save on job completion
	return reward
}

// TryDeliver hands over items for a delivery job at station, if possible.
func (a *App) TryDeliver(station string) (jobs.Job, int, bool) {
	b := a.Board()
	if b == nil {
		return jobs.Job{}, 0, false
	}
	j := b.DeliveryFor(station, a.State.Backpack.Count)
	if j == nil {
		return jobs.Job{}, 0, false
	}
	if err := a.State.Backpack.Remove(j.ItemID, j.Qty); err != nil {
		return jobs.Job{}, 0, false
	}
	return *j, a.PayJob(*j), true
}

// ExperimentDone pays any of today's jobs for finishing experiment. Returns total paid.
func (a *App) ExperimentDone(experiment string) int {
	b := a.Board()
	if b == nil {
		return 0
	}
	total := 0
	for _, j := range b.ExperimentJobs(experiment) {
		total += a.PayJob(j)
	}
	return total
}

// TalkToNova is what happens when you use Professor Emily's desk (C-1, P-4).
func (a *App) TalkToNova() {
	b := a.Board()
	switch {
	case b == nil:
		dealt := a.GiveJobs()
		var sb strings.Builder
		sb.WriteString(a.Tf("nova.jobs", a.State.Name))
		for _, j := range dealt {
			fmt.Fprintf(&sb, "\n• %s (%d)", j.Title, j.Reward)
		}
		a.Say("nova", "nova.name", sb.String())
	case b.AllDone():
		a.Say("nova", "nova.name", a.T("nova.allDone"))
	default:
		a.Say("nova", "nova.name", a.Tf("nova.keepGoing", len(b.Open())))
	}
}

// animalSpec is one entry in data/animals.json.
type animalSpec struct {
	ID    string  `json:"id"`
	Kind  string  `json:"kind"`
	Room  string  `json:"room"`
	Speed float64 `json:"speed"`
}

func (a *App) loadAnimals() error {
	var doc struct {
		Animals []animalSpec `json:"animals"`
	}
	if err := json.Unmarshal(data.Animals, &doc); err != nil {
		return err
	}
	a.Animals = doc.Animals
	return nil
}
