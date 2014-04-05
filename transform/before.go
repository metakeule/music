package transform

import "github.com/metakeule/music"

type before []*music.Event

func (b before) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	for _, e := range b {
		res = append(res, e.Clone())
	}

	for _, e := range events {
		res = append(res, e.Clone())
	}

	return res
}

func Before(events ...*music.Event) music.Transformer {
	return before(events)
}
