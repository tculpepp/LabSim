package sfx

import (
	"math"
	"testing"
)

// Every sound the game asks for must exist.
var used = []string{"boom", "buy", "coin", "discover", "fizzle", "jump", "pet", "pickup", "place", "success", "click", "door", "night"}

func TestSoundsExistAndAreSane(t *testing.T) {
	for _, n := range used {
		s, ok := Generate(n)
		if !ok || len(s) == 0 {
			t.Errorf("%s: missing", n)
			continue
		}
		if d := float64(len(s)) / SampleRate; d > 2 {
			t.Errorf("%s: %.1fs is too long", n, d)
		}
		peak := 0.0
		for _, v := range s {
			peak = math.Max(peak, math.Abs(float64(v)))
		}
		if peak > 1 || peak < 0.1 {
			t.Errorf("%s: peak %.2f", n, peak)
		}
		// Gentle on little ears: no hard click at the end.
		if last := math.Abs(float64(s[len(s)-1])); last > 0.05 {
			t.Errorf("%s: ends abruptly (%.2f)", n, last)
		}
	}
	if _, ok := Generate("nope"); ok {
		t.Error("unknown sound")
	}
}

func TestDeterministic(t *testing.T) {
	a, _ := Generate("boom")
	b, _ := Generate("boom")
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("boom differs between calls")
		}
	}
}

func TestPCM(t *testing.T) {
	s := []float32{0, 1, -1, 0.5}
	b := PCM16Stereo(s, 1)
	if len(b) != len(s)*4 {
		t.Fatalf("len = %d", len(b))
	}
	// Sample 1 at full volume = 32767 in both channels, little-endian.
	if b[4] != 0xff || b[5] != 0x7f || b[6] != 0xff || b[7] != 0x7f {
		t.Fatalf("bytes = % x", b[4:8])
	}
	half := PCM16Stereo(s, 0.5)
	if int16(uint16(half[4])|uint16(half[5])<<8) != 16384 {
		t.Fatal("volume not applied")
	}
	if z := PCM16Stereo(s, 0); z[4] != 0 || z[5] != 0 {
		t.Fatal("mute")
	}
}
