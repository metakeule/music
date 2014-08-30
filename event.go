package music

import "bytes"

type Parameter interface {
	Params() map[string]float64
}

type ParamsMap map[string]float64

func (p ParamsMap) Params() map[string]float64 {
	return map[string]float64(p)
}

type Event struct {
	Voice       *Voice
	Params      Parameter // a special parameter offset may be used to set a per event offset
	Runner      func(*Event)
	Type        string
	Tick        uint
	AbsPosition Measure // will be enabled when integrated
	Offset      float64 // offset added to the final position (includes instrument and sample offsets as well as offset set via parameter)
	SCCode      bytes.Buffer
}

/*
type Event struct {
	Voice Voice
	// parameters that might be modified by ParamModifiers
	Params map[string]float64
	// modifiers, such as returning a frequency for a position in a scale
	// ParamModifiers map[string]func(float64) float64
	Runner      func(*Event)
	Type        string
	Tick        uint
	AbsPosition Measure
	//Duration       Measure
}
*/

var fin = &Event{Runner: func(*Event) {}, Type: "fin"}
var start = &Event{Runner: func(*Event) {}, Type: "start"}

func newEvent(v *Voice, type_ string) *Event {
	return &Event{
		Voice: v,
		//Params: map[string]float64{},
		Params: ParamsMap(map[string]float64{}),
		// ParamModifiers: map[string]func(float64) float64{},
		Type: type_,
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
	n.Runner = voice.On
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
	n.Runner = voice.Change
	n.Type = "CHANGE"
	return n
}

func (ev *Event) Clone() *Event {
	//n := &Event{Voice: ev.Voice, Runner: ev.Runner, ParamModifiers: ev.ParamModifiers}
	n := &Event{Voice: ev.Voice, Runner: ev.Runner}
	n.Type = ev.Type
	n.AbsPosition = ev.AbsPosition
	n.Params = ev.Params
	return n
}

//func On(v Voice, params ...map[string]float64) *Event {
func On(v *Voice, params ...Parameter) *Event {
	return &Event{
		Voice:  v,
		Params: Params(params...),
		// ParamModifiers: map[string]func(float64) float64{},
		Runner: v.On,
		Type:   "ON",
	}
}

func Off(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.Off,
		Type:   "OFF",
	}
}

func Mute(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.Mute,
		Type:   "MUTE",
	}
}

func UnMute(v *Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.UnMute,
		Type:   "UNMUTE",
	}
}

//func Change(v Voice, params ...map[string]float64) *Event {
func Change(v *Voice, params ...Parameter) *Event {
	return &Event{
		Voice:  v,
		Params: Params(params...),
		// ParamModifiers: map[string]func(float64) float64{},
		Runner: v.Change,
		Type:   "CHANGE",
	}
}
