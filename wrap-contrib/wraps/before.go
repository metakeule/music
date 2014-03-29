package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// BeforeFunc is of the type http.HandlerFunc
// and provides a wrap.Wrapper that calls itself before
// the inner handler has been called
type BeforeFunc func(music.EventWriter, []*music.Event)

// ServeHandle serves the given request with the BeforeFunc and after that
// with the inner handler
func (a BeforeFunc) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	a(w, events)
	inner.Transform(w, events)
}

// Wrap wraps the given inner handler with the returned handler
func (a BeforeFunc) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(a, inner)
}

// Before returns an BeforeFunc for a http.Handler
func Before(h music.Transformer) wrap.Wrapper {
	return BeforeFunc(h.Transform)
}
