// Package inventory implements the player's slot-limited backpack.
package inventory

import (
	"errors"
	"sort"
)

var (
	ErrBackpackFull    = errors.New("backpack is full")
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrNotEnoughItems  = errors.New("not enough items")
)

// Stack is one backpack slot: an item ID and how many.
type Stack struct {
	ID  string
	Qty int
}

// Backpack holds item stacks. Each distinct item uses one slot; stacks are unbounded.
type Backpack struct {
	slots int
	items map[string]int
}

func NewBackpack(slots int) *Backpack {
	return &Backpack{slots: slots, items: map[string]int{}}
}

func (b *Backpack) Slots() int     { return b.slots }
func (b *Backpack) UsedSlots() int { return len(b.items) }
func (b *Backpack) Count(id string) int {
	return b.items[id]
}

// Upgrade adds extra slots (bought in the shop).
func (b *Backpack) Upgrade(extra int) {
	if extra > 0 {
		b.slots += extra
	}
}

func (b *Backpack) Add(id string, qty int) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	if _, ok := b.items[id]; !ok && len(b.items) >= b.slots {
		return ErrBackpackFull
	}
	b.items[id] += qty
	return nil
}

func (b *Backpack) Remove(id string, qty int) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	have := b.items[id]
	if have < qty {
		return ErrNotEnoughItems
	}
	if have == qty {
		delete(b.items, id)
	} else {
		b.items[id] = have - qty
	}
	return nil
}

// Items returns stacks sorted by ID for stable display.
func (b *Backpack) Items() []Stack {
	out := make([]Stack, 0, len(b.items))
	for id, q := range b.items {
		out = append(out, Stack{ID: id, Qty: q})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
