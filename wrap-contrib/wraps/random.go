package wraps

import (
	"math/rand"

	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// randomBefore calls a randomly chosen transformer from the slice and runs it before the
// inner transformer
type randomBefore []music.Transformer

func init() {
	rand.Seed(232349342)
}

func (e randomBefore) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	e[rand.Intn(len(e))].Transform(w, events)
	inner.Transform(w, events)
}

// Wrap wraps the given inner handler with the returned handler
func (e randomBefore) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(e, inner)
}

func RandomBefore(t ...music.Transformer) randomBefore {
	return randomBefore(t)
}

type randomAfter []music.Transformer

// randomBefore calls a randomly chosen transformer from the slice and runs it after the
// inner transformer
func (e randomAfter) Transform(inner music.Transformer, w music.EventWriter, events []*music.Event) {
	inner.Transform(w, events)
	e[rand.Intn(len(e))].Transform(w, events)
}

// Wrap wraps the given inner handler with the returned handler
func (e randomAfter) Wrap(inner music.Transformer) music.Transformer {
	return wrap.TransformTransformer(e, inner)
}

func RandomAfter(t ...music.Transformer) randomAfter {
	return randomAfter(t)
}
