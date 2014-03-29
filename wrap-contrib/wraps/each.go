package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// eachBefore calls for each of the given transformer
// the transformer, followed by the inner transformer
type eachBefore []music.Transformer

func (e eachBefore) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	for _, tr := range e {
		tr.Transform(w, events)
		inner.Transform(w, events)
	}
}

// Wrap wraps the given inner handler with the returned handler
func (e eachBefore) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(e, inner)
}

func EachBefore(t ...music.Transformer) eachBefore {
	return eachBefore(t)
}

type eachAfter []music.Transformer

// eachAfter calls for each of the given transformer
// the inner transformer, followed by the transformer
func (e eachAfter) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	for _, tr := range e {
		inner.Transform(w, events)
		tr.Transform(w, events)
	}
}

// Wrap wraps the given inner handler with the returned handler
func (e eachAfter) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(e, inner)
}

func EachAfter(t ...music.Transformer) eachAfter {
	return eachAfter(t)
}
