package music

import "bytes"

type Event struct {
	Voice                     *Voice
	Params                    Parameter // a special parameter offset may be used to set a per event offset
	Runner                    func(*Event)
	Type                      string
	tick                      uint
	absPosition               Measure // will be enabled when integrated
	offset                    float64 // offset added to the final position (includes instrument and sample offsets as well as offset set via parameter)
	sccode                    bytes.Buffer
	changedParamsPrepared     map[string]float64
	sampleInstrumentFrequency float64
}

var fin = &Event{Runner: func(*Event) {}, Type: "fin"}
var start = &Event{Runner: func(*Event) {}, Type: "start"}

func newEvent(v *Voice, type_ string) *Event {
	return &Event{
		Voice:  v,
		Params: ParamsMap(map[string]float64{}),
		Type:   type_,
	}
}

// merges the given params of the event into a clone
// of ev, returning the clone
// may be used with events that have modifiers, like Scale, Rhythm etc
// the given voice is set and we get an On event
//func (ev *Event) OnMerged(voice Voice, m map[string]float64) *Event {
func (ev *Event) OnMerged(voice *Voice, ps ...Parameter) *Event {
	n := ev.Clone()
	p := []Parameter{ev.Params}
	p = append(p, ps...)
	n.Params = Params(p...)
	n.Voice = voice
	n.Runner = voice.OnEvent
	n.Type = "ON"
	return n
}

// merges the given params of the event into a clone
// of ev, returning the clone
// may be used with events that have modifiers, like Scale, Rhythm etc
// the given voice is set and we get a change event
//func (ev *Event) ChangeMerged(voice Voice, m map[string]float64) *Event {
func (ev *Event) ChangeMerged(voice *Voice, ps ...Parameter) *Event {
	n := ev.Clone()
	p := []Parameter{ev.Params}
	p = append(p, ps...)
	n.Params = Params(p...)
	n.Voice = voice
	n.Runner = voice.ChangeEvent
	n.Type = "CHANGE"
	return n
}

func (ev *Event) Clone() *Event {
	n := &Event{Voice: ev.Voice, Runner: ev.Runner}
	n.Type = ev.Type
	n.absPosition = ev.absPosition
	n.Params = ev.Params
	return n
}

func OnEvent(v *Voice, params ...Parameter) *Event {
	return &Event{
		Voice:  v,
		Params: Params(params...),
		Runner: v.OnEvent,
		Type:   "ON",
	}
}

func OffEvent(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.OffEvent,
		Type:   "OFF",
	}
}

func MuteEvent(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.OffEvent,
		Type:   "MUTE",
	}
}

func UnMuteEvent(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.donothing,
		Type:   "UNMUTE",
	}
}

func ChangeEvent(v *Voice, params ...Parameter) *Event {
	return &Event{
		Voice:  v,
		Params: Params(params...),
		Runner: v.ChangeEvent,
		Type:   "CHANGE",
	}
}
