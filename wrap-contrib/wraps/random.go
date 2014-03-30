package wraps

import (
	"math/rand"

	"github.com/metakeule/music"
)

// randomBefore calls a randomly chosen transformer from the slice and runs it before the
// inner transformer
type randomBefore []music.Transformer

func init() {
	rand.Seed(232349342)
}

// Wrap wraps the given inner handler with the returned handler
func (e randomBefore) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		return inner.Transform(e[rand.Intn(len(e))].Transform(evts...)...)
	})
}

func RandomBefore(t ...music.Transformer) randomBefore {
	return randomBefore(t)
}

type randomAfter []music.Transformer

// randomBefore calls a randomly chosen transformer from the slice and runs it after the
// inner transformer
// Wrap wraps the given inner handler with the returned handler
func (e randomAfter) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		return e[rand.Intn(len(e))].Transform(inner.Transform(evts...)...)
	})
}

func RandomAfter(t ...music.Transformer) randomAfter {
	return randomAfter(t)
}
