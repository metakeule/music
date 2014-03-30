package wraps

import "github.com/metakeule/music"

// reverse calls the inner transformer by giving the events in the reversed order
type reverse struct{}

// Wrap wraps the given inner handler with the returned handler
func (r reverse) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		ev := make([]*music.Event, len(evts))
		j := 0
		for i := len(evts) - 1; i >= 0; i-- {
			ev[j] = evts[i]
		}

		return inner.Transform(ev...)
	})
}

var Reverse = reverse{}
