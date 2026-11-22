package scenes

import (
	"testing"

	"labsim/internal/content"
	"labsim/internal/save"
)

func playing(t *testing.T, store save.Store) (*App, *Play) {
	t.Helper()
	a := newApp(t, store)
	a.StartNewGame("x", content.Junior, nil)
	a.Pop() // close Gizmo's welcome dialog
	p, ok := a.scenes.Top()
	if !ok {
		t.Fatal("no scene")
	}
	return a, p.(*Play)
}

func TestGrabPickup(t *testing.T) {
	store := &save.MemStore{}
	a, p := playing(t, store)
	p.refreshPickups(a)
	if len(p.pickups) == 0 {
		t.Fatal("no pickups generated")
	}
	pk := p.pickups[0]
	p.s.X, p.s.Y = pk.X, pk.Y
	p.grab(a)
	if p.s.Backpack.Count(pk.ItemID) != 1 || !p.s.Picked(pk.Key) {
		t.Fatalf("did not grab %+v", pk)
	}
	p.grab(a) // standing still must not grab it twice
	if p.s.Backpack.Count(pk.ItemID) != 1 {
		t.Fatal("grabbed twice")
	}
	a.SaveNow()
	b := newApp(t, store)
	b.ContinueGame()
	if !b.State.Picked(pk.Key) {
		t.Fatal("picked item came back after reload on the same day")
	}
}

func TestGrabWithFullBackpack(t *testing.T) {
	a, p := playing(t, &save.MemStore{})
	for i := 0; i < p.s.Backpack.Slots(); i++ {
		p.s.Backpack.Add(string(rune('a'+i)), 1)
	}
	p.refreshPickups(a)
	pk := p.pickups[0]
	p.s.X, p.s.Y = pk.X, pk.Y
	p.grab(a)
	if p.s.Picked(pk.Key) || p.banner != a.T("bag.full") {
		t.Fatalf("full backpack: picked=%v banner=%q", p.s.Picked(pk.Key), p.banner)
	}
}

func TestSellToTom(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Backpack.Add("moon_rock", 2)
	sh := NewShop(a)
	sh.tab = 1
	rows := sh.sellRows(a)
	if len(rows) != 1 {
		t.Fatalf("rows = %d", len(rows))
	}
	rows[0].act(a)
	it, _ := a.Items.Get("moon_rock")
	if a.State.Wallet.Crystals() != it.Sell || a.State.Backpack.Count("moon_rock") != 1 {
		t.Fatalf("crystals=%d rocks=%d", a.State.Wallet.Crystals(), a.State.Backpack.Count("moon_rock"))
	}
}
