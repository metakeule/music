package music

// Voice can only play one sound at a time

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type ParamsModifier interface {
	Modify(Parameter) Parameter
}

type ParamsModifierFunc func(Parameter) Parameter

func (p ParamsModifierFunc) Modify(param Parameter) Parameter {
	return p(param)
}

type Voice struct {
	generator
	instrument          instrument
	scnode              int // the node id of the voice
	Group               int
	Bus                 int
	mute                bool
	lastSampleFrequency float64 // frequency of the last played sample
	ParamsModifier      ParamsModifier
}

type playDur struct {
	pos Measure
	dur Measure
	*Voice
	Params Parameter
}

func (p *playDur) Pattern(t Tracker) {
	t.At(p.pos, OnEvent(p.Voice, p.Params))
	t.At(p.pos+p.dur, OffEvent(p.Voice))
}

func (v *Voice) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return &playDur{M(pos), M(dur), v, MixParams(params...)}
}

type play struct {
	pos Measure
	*Voice
	Params Parameter
}

func (v *Voice) Sequencer(s Sequencer) Pattern {
	return &sequencer{
		seq: s,
		v:   v,
	}
}

func (p *play) Pattern(t Tracker) {
	t.At(p.pos, OnEvent(p.Voice, p.Params))
}

func (v *Voice) Phrase(pos string) *phrase {
	return newPhrase(v, pos)
}

func (v *Voice) Rhythm(start string, positions ...string) *rhythm {
	if len(positions) < 1 {
		panic("number of positions must be 1 at least")
	}
	return newRhythm(v, start, positions...)
}

func (v *Voice) Play(pos string, params ...Parameter) Pattern {
	return &play{M(pos), v, MixParams(params...)}
}

type exec_ struct {
	pos   Measure
	fn    func(t Tracker) (EventGenerator, Parameter)
	voice *Voice
}

func (e *exec_) Pattern(t Tracker) {
	evGen, param := e.fn(t)

	//ev := newEvent(e.voice, "CUSTOM")
	// ev.Runner = e.fn
	t.At(e.pos, evGen(e.voice, param))
}

func (v *Voice) Exec(pos string, fn func(t Tracker) (EventGenerator, Parameter)) Pattern {
	return &exec_{M(pos), fn, v}
}

type stop struct {
	pos Measure
	*Voice
}

func (p *stop) Pattern(t Tracker) {
	t.At(p.pos, OffEvent(p.Voice))
}

func (v *Voice) Stop(pos string) Pattern {
	return &stop{M(pos), v}
}

type mod struct {
	pos Measure
	*Voice
	Params Parameter
}

func (p *mod) Pattern(t Tracker) {
	t.At(p.pos, ChangeEvent(p.Voice, p.Params))
}

type mute struct {
	v    *Voice
	pos  Measure
	mute bool
}

func (m *mute) Pattern(t Tracker) {
	if m.mute {
		t.At(m.pos, MuteEvent(m.v))
		return
	}
	t.At(m.pos, UnMuteEvent(m.v))
}

func (v *Voice) Mute(pos string) Pattern {
	return &mute{v, M(pos), true}
}

func (v *Voice) UnMute(pos string) Pattern {
	return &mute{v, M(pos), false}
}

func (v *Voice) Modify(pos string, params ...Parameter) Pattern {
	return &mod{M(pos), v, MixParams(params...)}
}

func (v *Voice) Metronome(unit Measure, parameter ...Parameter) Pattern {
	return &metronome{voice: v, unit: unit, eventProps: MixParams(parameter...)}
}

func (v *Voice) Bar(parameter ...Parameter) Pattern {
	return &bar{voice: v, eventProps: MixParams(parameter...)}
}

type metronome struct {
	last       Measure
	voice      *Voice
	unit       Measure
	eventProps Parameter
}

func (m *metronome) Pattern(t Tracker) {
	n := int(t.CurrentBar() / m.unit)
	half := m.unit / 2
	for i := 0; i < n; i++ {
		t.At(m.unit*Measure(i), OnEvent(m.voice, m.eventProps))
		t.At(m.unit*Measure(i)+half, OffEvent(m.voice))
	}
}

type bar struct {
	voice      *Voice
	counter    float64
	eventProps Parameter
}

func (m *bar) Pattern(t Tracker) {
	t.At(M("0"), OnEvent(m.voice, m.eventProps))
	t.At(M("1/8"), OffEvent(m.voice))
	m.counter++
}

func (v *Voice) paramsStr(params map[string]float64) string {
	var buf bytes.Buffer

	for k, v := range params {
		if k[0] != '_' {
			fmt.Fprintf(&buf, `, \%s, %v`, k, float32(v))
		}
	}

	return buf.String()
}

func (v *Voice) donothing(ev *Event) {}

/*
func (v *Voice) setMute(ev *Event) {
	v.mute = true
	v.OffEvent(ev)
}

func (v *Voice) unsetMute(*Event) {
	v.mute = false
}
*/

func ratedOffset(sampleOffset float64, params map[string]float64) float64 {
	rate, hasRate := params["rate"]
	if !hasRate || rate == 1 {
		return sampleOffset * (-1)
	}
	return (-1) * sampleOffset / rate
}

func (v *Voice) ptr() string {
	return fmt.Sprintf("%p", v)[6:]
}

// getCode is executed after the events have been sorted, respecting their offset
func (v *Voice) getCode(ev *Event) string {
	//fmt.Println(ev.Type)
	res := ""
	switch ev.Type {
	case "CUSTOM":
		// fmt.Println("running custom event")
		//ev.Runner(ev)
		return ev.sccode.String()
	case "MUTE":
		// println("muted")
		// fmt.Printf("muting %s\n", v.ptr())
		v.mute = true
	case "UNMUTE":
		// println("unmuted")
		v.mute = false
	case "ON":
		var bf bytes.Buffer
		oldNode := v.scnode
		_, isSample := v.instrument.(*sCSample)
		_, isSampleInstrument := v.instrument.(*sCSampleInstrument)

		if oldNode != 0 && oldNode > 2000 {
			if isSample || isSampleInstrument {
				// is freed automatically
				fmt.Fprintf(&bf, `, [\n_set, %d, \gate, -1]`, oldNode)
			} else {
				fmt.Fprintf(&bf, `, [\n_free, %d]`, oldNode)
			}
			// if oldNode != 0 {
			// fmt.Fprintf(&bf, `, [\n_free, %d]`, oldNode)
		}

		if isSample || isSampleInstrument {
			v.lastSampleFrequency = ev.sampleInstrumentFrequency
		}

		if v.mute {
			// println("muted (On)")
			v.scnode = 0
			return bf.String()
		}
		// fmt.Printf("ON %s\n", v.ptr())

		v.scnode = v.newNodeId()
		//s := strings.Replace(ev.sccode.String(), "##OLD_NODE##", fmt.Sprintf("%d", v.scnode), -1)
		bf.WriteString(strings.Replace(ev.sccode.String(), "##NODE##", fmt.Sprintf("%d", v.scnode), -1))
		//bf.WriteString(ev.sccode.String())
		res = bf.String()
		//fmt.Sprintf(ev.sccode.String(), ...)
	case "OFF":
		v.lastSampleFrequency = 0
		if v.scnode == 0 {
			return ""
		}
		res = strings.Replace(ev.sccode.String(), "##NODE##", fmt.Sprintf("%d", v.scnode), -1)
		//res = ev.sccode.String()
	case "CHANGE":
		if _, isBus := v.instrument.(*bus); isBus {
			return ev.sccode.String()
		}

		if _, isGroup := v.instrument.(group); isGroup {
			return ev.sccode.String()
		}

		if v.scnode == 0 || v.mute {
			return ""
		}

		isSample := false

		if _, ok := v.instrument.(*sCSample); ok {
			isSample = true
		}

		if _, ok := v.instrument.(*sCSampleInstrument); ok {
			isSample = true
		}
		_ = isSample

		if isSample {
			if v.lastSampleFrequency != 0 && ev.changedParamsPrepared["freq"] != 0 && v.lastSampleFrequency != ev.changedParamsPrepared["freq"] {
				if _, isSet := ev.changedParamsPrepared["rate"]; !isSet {
					ev.changedParamsPrepared["rate"] = ev.changedParamsPrepared["freq"] / v.lastSampleFrequency
				}
			}
		}

		var res bytes.Buffer

		res.WriteString(strings.Replace(ev.sccode.String(), "##NODE##", fmt.Sprintf("%d", v.scnode), -1))

		//fmt.Fprintf(&ev.sccode, `, [\n_set, %d%s]`, v.scnode, v.paramsStr(params))
		fmt.Fprintf(&res, `, [\n_set, %d%s]`, v.scnode, v.paramsStr(ev.changedParamsPrepared))

		return res.String()

		//res = ev.sccode.String()
	}

	// fmt.Printf("%s %p %s %s\n", v.instrument.Name(), v, ev.Type, res)
	return res
}

func (v *Voice) OnEvent(ev *Event) {

	if _, isBus := v.instrument.(*bus); isBus {
		panic("On not supported for busses")
	}

	if _, isGroup := v.instrument.(group); isGroup {
		panic("On not supported for groups")
	}

	/*
		if v.mute {
			return
		}
	*/

	if cl, ok := v.instrument.(codeLoader); ok {
		cl.Use()
	}

	params := ev.Params.Params()

	if v.ParamsModifier != nil {
		params = v.ParamsModifier.Modify(ParamsMap(params)).Params()
	}

	groupParam, hasGroupParam := params["group"]

	if hasGroupParam {
		v.Group = int(groupParam)
		delete(params, "group")
	}

	group := 1010

	if v.Group != 0 {
		group = v.Group
	}

	offsetParam, hasOffsetParam := params["offset"]

	if hasOffsetParam {
		delete(params, "offset")
	}

	/*
		oldNode := v.scnode
		_ = oldNode
		v.scnode = v.newNodeId()
	*/

	//if oldNode != 0 && oldNode > 2000 {
	// fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, oldNode)
	//fmt.Fprintf(&ev.sccode, `, [\n_free, %d]`, oldNode)
	//}

	//fmt.Fprintf(&ev.sccode, `, [\n_free, %d]`, oldNode)

	switch i := v.instrument.(type) {
	case *sCInstrument:
		ev.offset = i.Offset + offsetParam
	case *sCSample:
		if i.Sample.Frequency != 0 && params["freq"] != 0 && i.Sample.Frequency != params["freq"] {
			if _, isSet := params["rate"]; !isSet {
				params["rate"] = params["freq"] / i.Sample.Frequency
			}
		}
		bufnum := i.Sample.sCBuffer
		ev.sampleInstrumentFrequency = i.Sample.Frequency
		fmt.Fprintf(
			&ev.sccode,
			//`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			`, [\s_new, \%s, ##NODE##, 0, 0, \bufnum, %d%s]`,
			v.instrument.Name(),
			// v.scnode,
			bufnum,
			v.paramsStr(params),
		)
		ev.offset = ratedOffset(i.Sample.Offset, params) + offsetParam
		return

	case *sCSampleInstrument:
		sample := i.Sample(params)
		if sampleFreq, hasSampleFreq := params["samplefreq"]; hasSampleFreq {
			sample.Frequency = sampleFreq
			delete(params, "samplefreq")
		}

		bufnum := sample.sCBuffer
		ev.sampleInstrumentFrequency = sample.Frequency
		// v.lastInstrumentSample = sample
		fmt.Fprintf(
			&ev.sccode,
			//`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			`, [\s_new, \%s, ##NODE##, 0, 0, \bufnum, %d%s]`,
			fmt.Sprintf("sample%d", sample.Channels),
			// v.scnode,
			bufnum,
			v.paramsStr(params),
		)

		ev.offset = ratedOffset(sample.Offset, params) + offsetParam
		return
	}

	// fmt.Fprintf(&ev.sccode, `, [\s_new, \%s, %d, 1, %d%s]`, v.instrument.Name(), v.scnode, group, v.paramsStr(params))
	fmt.Fprintf(&ev.sccode, `, [\s_new, \%s, ##NODE##, 1, %d%s]`, v.instrument.Name(), group, v.paramsStr(params))

}

func (v *Voice) ChangeEvent(ev *Event) {

	params := ev.Params.Params()

	groupParam, hasGroupParam := params["group"]

	if hasGroupParam {
		v.Group = int(groupParam)
		delete(params, "group")
	}

	offsetParam, hasOffsetParam := params["offset"]

	if hasOffsetParam {
		delete(params, "offset")
	}

	// only respect offset per parameter in change events
	ev.offset = offsetParam

	if _, isBus := v.instrument.(*bus); isBus {
		for name, val := range params {
			busno, has := busses[name]

			if !has {
				panic("unknown bus " + name)
			}
			fmt.Fprintf(&ev.sccode, `, [\c_set, \%d, %v]`, busno, val)
		}
		return
	}

	if _, isGroup := v.instrument.(group); isGroup {
		fmt.Fprintf(&ev.sccode, `, [\n_set, %d%s]`, v.Group, v.paramsStr(params))
		return
	}

	for k, val := range params {
		if k[0] == '_' {
			idx := strings.Index(k, "-")

			if idx == -1 {
				panic("invalid special parameter must be '_map-[key] or _mapa-[key]")
			}

			pre := k[:idx]
			param := k[idx+1:]

			switch pre {
			case "_map":
				//fmt.Fprintf(&ev.sccode, `, [\n_map, %d, \%s, %d]`, v.scnode, param, int(val))
				fmt.Fprintf(&ev.sccode, `, [\n_map, ##NODE##, \%s, %d]`, param, int(val))
			case "_mapa":
				//fmt.Fprintf(&ev.sccode, `, [\n_mapa, %d, \%s, %d]`, v.scnode, param, int(val))
				fmt.Fprintf(&ev.sccode, `, [\n_mapa, ##NODE##, \%s, %d]`, param, int(val))
			default:
				panic("unknown special parameter must be '_map-[key] or _mapa-[key]")
			}
			delete(params, k)
		}
	}

	ev.changedParamsPrepared = params

	/*
		if i, ok := v.instrument.(*sCSample); ok {
			if i.Sample.Frequency != 0 && params["freq"] != 0 && i.Sample.Frequency != params["freq"] {
				if _, isSet := params["rate"]; !isSet {
					params["rate"] = params["freq"] / i.Sample.Frequency
				}
			}
		}

		if _, ok := v.instrument.(*sCSampleInstrument); ok {
			if v.lastInstrumentSample != nil && v.lastInstrumentSample.Frequency != 0 && params["freq"] != 0 && v.lastInstrumentSample.Frequency != params["freq"] {
				if _, isSet := params["rate"]; !isSet {
					params["rate"] = params["freq"] / v.lastInstrumentSample.Frequency
				}
			}
		}

		//fmt.Fprintf(&ev.sccode, `, [\n_set, %d%s]`, v.scnode, v.paramsStr(params))
		fmt.Fprintf(&ev.sccode, `, [\n_set, ##NODE##%s]`, v.paramsStr(params))
	*/
}

func (v *Voice) OffEvent(ev *Event) {
	if _, isBus := v.instrument.(*bus); isBus {
		panic("Off not supported for busses")
	}

	if _, isGroup := v.instrument.(group); isGroup {
		panic("Off not supported for groups")
	}

	// v.lastInstrumentSample = nil
	//fmt.Fprintf(&ev.sccode, `, [\n_set, %d, \gate, -1]`, v.scnode)
	fmt.Fprintf(&ev.sccode, `, [\n_set, ##NODE##, \gate, -1]`)
}

type codeLoader interface {
	IsUsed() bool
	Use()
}

type voices []*Voice

// v may be []*Voice or *Voice
func Voices(v ...interface{}) voices {
	vs := []*Voice{}

	for _, x := range v {
		switch t := x.(type) {
		case *Voice:
			vs = append(vs, t)
		case []*Voice:
			vs = append(vs, t...)
		default:
			panic(fmt.Sprintf("unsupported type %T, supported are *Voice and []*Voice", x))
		}
	}

	return voices(vs)
}

func (vs voices) Exec(pos string, fn func(t Tracker) (EventGenerator, Parameter)) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Exec(pos, fn))
	}
	return MixPatterns(ps...)
}

func (vs voices) Modify(pos string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Modify(pos, params...))
	}
	return MixPatterns(ps...)
}

func (vs voices) PlayDur(pos, dur string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.PlayDur(pos, dur, params...))
	}
	return MixPatterns(ps...)
}

func (vs voices) Stop(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Stop(pos))
	}
	return MixPatterns(ps...)
}

func (vs voices) Play(pos string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Play(pos, params...))
	}
	return MixPatterns(ps...)
}

func (vs voices) Mute(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Mute(pos))
	}
	return MixPatterns(ps...)
}

func (vs voices) UnMute(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.UnMute(pos))
	}
	return MixPatterns(ps...)
}

func (vs voices) SetBus(bus int) {
	if bus < 1 {
		panic("bus number must be > 0")
	}
	for _, v := range vs {
		v.Bus = bus
	}
}

func (vs voices) SetGroup(group int) {
	for _, v := range vs {
		v.Group = group
	}
}

func (vs voices) SetParamsModifier(m ParamsModifier) {
	for _, v := range vs {
		v.ParamsModifier = m
	}
}

/*
MultiSet(Random(20), Offset(15), Random(1.4), Amp(1.5))
*/
// a simple idea to humanize offset by +/10 and amp by +- 0.1
type humanize_V1 struct {
	params       Parameter
	offsetFactor float64
	ampFactor    float64
	freqFactor   float64
}

func (h *humanize_V1) Params() map[string]float64 {
	p := h.params.Params()
	// between 0 and 1

	if h.offsetFactor > 0 {
		src1 := rand.NewSource(time.Now().UTC().UnixNano())
		r1 := rand.New(src1).Float64()

		offsetAdd := (r1 - 0.5) * h.offsetFactor
		if r1 <= 0.5 {
			offsetAdd = r1 * h.offsetFactor * (-1)
		}
		p["offset"] = p["offset"] + offsetAdd
	}

	if h.ampFactor > 0 {
		src2 := rand.NewSource(time.Now().UTC().UnixNano() * time.Now().UTC().UnixNano())
		r2 := rand.New(src2).Float64()

		ampAdd := r2 * h.ampFactor

		p["amp"] = p["amp"] + ampAdd
	}

	/*
		TODO also for freq
		offsetAdd := (r1 - 0.5) * h.offsetFactor
			if r1 < 0.5 {
				offsetAdd = r1 * h.offsetFactor * (-1)
			}
			p["offset"] = p["offset"] + offsetAdd
	*/

	if h.freqFactor > 0 && p["freq"] != 0 {
		// was := p["freq"]
		src3 := rand.NewSource(time.Now().UTC().UnixNano() * time.Now().UTC().UnixNano())
		r3 := rand.New(src3).Float64()

		freqAdd := (r3 - 0.5) * h.freqFactor

		if r3 <= 0.5 {
			freqAdd = r3 * h.freqFactor * (-1)
		}

		if x := p["freq"] + freqAdd; x > 0 {
			p["freq"] = x
			// fmt.Printf("freq: %v => %v \n", was, p["freq"])
		}

	}

	return p
}

type HumanizeV1 struct {
	OffsetFactor float64
	AmpFactor    float64
	FreqFactor   float64
}

func (h HumanizeV1) Modify(params Parameter) Parameter {
	return &humanize_V1{params, h.OffsetFactor, h.AmpFactor, h.FreqFactor}
}
