package experiment

// Circuit Builder (E-PHYS-1): put a material in the gap of a simple series
// circuit (battery, wires, bulb). Predict, then test, whether the bulb lights.

type Mat struct {
	ID       string
	Conducts bool
	Kit      bool // needs the Spark Kit tool (E-7)
}

var materials = []Mat{
	{"paperclip", true, false},
	{"coin", true, false},
	{"foil", true, false},
	{"rubber_band", false, false},
	{"plastic_spoon", false, false},
	{"wood_stick", false, false},
	{"pencil", true, true}, // graphite conducts: a classic surprise!
	{"key", true, true},
	{"eraser", false, true},
	{"marble", false, true},
}

// Materials returns the materials on the tray; kit adds the Spark Kit set.
func Materials(kit bool) []Mat {
	var out []Mat
	for _, m := range materials {
		if kit || !m.Kit {
			out = append(out, m)
		}
	}
	return out
}

func Material(id string) (Mat, bool) {
	for _, m := range materials {
		if m.ID == id {
			return m, true
		}
	}
	return Mat{}, false
}

// Lights reports whether the bulb lights: a series circuit needs every part
// to conduct and the switch to be closed.
func Lights(material string, switchOn bool) bool {
	m, ok := Material(material)
	return ok && m.Conducts && switchOn
}

type CircuitResult struct {
	Lit       bool
	Correct   bool // prediction matched
	Discovery string
	Hint      string
}

// TestCircuit runs the circuit and checks the player's prediction.
func TestCircuit(material string, switchOn, predictLit bool) CircuitResult {
	lit := Lights(material, switchOn)
	r := CircuitResult{Lit: lit, Correct: lit == predictLit, Discovery: "circuit." + material}
	if !r.Correct {
		m, _ := Material(material)
		switch {
		case !switchOn && m.Conducts:
			r.Hint = "circuit.hint.switch"
		case m.Conducts:
			r.Hint = "circuit.hint.conductor"
		default:
			r.Hint = "circuit.hint.insulator"
		}
	}
	return r
}
