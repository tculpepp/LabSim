package experiment

import "math"

// Planet Gravity (E-SPACE-1): real surface gravity values (m/s²).

type PlanetInfo struct {
	ID        string
	G         float64
	Telescope bool // needs the Big Telescope (E-7)
}

var planets = []PlanetInfo{
	{"moon", 1.62, false},
	{"mars", 3.71, false},
	{"earth", 9.81, false},
	{"jupiter", 24.79, true},
	{"pluto", 0.62, true},
}

// JumpSpeed is a kid's take-off speed in m/s (about a 32 cm jump on Earth).
const JumpSpeed = 2.5

func Planets(telescope bool) []PlanetInfo {
	var out []PlanetInfo
	for _, p := range planets {
		if telescope || !p.Telescope {
			out = append(out, p)
		}
	}
	return out
}

func Planet(id string) (PlanetInfo, bool) {
	for _, p := range planets {
		if p.ID == id {
			return p, true
		}
	}
	return PlanetInfo{}, false
}

// JumpHeight is how high you rise: h = v² / 2g.
func JumpHeight(g float64) float64 { return JumpSpeed * JumpSpeed / (2 * g) }

// FallTime is how long a drop from h metres takes: t = sqrt(2h / g).
func FallTime(h, g float64) float64 { return math.Sqrt(2 * h / g) }

// HighestJump is the planet (of those given) where you jump highest.
func HighestJump(ps []PlanetInfo) string {
	best := ps[0]
	for _, p := range ps[1:] {
		if p.G < best.G {
			best = p
		}
	}
	return best.ID
}

// MoonEarthRatio is how many times higher you jump on the Moon (rounded).
func MoonEarthRatio() int {
	m, _ := Planet("moon")
	e, _ := Planet("earth")
	return int(math.Round(JumpHeight(m.G) / JumpHeight(e.G)))
}
