package transform

import "github.com/metakeule/music"

type reverse struct{}

func (r reverse) Transform(events ...*music.Event) []*music.Event {
	res := make([]*music.Event, len(events))

	for i := 0; i < len(events); i++ {
		res[i] = events[len(events)-1-i]
	}

	return res
}

var Reverse = reverse{}
