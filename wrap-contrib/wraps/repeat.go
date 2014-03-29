package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// AfterFunc is of the type music.Transformer
// and provides a wrap.Wrapper that calls itself after
// the inner handler has been called
type Repeat uint

// ServeHandle serves the given request with the inner handler and after that
// with the AfterFunc
func (r Repeat) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	for i := 0; i < int(r); i++ {
		inner.Transform(w, events)
	}
}

// Wrap wraps the given inner handler with the returned handler
func (r Repeat) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(r, inner)
}
