package data

import (
	"testing"

	"labsim/internal/shop"
	"labsim/internal/world"
)

func TestShopCatalog(t *testing.T) {
	c, err := shop.Load(Shop)
	if err != nil {
		t.Fatal(err)
	}
	// Experiments look for these tool keys (E-7).
	for _, id := range []string{"spark_kit", "super_microscope", "telescope"} {
		if e, ok := c.Get(id); !ok || e.Kind != shop.KindTool {
			t.Errorf("missing tool %s", id)
		}
	}
	m, err := world.Load(Map)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range c.OfKind(shop.KindRoom) {
		r := m.Room(e.ID)
		if r == nil || !r.Locked {
			t.Errorf("room entry %s is not a locked room", e.ID)
			continue
		}
		if r.UnlockPrice != e.Price {
			t.Errorf("%s: shop price %d, door says %d", e.ID, e.Price, r.UnlockPrice)
		}
	}
	if len(c.OfKind(shop.KindDecor)) < 6 {
		t.Error("My Lab needs some decorations to buy")
	}
}
