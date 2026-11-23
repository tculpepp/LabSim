package experiment

// Crystal Growing (E-CHEM-2): dissolve spoons of powder in water, let it
// cool, and crystals grow. Hot water holds much more of most powders.

type CrystalInfo struct {
	ID      string
	Need    int // fewest spoons that make a strong enough solution
	MaxCold int // most spoons that dissolve in cold water
	MaxHot  int // most spoons that dissolve in hot water
}

// MaxSpoons is the most the spoon counter goes up to.
const MaxSpoons = 6

var crystals = []CrystalInfo{
	{"salt", 3, 3, 4},  // salt dissolves about the same hot or cold
	{"sugar", 4, 2, 6}, // rock candy needs lots of sugar in hot water
	{"alum", 2, 1, 4},
	{"epsom", 3, 2, 5},
}

func Crystals() []CrystalInfo { return crystals }

func Crystal(id string) (CrystalInfo, bool) {
	for _, c := range crystals {
		if c.ID == id {
			return c, true
		}
	}
	return CrystalInfo{}, false
}

type CrystalResult struct {
	OK        bool
	Discovery string
	Hint      string // text key explaining what to change
}

// GrowCrystal tries a solution of spoons of powder id in hot or cold water.
func GrowCrystal(id string, hot bool, spoons int) CrystalResult {
	c, ok := Crystal(id)
	switch {
	case !ok:
		return CrystalResult{Hint: "crystal.hint.pick"}
	case spoons < c.Need:
		return CrystalResult{Hint: "crystal.hint.weak"}
	case !hot && spoons > c.MaxCold:
		return CrystalResult{Hint: "crystal.hint.cold"}
	case spoons > c.MaxHot:
		return CrystalResult{Hint: "crystal.hint.pile"}
	}
	return CrystalResult{OK: true, Discovery: "crystal." + id}
}
