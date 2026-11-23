package experiment

import "math/rand/v2"

// Microscope Explorer (E-BIO-1): look at a slide and say what you see.

type SlideInfo struct {
	ID     string
	Answer string // option ID
	Super  bool   // needs the Super Microscope (E-7)
	Needs  string // item used up to make the slide, if any
}

var slides = []SlideInfo{
	{ID: "pond", Answer: "creatures", Needs: "pond_sample"},
	{ID: "leaf", Answer: "plant_cells"},
	{ID: "hair", Answer: "scales"},
	{ID: "onion", Answer: "brick_cells"},
	{ID: "salt", Answer: "cubes", Super: true},
	{ID: "sand", Answer: "rocks", Super: true},
	{ID: "wing", Answer: "wing_scales", Super: true},
}

// Options is every answer that can appear as a choice.
var Options = []string{"creatures", "plant_cells", "scales", "brick_cells", "cubes", "rocks", "wing_scales"}

func Slides(super bool) []SlideInfo {
	var out []SlideInfo
	for _, s := range slides {
		if super || !s.Super {
			out = append(out, s)
		}
	}
	return out
}

func Slide(id string) (SlideInfo, bool) {
	for _, s := range slides {
		if s.ID == id {
			return s, true
		}
	}
	return SlideInfo{}, false
}

// Choices returns n options including the answer, shuffled by seed.
// Junior uses 3; Expert uses 4 (D-2).
func Choices(s SlideInfo, n int, seed uint64) []string {
	r := rand.New(rand.NewPCG(seed, 31))
	out := []string{s.Answer}
	for _, i := range r.Perm(len(Options)) {
		if len(out) == n {
			break
		}
		if Options[i] != s.Answer {
			out = append(out, Options[i])
		}
	}
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

type LookResult struct {
	Correct   bool
	Discovery string
	Hint      string
}

func LookAt(s SlideInfo, choice string) LookResult {
	if choice == s.Answer {
		return LookResult{Correct: true, Discovery: "micro." + s.ID}
	}
	return LookResult{Hint: "micro.hint." + s.ID}
}
