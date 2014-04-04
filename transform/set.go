package transform

import "github.com/metakeule/music"

type SetRest bool

func (s SetRest) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Rest = bool(s)
		res[i] = n
	}

	return res
}

type SetAccent bool

func (s SetAccent) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Accent = bool(s)
		res[i] = n
	}

	return res
}

type SetRhythm struct {
	music.Rhythm
}

func (s SetRhythm) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Rhythm = s.Rhythm
		res[i] = n
	}

	return res
}

type SetInstrument string

func (s SetInstrument) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Instrument = string(s)
		res[i] = n
	}

	return res
}

type SetBar struct {
	*music.Bar
}

func (s SetBar) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Bar = s.Bar
		res[i] = n
	}

	return res
}

type SetScale struct {
	music.Scale
}

func (s SetScale) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Scale = s.Scale
		res[i] = n
	}

	return res
}

type SetVolume float32

func (a SetVolume) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Volume = float32(a)
		res[i] = n
	}

	return res
}

type SetHeight int

func (a SetHeight) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Height = int(a)
		res[i] = n
	}

	return res
}

type SetLength int

func (a SetLength) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Length = uint(int(a))
		res[i] = n
	}

	return res
}

type SetInstrumentParams map[string]float64

func (a SetInstrumentParams) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()

		if n.InstrumentParams == nil {
			n.InstrumentParams = map[string]float64{}
		}

		for k, v := range a {
			n.InstrumentParams[k] = v
		}

		res[i] = n
	}

	return res
}

type SetTempo int

func (a SetTempo) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Tempo = music.Tempo(uint(int(a)))
		res[i] = n
	}

	return res
}
