package music

type Parameter interface {
	Params() map[string]float64
}

type ParamsMap map[string]float64

func (p ParamsMap) Params() map[string]float64 {
	return map[string]float64(p)
}

type Event struct {
	Voice Voice
	// parameters that might be modified by ParamModifiers
	Params map[string]float64
	// modifiers, such as returning a frequency for a position in a scale
	ParamModifiers map[string]func(float64) float64
	Runner         func(*Event)
	Type           string
	Tick           uint
	AbsPosition    Measure
	//Duration       Measure
}

var fin = &Event{Runner: func(*Event) {}, Type: "fin"}
var start = &Event{Runner: func(*Event) {}, Type: "start"}

func newEvent(v Voice, type_ string) *Event {
	return &Event{
		Voice:          v,
		Params:         map[string]float64{},
		ParamModifiers: map[string]func(float64) float64{},
		Type:           type_,
	}
}

// merges the given params of the event into a clone
// of ev, returning the clone
// may be used with events that have modifiers, like Scale, Rhythm etc
// the given voice is set and we get an On event
//func (ev *Event) OnMerged(voice Voice, m map[string]float64) *Event {
func (ev *Event) OnMerged(voice Voice, ps ...Parameter) *Event {
	n := ev.Clone()

	for _, p := range ps {
		for k, v := range p.Params() {
			n.Params[k] = v
		}
	}

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
func (ev *Event) ChangeMerged(voice Voice, ps ...Parameter) *Event {
	n := ev.Clone()

	for _, p := range ps {
		for k, v := range p.Params() {
			n.Params[k] = v
		}
	}

	n.Voice = voice
	n.Runner = voice.Change
	n.Type = "CHANGE"
	return n
}

func (ev *Event) FinalParams() map[string]float64 {
	res := map[string]float64{}

	for k, v := range ev.Params {
		modifier, exists := ev.ParamModifiers[k]
		if exists {
			v = modifier(v)
		}
		res[k] = v
	}
	return res
}

func (ev *Event) Clone() *Event {
	n := &Event{Voice: ev.Voice, Runner: ev.Runner, ParamModifiers: ev.ParamModifiers}
	n.Type = ev.Type
	n.AbsPosition = ev.AbsPosition
	//n.Duration = ev.Duration
	if len(ev.Params) > 0 {
		n.Params = map[string]float64{}
		for k, v := range ev.Params {
			n.Params[k] = v
		}
	}
	return n
}

//func On(v Voice, params ...map[string]float64) *Event {
func On(v Voice, params ...Parameter) *Event {
	p := map[string]float64{}

	for _, ps := range params {

		for k, v := range ps.Params() {
			p[k] = v
		}

	}
	return &Event{
		Voice:          v,
		Params:         p,
		ParamModifiers: map[string]func(float64) float64{},
		Runner:         v.On,
		Type:           "ON",
	}
}

func Off(v Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.Off,
		Type:   "OFF",
	}
}

func Mute(v Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.Mute,
		Type:   "MUTE",
	}
}

func UnMute(v Voice) *Event {
	return &Event{
		Voice:  v,
		Runner: v.UnMute,
		Type:   "UNMUTE",
	}
}

//func Change(v Voice, params ...map[string]float64) *Event {
func Change(v Voice, params ...Parameter) *Event {
	p := map[string]float64{}

	for _, ps := range params {

		for k, v := range ps.Params() {
			p[k] = v
		}

	}
	return &Event{
		Voice:          v,
		Params:         p,
		ParamModifiers: map[string]func(float64) float64{},
		Runner:         v.Change,
		Type:           "CHANGE",
	}
}
