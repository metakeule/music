package transform

import "github.com/metakeule/music"

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
		c := int(uint(s))
		j := i + 1
		n := e.Clone()
		if j%c == 0 {
			continue
		}
		res = append(res, n)
	}

	return res
}
