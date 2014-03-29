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

// Transform writes the current event after the existing ones
func (ev *Event) Transform(w EventWriter, events []*Event) {
	for _, e := range events {
		w.Write(e)
	}
	w.Write(ev)
}

// todo: check that all properties are set, or raise an error
func (ev *Event) EventWriter(tw ToneWriter) EventWriter {
	return newSequencer(tw, ev)
}

// a typical tracker, should fullfill the Sequenceer interface
type sequencer struct {
	Current     *Event
	CurrentTick uint
	// die aktuelle position im takt in % von der basiseinheit, vom start des taktes aus gezählt
	// z.b. im 4/4 takt wäre 300 auf der dritten 4tel und 350 auf der 4+
	CurrentPositionInBar uint
	// Events               []*Event
	// EventCounter         uint
	ToneWriter ToneWriter
	globals    Globals
}

func (st *sequencer) Instrument() string { return st.Current.Instrument }

func (st *sequencer) InstrumentParameters() map[string]float64 { return st.Current.InstrumentParams }

func (st *sequencer) Frequency(ev *Event) float64 {
	return st.Current.Scale.Frequency(st.Current.Height)
}

func (st *sequencer) Start(ev *Event) uint {
	if ev.Rest {
		return st.CurrentTick
	}

	// TODO: check with negativ delay, should keep the negative sign upto delayTicks
	// but we are using uint everywhere, so check it (maybe take the neg. sign out) and
	// respect it at the end
	delay := st.Current.Rhythm.Delay(st.Current.Bar, st.CurrentPositionInBar)
	delayPerMinute := ((float64(delay) / 100.0) / float64(st.Current.Bar.NumBeats)) * float64(st.Current.Bar.TempoBar)
	delayTicks := uint(delayPerMinute * float64(st.Current.Tempo))

	return st.CurrentTick + delayTicks
}

func (st *sequencer) Duration(ev *Event) uint {
	notePerMinute := float64(ev.Length) / float64(100) * float64(st.Current.Bar.TempoBar) / float64(st.Current.Bar.NumBeats)
	noteTicks := notePerMinute * float64(st.Current.Tempo)
	st.CurrentTick = st.CurrentTick + uint(noteTicks)
	st.CurrentPositionInBar = (st.CurrentPositionInBar + ev.Length) % (st.Current.Bar.NumBeats * 100)
	return uint(noteTicks)
}

func (s *sequencer) Globals() Globals {
	return s.globals
}

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

func newSequencer(tw ToneWriter, ev *Event) *sequencer {
	ev.MustBeComplete()
	st := &sequencer{}
	st.Current = &Event{}
	st.Current.Bar = ev.Bar
	st.Current.Scale = ev.Scale
	st.Current.Tempo = ev.Tempo
	st.Current.Rhythm = ev.Rhythm
	st.Current.Volume = ev.Volume
	st.Current.Instrument = ev.Instrument
	st.Current.InstrumentParams = ev.InstrumentParams
	st.Current.Height = ev.Height
	st.ToneWriter = tw
	st.globals = NewGlobals()
	return st
}

func (st *sequencer) Write(ev *Event) {

	if bar := ev.Bar; bar != nil {
		st.CurrentPositionInBar = 0
		st.Current.Bar = bar
	}

	if scale := ev.Scale; scale != nil {
		st.Current.Height = 0
		st.Current.Scale = scale
	}

	if tempo := ev.Tempo; tempo > 0 {
		st.Current.Tempo = tempo
	}

	if rhythm := ev.Rhythm; rhythm != nil {
		st.Current.Rhythm = rhythm
	}

	if vol := ev.Volume; vol > 0 {
		st.Current.Volume = vol
	}

	if instr := ev.Instrument; instr != "" {
		st.Current.Instrument = instr
	}

	if instrParams := ev.InstrumentParams; instrParams != nil {
		st.Current.InstrumentParams = instrParams
	}

	st.Current.Height = st.Current.Height + ev.Height

	tone := &Tone{}
	tone.Start = st.Start(ev)
	tone.Duration = st.Duration(ev)

	if !ev.Rest {
		tone.Amplitude = st.Amplitude(ev)
		tone.Frequency = st.Frequency(ev)
		tone.Instrument = st.Instrument()
		tone.InstrumentParameters = st.InstrumentParameters()
	}

	st.ToneWriter.Write(tone)
}

// return a new eventwriter starting from this point as a copy
// (same current properties and a copy of globals)
func (s *sequencer) Fork() EventWriter {
	s2 := newSequencer(s.ToneWriter, s.Current)
	s2.CurrentTick = s.CurrentTick
	s2.CurrentPositionInBar = s.CurrentPositionInBar
	// s2.Globals = s.Globals

	// TODO make it threadsafe
	for k, v := range s.globals {
		s2.globals.Set(k, v)
	}

	return s2
}

func (st *sequencer) Amplitude(ev *Event) float32 {
	rhythmAmp := st.Current.Rhythm.Amplitude(st.Current.Bar, st.CurrentPositionInBar, ev.Accent)
	// fmt.Println(rhythmAmp, st.CurrentVolume)
	return st.Current.Volume * rhythmAmp
}

// TODO:
// - umarbeiten eines Sequencers zu einem EventWriter
// - erstellen eines complete events, dass sich wie ein EventWriter verhalten kann
// - erstellen eines helpers (eines fake EventWriters zum tracken, wie bei den wrap-contrib helpers)

// TODO: make it threadsafe with a mutex
type Globals map[string]float64

func NewGlobals() Globals {
	return Globals(map[string]float64{})
}

func (g Globals) Get(name string) float64 {
	return g[name]
}

func (g Globals) Set(name string, value float64) {
	g[name] = value
}

func (g Globals) UnSet(name string) {
	delete(g, name)
}

func (g Globals) Has(name string) bool {
	_, has := g[name]
	return has
}

type EventWriter interface {
	Globals() Globals
	Write(*Event)
	// get new EventWriter that start with the current values
	// can be used to create parallel events (for performance reasons they should be reduced later on when submitting
	// to a polyphon instrument)
	// maybe not needed and can be implement via transformer wrapper, we will see
	Fork() EventWriter
}

type ToneWriter interface {
	Write(*Tone)
}

type ToneWriterFunc func(*Tone)

func (t ToneWriterFunc) Write(tone *Tone) { t(tone) }

type Transformer interface {
	// Transform transforms a Slice of musical events
	// that slice can be thought of a parallel or consecutive events, depending on the execution context
	Transform(EventWriter, []*Event)
}

type TransformerFunc func(EventWriter, []*Event)

func (t TransformerFunc) Transform(w EventWriter, evs []*Event) { t(w, evs) }
