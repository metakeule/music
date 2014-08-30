package music

type Generator interface {
	NewNodeId() int
	NewBusId() int
	NewGroupId() int
	NewSampleBuffer() int
}

type Instrument interface {
	Name() string
}

type SCInstrument struct {
	name   string  // name of the instrument must not have characters other than [a-zA-Z0-9_]
	Path   string  // path of the instrument
	Offset float64 // offset applied to every instance of the instrument
	used   bool
}

// the outer invoker may use the first voices instrument to query loadcode etc
func NewSCInstrument(g Generator, name, path string, numVoices int) []*Voice {
	i := &SCInstrument{
		name: name,
		Path: path,
	}
	return Voices(numVoices, g, i, -1)
}

func (i *SCInstrument) Name() string {
	return i.name
}

func (i *SCInstrument) IsUsed() bool {
	return i.used
}

func (i *SCInstrument) Use() {
	i.used = true
}

/*
func (i *SCInstrument) LoadCode() []byte {
	data, err := ioutil.ReadFile(metapath)
	if err != nil {
		panic("can't read instrument " + i.Path)
	}
	return data
}
*/

func Voices(num int, g Generator, instr Instrument, groupid int) []*Voice {
	voices := make([]*Voice, num)

	for i := 0; i < num; i++ {
		voices[i] = &Voice{Generator: g, Instrument: instr}
		if groupid > -1 {
			voices[i].SCGroup = groupid
		}
	}

	return voices
}

/*
import (
	"fmt"
	"io"
)

type SampleOrchestra interface {
	SampleForParams(instrument string, params map[string]float64) string
}

// TODO: perhaps use groups instead of instrNumbers

type scer interface {
	io.Writer
	IncrSampleNumber() int
	IncrInstrNumber() int
	SetVoiceToNum(name string, num int)
	SetNumToVoices(num int, v Voice)
	GetBus(name string) int
	UseSample(name string)
	UseInstrument(name string)
	GetSampleOffset(name string) int
}

type instrument struct {
	name string
	bus  bool
	//sc     *sc
	sc     scer
	offset int
}

func (in *instrument) Name() string {
	return in.name
}

func (in *instrument) Voices(num int) []Voice {
	v := make([]Voice, num)
	for i := 0; i < num; i++ {
		// in.sc.instrNumber++
		instrn := in.sc.IncrInstrNumber()
		name := fmt.Sprintf("%s-%d", in.name, i)
		vc := &voice{
			name:       name,
			instrument: in,
			num:        i,
			// instrNum:   in.sc.instrNumber,
			instrNum: instrn,
		}
		v[i] = vc
		// in.sc.voicesToNum[name] = in.sc.instrNumber
		// in.sc.voicesToNum[name] = instrn
		in.sc.SetVoiceToNum(name, instrn)
		//in.sc.numToVoices[in.sc.instrNumber] = vc
		// in.sc.numToVoices[instrn] = vc
		in.sc.SetNumToVoices(instrn, vc)
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
*/

/*
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

func (v *voice) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return PlayDur(pos, dur, v, params...)
}

func (v *voice) Play(pos string, params ...Parameter) Pattern {
	return Play(pos, v, params...)
}

func (v *voice) Stop(pos string) Pattern {
	return Stop(pos, v)
}

func (v *voice) Modify(pos string, params ...Parameter) Pattern {
	return Modify(pos, v, params...)
}

func (v *voice) On(ev *Event) {
	v.instrument.sc.UseInstrument(v.instrument.name)
	instrnum := v.instrument.sc.IncrInstrNumber()
	// v.instrument.sc.instrNumber++
	if v.instrument.bus {
		//fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, 1200%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
		fmt.Fprintf(v.instrument.sc, `, [\s_new, \%s, %d, 1, 1200%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
		return
	}
	if v.instrNum > 2000 {
		//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_free, %d]`, v.instrNum)
		fmt.Fprintf(v.instrument.sc, `, [\n_free, %d]`, v.instrNum)
	}
	//v.instrNum = v.instrument.sc.instrNumber
	v.instrNum = instrnum

	if v.mute {
		return
	}

	group := 1010

	if v.group != nil {
		group = v.group.Id()
	}

	//fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, %d%s]`, v.instrument.name, v.instrNum, group, v.paramsStr(ev))
	fmt.Fprintf(v.instrument.sc, `, [\s_new, \%s, %d, 1, %d%s]`, v.instrument.name, v.instrNum, group, v.paramsStr(ev))
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
				// fmt.Fprintf(v.instrument.sc.buffer, `, [\n_map, %d, \%s, %d]`, v.instrNum, param, int(val))
				fmt.Fprintf(v.instrument.sc, `, [\n_map, %d, \%s, %d]`, v.instrNum, param, int(val))
			case "_mapa":
				// fmt.Fprintf(v.instrument.sc.buffer, `, [\n_mapa, %d, \%s, %d]`, v.instrNum, param, int(val))
				fmt.Fprintf(v.instrument.sc, `, [\n_mapa, %d, \%s, %d]`, v.instrNum, param, int(val))
			default:
				panic("unknown special parameter must be '_map-[key] or _mapa-[key]")
			}

		}
	}
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d%s]`, v.instrNum, v.paramsStr(ev))
	fmt.Fprintf(v.instrument.sc, `, [\n_set, %d%s]`, v.instrNum, v.paramsStr(ev))
}

func (v *voice) Off(ev *Event) {
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, 0]`, v.instrNum)
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, -1]`, v.instrNum)
	fmt.Fprintf(v.instrument.sc, `, [\n_set, %d, \gate, -1]`, v.instrNum)
}

func (v *voice) Mute(*Event)   { v.mute = true }
func (v *voice) UnMute(*Event) { v.mute = false }
func (v *voice) Name() string  { return v.name }
func (v *voice) Offset() int   { return v.instrument.offset }
*/
