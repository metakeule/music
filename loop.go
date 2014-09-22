package music

func RepeatPattern(n int, loop Pattern) Pattern {
	ls := make([]Pattern, n)

	for i := 0; i < n; i++ {
		ls[i] = loop
	}

	return SeqPatterns(ls...)
}

func Loop(patterns ...Pattern) *loop {
	l := &loop{}
	l.Next(patterns...)
	return l
}

type loop struct {
	currentBar int
	patterns   []Pattern
}

func (l *loop) Events(barNum int, t Tracker) map[Measure][]*Event {
	num := 0

	for _, p := range l.patterns {
		next := num + p.NumBars()
		if barNum < next {
			return p.Events(barNum-num, t)
		}
		num = next
	}
	return nil
}

func (l *loop) NumBars() int {
	num := 0

	for _, p := range l.patterns {
		num += p.NumBars()
	}
	return num
}

func (l *loop) Next(patterns ...Pattern) *loop {
	// l.nextBar()
	l.patterns = append(l.patterns, MixPatterns(patterns...))
	return l
}

type loopInTrack struct {
	start        Measure
	currentIndex int
	loop         Pattern
}

func (l *loopInTrack) setEventsForBar(t *Track) {
	for pos, events := range l.loop.Events(l.currentIndex, t) {
		t.At(l.start+pos, events...)
	}
	if l.currentIndex >= l.loop.NumBars()-1 {
		l.currentIndex = 0
	} else {
		l.currentIndex++
	}
}
