package music

import "math/rand"

func Random1(pos string, v *Voice, key string, add float64, m ...Parameter) *randomized {
	return &randomized{
		Voice:     v,
		vals:      Params(m...).Params(),
		randomKey: key,
		randomAdd: add,
		pos:       pos,
	}
}

type randomized struct {
	*Voice
	vals      map[string]float64
	randomKey string
	randomAdd float64
	pos       string
	//randFunc     func(add float64) float64
}

func (r *randomized) Transform(tr Tracker) {
	val := rand.Float64() + r.randomAdd
	if val > 1 {
		val = 1
	}

	if val < 0 {
		val = 0
	}
	r.vals[r.randomKey] = val
	tr.At(M(r.pos), On(r.Voice, ParamsMap(r.vals)))
}

type param struct {
	name  string
	value float64
}

func (p param) Params() map[string]float64 {
	return map[string]float64{p.name: p.value}
}

func Param(name string, value float64) Parameter {
	return param{name, value}
}

type Freq float64

func (f Freq) Params() map[string]float64 {
	return map[string]float64{"freq": float64(f)}
}

type Amp float64

func (f Amp) Params() map[string]float64 {
	return map[string]float64{"amp": float64(f)}
}

// Dynamic (how it is player in contrast to amplitude)
type Dyn float64

func (f Dyn) Params() map[string]float64 {
	return map[string]float64{"dyn": float64(f)}
}

type Out float64

func (f Out) Params() map[string]float64 {
	return map[string]float64{"out": float64(f)}
}

type In float64

func (f In) Params() map[string]float64 {
	return map[string]float64{"in": float64(f)}
}

type Gate float64

func (f Gate) Params() map[string]float64 {
	return map[string]float64{"gate": float64(f)}
}

type Pan float64

func (f Pan) Params() map[string]float64 {
	return map[string]float64{"pan": float64(f)}
}

type Dur float64

func (f Dur) Params() map[string]float64 {
	return map[string]float64{"dur": float64(f)}
}

type Rate float64

func (f Rate) Params() map[string]float64 {
	return map[string]float64{"rate": float64(f)}
}

type Offset float64

func (f Offset) Params() map[string]float64 {
	return map[string]float64{"offset": float64(f)}
}
