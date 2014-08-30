package music

// the outer invoker may use the first voices instrument to query loadcode etc
func NewRoute(g Generator, name, path string, numVoices int) []*Voice {
	instr := &SCInstrument{
		name: name,
		Path: path,
	}
	return Voices(numVoices, g, instr, 1200)
}
