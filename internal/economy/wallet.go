// Package economy tracks the player's crystals (in-game currency).
package economy

import "errors"

var (
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrNotEnoughCrystals = errors.New("not enough crystals")
)

// Wallet holds the player's crystals.
type Wallet struct {
	crystals int
}

func NewWallet(start int) *Wallet {
	if start < 0 {
		start = 0
	}
	return &Wallet{crystals: start}
}

func (w *Wallet) Crystals() int { return w.crystals }

func (w *Wallet) CanAfford(price int) bool { return price <= w.crystals }

func (w *Wallet) Earn(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	w.crystals += amount
	return nil
}

func (w *Wallet) Spend(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !w.CanAfford(amount) {
		return ErrNotEnoughCrystals
	}
	w.crystals -= amount
	return nil
}
