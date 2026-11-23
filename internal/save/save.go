// Package save handles the versioned save format, storage, and autosave timing (SV-1..SV-3).
package save

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CurrentVersion of the save format. Bump it (and add a Migration) when a
// change would break old saves. Adding optional fields does not need a bump.
const CurrentVersion = 1

var (
	ErrCorrupt       = errors.New("save data is corrupt")
	ErrFutureVersion = errors.New("save is from a newer version of the game")
	ErrNoSave        = errors.New("no save found")
)

type Item struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

// Data is the on-disk save. One scientist per device (SV-2).
type Data struct {
	Version       int      `json:"version"`
	Name          string   `json:"name"`
	Level         string   `json:"level"`
	X             float64  `json:"x"`
	Y             float64  `json:"y"`
	Room          string   `json:"room,omitempty"`
	Crystals      int      `json:"crystals"`
	BackpackSlots int      `json:"backpackSlots"`
	Items         []Item   `json:"items,omitempty"`
	Day           int      `json:"day"`
	DayElapsedMs  int64    `json:"dayElapsedMs"`
	Unlocked      []string `json:"unlocked,omitempty"`
	Flags         []string `json:"flags,omitempty"`

	Avatar map[string]string `json:"avatar,omitempty"`
	Owned  []string          `json:"owned,omitempty"`

	PickedDay int      `json:"pickedDay,omitempty"`
	Picked    []string `json:"picked,omitempty"` // pickup keys grabbed on PickedDay

	JobsDay  int      `json:"jobsDay,omitempty"` // day Dr. Nova handed out Jobs
	Jobs     []string `json:"jobs,omitempty"`
	JobsDone []string `json:"jobsDone,omitempty"`

	Discoveries []string `json:"discoveries,omitempty"`
}

// Migration upgrades raw save JSON from version v to v+1 in place.
type Migration func(raw map[string]any) error

// Codec encodes and decodes saves, migrating old versions.
type Codec struct {
	Version    int
	Migrations map[int]Migration // keyed by the version they upgrade FROM
}

// migrations is the real migration chain. Empty while only v1 exists.
var migrations = map[int]Migration{}

func DefaultCodec() Codec { return Codec{Version: CurrentVersion, Migrations: migrations} }

func corrupt(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrCorrupt, fmt.Sprintf(format, a...))
}

func (c Codec) Encode(d Data) ([]byte, error) {
	d.Version = c.Version
	return json.Marshal(d)
}

func (c Codec) Decode(b []byte) (Data, error) {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return Data{}, corrupt("%v", err)
	}
	vf, ok := raw["version"].(float64)
	if !ok || vf < 1 || vf != float64(int(vf)) {
		return Data{}, corrupt("missing or bad version")
	}
	v := int(vf)
	if v > c.Version {
		return Data{}, fmt.Errorf("%w: v%d > v%d", ErrFutureVersion, v, c.Version)
	}
	for ; v < c.Version; v++ {
		m, ok := c.Migrations[v]
		if !ok {
			return Data{}, corrupt("no migration from v%d", v)
		}
		if err := m(raw); err != nil {
			return Data{}, corrupt("migrating from v%d: %v", v, err)
		}
		raw["version"] = v + 1
	}
	norm, err := json.Marshal(raw)
	if err != nil {
		return Data{}, corrupt("%v", err)
	}
	var d Data
	if err := json.Unmarshal(norm, &d); err != nil {
		return Data{}, corrupt("%v", err)
	}
	return d, validate(d)
}

func validate(d Data) error {
	switch {
	case d.Crystals < 0:
		return corrupt("negative crystals")
	case d.Day < 1:
		return corrupt("day must be at least 1")
	case d.DayElapsedMs < 0:
		return corrupt("negative day time")
	case d.BackpackSlots < 1:
		return corrupt("backpack needs at least one slot")
	case len(d.Items) > d.BackpackSlots:
		return corrupt("more items than backpack slots")
	}
	for _, it := range d.Items {
		if it.ID == "" || it.Qty < 1 {
			return corrupt("bad item %+v", it)
		}
	}
	return nil
}
