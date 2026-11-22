package data

import (
	"encoding/json"
	"testing"

	"labsim/internal/geom"
	"labsim/internal/items"
	"labsim/internal/jobs"
)

// DailySupply is how many of each item spawn per day (used to deal jobs).
func dailySupply(t *testing.T) map[string]int {
	t.Helper()
	cat, _ := items.Load(Items)
	m := loadMap(t)
	rooms := map[string]geom.Rect{}
	for _, r := range m.Rooms {
		rooms[r.ID] = r.Bounds
	}
	rules, err := items.LoadRules(Spawns, cat, rooms)
	if err != nil {
		t.Fatal(err)
	}
	s := map[string]int{}
	for _, r := range rules {
		s[r.Item] += r.Count
	}
	return s
}

func TestJobsReferToRealThings(t *testing.T) {
	p, err := jobs.LoadPool(Jobs)
	if err != nil {
		t.Fatalf("jobs.json: %v", err)
	}
	m := loadMap(t)
	cat, _ := items.Load(Items)
	supply := dailySupply(t)
	for _, j := range p.Jobs {
		if m.Station(j.Station) == nil {
			t.Errorf("job %q: unknown station %q", j.ID, j.Station)
		}
		if j.Kind == jobs.KindDeliver {
			if _, ok := cat.Get(j.ItemID); !ok {
				t.Errorf("job %q: unknown item %q", j.ID, j.ItemID)
			}
			if supply[j.ItemID] < j.Qty {
				t.Errorf("job %q needs %d %s but only %d spawn a day", j.ID, j.Qty, j.ItemID, supply[j.ItemID])
			}
		}
	}
	need := map[string]int{}
	for _, id := range p.FirstDay {
		j, _ := p.Get(id)
		need[j.ItemID] += j.Qty
	}
	for item, n := range need {
		if item != "" && n > supply[item] {
			t.Errorf("first day needs %d %s, only %d spawn", n, item, supply[item])
		}
	}
}

func TestAnimals(t *testing.T) {
	var doc struct {
		Animals []struct {
			ID, Kind, Room string
			Speed          float64
		}
	}
	if err := json.Unmarshal(Animals, &doc); err != nil {
		t.Fatal(err)
	}
	m := loadMap(t)
	kinds := map[string]bool{}
	for _, an := range doc.Animals {
		if m.Room(an.Room) == nil {
			t.Errorf("animal %q in unknown room %q", an.ID, an.Room)
		}
		kinds[an.Kind] = true
	}
	// The designer picked these lab animals.
	for _, k := range []string{"hamster", "axolotl", "bunny", "butterfly"} {
		if !kinds[k] {
			t.Errorf("missing %s", k)
		}
	}
}
