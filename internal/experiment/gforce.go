package experiment

// G-Force Explorer (E-PHYS-2): how hard rides push you, in g (1 g is the
// push you feel sitting still on Earth). Typical peak values.

type RideInfo struct {
	ID string
	G  float64
}

var rides = []RideInfo{
	{"still", 1},
	{"elevator", 1.2},
	{"rocket", 3},  // Space Shuttle launch peaked at about 3 g
	{"coaster", 4}, // bottom of a big loop
	{"jet", 9},     // fighter pilot in a tight turn
}

func Rides() []RideInfo { return rides }

func Ride(id string) (RideInfo, bool) {
	for _, r := range rides {
		if r.ID == id {
			return r, true
		}
	}
	return RideInfo{}, false
}

// MostG is the ride (of those given) that pushes hardest.
func MostG(rs []RideInfo) string {
	best := rs[0]
	for _, r := range rs[1:] {
		if r.G > best.G {
			best = r
		}
	}
	return best.ID
}

// FeltWeight is how heavy someone of mass kg feels at g g's.
func FeltWeight(kg, g float64) float64 { return kg * g }
