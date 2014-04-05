package transform

import "github.com/metakeule/music"

type around struct {
	before []*music.Event
	after  []*music.Event
}

func (a *around) Transform(events ...*music.Event) []*music.Event {
	// res := make([]*music.Event, len(events)+len(a.before)+len(a.after))
	res := []*music.Event{}

	for _, e := range a.before {
		// res[i] = e.Clone()
		res = append(res, e.Clone())
	}

	for _, e := range events {
		// res[i+len(a.before)-1] = e.Clone()
		res = append(res, e.Clone())
	}

	for _, e := range a.after {
		// res[i+len(events)-1+len(a.before)-1] = e.Clone()
		res = append(res, e.Clone())
	}

	return res
}

func Around(before []*music.Event, after []*music.Event) music.Transformer {
	return &around{before, after}
}
