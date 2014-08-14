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
}

type Scale interface {
	Frequency(degree int) float64
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

//func New(bar string, tr ...Transformer) *Track {
func New(bar string, tempo Tempo, tr ...Transformer) *Track {
	//t := NewTrack(BPM(120), M(bar))
	t := NewTrack(tempo, M(bar))
	t.Compose(tr...)
	return t
}

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

func Metronome(voice Voice, unit Measure, eventProps ...map[string]float64) *metronome {
	return &metronome{voice: voice, unit: unit, eventProps: eventProps}
}

func Bar(voice Voice, eventProps ...map[string]float64) *bar {
	return &bar{voice: voice, eventProps: eventProps}
}

type metronome struct {
	last       Measure
	voice      Voice
	unit       Measure
	eventProps []map[string]float64
}

func (m *metronome) Transform(t Tracker) {
	n := int(t.CurrentBar() / m.unit)
	half := m.unit / 2
	for i := 0; i < n; i++ {
		t.At(m.unit*Measure(i), On(m.voice, m.eventProps...))
		t.At(m.unit*Measure(i)+half, Off(m.voice))
	}
}

type bar struct {
	voice      Voice
	counter    float64
	eventProps []map[string]float64
}

func (m *bar) Transform(t Tracker) {
	t.At(M("0"), On(m.voice, m.eventProps...))
	t.At(M("1/8"), Off(m.voice))
	m.counter++
}
