package inventory

import (
	"errors"
	"testing"
)

func TestNewBackpackEmpty(t *testing.T) {
	b := NewBackpack(4)
	if b.Slots() != 4 || b.UsedSlots() != 0 {
		t.Fatalf("slots=%d used=%d", b.Slots(), b.UsedSlots())
	}
	if b.Count("star_dust") != 0 {
		t.Fatal("expected zero star dust")
	}
}

func TestAddStacksSameItem(t *testing.T) {
	b := NewBackpack(2)
	mustAdd(t, b, "star_dust", 1)
	mustAdd(t, b, "star_dust", 2)
	if got := b.Count("star_dust"); got != 3 {
		t.Fatalf("Count = %d, want 3", got)
	}
	if b.UsedSlots() != 1 {
		t.Fatalf("UsedSlots = %d, want 1", b.UsedSlots())
	}
}

func TestAddFullBackpackRejectsNewItem(t *testing.T) {
	b := NewBackpack(1)
	mustAdd(t, b, "star_dust", 1)
	if err := b.Add("battery", 1); !errors.Is(err, ErrBackpackFull) {
		t.Fatalf("err = %v, want ErrBackpackFull", err)
	}
	// stacking an existing item still works when full
	mustAdd(t, b, "star_dust", 1)
}

func TestAddRejectsNonPositive(t *testing.T) {
	b := NewBackpack(1)
	if err := b.Add("star_dust", 0); !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("err = %v, want ErrInvalidQuantity", err)
	}
}

func TestRemove(t *testing.T) {
	b := NewBackpack(2)
	mustAdd(t, b, "star_dust", 3)
	if err := b.Remove("star_dust", 2); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if b.Count("star_dust") != 1 {
		t.Fatalf("Count = %d, want 1", b.Count("star_dust"))
	}
}

func TestRemoveAllFreesSlot(t *testing.T) {
	b := NewBackpack(1)
	mustAdd(t, b, "star_dust", 2)
	if err := b.Remove("star_dust", 2); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if b.UsedSlots() != 0 {
		t.Fatalf("UsedSlots = %d, want 0", b.UsedSlots())
	}
	mustAdd(t, b, "battery", 1)
}

func TestRemoveTooManyFails(t *testing.T) {
	b := NewBackpack(1)
	mustAdd(t, b, "star_dust", 1)
	if err := b.Remove("star_dust", 2); !errors.Is(err, ErrNotEnoughItems) {
		t.Fatalf("err = %v, want ErrNotEnoughItems", err)
	}
	if b.Count("star_dust") != 1 {
		t.Fatal("count changed on failed remove")
	}
}

func TestUpgradeAddsSlots(t *testing.T) {
	b := NewBackpack(2)
	b.Upgrade(2)
	if b.Slots() != 4 {
		t.Fatalf("Slots = %d, want 4", b.Slots())
	}
}

func TestItemsSortedByID(t *testing.T) {
	b := NewBackpack(3)
	mustAdd(t, b, "star_dust", 2)
	mustAdd(t, b, "battery", 1)
	items := b.Items()
	if len(items) != 2 || items[0].ID != "battery" || items[1].ID != "star_dust" || items[1].Qty != 2 {
		t.Fatalf("Items = %+v", items)
	}
}

func mustAdd(t *testing.T, b *Backpack, id string, qty int) {
	t.Helper()
	if err := b.Add(id, qty); err != nil {
		t.Fatalf("Add(%q,%d): %v", id, qty, err)
	}
}
