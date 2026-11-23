// Package shop holds Tom's catalog and the rules for buying (P-5, E-7).
package shop

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalid    = errors.New("invalid shop data")
	ErrOwned      = errors.New("already owned")
	ErrTooPoor    = errors.New("not enough crystals")
	ErrNeedsFirst = errors.New("needs an earlier upgrade first")
)

type Kind string

const (
	KindTool     Kind = "tool"     // unlocks experiment extras (E-7)
	KindBackpack Kind = "backpack" // extra slots
	KindRoom     Kind = "room"     // unlocks a locked room
	KindDecor    Kind = "decor"    // placed in My Lab (P-6)
	KindOutfit   Kind = "outfit"   // priced avatar options (built from the avatar catalog)
)

type Entry struct {
	ID       string `json:"id"`
	Kind     Kind   `json:"kind"`
	Name     string `json:"name"`
	Desc     string `json:"desc,omitempty"`
	Price    int    `json:"price"`
	Slots    int    `json:"slots,omitempty"`
	Requires string `json:"requires,omitempty"`
	Part     string `json:"-"`
	Option   string `json:"-"`
}

// Key is the ownership key stored in the save.
func (e Entry) Key() string {
	if e.Kind == KindOutfit {
		return "avatar:" + e.Part + ":" + e.Option
	}
	return string(e.Kind) + ":" + e.ID
}

type Catalog struct {
	Entries []Entry `json:"entries"`
}

func Load(b []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	seen := map[string]bool{}
	for _, e := range c.Entries {
		switch {
		case e.ID == "" || e.Name == "":
			return nil, fmt.Errorf("%w: entry needs id and name: %+v", ErrInvalid, e)
		case seen[e.ID]:
			return nil, fmt.Errorf("%w: duplicate %q", ErrInvalid, e.ID)
		case e.Price <= 0:
			return nil, fmt.Errorf("%w: %q has no price", ErrInvalid, e.ID)
		case e.Kind != KindTool && e.Kind != KindBackpack && e.Kind != KindRoom && e.Kind != KindDecor:
			return nil, fmt.Errorf("%w: %q has kind %q", ErrInvalid, e.ID, e.Kind)
		case e.Kind == KindBackpack && e.Slots <= 0:
			return nil, fmt.Errorf("%w: backpack %q adds no slots", ErrInvalid, e.ID)
		}
		seen[e.ID] = true
	}
	for _, e := range c.Entries {
		if e.Requires != "" && !seen[e.Requires] {
			return nil, fmt.Errorf("%w: %q requires unknown %q", ErrInvalid, e.ID, e.Requires)
		}
	}
	return &c, nil
}

func (c *Catalog) Get(id string) (Entry, bool) {
	for _, e := range c.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

func (c *Catalog) OfKind(k Kind) []Entry {
	var out []Entry
	for _, e := range c.Entries {
		if e.Kind == k {
			out = append(out, e)
		}
	}
	return out
}

// Check says whether e can be bought now. Requires refers to an entry of the
// same kind (e.g. backpack upgrades in order).
func Check(e Entry, crystals int, owns func(key string) bool) error {
	if owns(e.Key()) {
		return ErrOwned
	}
	if e.Requires != "" && !owns(Entry{ID: e.Requires, Kind: e.Kind}.Key()) {
		return ErrNeedsFirst
	}
	if crystals < e.Price {
		return ErrTooPoor
	}
	return nil
}
