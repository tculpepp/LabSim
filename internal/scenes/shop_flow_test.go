package scenes

import (
	"errors"
	"testing"

	"labsim/internal/decor"
	"labsim/internal/save"
	"labsim/internal/shop"
)

func TestBoughtKeys(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	for _, c := range shopCats {
		if !a.Text.Has(boughtKey(c.kind)) {
			t.Errorf("no %s text", boughtKey(c.kind))
		}
	}
}

func TestBuyTool(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	kit, _ := a.ShopCat.Get("spark_kit")
	if err := a.Buy(kit); !errors.Is(err, shop.ErrTooPoor) {
		t.Fatalf("broke buy: %v", err)
	}
	a.State.Wallet.Earn(60)
	if err := a.Buy(kit); err != nil {
		t.Fatal(err)
	}
	if a.State.Wallet.Crystals() != 10 || !a.State.Owns("tool:spark_kit") {
		t.Fatal("not granted")
	}
	if len(NewCircuitLab().mats(a)) != 10 {
		t.Fatal("spark kit should add materials")
	}
}

func TestBuyBackpacksInOrder(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Wallet.Earn(1000)
	big, _ := a.ShopCat.Get("bag_big")
	if err := a.Buy(big); !errors.Is(err, shop.ErrNeedsFirst) {
		t.Fatalf("out of order: %v", err)
	}
	plus, _ := a.ShopCat.Get("bag_plus")
	slots := a.State.Backpack.Slots()
	a.Buy(plus)
	a.Buy(big)
	if a.State.Backpack.Slots() != slots+plus.Slots+big.Slots {
		t.Fatalf("slots = %d", a.State.Backpack.Slots())
	}
	if got := a.State.Snapshot().BackpackSlots; got != a.State.Backpack.Slots() {
		t.Fatal("slots not saved")
	}
}

func TestBuyRoomOpensDoor(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Wallet.Earn(500)
	room, _ := a.ShopCat.Get("robot_workshop")
	if err := a.Buy(room); err != nil {
		t.Fatal(err)
	}
	for _, d := range a.Map.Doors() {
		if d.RoomID == "robot_workshop" && a.Map.DoorLocked(d, a.State.IsUnlocked) {
			t.Fatal("door still locked")
		}
	}
}

func TestBuyOutfitWearable(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Wallet.Earn(500)
	var crown shop.Entry
	for _, e := range a.Outfits() {
		if e.Part == "hat" && e.Option == "crown" {
			crown = e
		}
	}
	if crown.Price == 0 {
		t.Fatal("crown not in shop")
	}
	if err := a.Buy(crown); err != nil {
		t.Fatal(err)
	}
	av := a.Avatars.Normalize(map[string]string{"hat": "crown"}, a.State.Owns)
	if av["hat"] != "crown" {
		t.Fatal("bought crown not wearable")
	}
}

func TestShopRowsMarkOwned(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Own("tool:telescope")
	sh := NewShop(a)
	for _, r := range sh.buyRows(a) {
		if r.name == "Big Telescope" && r.ok {
			t.Fatal("owned item still buyable")
		}
	}
}

func TestDecorateSavesLayout(t *testing.T) {
	a, _ := playing(t, &save.MemStore{})
	a.State.Own("decor:plant")
	a.State.Decor = []decor.Placement{{ID: "lamp", Cell: decor.Cell{X: 4, Y: 3}}} // not owned: dropped
	d := NewDecorate(a)
	a.Push(d)
	if len(d.layout.Placements()) != 0 {
		t.Fatal("unowned decor restored")
	}
	_, _, w, h, blocked := a.labGrid()
	if w < 6 || h < 4 || len(blocked) == 0 {
		t.Fatalf("grid %dx%d blocked %d", w, h, len(blocked))
	}
	if err := d.layout.Place("plant", decor.Cell{X: 4, Y: 3}); err != nil {
		t.Fatal(err)
	}
	d.close(a)
	if len(a.State.Decor) != 1 || a.State.Decor[0].ID != "plant" {
		t.Fatalf("decor = %v", a.State.Decor)
	}
	for _, c := range blocked {
		if d.layout.Place("plant", c) == nil {
			t.Fatalf("placed on blocked cell %v", c)
		}
	}
}
