// Package content holds player-facing text with Junior/Expert variants (D-1, D-2).
package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
)

var ErrInvalid = errors.New("invalid text catalog")

// Level is the difficulty chosen by the player.
type Level string

const (
	Junior Level = "junior"
	Expert Level = "expert"
)

// ParseLevel returns the level for s, or Junior and false if unknown.
func ParseLevel(s string) (Level, bool) {
	switch Level(s) {
	case Junior, Expert:
		return Level(s), true
	}
	return Junior, false
}

type entry struct {
	Junior string `json:"junior"`
	Expert string `json:"expert,omitempty"`
}

// Catalog maps text keys to per-level strings.
type Catalog struct{ m map[string]entry }

var verbRe = regexp.MustCompile(`%[a-z]`)

// Load parses {"key": {"junior": "...", "expert": "..."}}. Junior is required;
// both variants must use the same fmt verbs in the same order.
func Load(b []byte) (*Catalog, error) {
	m := map[string]entry{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	for k, e := range m {
		if e.Junior == "" {
			return nil, fmt.Errorf("%w: %q has no junior text", ErrInvalid, k)
		}
		if e.Expert != "" && !slices.Equal(verbRe.FindAllString(e.Junior, -1), verbRe.FindAllString(e.Expert, -1)) {
			return nil, fmt.Errorf("%w: %q junior and expert use different %% verbs", ErrInvalid, k)
		}
	}
	return &Catalog{m: m}, nil
}

// Get returns the text for key at level; Expert falls back to Junior.
// Missing keys return "[key]" so they are easy to spot while playing.
func (c *Catalog) Get(key string, lvl Level) string {
	e, ok := c.m[key]
	if !ok {
		return "[" + key + "]"
	}
	if lvl == Expert && e.Expert != "" {
		return e.Expert
	}
	return e.Junior
}

// Has reports whether key exists.
func (c *Catalog) Has(key string) bool { _, ok := c.m[key]; return ok }

func (c *Catalog) Keys() []string {
	out := make([]string, 0, len(c.m))
	for k := range c.m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
