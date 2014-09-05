package music

type phrase struct {
	voice      *Voice
	startPos   float64
	currentPos float64
	patterns   []Pattern
}

func (m *phrase) Pattern(t Tracker) {
	for _, p := range m.patterns {
		p.Pattern(t)
	}
}

func newPhrase(v *Voice, pos string) *phrase {
	return &phrase{
		voice:    v,
		startPos: _M(pos),
	}
}

func (m *phrase) currentMeasure() Measure {
	return Measure(int(m.startPos + m.currentPos))
}

func (m *phrase) Play(distance string, params ...Parameter) *phrase {
	m.currentPos += _M(distance)
	m.patterns = append(m.patterns, &play{m.currentMeasure(), m.voice, MixParams(params...)})
	return m
}

func (m *phrase) Stop(distance string) *phrase {
	m.currentPos += _M(distance)
	m.patterns = append(m.patterns, &stop{m.currentMeasure(), m.voice})
	return m
}

func (m *phrase) Modify(distance string, params ...Parameter) *phrase {
	m.currentPos += _M(distance)
	m.patterns = append(m.patterns, &mod{m.currentMeasure(), m.voice, MixParams(params...)})
	return m
}
