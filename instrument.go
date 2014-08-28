package music

import (
	"bytes"
	"fmt"

	"strings"
)

// TODO: perhaps use groups instead of instrNumbers

type instrument struct {
	name   string
	bus    bool
	sc     *sc
	offset int
}

func (in *instrument) Name() string {
	return in.name
}

func (in *instrument) New(num int) []Voice {
	v := make([]Voice, num)
	for i := 0; i < num; i++ {
		in.sc.instrNumber++
		name := fmt.Sprintf("%s-%d", in.name, i)
		vc := &voice{
			name:       name,
			instrument: in,
			num:        i,
			instrNum:   in.sc.instrNumber,
		}
		v[i] = vc
		in.sc.voicesToNum[name] = in.sc.instrNumber
		in.sc.numToVoices[in.sc.instrNumber] = vc
	}
	return v
}

type voice struct {
	instrument *instrument
	name       string
	// voice number
	num         int
	instrNum    int
	initialized bool
	mute        bool
	group       *group
}

type Groupable interface {
	SetGroup(group *group)
}

func (v *voice) SetGroup(group *group) {
	v.group = group
}

func (v *voice) paramsStr(ev *Event) string {
	var buf bytes.Buffer

	for k, v := range ev.FinalParams() {
		if k[0] != '_' {
			fmt.Fprintf(&buf, `, \%s, %v`, k, float32(v))
		}
	}

	return buf.String()

}

func (v *voice) PlayDur(pos, dur string, params ...Parameter) Transformer {
	return PlayDur(pos, dur, v, params...)
}

func (v *voice) Play(pos string, params ...Parameter) Transformer {
	return Play(pos, v, params...)
}

func (v *voice) Stop(pos string) Transformer {
	return Stop(pos, v)
}

func (v *voice) Modify(pos string, params ...Parameter) Transformer {
	return Modify(pos, v, params...)
}

func (v *voice) On(ev *Event) {
	v.instrument.sc.instrNumber++
	if v.instrument.bus {
		fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, 1200%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
		return
	}
	if v.instrNum > 2000 {
		fmt.Fprintf(v.instrument.sc.buffer, `, [\n_free, %d]`, v.instrNum)
	}
	v.instrNum = v.instrument.sc.instrNumber

	if v.mute {
		return
	}

	group := 1010

	if v.group != nil {
		group = v.group.Id()
	}

	fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, %d%s]`, v.instrument.name, v.instrNum, group, v.paramsStr(ev))
}

func (v *voice) Change(ev *Event) {
	// handle bus mapping
	for k, val := range ev.FinalParams() {
		if k[0] == '_' {
			idx := strings.Index(k, "-")

			if idx == -1 {
				panic("invalid special parameter must be '_map-[key] or _mapa-[key]")
			}

			pre := k[:idx]
			param := k[idx+1:]

			switch pre {
			case "_map":
				fmt.Fprintf(v.instrument.sc.buffer, `, [\n_map, %d, \%s, %d]`, v.instrNum, param, int(val))
			case "_mapa":
				fmt.Fprintf(v.instrument.sc.buffer, `, [\n_mapa, %d, \%s, %d]`, v.instrNum, param, int(val))
			default:
				panic("unknown special parameter must be '_map-[key] or _mapa-[key]")
			}

		}
	}
	fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d%s]`, v.instrNum, v.paramsStr(ev))
}

func (v *voice) Off(ev *Event) {
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, 0]`, v.instrNum)
	fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, -1]`, v.instrNum)
}

func (v *voice) Mute(*Event)   { v.mute = true }
func (v *voice) UnMute(*Event) { v.mute = false }
func (v *voice) Name() string  { return v.name }
func (v *voice) Offset() int   { return v.instrument.offset }
