package transform

import "github.com/metakeule/music"

type after []*music.Event

func (a after) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	for _, e := range events {
		res = append(res, e.Clone())
	}

	for _, e := range a {
		res = append(res, e.Clone())
	}

	return res
}

func After(events ...*music.Event) music.Transformer {
	return after(events)
}
