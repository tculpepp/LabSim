package economy

import (
	"errors"
	"testing"
)

func TestNewWalletStartsWithBalance(t *testing.T) {
	w := NewWallet(10)
	if got := w.Crystals(); got != 10 {
		t.Fatalf("Crystals() = %d, want 10", got)
	}
}

func TestEarnAddsCrystals(t *testing.T) {
	w := NewWallet(0)
	if err := w.Earn(25); err != nil {
		t.Fatalf("Earn: %v", err)
	}
	if got := w.Crystals(); got != 25 {
		t.Fatalf("Crystals() = %d, want 25", got)
	}
}

func TestEarnRejectsNonPositive(t *testing.T) {
	w := NewWallet(5)
	for _, amt := range []int{0, -3} {
		if err := w.Earn(amt); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("Earn(%d) err = %v, want ErrInvalidAmount", amt, err)
		}
	}
	if got := w.Crystals(); got != 5 {
		t.Fatalf("balance changed to %d", got)
	}
}

func TestSpendDeducts(t *testing.T) {
	w := NewWallet(30)
	if err := w.Spend(12); err != nil {
		t.Fatalf("Spend: %v", err)
	}
	if got := w.Crystals(); got != 18 {
		t.Fatalf("Crystals() = %d, want 18", got)
	}
}

func TestSpendInsufficientLeavesBalance(t *testing.T) {
	w := NewWallet(5)
	if err := w.Spend(6); !errors.Is(err, ErrNotEnoughCrystals) {
		t.Fatalf("err = %v, want ErrNotEnoughCrystals", err)
	}
	if got := w.Crystals(); got != 5 {
		t.Fatalf("balance changed to %d", got)
	}
}

func TestSpendExactBalance(t *testing.T) {
	w := NewWallet(7)
	if err := w.Spend(7); err != nil {
		t.Fatalf("Spend: %v", err)
	}
	if w.Crystals() != 0 {
		t.Fatalf("want 0, got %d", w.Crystals())
	}
}

func TestSpendRejectsNonPositive(t *testing.T) {
	w := NewWallet(5)
	if err := w.Spend(0); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("err = %v, want ErrInvalidAmount", err)
	}
}

func TestCanAfford(t *testing.T) {
	w := NewWallet(10)
	if !w.CanAfford(10) || w.CanAfford(11) {
		t.Fatal("CanAfford wrong at boundary")
	}
}
