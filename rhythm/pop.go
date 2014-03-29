package rhythm

import (
	"github.com/metakeule/music"
)

type popRhythm struct {
	FirstAccentFactor  float32
	SecondAccentFactor float32
	NoteAccentFactor   float32
	NormalFactor       float32
}

func NewPop(firstAccentFactor, secondAccentFactor, noteAccentFactor, normalFactor float32) music.Rhythm {
	return &popRhythm{firstAccentFactor, secondAccentFactor, noteAccentFactor, normalFactor}
}

func (p *popRhythm) Amplitude(bar *music.Bar, pos uint, accent bool) float32 {
	if pos == 0 {
		if accent {
			return p.NoteAccentFactor * p.FirstAccentFactor
		}
		return p.FirstAccentFactor
	}

	if pos*2 == bar.NumBeats {
		if accent {
			return p.NoteAccentFactor * p.SecondAccentFactor
		}
		return p.SecondAccentFactor
	}

	if accent {
		return p.NoteAccentFactor
	}

	return p.NormalFactor
}
func (p *popRhythm) Delay(bar *music.Bar, pos uint) int {
	return 0
}
