package wrap

// package github.com/metakeule/music/wrap is developed in analogy to github.com/go-on/wrap

import "github.com/metakeule/music"

// Wrapper can wrap a Modifier with another one
type Wrapper interface {
	// Wrap wraps an inner http.Handler with a new http.Handler that
	// is returned. The inner handler might be used in the scope of a
	// returned http.HandlerFunc.
	Wrap(inner music.Transformer) (outer music.Transformer)
}

// noop is an EventWriter that does nothing
var noop = music.TransformerFunc(func(events ...*music.Event) []*music.Event { return events })

// New returns a Modifier that runs a stack of the given wrappers.
// When the handler serves the request the first wrapper
// serves the request and may let the second wrapper (its "inner" wrapper) serve.
// The second wrapper may let the third wrapper serve and so on.
// The last wrapper has as "inner" wrapper the not exported noop handler that does nothing.
func New(wrapper ...Wrapper) (h music.Transformer) {
	h = noop
	for i := len(wrapper) - 1; i >= 0; i-- {
		// for i := 0; i < len(wrapper); i++ {
		h = wrapper[i].Wrap(h)
	}
	return
}
