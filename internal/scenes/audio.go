package scenes

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"

	"labsim/internal/sfx"
	"labsim/internal/ui"
)

// soundBank holds the synthesized effects ready to play.
type soundBank struct {
	ctx  *audio.Context
	pcm  map[string][]byte
	last map[string]int64 // tick each sound last played (avoid stacking)
}

func newSoundBank() *soundBank {
	b := &soundBank{ctx: audio.NewContext(sfx.SampleRate), pcm: map[string][]byte{}, last: map[string]int64{}}
	for _, n := range sfx.Names() {
		s, _ := sfx.Generate(n)
		b.pcm[n] = sfx.PCM16Stereo(s, 1)
	}
	return b
}

func (b *soundBank) play(name string, vol float64) {
	pcm, ok := b.pcm[name]
	if !ok || vol <= 0 {
		return
	}
	now := ebiten.Tick()
	if t, ok := b.last[name]; ok && now-t < 4 {
		return
	}
	b.last[name] = now
	p := b.ctx.NewPlayerFromBytes(pcm)
	p.SetVolume(vol)
	p.Play()
}

// EnableSound turns on sound effects (AU-1). Called by main, not tests:
// there is only one audio context per program.
func (a *App) EnableSound() {
	bank := newSoundBank()
	a.sfx = func(name string) {
		if a.State == nil {
			bank.play(name, 0.7)
			return
		}
		s := a.State.Settings
		if s.Muted {
			return
		}
		bank.play(name, float64(s.Volume)/100)
	}
	ui.OnActivate = func() { a.Sound("click") }
}
