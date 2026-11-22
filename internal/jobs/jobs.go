// Package jobs implements Dr. Nova's daily job board (P-2, P-4).
package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
)

var (
	ErrUnknownJob  = errors.New("unknown job")
	ErrAlreadyDone = errors.New("job already done")
	ErrInvalid     = errors.New("invalid jobs data")
)

type Kind string

const (
	KindDeliver    Kind = "deliver"    // bring Qty of ItemID to Station
	KindExperiment Kind = "experiment" // complete Experiment at Station
)

// Job is one task on the board. Reward is paid in crystals.
type Job struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Kind       Kind   `json:"kind"`
	ItemID     string `json:"item,omitempty"`
	Qty        int    `json:"qty,omitempty"`
	Experiment string `json:"experiment,omitempty"`
	Station    string `json:"station"`
	Reward     int    `json:"reward"`
}

// Pool is every job Dr. Nova can hand out, plus the gentle first-day set.
type Pool struct {
	Jobs     []Job    `json:"pool"`
	FirstDay []string `json:"firstDay"`
}

func LoadPool(b []byte) (*Pool, error) {
	var p Pool
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	seen := map[string]bool{}
	for _, j := range p.Jobs {
		switch {
		case j.ID == "" || j.Title == "" || j.Station == "":
			return nil, fmt.Errorf("%w: job needs id, title, station: %+v", ErrInvalid, j)
		case seen[j.ID]:
			return nil, fmt.Errorf("%w: duplicate job %q", ErrInvalid, j.ID)
		case j.Reward <= 0:
			return nil, fmt.Errorf("%w: %q has no reward", ErrInvalid, j.ID)
		case j.Kind == KindDeliver && (j.ItemID == "" || j.Qty < 1):
			return nil, fmt.Errorf("%w: delivery %q needs item and qty", ErrInvalid, j.ID)
		case j.Kind == KindExperiment && j.Experiment == "":
			return nil, fmt.Errorf("%w: %q needs an experiment", ErrInvalid, j.ID)
		case j.Kind != KindDeliver && j.Kind != KindExperiment:
			return nil, fmt.Errorf("%w: %q has unknown kind %q", ErrInvalid, j.ID, j.Kind)
		}
		seen[j.ID] = true
	}
	for _, id := range p.FirstDay {
		if !seen[id] {
			return nil, fmt.Errorf("%w: firstDay lists unknown job %q", ErrInvalid, id)
		}
	}
	return &p, nil
}

func (p *Pool) Get(id string) (Job, bool) {
	for _, j := range p.Jobs {
		if j.ID == id {
			return j, true
		}
	}
	return Job{}, false
}

// Board holds today's jobs and which are done.
type Board struct {
	jobs []Job
	done map[string]bool
}

func NewBoard(jobs []Job, done []string) *Board {
	b := &Board{jobs: jobs, done: map[string]bool{}}
	for _, id := range done {
		b.done[id] = true
	}
	return b
}

func (b *Board) All() []Job { return b.jobs }

func (b *Board) Open() []Job {
	var out []Job
	for _, j := range b.jobs {
		if !b.done[j.ID] {
			out = append(out, j)
		}
	}
	return out
}

func (b *Board) IsDone(id string) bool { return b.done[id] }

func (b *Board) Done() []string {
	var out []string
	for _, j := range b.jobs {
		if b.done[j.ID] {
			out = append(out, j.ID)
		}
	}
	return out
}

func (b *Board) AllDone() bool { return len(b.Open()) == 0 }

// Complete marks a job done and returns its reward.
func (b *Board) Complete(id string) (int, error) {
	for _, j := range b.jobs {
		if j.ID != id {
			continue
		}
		if b.done[id] {
			return 0, ErrAlreadyDone
		}
		b.done[id] = true
		return j.Reward, nil
	}
	return 0, ErrUnknownJob
}

// DeliveryFor returns an open delivery to station that the player can
// fulfil now (count reports how many of an item they carry).
func (b *Board) DeliveryFor(station string, count func(item string) int) *Job {
	for _, j := range b.Open() {
		if j.Kind == KindDeliver && j.Station == station && count(j.ItemID) >= j.Qty {
			j := j
			return &j
		}
	}
	return nil
}

// ExperimentJobs returns open jobs completed by finishing experiment.
func (b *Board) ExperimentJobs(experiment string) []Job {
	var out []Job
	for _, j := range b.Open() {
		if j.Kind == KindExperiment && j.Experiment == experiment {
			out = append(out, j)
		}
	}
	return out
}

// Deal picks up to n distinct jobs from pool. The same seed gives the same
// jobs (seed with the day number). Deliveries dealt together never ask for
// more of an item than supply says spawns in a day.
func Deal(pool []Job, n int, seed uint64, supply map[string]int) []Job {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	left := map[string]int{}
	for k, v := range supply {
		left[k] = v
	}
	var out []Job
	for _, i := range r.Perm(len(pool)) {
		if len(out) == n {
			break
		}
		j := pool[i]
		if j.Kind == KindDeliver {
			if left[j.ItemID] < j.Qty {
				continue
			}
			left[j.ItemID] -= j.Qty
		}
		out = append(out, j)
	}
	return out
}
