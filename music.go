package music

type Instrument interface {
	New(num int) []Voice
	Name() string
}

type Voice interface {
	On(*Event)
	Off(*Event)
	Change(*Event)
	Name() string
	Mute(*Event)
	UnMute(*Event)
	Offset() int // offset in millisecs, may be negative
	EventPlayer
}

type Scale interface {
	Degree(degree int) Parameter
}

type Tracker interface {
	EachBar() []Transformer
	TempoAt(abspos Measure) Tempo
	CurrentBar() Measure
	BarNum() int
	SetEachBar(eachBar ...Transformer)
	SetTempo(position Measure, tempo Tempo)
	At(position Measure, events ...*Event)
}

/*
//func New(bar string, tr ...Transformer) *Track {
func New(bar string, tempo Tempo, tr ...Transformer) *Track {
	//t := NewTrack(BPM(120), M(bar))
	t := NewTrack(tempo, M(bar))
	t.Compose(tr...)
	return t
}
*/

type eachBar struct {
	trafos []Transformer
}

func (e *eachBar) Transform(t Tracker) {
	for _, tr := range e.trafos {
		tr.Transform(t)
	}
	t.SetEachBar(e.trafos...)
}

func EachBar(tr ...Transformer) *eachBar {
	return &eachBar{tr}
}

func Params(parameter ...Parameter) Parameter {
	params := map[string]float64{}

	for _, p := range parameter {
		for k, v := range p.Params() {
			params[k] = v
		}
	}
	return ParamsMap(params)
}

func Metronome(voice Voice, unit Measure, parameter ...Parameter) *metronome {
	return &metronome{voice: voice, unit: unit, eventProps: Params(parameter...)}
}

func Bar(voice Voice, parameter ...Parameter) *bar {
	return &bar{voice: voice, eventProps: Params(parameter...)}
}

type metronome struct {
	last       Measure
	voice      Voice
	unit       Measure
	eventProps Parameter
}

func (m *metronome) Transform(t Tracker) {
	n := int(t.CurrentBar() / m.unit)
	half := m.unit / 2
	for i := 0; i < n; i++ {
		t.At(m.unit*Measure(i), On(m.voice, m.eventProps))
		t.At(m.unit*Measure(i)+half, Off(m.voice))
	}
}

type bar struct {
	voice      Voice
	counter    float64
	eventProps Parameter
}

func (m *bar) Transform(t Tracker) {
	t.At(M("0"), On(m.voice, m.eventProps))
	t.At(M("1/8"), Off(m.voice))
	m.counter++
}
