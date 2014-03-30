package music

/*
// Alles exemplarisch

// eine Note ist unabhängig von dem Instrument, auf dem sie gespielt wird,
// von dem Takt, in dem sie ist und von der Tonleiter, in der sie gespielt wird
// vielleicht kann Note auch vollständig in Event aufgehen und es gibt nur ein paar funktionen
// als shortcuts um unvollständige events zu erzeugen
type Note struct {
}
*/

func InstrumentNote(height int, length uint, instr string) *Event {
	return &Event{Height: height, Length: length, Instrument: instr}
}

func Note(height int, length uint) *Event {
	return &Event{Height: height, Length: length}
}

func AccentedNote(height int, length uint) *Event {
	return &Event{Height: height, Length: length, Accent: true}
}

func Rest(length uint) *Event {
	return &Event{Length: length, Rest: true}
}

// eine melodie iste eine abfolge von Noten
// type Sequence []*Note

// ein Tone ist eine konkretisierung einer Note
type Tone struct {
	Instrument           string             // instrument, welches die Note spielt, wenn leerer string: pause
	Start                uint               // startpunkt des tones in Ticks (feinste auflösung des stücks)
	Duration             uint               // dauer des tones in ticks
	Frequency            float64            // frequenz, mit der das instrument angesteuert wird
	InstrumentParameters map[string]float64 // parameter, mit denen das instrument angesteuert wird, dazu zählen auch die
	// ausgabe kanäle
	Amplitude float32 // amplitude, mit der das instrument angesteuert wird
}

// ein track ist eine abfolge von Tone (keine überlappung)
type Sequence []*Tone

// makes sound from tracks
// all sequences are considered to run in parallel
type Player interface {
	Play([]Sequence)
}

// eine skala liefert zu einer scalen position eine frequenz
// bei z.B. C-Dur müssen alle möglichen positionen aller möglichen frequencen berücksichtigt werden
// 0 ist die referenz-position, z.b. bei C-Dur das eingestrichene C. -8 wäre dann eine Oktave darunter
type Scale interface {
	Frequency(scalePosition int) float64
}

type Bar struct {
	NumBeats uint // the number of base units that fits into a bar
	// der kehrwert davon ist die länge einer basiseinheit (die in der Note verwendet wird)
	// in einem 4/4 takt ist NumBeats 4 in einem 6/8 takt 6
	TempoBar uint // number of Bars (not beats!) per Minute
	// TempoBar * NumBeats = Beats per Minute
}

type Tempo uint // geschwindigkeit in Ticks per Minute

type Rhythm interface {
	// Amplitude returns an amplitude factor that is multiplied by the current volume and passed to the instrument
	// depending on the position in a Bar and the question if it has an accent
	Amplitude(bar *Bar, pos uint, accent bool) float32

	// verzögerung in % der basiseinheit des takes (für den groove)
	// positiv (laid back) oder negativ (vorgezogen)
	// in abhänigkeit vom takt und von der position des taktes
	// verändert die startposition
	Delay(bar *Bar, pos uint) int
}

// a typical tracker, should fullfill the Sequenceer interface
type sequencer struct {
	current     *Event
	currentTick uint
	// die aktuelle position im takt in % von der basiseinheit, vom start des taktes aus gezählt
	// z.b. im 4/4 takt wäre 300 auf der dritten 4tel und 350 auf der 4+
	currentPositionInBar uint
	// Events               []*Event
	// EventCounter         uint
	toneWriter ToneWriter
}

func (st *sequencer) instrument() string { return st.current.Instrument }

func (st *sequencer) instrumentParameters() map[string]float64 { return st.current.InstrumentParams }

func (st *sequencer) frequency(ev *Event) float64 {
	return st.current.Scale.Frequency(st.current.Height)
}

func (st *sequencer) start(ev *Event) uint {
	if ev.Rest {
		return st.currentTick
	}

	// TODO: check with negativ delay, should keep the negative sign upto delayTicks
	// but we are using uint everywhere, so check it (maybe take the neg. sign out) and
	// respect it at the end
	delay := st.current.Rhythm.Delay(st.current.Bar, st.currentPositionInBar)
	delayPerMinute := ((float64(delay) / 100.0) / float64(st.current.Bar.NumBeats)) * float64(st.current.Bar.TempoBar)
	delayTicks := uint(delayPerMinute * float64(st.current.Tempo))

	return st.currentTick + delayTicks
}

func (st *sequencer) duration(ev *Event) uint {
	notePerMinute := float64(ev.Length) / float64(100) * float64(st.current.Bar.TempoBar) / float64(st.current.Bar.NumBeats)
	noteTicks := notePerMinute * float64(st.current.Tempo)
	st.currentTick = st.currentTick + uint(noteTicks)
	st.currentPositionInBar = (st.currentPositionInBar + ev.Length) % (st.current.Bar.NumBeats * 100)
	return uint(noteTicks)
}

func newSequencer(tw ToneWriter, ev *Event) *sequencer {
	ev.MustBeComplete()
	st := &sequencer{}
	st.current = &Event{}
	st.current.Bar = ev.Bar
	st.current.Scale = ev.Scale
	st.current.Tempo = ev.Tempo
	st.current.Rhythm = ev.Rhythm
	st.current.Volume = ev.Volume
	st.current.Instrument = ev.Instrument
	st.current.InstrumentParams = ev.InstrumentParams
	st.current.Height = ev.Height
	st.toneWriter = tw
	return st
}

// all given events are considered to start at the same time and from the same
// current values
// the first event advances the current of this sequencer
func (st *sequencer) Write(events ...*Event) {
	for i, ev := range events {
		// skip the first, we did it already
		if i == 0 {
			continue
		}
		s := st.Fork()
		s.writeSingleEvent(ev)
	}
	st.writeSingleEvent(events[0])
}

func (st *sequencer) writeSingleEvent(ev *Event) {

	if bar := ev.Bar; bar != nil {
		st.currentPositionInBar = 0
		st.current.Bar = bar
	}

	if scale := ev.Scale; scale != nil {
		st.current.Height = 0
		st.current.Scale = scale
	}

	if tempo := ev.Tempo; tempo > 0 {
		st.current.Tempo = tempo
	}

	if rhythm := ev.Rhythm; rhythm != nil {
		st.current.Rhythm = rhythm
	}

	if vol := ev.Volume; vol > 0 {
		st.current.Volume = vol
	}

	if instr := ev.Instrument; instr != "" {
		st.current.Instrument = instr
	}

	if instrParams := ev.InstrumentParams; instrParams != nil {
		st.current.InstrumentParams = instrParams
	}

	// st.current.Height = st.current.Height + ev.Height

	st.current.Height = ev.Height

	tone := &Tone{}
	tone.Start = st.start(ev)
	tone.Duration = st.duration(ev)

	if !ev.Rest {
		tone.Amplitude = st.amplitude(ev)
		tone.Frequency = st.frequency(ev)
		tone.Instrument = st.instrument()
		tone.InstrumentParameters = st.instrumentParameters()
	}

	st.toneWriter.Write(tone)
}

// return a new eventwriter starting from this point as a copy
// (same current properties and a copy of globals)
func (s *sequencer) Fork() *sequencer {
	s2 := newSequencer(s.toneWriter, s.current)
	s2.currentTick = s.currentTick
	s2.currentPositionInBar = s.currentPositionInBar
	return s2
}

func (st *sequencer) amplitude(ev *Event) float32 {
	rhythmAmp := st.current.Rhythm.Amplitude(st.current.Bar, st.currentPositionInBar, ev.Accent)
	// fmt.Println(rhythmAmp, st.CurrentVolume)
	return st.current.Volume * rhythmAmp
}

// TODO:
// - umarbeiten eines Sequencers zu einem EventWriter
// - erstellen eines complete events, dass sich wie ein EventWriter verhalten kann
// - erstellen eines helpers (eines fake EventWriters zum tracken, wie bei den wrap-contrib helpers)

type EventWriter interface {
	Write(events ...*Event)
}

type EventWriterFunc func(events ...*Event)

func (e EventWriterFunc) Write(events ...*Event) {
	e(events...)
}

// the ToneWriter writes all given tones at the same time (in parallel)
type ToneWriter interface {
	Write(tones ...*Tone)
}

type ToneWriterFunc func(...*Tone)

func (t ToneWriterFunc) Write(tones ...*Tone) { t(tones...) }

type Transformer interface {
	// Transform transforms a Slice of musical events
	Transform(events ...*Event) []*Event
}

type TransformerFunc func(events ...*Event) []*Event

func (t TransformerFunc) Transform(events ...*Event) []*Event {
	return t(events...)
}
