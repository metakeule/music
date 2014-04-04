package transform

import (
	"github.com/metakeule/music"
)

type AddVolume float32

// adds volume to each event
func (a AddVolume) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Volume = n.Volume + float32(a)
		res[i] = n
	}

	return res
}

type AddHeight int

// adds volume to each event
func (a AddHeight) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Height = n.Height + int(a)
		res[i] = n
	}

	return res
}

type AddLength int

// adds length to each event
func (a AddLength) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Length = uint(int(n.Length) + int(a))
		res[i] = n
	}

	return res
}

type AddInstrumentParams map[string]float64

// adds length to each event
func (a AddInstrumentParams) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()

		if n.InstrumentParams == nil {
			n.InstrumentParams = map[string]float64{}
		}

		for k, v := range a {
			n.InstrumentParams[k] = n.InstrumentParams[k] + v
		}

		res[i] = n
	}

	return res
}

type AddTempo int

// adds length to each event
func (a AddTempo) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i, e := range events {
		n := e.Clone()
		n.Tempo = music.Tempo(uint(int(uint(n.Tempo)) + int(a)))
		res[i] = n
	}

	return res
}
