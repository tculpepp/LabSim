package experiment

// Greenhouse Garden (E-BIO-2): plants need water AND light to grow. Too
// little water and they're thirsty; too much and the roots get soggy.

var Flowers = []string{"sunflower", "tulip", "rose", "daisy"}

const (
	BloomStage = 4 // seed, sprout, leaves, bud, flower
	MaxWater   = 3 // more than this and the pot is soggy
	GardenSize = 6
)

type Pot struct {
	Seed  string
	Stage int
	Water int
}

type Growth int

const (
	GrowOK Growth = iota
	GrowDone
	GrowEmpty
	GrowDry
	GrowDark
	GrowSoggy
)

func (p *Pot) AddWater()    { p.Water++ }
func (p Pot) Soggy() bool   { return p.Water > MaxWater }
func (p Pot) Bloomed() bool { return p.Seed != "" && p.Stage >= BloomStage }
func (p Pot) Planted() bool { return p.Seed != "" }

// NextDay grows the plant one stage if it has water and light. It drinks
// one splash of water a day, and extra water drains away.
func (p *Pot) NextDay(sun bool) Growth {
	switch {
	case p.Seed == "":
		return GrowEmpty
	case p.Bloomed():
		return GrowDone
	case p.Soggy():
		p.Water = MaxWater
		return GrowSoggy
	case p.Water == 0:
		return GrowDry
	case !sun:
		return GrowDark
	}
	p.Water--
	p.Stage++
	return GrowOK
}

// GardenBloomed is true when every pot has a flower.
func GardenBloomed(pots []Pot) bool {
	if len(pots) == 0 {
		return false
	}
	for _, p := range pots {
		if !p.Bloomed() {
			return false
		}
	}
	return true
}
