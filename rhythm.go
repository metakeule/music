package music

type rhythm struct {
	positions      []string
	v              *Voice
	positionsIndex int
	currentPos     float64
	startPos       float64
	patterns       []Pattern
}

func newRhythm(v *Voice, start string, pos ...string) *rhythm {
	return &rhythm{positions: pos, v: v, startPos: _M(start)}
}

func (r *rhythm) Pattern(t Tracker) {
	for _, p := range r.patterns {
		p.Pattern(t)
	}
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

func (r *rhythm) Play(params ...Parameter) *rhythm {
	r.patterns = append(r.patterns, &play{r.currentMeasure(), r.v, MixParams(params...)})
	r.currentPos += _M(r.pos())
	return r
}

func (r *rhythm) Stop() *rhythm {
	r.patterns = append(r.patterns, &stop{r.currentMeasure(), r.v})
	r.currentPos += _M(r.pos())
	return r
}

func (r *rhythm) Modify(params ...Parameter) *rhythm {
	r.patterns = append(r.patterns, &mod{r.currentMeasure(), r.v, MixParams(params...)})
	r.currentPos += _M(r.pos())
	return r
}
