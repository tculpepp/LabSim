// Package items holds the item catalog and daily spawn placement (I-1..I-4).
package items

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"

	"labsim/internal/geom"
	"labsim/internal/world"
)

var ErrInvalid = errors.New("invalid item data")

const (
	SpawnMargin = 36.0 // keep pickups away from walls
	MinSpacing  = 40.0 // keep pickups apart so each is easy to grab
	tries       = 40
)

type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Sell  int    `json:"sell"` // crystals Tom pays for one
}

type Catalog struct {
	list []Item
	byID map[string]Item
}

func Load(b []byte) (*Catalog, error) {
	var doc struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	c := &Catalog{list: doc.Items, byID: map[string]Item{}}
	for _, it := range doc.Items {
		switch {
		case it.ID == "" || it.Name == "":
			return nil, fmt.Errorf("%w: item needs id and name: %+v", ErrInvalid, it)
		case c.byID[it.ID].ID != "":
			return nil, fmt.Errorf("%w: duplicate item %q", ErrInvalid, it.ID)
		case it.Sell < 0:
			return nil, fmt.Errorf("%w: %q has negative price", ErrInvalid, it.ID)
		}
		if _, err := world.ParseHexColor(it.Color); err != nil {
			return nil, fmt.Errorf("%w: %q: %v", ErrInvalid, it.ID, err)
		}
		c.byID[it.ID] = it
	}
	return c, nil
}

func (c *Catalog) Get(id string) (Item, bool) { it, ok := c.byID[id]; return it, ok }
func (c *Catalog) All() []Item                { return c.list }

// Rule spawns Count of Item in Room every in-game day.
type Rule struct {
	Room  string `json:"room"`
	Item  string `json:"item"`
	Count int    `json:"count"`
}

func LoadRules(b []byte, c *Catalog, rooms map[string]geom.Rect) ([]Rule, error) {
	var doc struct {
		Rules []Rule `json:"rules"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	for _, r := range doc.Rules {
		if _, ok := c.Get(r.Item); !ok {
			return nil, fmt.Errorf("%w: rule uses unknown item %q", ErrInvalid, r.Item)
		}
		if _, ok := rooms[r.Room]; !ok {
			return nil, fmt.Errorf("%w: rule uses unknown room %q", ErrInvalid, r.Room)
		}
		if r.Count < 1 {
			return nil, fmt.Errorf("%w: rule count must be positive", ErrInvalid)
		}
	}
	return doc.Rules, nil
}

// Pickup is one item lying on the floor. Key is unique per day.
type Pickup struct {
	Key    string
	ItemID string
	X, Y   float64
}

// Generate places the day's pickups. The same day always gives the same
// result, so pickups don't move when the game is reloaded. free reports
// whether a spot is clear of walls and furniture.
func Generate(rules []Rule, rooms map[string]geom.Rect, day int, free func(x, y float64) bool) []Pickup {
	r := rand.New(rand.NewPCG(uint64(day), 0x5eed))
	var out []Pickup
	for ri, rule := range rules {
		area := rooms[rule.Room].Expand(-SpawnMargin)
		for n := 0; n < rule.Count; n++ {
			for try := 0; try < tries; try++ {
				x := area.X + r.Float64()*area.W
				y := area.Y + r.Float64()*area.H
				if !free(x, y) || tooClose(out, x, y) {
					continue
				}
				out = append(out, Pickup{Key: fmt.Sprintf("d%d-r%d-%d", day, ri, n), ItemID: rule.Item, X: x, Y: y})
				break
			}
		}
	}
	return out
}

func tooClose(ps []Pickup, x, y float64) bool {
	for _, p := range ps {
		dx, dy := p.X-x, p.Y-y
		if dx*dx+dy*dy < MinSpacing*MinSpacing {
			return true
		}
	}
	return false
}
