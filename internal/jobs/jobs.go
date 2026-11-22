// Package jobs implements Dr. Nova's daily job board.
package jobs

import (
	"errors"
	"math/rand/v2"
)

var (
	ErrUnknownJob  = errors.New("unknown job")
	ErrAlreadyDone = errors.New("job already done")
)

type Kind string

const (
	KindDeliver    Kind = "deliver"    // bring Qty of ItemID to Room
	KindExperiment Kind = "experiment" // complete an experiment in Room
)

// Job is one task on the board. Reward is paid in crystals.
type Job struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Kind   Kind   `json:"kind"`
	ItemID string `json:"item,omitempty"`
	Qty    int    `json:"qty,omitempty"`
	Room   string `json:"room"`
	Reward int    `json:"reward"`
}

// Board holds today's jobs and which are done.
type Board struct {
	jobs []Job
	done map[string]bool
}

func NewBoard(jobs []Job) *Board {
	return &Board{jobs: jobs, done: map[string]bool{}}
}

func (b *Board) Open() []Job {
	var out []Job
	for _, j := range b.jobs {
		if !b.done[j.ID] {
			out = append(out, j)
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

// Deal picks up to n distinct jobs from pool. Same seed → same jobs (seed with the day number).
func Deal(pool []Job, n int, seed uint64) []Job {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	idx := r.Perm(len(pool))
	if n > len(pool) {
		n = len(pool)
	}
	out := make([]Job, n)
	for i := 0; i < n; i++ {
		out[i] = pool[idx[i]]
	}
	return out
}
