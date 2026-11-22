// Package avatar holds the avatar part catalog and customization rules (A-1..A-6).
package avatar

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalid = errors.New("invalid avatar catalog")

// Avatar maps part ID → option ID, e.g. {"hat": "cap"}.
type Avatar map[string]string

// Clone returns a copy.
func (a Avatar) Clone() Avatar {
	out := make(Avatar, len(a))
	for k, v := range a {
		out[k] = v
	}
	return out
}

type Option struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
	Price int    `json:"price,omitempty"` // 0 = free; otherwise bought in the shop
}

type Part struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Options []Option `json:"options"`
}

type Catalog struct {
	Parts []Part `json:"parts"`
}

// Key is the ownership key for a paid option, used by the shop.
func Key(part, option string) string { return "avatar:" + part + ":" + option }

func Load(b []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if len(c.Parts) == 0 {
		return nil, fmt.Errorf("%w: no parts", ErrInvalid)
	}
	parts := map[string]bool{}
	for _, p := range c.Parts {
		if parts[p.ID] {
			return nil, fmt.Errorf("%w: duplicate part %q", ErrInvalid, p.ID)
		}
		parts[p.ID] = true
		opts, free := map[string]bool{}, false
		for _, o := range p.Options {
			if opts[o.ID] {
				return nil, fmt.Errorf("%w: %s: duplicate option %q", ErrInvalid, p.ID, o.ID)
			}
			opts[o.ID] = true
			if o.Price == 0 {
				free = true
			}
			if o.Color != "" && !validHex(o.Color) {
				return nil, fmt.Errorf("%w: %s.%s: bad color %q", ErrInvalid, p.ID, o.ID, o.Color)
			}
		}
		if !free {
			return nil, fmt.Errorf("%w: part %q has no free option", ErrInvalid, p.ID)
		}
	}
	return &c, nil
}

func validHex(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	_, err := strconv.ParseUint(s[1:], 16, 32)
	return err == nil
}

func (c *Catalog) part(id string) *Part {
	for i := range c.Parts {
		if c.Parts[i].ID == id {
			return &c.Parts[i]
		}
	}
	return nil
}

// Option looks up one option.
func (c *Catalog) Option(part, id string) (Option, bool) {
	if p := c.part(part); p != nil {
		for _, o := range p.Options {
			if o.ID == id {
				return o, true
			}
		}
	}
	return Option{}, false
}

func available(part string, o Option, owned func(string) bool) bool {
	return o.Price == 0 || (owned != nil && owned(Key(part, o.ID)))
}

// Default picks the first free option of every part.
func (c *Catalog) Default() Avatar {
	av := Avatar{}
	for _, p := range c.Parts {
		for _, o := range p.Options {
			if o.Price == 0 {
				av[p.ID] = o.ID
				break
			}
		}
	}
	return av
}

// Normalize drops unknown parts and replaces missing, invalid, or unowned choices with defaults.
func (c *Catalog) Normalize(av Avatar, owned func(string) bool) Avatar {
	out := c.Default()
	for _, p := range c.Parts {
		if o, ok := c.Option(p.ID, av[p.ID]); ok && available(p.ID, o, owned) {
			out[p.ID] = o.ID
		}
	}
	return out
}

// Cycle moves part to the next (dir > 0) or previous available option, wrapping.
func (c *Catalog) Cycle(av Avatar, part string, dir int, owned func(string) bool) Avatar {
	out := av.Clone()
	p := c.part(part)
	if p == nil {
		return out
	}
	n := len(p.Options)
	cur := 0
	for i, o := range p.Options {
		if o.ID == av[part] {
			cur = i
		}
	}
	step := 1
	if dir < 0 {
		step = -1
	}
	for i := 1; i <= n; i++ {
		o := p.Options[((cur+step*i)%n+n)%n]
		if available(part, o, owned) {
			out[part] = o.ID
			return out
		}
	}
	return out
}

// PricedOption is a paid option, listed in the shop.
type PricedOption struct {
	Part, PartName, ID, Name string
	Price                    int
}

func (c *Catalog) Priced() []PricedOption {
	var out []PricedOption
	for _, p := range c.Parts {
		for _, o := range p.Options {
			if o.Price > 0 {
				out = append(out, PricedOption{Part: p.ID, PartName: p.Name, ID: o.ID, Name: o.Name, Price: o.Price})
			}
		}
	}
	return out
}
