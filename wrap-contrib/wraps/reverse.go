package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// reverse calls the inner transformer by giving the events in the reversed order
type reverse struct{}

func (r reverse) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	ev := make([]*music.Event, len(events))
	j := 0
	for i := len(events) - 1; i >= 0; i-- {
		ev[j] = events[i]
	}

	inner.Transform(w, ev)
}

// Wrap wraps the given inner handler with the returned handler
func (r reverse) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(r, inner)
}

var Reverse = reverse{}
