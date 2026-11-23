package scenes

import (
	"testing"

	"labsim/internal/save"
)

// standAtWorkshopDoor puts the player next to the Robot Workshop's locked door.
func standAtWorkshopDoor(t *testing.T, a *App, p *Play) {
	t.Helper()
	for _, d := range a.Map.Doors() {
		if d.RoomID == "robot_workshop" {
			p.s.X, p.s.Y = d.Rect.Center()
			return
		}
	}
	t.Fatal("no workshop door")
}

func TestLockedDoorOffersToOpenWhenAffordable(t *testing.T) {
	a, p := playing(t, &save.MemStore{})
	a.State.Wallet.Earn(141)
	standAtWorkshopDoor(t, a, p)
	p.interact(a)
	top, _ := a.scenes.Top()
	c, ok := top.(*Confirm)
	if !ok {
		t.Fatalf("top scene = %T, want a Confirm to open the door", top)
	}
	c.onYes(a)
	if !a.State.IsUnlocked("robot_workshop") || a.State.Wallet.Crystals() != 141-135 {
		t.Fatalf("unlocked=%v crystals=%d", a.State.IsUnlocked("robot_workshop"), a.State.Wallet.Crystals())
	}
}

func TestLockedDoorSaysHowManyMore(t *testing.T) {
	a, p := playing(t, &save.MemStore{})
	a.State.Wallet.Earn(100)
	standAtWorkshopDoor(t, a, p)
	p.interact(a)
	top, _ := a.scenes.Top()
	d, ok := top.(*Dialog)
	if !ok {
		t.Fatalf("top scene = %T, want Gizmo's dialog", top)
	}
	if want := a.Tf("door.locked", "Robot Workshop", 135, 35); d.text != want {
		t.Fatalf("said %q, want %q", d.text, want)
	}
}
