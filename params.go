package music

import (
	"fmt"
	"math/rand"
)

/*
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
*/

type Valuer interface {
	Value(current float64) float64
}

type valuePipe []Valuer

func (v valuePipe) Value(current float64) float64 {
	for _, vl := range v {
		current = vl.Value(current)
	}
	return current
}

func ValuePipe(v ...Valuer) Valuer {
	return valuePipe(v)
}

type Min float64

func (m Min) Value(current float64) float64 {
	if current < float64(m) {
		return float64(m)
	}
	return current
}

type Max float64

func (m Max) Value(current float64) float64 {
	if current > float64(m) {
		return float64(m)
	}
	return current
}

type Add float64

func (m Add) Value(current float64) float64 {
	return current + float64(m)
}

type Multiply float64

func (m Multiply) Value(current float64) float64 {
	return current * float64(m)
}

type setter struct {
	params []Parameter
	valuer []Valuer
}

func (s *setter) Params() map[string]float64 {
	ps := map[string]float64{}

	singleValuer := len(s.valuer) == 1

	for i, p := range s.params {

		for k, v := range p.Params() {
			if singleValuer {
				ps[k] = s.valuer[0].Value(v)
			} else {
				ps[k] = s.valuer[i].Value(v)
			}
		}
	}

	return ps
}

func Set(v Valuer, params ...Parameter) Parameter {
	return &setter{params, []Valuer{v}}
}

func MultiSet(valuerParamsPair ...interface{}) Parameter {
	if len(valuerParamsPair)%2 != 0 {
		panic("must be pairs of a Valuer followed by the param")
	}

	s := &setter{}

	s.valuer = make([]Valuer, len(valuerParamsPair)/2)
	s.params = make([]Parameter, len(valuerParamsPair)/2)

	for i := 0; i < len(valuerParamsPair); {
		// fmt.Printf("i: %d/%d\n", i, len(valuerParamsPair))
		val, valOk := valuerParamsPair[i].(Valuer)
		if !valOk {
			panic(fmt.Sprintf("argument no %d must be Valuer but is %T", i*2, valuerParamsPair[i]))
		}
		param, paramOk := valuerParamsPair[i+1].(Parameter)
		if !paramOk {
			panic(fmt.Sprintf("argument no %d must be Parameter but is %T", (i*2)+1, valuerParamsPair[i+1]))
		}
		s.valuer[i/2] = val
		s.params[i/2] = param
		i += 2
		/*
			if i < len(valuerParamsPair) {
				break
			}
		*/
	}

	return s
}

// Random adds a random float multiplied by the given value
// to the existing value
type Random float64

func (r Random) Value(current float64) float64 {
	return current + (rand.Float64() * float64(r))
}

/*
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
*/

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
