package param

import (
	"math/rand"

	"github.com/metakeule/music"
	"github.com/metakeule/music/note"
)

func Merge(m ...map[string]float64) map[string]float64 {
	r := map[string]float64{}
	for _, mm := range m {
		for k, v := range mm {
			r[k] = v
		}
	}
	return r
}

func SetRandom1(pos string, v music.Voice, key string, add float64, m ...map[string]float64) *randomized {
	return &randomized{
		Voice:     v,
		vals:      Merge(m...),
		randomKey: key,
		randomAdd: add,
		pos:       pos,
	}
}

type randomized struct {
	music.Voice
	vals      map[string]float64
	randomKey string
	randomAdd float64
	pos       string
	//randFunc     func(add float64) float64
}

func (r *randomized) Transform(tr music.Tracker) {
	val := rand.Float64() + r.randomAdd
	if val > 1 {
		val = 1
	}

	if val < 0 {
		val = 0
	}
	r.vals[r.randomKey] = val
	tr.At(music.M(r.pos), music.On(r.Voice, r.vals))
}

var (
	Freq = "freq"
	Amp  = "amp"
	Out  = "out"
	In   = "in"
	Gate = "gate"
	Pan  = "pan"
	Dur  = "dur"
)

func SetFreq(n float64) map[string]float64 {
	return map[string]float64{Freq: n}
}

func SetAmp(v float64) map[string]float64 {
	return map[string]float64{Amp: v}
}

func SetOut(v int) map[string]float64 {
	return map[string]float64{Out: float64(v)}
}

func SetIn(v int) map[string]float64 {
	return map[string]float64{In: float64(v)}
}

func SetGate(v float64) map[string]float64 {
	return map[string]float64{Gate: v}
}

func SetPan(v float64) map[string]float64 {
	return map[string]float64{Pan: v}
}

func SetDur(v float64) map[string]float64 {
	return map[string]float64{Dur: v}
}

// gets a midi note but sets a frequency
func SetNote(n note.Note) map[string]float64 {
	return map[string]float64{Freq: n.Frequency()}
}
