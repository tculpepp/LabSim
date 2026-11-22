// Package names builds safe preset names like "Doctor Sparkle Beaker" (A-7, S-3).
package names

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
)

var ErrInvalid = errors.New("invalid names data")

type Builder struct {
	Titles     []string `json:"titles"`
	Adjectives []string `json:"adjectives"`
	Nouns      []string `json:"nouns"`
}

// Pick is an index into each list; out-of-range values wrap.
type Pick struct{ Title, Adjective, Noun int }

func Load(b []byte) (*Builder, error) {
	var nb Builder
	if err := json.Unmarshal(b, &nb); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if len(nb.Titles) == 0 || len(nb.Adjectives) == 0 || len(nb.Nouns) == 0 {
		return nil, fmt.Errorf("%w: every list needs at least one word", ErrInvalid)
	}
	return &nb, nil
}

func wrap(i, n int) int { return ((i % n) + n) % n }

// Normalize wraps a pick into range.
func (b *Builder) Normalize(p Pick) Pick {
	return Pick{wrap(p.Title, len(b.Titles)), wrap(p.Adjective, len(b.Adjectives)), wrap(p.Noun, len(b.Nouns))}
}

func (b *Builder) Build(p Pick) string {
	p = b.Normalize(p)
	return b.Titles[p.Title] + " " + b.Adjectives[p.Adjective] + " " + b.Nouns[p.Noun]
}

// Random returns a pick chosen from seed ("Surprise me!").
func (b *Builder) Random(seed uint64) Pick {
	r := rand.New(rand.NewPCG(seed, seed*31+7))
	return Pick{r.IntN(len(b.Titles)), r.IntN(len(b.Adjectives)), r.IntN(len(b.Nouns))}
}

// All lists every possible name (used to test them all against the filter).
func (b *Builder) All() []string {
	var out []string
	for i := range b.Titles {
		for j := range b.Adjectives {
			for k := range b.Nouns {
				out = append(out, b.Build(Pick{i, j, k}))
			}
		}
	}
	return out
}
