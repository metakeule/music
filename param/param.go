package param

import (
	"math/rand"

	"github.com/metakeule/music"
	"github.com/metakeule/music/note"
)

func Set(m ...map[string]float64) map[string]float64 {
	r := map[string]float64{}
	for _, mm := range m {
		for k, v := range mm {
			r[k] = v
		}
	}
	return r
}

func Random1(pos string, v music.Voice, key string, add float64, m ...map[string]float64) *randomized {
	return &randomized{
		Voice:     v,
		vals:      Set(m...),
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
	Freq_ = "freq"
	Amp_  = "amp"
	Out_  = "out"
	In_   = "in"
	Gate_ = "gate"
	Pan_  = "pan"
	Dur_  = "dur"
)

func Freq(n float64) map[string]float64 {
	return map[string]float64{Freq_: n}
}

func Amp(v float64) map[string]float64 {
	return map[string]float64{Amp_: v}
}

func Out(v int) map[string]float64 {
	return map[string]float64{Out_: float64(v)}
}

func In(v int) map[string]float64 {
	return map[string]float64{In_: float64(v)}
}

func Gate(v float64) map[string]float64 {
	return map[string]float64{Gate_: v}
}

func Pan(v float64) map[string]float64 {
	return map[string]float64{Pan_: v}
}

func Dur(v float64) map[string]float64 {
	return map[string]float64{Dur_: v}
}

// gets a midi note but sets a frequency
func Note(n note.Note) map[string]float64 {
	return map[string]float64{Freq_: n.Frequency()}
}

type ScaleStepper struct {
	music.Scale
}

func (s *ScaleStepper) At(degree int) map[string]float64 {
	return Freq(s.Frequency(degree))
}

func Degree(s music.Scale, degree int) map[string]float64 {
	return Freq(s.Frequency(degree))
}

var (
	FFFF_ float64 = 0.5
	FFF_  float64 = 0.45
	FF_   float64 = 0.4
	F_    float64 = 0.35
	MF_   float64 = 0.3
	MP_   float64 = 0.25
	P_    float64 = 0.2
	PP_   float64 = 0.15
	PPP_  float64 = 0.1
	PPPP_ float64 = 0.05
)

func FFFF() map[string]float64 { return Amp(FFFF_) }

// forte fortissimo
func FFF() map[string]float64 { return Amp(FFF_) }

// fortissimo
func FF() map[string]float64 { return Amp(FF_) }

// forte
func F() map[string]float64 { return Amp(F_) }

// mezzoforte
func MF() map[string]float64 { return Amp(MF_) }

// mezzopiano
func MP() map[string]float64 { return Amp(MP_) }

// piano
func P() map[string]float64 { return Amp(P_) }

// pianissimo
func PP() map[string]float64 { return Amp(PP_) }

// piano pianissimo
func PPP() map[string]float64 { return Amp(PPP_) }

func PPPP() map[string]float64 { return Amp(PPPP_) }
