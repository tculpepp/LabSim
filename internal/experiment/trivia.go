package experiment

import "math/rand/v2"

// Space Trivia (E-SPACE-2): quick quiz rounds. Question text lives in
// text.json as trivia.<id>.q and trivia.<id>.a0..a<N-1>; a0 is always the
// right answer and Order shuffles it for display.

type TriviaQ struct {
	ID     string
	N      int  // number of answers
	Expert bool // only in Expert rounds
	Order  []int
}

var triviaBank = []TriviaQ{
	{ID: "sun", N: 4},
	{ID: "red_planet", N: 4},
	{ID: "biggest", N: 4},
	{ID: "rings", N: 4},
	{ID: "planets", N: 4},
	{ID: "closest", N: 4},
	{ID: "moonlight", N: 4},
	{ID: "hottest", N: 4},
	{ID: "light_time", N: 4, Expert: true},
	{ID: "venus_day", N: 4, Expert: true},
	{ID: "volcano", N: 4, Expert: true},
	{ID: "black_hole", N: 4, Expert: true},
}

// TriviaBank lists every question (for the notebook and text checks).
func TriviaBank() []TriviaQ { return triviaBank }

// RoundSize is how many questions a round asks.
const RoundSize = 5

// TriviaRound picks n different questions with shuffled answers.
func TriviaRound(seed uint64, expert bool, n int) []TriviaQ {
	r := rand.New(rand.NewPCG(seed, 77))
	var pool []TriviaQ
	for _, q := range triviaBank {
		if expert || !q.Expert {
			pool = append(pool, q)
		}
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if n > len(pool) {
		n = len(pool)
	}
	out := pool[:n]
	for i := range out {
		out[i].Order = r.Perm(out[i].N)
	}
	return out
}

// IsRight reports whether the answer shown in position pos is correct.
func (q TriviaQ) IsRight(pos int) bool { return pos >= 0 && pos < len(q.Order) && q.Order[pos] == 0 }

// TriviaPassed: get more than half right.
func TriviaPassed(right, asked int) bool { return right*2 > asked }
