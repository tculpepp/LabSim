package scenes

import (
	"labsim/data"
	"labsim/internal/decor"
	"labsim/internal/geom"
	"labsim/internal/shop"
)

func (a *App) loadShop() (err error) {
	a.ShopCat, err = shop.Load(data.Shop)
	return err
}

// Outfits lists the paid avatar options as shop entries (A-5).
func (a *App) Outfits() []shop.Entry {
	var out []shop.Entry
	for _, o := range a.Avatars.Priced() {
		out = append(out, shop.Entry{ID: o.Part + ":" + o.ID, Kind: shop.KindOutfit, Name: o.Name, Desc: o.PartName,
			Price: o.Price, Part: o.Part, Option: o.ID})
	}
	return out
}

// Buy spends crystals and grants e, applying its effect.
func (a *App) Buy(e shop.Entry) error {
	s := a.State
	if err := shop.Check(e, s.Wallet.Crystals(), s.Owns); err != nil {
		return err
	}
	if err := s.Wallet.Spend(e.Price); err != nil {
		return err
	}
	s.Own(e.Key())
	switch e.Kind {
	case shop.KindBackpack:
		s.Backpack.Upgrade(e.Slots)
	case shop.KindRoom:
		s.Unlock(e.ID)
	}
	a.Sound("buy")
	a.SaveNow()
	return nil
}

// ---- My Lab grid ----

const labCell = 40.0

// labGrid returns the decorating grid inside My Lab: its top-left in world
// coordinates, size in cells, and cells blocked by furniture or the doorway.
func (a *App) labGrid() (ox, oy float64, w, h int, blocked []decor.Cell) {
	r := a.Map.Room("my_lab")
	if r == nil {
		return 0, 0, 0, 0, nil
	}
	b := r.Bounds
	ox, oy = b.X+20, b.Y+40
	w, h = int((b.W-40)/labCell), int((b.H-60)/labCell)
	var keepOut []geom.Rect
	for _, st := range a.Map.Stations() {
		if st.Room == "my_lab" {
			keepOut = append(keepOut, st.Rect().Expand(10))
		}
	}
	for _, d := range a.Map.Doors() {
		keepOut = append(keepOut, geom.Centered(d.Rect.X+d.Rect.W/2, d.Rect.Y+d.Rect.H/2, d.Rect.W+40, d.Rect.H+70))
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := geom.Rect{X: ox + float64(x)*labCell, Y: oy + float64(y)*labCell, W: labCell, H: labCell}
			for _, k := range keepOut {
				if c.Intersects(k) {
					blocked = append(blocked, decor.Cell{X: x, Y: y})
					break
				}
			}
		}
	}
	return
}

// LabLayout builds the My Lab layout from the save, dropping anything not owned.
func (a *App) LabLayout() *decor.Layout {
	_, _, w, h, blocked := a.labGrid()
	var saved []decor.Placement
	for _, p := range a.State.Decor {
		if a.State.Owns("decor:" + p.ID) {
			saved = append(saved, p)
		}
	}
	return decor.NewLayout(w, h, blocked, saved)
}
