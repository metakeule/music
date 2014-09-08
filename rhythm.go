package music

type rhythm struct {
	positions      []string
	v              *Voice
	positionsIndex int
	currentPos     float64
	startPos       float64
	Patterns       []Pattern
	//barIndex       int
	Loop *loop
}

func newRhythm(v *Voice, start string, pos ...string) *rhythm {
	l := &loop{}
	l.events = append(l.events, map[Measure][]*Event{})

	return &rhythm{positions: pos, v: v, startPos: _M(start), Patterns: []Pattern{nil}, Loop: l}
}

/*
func (r *rhythm) Pattern(t Tracker) {
	for _, p := range r.patterns {
		p.Pattern(t)
	}
}
*/

func (r *rhythm) BarEvents(bar int) map[Measure][]*Event {
	return r.Loop.BarEvents(bar)
}

func (r *rhythm) NumBars() int {
	return r.Loop.NumBars()
}

func (r *rhythm) currentMeasure() Measure {
	return Measure(int(r.startPos + r.currentPos))
}

func (r *rhythm) pos() (pos_ string) {
	pos_ = r.positions[r.positionsIndex]
	if r.positionsIndex < len(r.positions)-1 {
		r.positionsIndex++
	} else {
		r.positionsIndex = 0
	}
	return pos_
}

func (r *rhythm) Next() *rhythm {
	/*
		r.barIndex++
		r.Patterns = append(r.Patterns, nil)
	*/
	r.Loop.nextBar()
	r.currentPos = 0
	return r
}

/*
func (r *rhythm) addPattern(p Pattern) {
	r.Patterns[r.barIndex] = MixPatterns(r.Patterns[r.barIndex], p)
}
*/

func (r *rhythm) Play(params ...Parameter) *rhythm {
	//r.patterns = append(r.patterns, &play{r.currentMeasure(), r.v, MixParams(params...)})
	r.Loop.At(r.currentMeasure(), OnEvent(r.v, MixParams(params...)))
	// r.addPattern(&play{r.currentMeasure(), r.v, MixParams(params...)})
	r.currentPos += _M(r.pos())
	return r
}

func (r *rhythm) Stop() *rhythm {
	// r.patterns = append(r.patterns, &stop{r.currentMeasure(), r.v})
	r.Loop.At(r.currentMeasure(), OffEvent(r.v))
	// r.addPattern(&stop{r.currentMeasure(), r.v})
	r.currentPos += _M(r.pos())
	return r
}

func (r *rhythm) Modify(params ...Parameter) *rhythm {

	r.Loop.At(r.currentMeasure(), ChangeEvent(r.v, MixParams(params...)))
	// r.patterns = append(r.patterns, &mod{r.currentMeasure(), r.v, MixParams(params...)})
	// r.addPattern(&mod{r.currentMeasure(), r.v, MixParams(params...)})
	r.currentPos += _M(r.pos())
	return r
}
