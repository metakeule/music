package music

/*
type Loop struct {
	Pattern
	// Number of Bars that correspond to the length of the loop
	// must be > 0
	NumBars uint
}
*/

/*
type Tracker interface {
	SetTempo(pos Measure, tempo Tempo)
	TempoAt(abspos Measure) Tempo
	At(pos Measure, events ...*Event)
	MixPatterns(tf ...Pattern)
	CurrentBar() Measure
}
*/
// At(pos Measure, events ...*Event)

/*
type Stationer interface {
	At(pos Measure, events ...*Event)
}
*/

func RepeatLoop(n int, loop Looper) Looper {
	ls := make([]Looper, n)

	for i := 0; i < n; i++ {
		ls[i] = loop
	}

	return SeqLoop(ls...)
}

// LoopPattern creates a Looper based on the given Patterns.
// After each Pattern a bar change is introduced
func LoopPattern(patterns ...Pattern) Looper {
	if len(patterns) < 1 {
		panic("need at least 1 pattern")
	}
	l := Loop(patterns[0])
	if len(patterns) == 1 {
		return l
	}
	for _, pattern := range patterns[1:] {
		l.Next(pattern)
	}
	return l
}

type seqLoop []Looper

func SeqLoop(l ...Looper) Looper {
	return seqLoop(l)
}

func (l seqLoop) NumBars() int {
	//return len(l.events)
	num := 0

	for _, lo := range l {
		num += lo.NumBars()
	}
	return num
}

func (l seqLoop) BarEvents(bar int) map[Measure][]*Event {
	i := 0
	for _, lo := range l {
		if i+lo.NumBars() > bar {
			return lo.BarEvents(bar - i)
		}
		i += lo.NumBars()
	}
	return nil
}

type loop struct {
	// *Track
	currentBar int
	events     []map[Measure][]*Event
}

func (l *loop) BarEvents(bar int) map[Measure][]*Event {
	if bar > len(l.events)-1 {
		return nil
	}
	return l.events[bar]
}

func (l *loop) NumBars() int {
	return len(l.events)
}

func Loop(tr ...Pattern) *loop {
	l := &loop{}
	l.Next(tr...)
	return l
}

type Looper interface {
	BarEvents(bar int) map[Measure][]*Event
	NumBars() int
}

func (l *loop) At(pos Measure, events ...*Event) {
	l.events[l.currentBar][pos] = append(l.events[l.currentBar][pos], events...)
}

func (l *loop) nextBar() {
	l.events = append(l.events, map[Measure][]*Event{})
	l.currentBar = len(l.events) - 1
}

func (l *loop) Next(patterns ...Pattern) *loop {
	l.nextBar()

	for _, p := range patterns {
		p.Pattern(l)
	}
	return l
}

type loopInTrack struct {
	start        Measure
	currentIndex int
	loop         Looper
}

func (l *loopInTrack) setEventsForBar(t *Track) {
	for pos, events := range l.loop.BarEvents(l.currentIndex) {
		t.At(l.start+pos, events...)
	}
	if l.currentIndex >= l.loop.NumBars()-1 {
		l.currentIndex = 0
	} else {
		l.currentIndex++
	}
}

/*
type loopPattern struct {
	events   []*Event
	position Measure
}

func (l *loopPattern) Pattern(t Tracker) {
	for _, ev := range l.events {
		t.At(ev.absPosition+l.position, ev.Clone())
	}
}

func (l *Loop) Play(pos string) Pattern {
	return &loop2Pattern{l.Events, M(pos)}
}

func newLoop(tempo Tempo, m Measure) *Loop {
	newTrack(tempo, m)
}

var _Stationer = Loop2{}
*/
