// Package sfx synthesizes LabSim's sound effects (AU-1). No audio files:
// every sound is built from simple waves, so it's tiny and CC0 by default.
package sfx

import (
	"math"
	"math/rand/v2"
)

const SampleRate = 44100

type note struct {
	freq, start, dur float64 // Hz, seconds, seconds
	wave             func(phase float64) float64
	vol              float64
	slide            float64 // Hz per second
}

func sine(p float64) float64     { return math.Sin(2 * math.Pi * p) }
func triangle(p float64) float64 { return 4*math.Abs(p-math.Floor(p+0.5)) - 1 }
func square(p float64) float64 {
	if p-math.Floor(p) < 0.5 {
		return 0.6
	}
	return -0.6
}

// render mixes notes, each with a quick attack and smooth release.
func render(total float64, notes []note) []float32 {
	n := int(total * SampleRate)
	out := make([]float32, n)
	for _, nt := range notes {
		phase := 0.0
		s0 := int(nt.start * SampleRate)
		len_ := int(nt.dur * SampleRate)
		for i := 0; i < len_ && s0+i < n; i++ {
			t := float64(i) / SampleRate
			f := nt.freq + nt.slide*t
			phase += f / SampleRate
			env := envelope(t, nt.dur)
			out[s0+i] += float32(nt.wave(phase) * env * nt.vol)
		}
	}
	return clip(fadeOut(out))
}

func envelope(t, dur float64) float64 {
	const attack = 0.005
	if t < attack {
		return t / attack
	}
	return math.Pow(1-t/dur, 2)
}

func fadeOut(s []float32) []float32 {
	k := min(len(s), SampleRate/100)
	for i := 0; i < k; i++ {
		s[len(s)-1-i] *= float32(i) / float32(k)
	}
	return s
}

func clip(s []float32) []float32 {
	for i, v := range s {
		s[i] = float32(math.Max(-1, math.Min(1, float64(v))))
	}
	return s
}

func noise(dur, vol, lowpass float64, seed uint64) []float32 {
	r := rand.New(rand.NewPCG(seed, 7))
	n := int(dur * SampleRate)
	out := make([]float32, n)
	y := 0.0
	for i := range out {
		t := float64(i) / SampleRate
		y += lowpass * (r.Float64()*2 - 1 - y)
		out[i] = float32(y * vol * envelope(t, dur))
	}
	return out
}

func mix(a, b []float32) []float32 {
	if len(b) > len(a) {
		a, b = b, a
	}
	out := append([]float32(nil), a...)
	for i, v := range b {
		out[i] += v
	}
	return clip(fadeOut(out))
}

// Note frequencies.
const (
	C5 = 523.25
	E5 = 659.25
	G5 = 783.99
	C6 = 1046.5
	E6 = 1318.5
	G6 = 1568.0
	A4 = 440.0
)

var sounds = map[string]func() []float32{
	"click":  func() []float32 { return render(0.06, []note{{1200, 0, 0.05, sine, 0.35, 0}}) },
	"pickup": func() []float32 { return render(0.18, []note{{600, 0, 0.16, triangle, 0.5, 2400}}) },
	"coin": func() []float32 {
		return render(0.35, []note{{E6, 0, 0.08, square, 0.25, 0}, {G6 * 1.0, 0.07, 0.26, square, 0.25, 0}})
	},
	"buy": func() []float32 {
		return render(0.5, []note{{G5, 0, 0.1, triangle, 0.45, 0}, {C6, 0.08, 0.12, triangle, 0.45, 0}, {E6, 0.16, 0.3, triangle, 0.45, 0}})
	},
	"success": func() []float32 {
		return render(0.7, []note{{C5, 0, 0.15, triangle, 0.5, 0}, {E5, 0.1, 0.15, triangle, 0.5, 0}, {G5, 0.2, 0.15, triangle, 0.5, 0}, {C6, 0.3, 0.4, triangle, 0.55, 0}})
	},
	"discover": func() []float32 {
		ns := []note{}
		for i, f := range []float64{C6, E6, G6, C6 * 2, G6, C6 * 2} {
			ns = append(ns, note{f, float64(i) * 0.07, 0.25, sine, 0.35, 0})
		}
		return render(0.7, ns)
	},
	"boom": func() []float32 {
		thump := render(0.6, []note{{110, 0, 0.5, sine, 0.8, -150}})
		return mix(thump, noise(0.9, 0.9, 0.08, 1))
	},
	"fizzle": func() []float32 {
		return render(0.45, []note{{500, 0, 0.42, triangle, 0.4, -700}, {506, 0, 0.42, triangle, 0.3, -700}})
	},
	"jump": func() []float32 { return render(0.3, []note{{300, 0, 0.28, square, 0.35, 1500}}) },
	"pet": func() []float32 {
		return render(0.3, []note{{A4 * 2, 0, 0.12, sine, 0.45, 300}, {A4 * 2.5, 0.1, 0.18, sine, 0.45, 200}})
	},
	"place": func() []float32 {
		return mix(render(0.15, []note{{180, 0, 0.12, sine, 0.6, -300}}), noise(0.08, 0.3, 0.2, 2))
	},
	"door": func() []float32 {
		return render(0.35, []note{{220, 0, 0.3, triangle, 0.45, 180}, {330, 0.05, 0.28, triangle, 0.3, 150}})
	},
	"night": func() []float32 {
		return render(1.2, []note{{G5, 0, 0.5, sine, 0.35, 0}, {E5, 0.3, 0.5, sine, 0.35, 0}, {C5, 0.6, 0.6, sine, 0.4, 0}})
	},
}

// Names lists every sound.
func Names() []string {
	out := make([]string, 0, len(sounds))
	for n := range sounds {
		out = append(out, n)
	}
	return out
}

// Generate returns mono samples in [-1, 1] at SampleRate.
func Generate(name string) ([]float32, bool) {
	f, ok := sounds[name]
	if !ok {
		return nil, false
	}
	return f(), true
}

// PCM16Stereo converts samples to 16-bit little-endian stereo bytes, the
// format Ebitengine's audio player expects. vol is 0..1.
func PCM16Stereo(s []float32, vol float64) []byte {
	out := make([]byte, len(s)*4)
	for i, v := range s {
		x := int16(math.Round(float64(v) * vol * 32767))
		lo, hi := byte(uint16(x)), byte(uint16(x)>>8)
		out[i*4], out[i*4+1], out[i*4+2], out[i*4+3] = lo, hi, lo, hi
	}
	return out
}
