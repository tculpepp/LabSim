package experiment

// Glowing Slime (E-CHEM-1, magic): slime base + one add-in makes a slime.

var slimeAddIns = map[string]string{
	"glow_powder":   "glow",
	"bouncy_beads":  "bouncy",
	"rainbow_drops": "rainbow",
	"stretch_gel":   "stretchy",
}

// SlimeKinds lists the slimes in notebook order.
var SlimeKinds = []string{"glow", "bouncy", "rainbow", "stretchy"}

type SlimeResult struct {
	OK        bool
	Slime     string
	Discovery string
	Hint      string // text key explaining what to change
	Boom      bool   // funny explosion (E-4)
}

func IsSlimeIngredient(id string) bool {
	_, addIn := slimeAddIns[id]
	return id == "slime_base" || addIn
}

// MixSlime mixes the chosen ingredients.
func MixSlime(ingredients []string) SlimeResult {
	if len(ingredients) == 0 {
		return SlimeResult{Hint: "slime.hint.empty"}
	}
	base := 0
	var adds []string
	for _, id := range ingredients {
		switch {
		case id == "slime_base":
			base++
		case slimeAddIns[id] != "":
			adds = append(adds, id)
		default:
			return SlimeResult{Hint: "slime.hint.odd", Boom: true}
		}
	}
	switch {
	case base == 0:
		return SlimeResult{Hint: "slime.hint.base", Boom: true}
	case len(adds) == 0:
		return SlimeResult{Hint: "slime.hint.addin"}
	case len(adds) > 1:
		return SlimeResult{Hint: "slime.hint.tooMany", Boom: true}
	}
	kind := slimeAddIns[adds[0]]
	return SlimeResult{OK: true, Slime: kind, Discovery: "slime." + kind}
}
