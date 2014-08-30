package music

// Voice can only play one sound at a time

import (
	"bytes"
	"fmt"
	"strings"
)

type Voice struct {
	Generator
	Instrument
	SCNode  int // the node id of the voice
	SCGroup int
	mute    bool
	Bus     int
}

func (v *Voice) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return PlayDur(pos, dur, v, params...)
}

func (v *Voice) Play(pos string, params ...Parameter) Pattern {
	return Play(pos, v, params...)
}

func (v *Voice) Stop(pos string) Pattern {
	return Stop(pos, v)
}

func (v *Voice) Modify(pos string, params ...Parameter) Pattern {
	return Modify(pos, v, params...)
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

func (v *Voice) Mute(ev *Event) {
	v.mute = true
	v.Off(ev)
}

func (v *Voice) UnMute(*Event) {
	v.mute = false
}

func ratedOffset(sampleOffset float64, params map[string]float64) float64 {
	rate, hasRate := params["rate"]
	if !hasRate || rate == 1 {
		return sampleOffset * (-1)
	}
	return (-1) * sampleOffset / rate
}

func (v *Voice) On(ev *Event) {

	if _, isBus := v.Instrument.(*Bus); isBus {
		panic("On not supported for busses")
	}

	if _, isGroup := v.Instrument.(Group); isGroup {
		panic("On not supported for groups")
	}

	if v.mute {
		return
	}

	if cl, ok := v.Instrument.(CodeLoader); ok {
		cl.Use()
	}

	params := ev.Params.Params()

	groupParam, hasGroupParam := params["group"]

	if hasGroupParam {
		v.SCGroup = int(groupParam)
		delete(params, "group")
	}

	group := 1010

	if v.SCGroup != 0 {
		group = v.SCGroup
	}

	offsetParam, hasOffsetParam := params["offset"]

	if hasOffsetParam {
		delete(params, "offset")
	}

	oldNode := v.SCNode
	_ = oldNode
	v.SCNode = v.NewNodeId()

	if oldNode != 0 && oldNode > 2000 {
		// if oldNode != 0 {
		// fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, oldNode)
		fmt.Fprintf(&ev.SCCode, `, [\n_free, %d]`, oldNode)
	}

	switch i := v.Instrument.(type) {
	case *SCInstrument:
		/*
			if oldNode != 0 && oldNode > 2000 {
				// fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, oldNode)
				fmt.Fprintf(&ev.SCCode, `, [\n_free, %d]`, oldNode)
			}
		*/
		ev.Offset = i.Offset + offsetParam
	case *SCSample:
		if i.Freq != 0 && params["freq"] != 0 && i.Freq != params["freq"] {
			if _, isSet := params["rate"]; !isSet {
				params["rate"] = params["freq"] / i.Freq
			}
		}

		/*
			if oldNode != 0 && oldNode > 2000 {
				// fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, oldNode)
				fmt.Fprintf(&ev.SCCode, `, [\n_free, %d]`, oldNode)
			}
		*/
		bufnum := i.Sample.SCBuffer
		fmt.Fprintf(
			&ev.SCCode,
			`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			v.Instrument.Name(),
			v.SCNode,
			bufnum,
			v.paramsStr(params),
		)
		ev.Offset = ratedOffset(i.Sample.Offset, params) + offsetParam
		return

	case *SCSampleInstrument:
		sample := i.Sample(params)
		bufnum := sample.SCBuffer
		fmt.Fprintf(
			&ev.SCCode,
			`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
			fmt.Sprintf("sample%d", sample.Channels),
			v.SCNode,
			bufnum,
			v.paramsStr(params),
		)

		ev.Offset = ratedOffset(sample.Offset, params) + offsetParam
		return
	}

	fmt.Fprintf(&ev.SCCode, `, [\s_new, \%s, %d, 1, %d%s]`, v.Instrument.Name(), v.SCNode, group, v.paramsStr(params))

}

func (v *Voice) Change(ev *Event) {
	if v.SCNode == 0 {
		// fmt.Println("can't change not existing node for instrument " + v.Instrument.Name())
		return
	}

	params := ev.Params.Params()

	groupParam, hasGroupParam := params["group"]

	if hasGroupParam {
		v.SCGroup = int(groupParam)
		delete(params, "group")
	}

	offsetParam, hasOffsetParam := params["offset"]

	if hasOffsetParam {
		delete(params, "offset")
	}

	// only respect offset per parameter in change events
	ev.Offset = offsetParam

	if _, isBus := v.Instrument.(*Bus); isBus {
		for name, val := range params {
			busno, has := busses[name]

			if !has {
				panic("unknown bus " + name)
			}
			fmt.Fprintf(&ev.SCCode, `, [\c_set, \%d, %v]`, busno, val)
		}
		return
	}

	if _, isGroup := v.Instrument.(Group); isGroup {
		fmt.Fprintf(&ev.SCCode, `, [\n_set, %d%s]`, v.SCGroup, v.paramsStr(params))
		return
	}

	// give it a chance to modify the params, e.g. rate
	/*
		if si, ok := v.Instrument.(*SCSampleInstrument); ok {
			si.SamplePath(si.instrument, ev.Params)
		}
	*/

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
				fmt.Fprintf(&ev.SCCode, `, [\n_map, %d, \%s, %d]`, v.SCNode, param, int(val))
			case "_mapa":
				fmt.Fprintf(&ev.SCCode, `, [\n_mapa, %d, \%s, %d]`, v.SCNode, param, int(val))
			default:
				panic("unknown special parameter must be '_map-[key] or _mapa-[key]")
			}

		}
	}

	if i, ok := v.Instrument.(*SCSample); ok {
		if i.Freq != 0 && params["freq"] != 0 && i.Freq != params["freq"] {
			if _, isSet := params["rate"]; !isSet {
				params["rate"] = params["freq"] / i.Freq
			}
		}
	}

	fmt.Fprintf(&ev.SCCode, `, [\n_set, %d%s]`, v.SCNode, v.paramsStr(params))
}

func (v *Voice) Off(ev *Event) {
	if _, isBus := v.Instrument.(*Bus); isBus {
		panic("Off not supported for busses")
	}

	if _, isGroup := v.Instrument.(Group); isGroup {
		panic("Off not supported for groups")
	}
	if v.SCNode == 0 {
		// fmt.Println("can't stop not existing node for instrument " + v.Instrument.Name())
		return
	}

	fmt.Fprintf(&ev.SCCode, `, [\n_set, %d, \gate, -1]`, v.SCNode)
}

type CodeLoader interface {
	IsUsed() bool
	Use()
}
