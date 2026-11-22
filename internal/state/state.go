// Package state is the single game-state struct (T-9) and its save mapping.
package state

import (
	"fmt"
	"sort"
	"time"

	"labsim/internal/clock"
	"labsim/internal/content"
	"labsim/internal/economy"
	"labsim/internal/inventory"
	"labsim/internal/save"
)

const (
	StartBackpackSlots = 6
	DayLength          = 10 * time.Minute // P-3; becomes a setting in M10 (ST-1)
)

// State is everything about one scientist's game.
type State struct {
	Name     string
	Level    content.Level
	X, Y     float64
	Room     string
	Wallet   *economy.Wallet
	Backpack *inventory.Backpack
	Clock    *clock.Clock

	unlocked map[string]bool
	flags    map[string]bool
}

// New starts a fresh scientist at (x, y).
func New(name string, lvl content.Level, x, y float64) *State {
	return &State{
		Name: name, Level: lvl, X: x, Y: y,
		Wallet:   economy.NewWallet(0),
		Backpack: inventory.NewBackpack(StartBackpackSlots),
		Clock:    clock.New(DayLength),
		unlocked: map[string]bool{},
		flags:    map[string]bool{},
	}
}

func (s *State) Unlock(roomID string)          { s.unlocked[roomID] = true }
func (s *State) IsUnlocked(roomID string) bool { return s.unlocked[roomID] }
func (s *State) SetFlag(name string)           { s.flags[name] = true }
func (s *State) Flag(name string) bool         { return s.flags[name] }

func keys(m map[string]bool) []string {
	var out []string
	for k, v := range m {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// Snapshot converts the state to save data.
func (s *State) Snapshot() save.Data {
	var items []save.Item
	for _, st := range s.Backpack.Items() {
		items = append(items, save.Item{ID: st.ID, Qty: st.Qty})
	}
	return save.Data{
		Name: s.Name, Level: string(s.Level), X: s.X, Y: s.Y, Room: s.Room,
		Crystals:      s.Wallet.Crystals(),
		BackpackSlots: s.Backpack.Slots(),
		Items:         items,
		Day:           s.Clock.Day(),
		DayElapsedMs:  s.Clock.Elapsed().Milliseconds(),
		Unlocked:      keys(s.unlocked),
		Flags:         keys(s.flags),
	}
}

// FromSave rebuilds state from validated save data.
func FromSave(d save.Data) (*State, error) {
	lvl, _ := content.ParseLevel(d.Level)
	s := New(d.Name, lvl, d.X, d.Y)
	s.Room = d.Room
	s.Wallet = economy.NewWallet(d.Crystals)
	s.Backpack = inventory.NewBackpack(d.BackpackSlots)
	for _, it := range d.Items {
		if err := s.Backpack.Add(it.ID, it.Qty); err != nil {
			return nil, fmt.Errorf("%w: backpack: %v", save.ErrCorrupt, err)
		}
	}
	s.Clock = clock.Restore(DayLength, d.Day, time.Duration(d.DayElapsedMs)*time.Millisecond)
	for _, id := range d.Unlocked {
		s.Unlock(id)
	}
	for _, f := range d.Flags {
		s.SetFlag(f)
	}
	return s, nil
}
