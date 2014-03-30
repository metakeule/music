package music

// ein event kann alle diese eigenschaften haben, oder aber nur
// manche. Note muss er immer haben
type Event struct {
	Height int // die tonhöhenveränderung gegenüber der vorigen Note in einheiten der umgebenden tonleiter
	// z.b. in der Tonleiter C-Dur, wenn der vorige Ton D war und Height -1, wäre der resultierende Ton C
	// in der Chromatischen Tonleiter, wenn der vorige Ton Cis war und Height -1, wäre der resultierende Ton C
	// die anfangs tonhöhe wenn noch keine note davor existierte, ist 0, das heisst, die erste position der
	// scala. wenn ein scalenwechsel erfolgt, wird der height wert wieder auf die erste position the scala
	// bezogen
	// bei pausen bleiben die vorigen tonhöhen erhalten
	// obligatorischer wert, Events mit der länge 0 werden ignoriert (ist eine möglichkeit, einen event zu löschen)
	Length uint // längeneinheit in % der basiseinheit des taktes
	// also z.b. im 4/4 Takt hat eine 4tel die Length 100
	// eine 8tel die Length 50, eine 16tel die Length 25
	// eine halbe die length 200 eine ganze die Length 400
	// im 6/8 takt hat eine 1tel die Length 100, eine 4tel die length 200, eine 16tel 50, eine 32tel 25 usw
	Accent bool // hat die Note einen Akzent, wie der sich dann konkret niederschlägt, hängt von dem umgebenden rhythmus
	// und der umgebungsdynamik ab

	Rest bool // if true, ist es eine pause

	Bar              *Bar               // falls kein Taktwechsel: nil
	Scale            Scale              // falls kein Tonartwechsel: nil
	Instrument       string             // falls kein instrumentenwechsel: ""
	InstrumentParams map[string]float64 // falls kein instrumentenparameterwechsel: nil
	Tempo            Tempo              // falls kein tempowechsel: 0
	Rhythm           Rhythm             // falls kein rhythmenwechsel: nil
	Volume           float32            // falls keinen volume wechsel: <= 0
}

func (ev *Event) Clone() *Event {
	clone := &Event{}
	clone.Accent = ev.Accent
	clone.Bar = ev.Bar
	clone.Height = ev.Height
	clone.Instrument = ev.Instrument
	clone.InstrumentParams = map[string]float64{}

	for k, v := range ev.InstrumentParams {
		clone.InstrumentParams[k] = v
	}

	clone.Length = ev.Length
	clone.Rest = ev.Rest
	clone.Rhythm = ev.Rhythm
	clone.Scale = ev.Scale
	clone.Tempo = ev.Tempo
	clone.Volume = ev.Volume
	return clone
}

// adds the event to the events
func (ev *Event) Transform(events ...*Event) []*Event {
	/*
		res := make([]*Event, len(events)+1)

		for i, e := range events {
			res[i] = e.Clone()
		}

		res[len(events)] = ev.Clone()
		return res
	*/
	return []*Event{ev.Clone()}
}

// returns the outer transformer before the inner
func (ev *Event) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(evts ...*Event) []*Event {
		evts = ev.Transform(evts...)
		return inner.Transform(evts...)
	})
}

type events []*Event

func (e events) Clone() []*Event {
	res := make([]*Event, len(e))

	for i, ev := range e {
		res[i] = ev.Clone()
	}

	return res
}

func (e events) Transform(evts ...*Event) []*Event {
	return e.Clone()
	//	return append(events(evts).Clone(), e.Clone()...)
}

func (e events) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(evts ...*Event) []*Event {
		evts = e.Transform(evts...)
		return inner.Transform(evts...)
	})
}

func Events(ev ...*Event) events {
	return events(ev)
}

// applies the events properties to the given event
// and returns the resulting event
func (ev *Event) ApplyTo(event *Event) *Event {

	// res := make([]*Event, len(events))

	// for i, e := range events {
	current := event.Clone()

	if ev.Length != 0 {
		current.Length = ev.Length
	}

	if ev.Accent {
		current.Accent = true
	}

	if ev.Bar != nil {
		current.Bar = ev.Bar
	}

	if ev.Height != 0 {
		current.Height = ev.Height
	}

	if ev.Scale != nil {
		current.Scale = ev.Scale
	}

	if ev.Tempo != 0 {
		current.Tempo = ev.Tempo
	}

	if ev.Rhythm != nil {
		current.Rhythm = ev.Rhythm
	}

	if ev.Volume != 0 {
		current.Volume = ev.Volume
	}

	if ev.Instrument != "" {
		current.Instrument = ev.Instrument
	}

	if ev.InstrumentParams != nil {
		for k, v := range ev.InstrumentParams {
			current.InstrumentParams[k] = v
		}
	}

	if ev.Rest {
		current.Rest = true
	}

	return current
	// res[i] = current
	// }

	// return res
}

// Transform writes the current event after the existing ones
/*
func (ev *Event) Transformx(w EventWriter, events []*Event) {
	for _, e := range events {
		w.Write(e)
	}
	w.Write(ev)
}
*/

func (ev *Event) MustBeComplete() {
	if ev.Bar == nil {
		panic("has no bar")
	}

	if ev.Scale == nil {
		panic("has no scale")
	}

	if ev.Tempo == 0 {
		panic("has no tempo")
	}

	if ev.Rhythm == nil {
		panic("has no rhythm")
	}

	if ev.Volume == 0 {
		panic("has no volume")
	}

	if ev.Instrument == "" {
		panic("has no instrument")
	}

	if ev.InstrumentParams == nil {
		panic("has no InstrumentParams")
	}
}

// todo: check that all properties are set, or raise an error
func (ev *Event) Sequencer(tw ToneWriter) *sequencer { return newSequencer(tw, ev) }

func (ev *Event) instrument() string                       { return ev.Instrument }
func (ev *Event) instrumentParameters() map[string]float64 { return ev.InstrumentParams }
func (ev *Event) frequency() float64                       { return ev.Scale.Frequency(ev.Height) }

func (ev *Event) start(currentTick uint, currentPositionInBar uint) uint {
	if ev.Rest {
		return currentTick
	}

	// TODO: check with negativ delay, should keep the negative sign upto delayTicks
	// but we are using uint everywhere, so check it (maybe take the neg. sign out) and
	// respect it at the end
	delay := ev.Rhythm.Delay(ev.Bar, currentPositionInBar)
	delayPerMinute := ((float64(delay) / 100.0) / float64(ev.Bar.NumBeats)) * float64(ev.Bar.TempoBar)
	delayTicks := uint(delayPerMinute * float64(ev.Tempo))

	return currentTick + delayTicks
}

func (ev *Event) duration(currentTick, currentPositionInBar uint) (newCurrentTick, newcurrentPositionInBar, durInTicks uint) {
	notePerMinute := float64(ev.Length) / float64(100) * float64(ev.Bar.TempoBar) / float64(ev.Bar.NumBeats)
	noteTicks := notePerMinute * float64(ev.Tempo)
	currentTick = currentTick + uint(noteTicks)
	currentPositionInBar = (currentPositionInBar + ev.Length) % (ev.Bar.NumBeats * 100)
	return currentTick, currentPositionInBar, uint(noteTicks)
}

func (ev *Event) Tone(currentTick uint, currentPositionInBar uint) (newCurrentTick, newcurrentPositionInBar uint, tone *Tone) {
	tone = &Tone{}
	tone.Start = ev.start(currentTick, currentPositionInBar)
	newCurrentTick, newcurrentPositionInBar, tone.Duration = ev.duration(currentTick, currentPositionInBar)
	if !ev.Rest {
		tone.Amplitude = ev.amplitude(currentPositionInBar)
		tone.Frequency = ev.frequency()
		tone.Instrument = ev.instrument()
		tone.InstrumentParameters = ev.instrumentParameters()
	}
	return newCurrentTick, newcurrentPositionInBar, tone
}

func (ev *Event) amplitude(currentPositionInBar uint) float32 {
	rhythmAmp := ev.Rhythm.Amplitude(ev.Bar, currentPositionInBar, ev.Accent)
	return ev.Volume * rhythmAmp
}

// TODO think over it, maybe we can get rid of Sequencer
func (e events) Tones(currentTick uint, currentPositionInBar uint) []*Tone {
	all := []*Tone{}
	for _, ev := range e {
		var t *Tone
		currentTick, currentPositionInBar, t = ev.Clone().Tone(currentTick, currentPositionInBar)
		all = append(all, t)
	}
	return all
}

type Ticker struct {
	Current uint
	ToneWriter
}

func NewTicker(startTick uint, w ToneWriter) *Ticker {
	return &Ticker{startTick, w}
}

func (t *Ticker) Serial(start *Event) *SerialSink {
	return &SerialSink{t, t.ToneWriter, start}
}

/*
func (t *Ticker) Parallel(start *Event) *ParallelSink {
	return &ParallelSink{t, t.ToneWriter, start}
}
*/

type SerialSink struct {
	Ticker *Ticker
	ToneWriter
	start *Event
}

type parallel struct {
	trs    []Transformer
	Ticker *Ticker
	ToneWriter
	start *Event
}

func (p *parallel) Transform(events ...*Event) []*Event {
	// first: take all transformer events
	allEvents := []*Event{}
	currentTicker := p.Ticker.Current

	for _, tr := range p.trs {
		p.Ticker.Current = currentTicker
		allEvents = append(allEvents, tr.Transform(Events(events...).Clone()...)...)
	}

	all := []*Tone{}
	for _, ev := range allEvents {
		var t *Tone
		_, _, t = ev.ApplyTo(p.start).Tone(currentTicker, 0)
		all = append(all, t)
	}
	p.ToneWriter.Write(all...)
	if len(allEvents) > 0 {
		p.Ticker.Current, _, _ = allEvents[0].ApplyTo(p.start).duration(currentTicker, 0)
	}
	return []*Event{}
}

func (p *parallel) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(events ...*Event) []*Event {
		p.Transform(events...)
		return inner.Transform()
	})
}

func (s *SerialSink) Parallel(trs ...Transformer) *parallel {
	return &parallel{
		trs:        trs,
		Ticker:     s.Ticker,
		ToneWriter: s.ToneWriter,
		start:      s.start,
	}
}

type serial struct {
	trs    []Transformer
	Ticker *Ticker
	ToneWriter
	start *Event
}

func (s *SerialSink) Serial(trs ...Transformer) *serial {
	return &serial{
		trs:        trs,
		Ticker:     s.Ticker,
		ToneWriter: s.ToneWriter,
		start:      s.start,
	}
}

func (s *serial) Transform(events ...*Event) []*Event {
	all := []*Tone{}
	currentPositionInBar := uint(0)

	for _, tr := range s.trs {
		for _, ev := range tr.Transform(Events(events...).Clone()...) {
			var t *Tone
			s.Ticker.Current, currentPositionInBar, t = ev.ApplyTo(s.start).Tone(s.Ticker.Current, currentPositionInBar)
			all = append(all, t)
		}

	}
	s.ToneWriter.Write(all...)
	return []*Event{}
}

func (s *serial) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(events ...*Event) []*Event {
		s.Transform(events...)
		return inner.Transform()
	})
}

func (s *SerialSink) WrapX(inner Transformer) Transformer {
	return TransformerFunc(func(evts ...*Event) []*Event {
		s.TransformX(evts...)
		return inner.Transform()
	})
}

// writes the event notes to the ToneWriter and returns an empty event list
func (s *SerialSink) TransformX(events ...*Event) []*Event {
	// fmt.Printf("transforming %d events\n", len(events))
	all := []*Tone{}
	currentPositionInBar := uint(0)
	for _, ev := range events {
		var t *Tone
		s.Ticker.Current, currentPositionInBar, t = ev.ApplyTo(s.start).Tone(s.Ticker.Current, currentPositionInBar)
		all = append(all, t)
	}
	s.ToneWriter.Write(all...)
	return []*Event{}
}

/*
type ParallelSink struct {
	Ticker *Ticker
	ToneWriter
	start *Event
}


// writes the event notes to the ToneWriter and returns an empty event list
func (s *ParallelSink) transform(events ...*Event) []*Event {
	all := []*Tone{}
	currentTicker := s.Ticker.Current
	for _, ev := range events {
		var t *Tone
		_, _, t = ev.ApplyTo(s.start).Tone(currentTicker, 0)
		all = append(all, t)
	}
	s.ToneWriter.Write(all...)
	return []*Event{}
}
*/
