package transform

import "github.com/metakeule/music"

type PickEvery uint

func (p PickEvery) Transform(events ...*music.Event) []*music.Event {
	res := []*music.Event{}

	for i, e := range events {
		c := int(uint(p))
		j := i + 1
		if j%c == 0 {
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
