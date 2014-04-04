package transform

import (
	"github.com/metakeule/music"
)

type Skip []uint

func (s Skip) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	has := map[uint]bool{}

	for _, sk := range s {
		has[sk] = true
	}

	for i, e := range events {
		if has[uint(i)] {
			continue
		}

		n := e.Clone()
		res = append(res, n)
	}

	return res
}

type SkipEvery uint

func (s SkipEvery) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	for i, e := range events {
		n := e.Clone()
		if i%int(uint(s)) == 0 {
			continue
		}
		res = append(res, n)
	}

	return res
}

type PickEvery uint

func (p PickEvery) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	for i, e := range events {
		if i%int(uint(p)) == 0 {
			res = append(res, e.Clone())
		}
	}

	return res
}

type Pick []uint

func (p Pick) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	has := map[uint]bool{}

	for _, pk := range p {
		has[pk] = true
	}

	for i, e := range events {
		if has[uint(i)] {
			res = append(res, e.Clone())
		}
	}

	return res
}

type reverse struct{}

func (r reverse) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i := 0; i < len(events); i++ {
		res[i] = events[len(events)-1-i]
	}

	return res
}

var Reverse = reverse{}

type before []*music.Event

func (b before) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events)+len(b))

	for i, e := range b {
		res[i] = e.Clone()
	}

	for i, e := range events {
		res[i+len(b)-1] = e.Clone()
	}

	return res
}

type after []*music.Event

func (a after) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events)+len(a))

	for i, e := range events {
		res[i] = e.Clone()
	}

	for i, e := range a {
		res[i+len(events)-1] = e.Clone()
	}

	return res
}

func Before(events ...*music.Event) music.Transformer {
	return before(events)
}

func After(events ...*music.Event) music.Transformer {
	return after(events)
}

type around struct {
	before []*music.Event
	after  []*music.Event
}

func (a *around) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events)+len(a.before)+len(a.after))

	for i, e := range a.before {
		res[i] = e.Clone()
	}

	for i, e := range events {
		res[i+len(a.before)-1] = e.Clone()
	}

	for i, e := range a.after {
		res[i+len(events)-1+len(a.before)-1] = e.Clone()
	}

	return res
}

func Around(before []*music.Event, after []*music.Event) music.Transformer {
	return &around{before, after}
}
