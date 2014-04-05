package music

type group []*Event

func (e group) Clone() []*Event {
	res := make([]*Event, len(e))

	for i, ev := range e {
		res[i] = ev.Clone()
	}

	return res
}

func (e group) Transform(evts ...*Event) []*Event {
	return e.Clone()
}

func Group(events ...*Event) group {
	return group(events)
}
