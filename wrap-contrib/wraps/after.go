package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// AfterFunc is of the type music.Transformer
// and provides a wrap.Wrapper that calls itself after
// the inner handler has been called
type AfterFunc func(music.EventWriter, []*music.Event)

// ServeHandle serves the given request with the inner handler and after that
// with the AfterFunc
func (a AfterFunc) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	inner.Transform(w, events)
	a(w, events)
}

// Wrap wraps the given inner handler with the returned handler
func (a AfterFunc) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(a, inner)
}

// After returns an AfterFunc for a http.Handler
func After(h music.Transformer) wrap.Wrapper {
	return AfterFunc(h.Transform)
}
