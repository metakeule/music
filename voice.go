package music

// Voice can only play one sound at a time

import (
	"bytes"
	"fmt"
	"strings"
)

type Voice struct {
	generator
	instrument           instrument
	scnode               int // the node id of the voice
	Group                int
	mute                 bool
	Bus                  int
	lastInstrumentSample *Sample // the last sample played by a sample instrument
}

type playDur struct {
	pos Measure
	dur Measure
	*Voice
	Params Parameter
}

func (p *playDur) Pattern(t *Track) {
	t.At(p.pos, OnEvent(p.Voice, p.Params))
	t.At(p.pos+p.dur, OffEvent(p.Voice))
}

func (v *Voice) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return &playDur{M(pos), M(dur), v, Params(params...)}
}

type play struct {
	pos Measure
	*Voice
	Params Parameter
}

func (p *play) Pattern(t *Track) {
	t.At(p.pos, OnEvent(p.Voice, p.Params))
}

func (v *Voice) Play(pos string, params ...Parameter) Pattern {
	return &play{M(pos), v, Params(params...)}
}

type exec_ struct {
	pos   Measure
	fn    func(e *Event)
	voice *Voice
	type_ string
}

func (e *exec_) Pattern(t *Track) {
	ev := newEvent(e.voice, e.type_)
	ev.Runner = e.fn
	t.At(e.pos, ev)
}

func (v *Voice) Exec(pos string, type_ string, fn func(t *Event)) Pattern {
	return &exec_{M(pos), fn, v, type_}
}

type stop struct {
	pos Measure
	*Voice
}

func (p *stop) Pattern(t *Track) {
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

func (p *mod) Pattern(t *Track) {
	t.At(p.pos, ChangeEvent(p.Voice, p.Params))
}

type mute struct {
	v    *Voice
	pos  Measure
	mute bool
}

func (m *mute) Pattern(t *Track) {
	if m.mute {
		t.At(m.pos, MuteEvent(m.v))
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
	return &mod{M(pos), v, Params(params...)}
}

func (v *Voice) Metronome(unit Measure, parameter ...Parameter) Pattern {
	return &metronome{voice: v, unit: unit, eventProps: Params(parameter...)}
}

func (v *Voice) Bar(parameter ...Parameter) Pattern {
	return &bar{voice: v, eventProps: Params(parameter...)}
}

type metronome struct {
	last       Measure
	voice      *Voice
	unit       Measure
	eventProps Parameter
}

func (m *metronome) Pattern(t *Track) {
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

func (m *bar) Pattern(t *Track) {
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

func (v *Voice) setMute(ev *Event) {
	v.mute = true
	v.OffEvent(ev)
}

func (v *Voice) unsetMute(*Event) {
	v.mute = false
}

func ratedOffset(sampleOffset float64, params map[string]float64) float64 {
	rate, hasRate := params["rate"]
	if !hasRate || rate == 1 {
		return sampleOffset * (-1)
	}
	return (-1) * sampleOffset / rate
}

func (v *Voice) OnEvent(ev *Event) {

	if _, isBus := v.instrument.(*bus); isBus {
		panic("On not supported for busses")
	}

	if _, isGroup := v.instrument.(group); isGroup {
		panic("On not supported for groups")
	}

	if v.mute {
		return
	}

	if cl, ok := v.instrument.(codeLoader); ok {
		cl.Use()
	}

	params := ev.Params.Params()

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

	oldNode := v.scnode
	_ = oldNode
	v.scnode = v.newNodeId()

	if oldNode != 0 && oldNode > 2000 {
		// fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, oldNode)
		fmt.Fprintf(&ev.sccode, `, [\n_free, %d]`, oldNode)
	}

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
		fmt.Fprintf(
			&ev.sccode,
			`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			v.instrument.Name(),
			v.scnode,
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
		v.lastInstrumentSample = sample
		fmt.Fprintf(
			&ev.sccode,
			`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			fmt.Sprintf("sample%d", sample.Channels),
			v.scnode,
			bufnum,
			v.paramsStr(params),
		)

		ev.offset = ratedOffset(sample.Offset, params) + offsetParam
		return
	}

	fmt.Fprintf(&ev.sccode, `, [\s_new, \%s, %d, 1, %d%s]`, v.instrument.Name(), v.scnode, group, v.paramsStr(params))

}

func (v *Voice) ChangeEvent(ev *Event) {
	if v.scnode == 0 {
		return
	}

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
				fmt.Fprintf(&ev.sccode, `, [\n_map, %d, \%s, %d]`, v.scnode, param, int(val))
			case "_mapa":
				fmt.Fprintf(&ev.sccode, `, [\n_mapa, %d, \%s, %d]`, v.scnode, param, int(val))
			default:
				panic("unknown special parameter must be '_map-[key] or _mapa-[key]")
			}

		}
	}

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

	fmt.Fprintf(&ev.sccode, `, [\n_set, %d%s]`, v.scnode, v.paramsStr(params))
}

func (v *Voice) OffEvent(ev *Event) {
	if _, isBus := v.instrument.(*bus); isBus {
		panic("Off not supported for busses")
	}

	if _, isGroup := v.instrument.(group); isGroup {
		panic("Off not supported for groups")
	}
	if v.scnode == 0 {
		// fmt.Println("can't stop not existing node for instrument " + v.Instrument.Name())
		return
	}

	v.lastInstrumentSample = nil
	fmt.Fprintf(&ev.sccode, `, [\n_set, %d, \gate, -1]`, v.scnode)
}

type codeLoader interface {
	IsUsed() bool
	Use()
}

type voices []*Voice

func Voices(v ...*Voice) voices {
	return voices(v)
}

func (vs voices) Exec(pos string, type_ string, fn func(t *Event)) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Exec(pos, type_, fn))
	}
	return Patterns(ps...)
}

func (vs voices) Modify(pos string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Modify(pos, params...))
	}
	return Patterns(ps...)
}

func (vs voices) PlayDur(pos, dur string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.PlayDur(pos, dur, params...))
	}
	return Patterns(ps...)
}

func (vs voices) Stop(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Stop(pos))
	}
	return Patterns(ps...)
}

func (vs voices) Play(pos string, params ...Parameter) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Play(pos, params...))
	}
	return Patterns(ps...)
}

func (vs voices) Mute(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.Mute(pos))
	}
	return Patterns(ps...)
}

func (vs voices) UnMute(pos string) Pattern {
	ps := []Pattern{}
	for _, v := range vs {
		ps = append(ps, v.UnMute(pos))
	}
	return Patterns(ps...)
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
