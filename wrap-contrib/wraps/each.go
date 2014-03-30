package wraps

import "github.com/metakeule/music"

// eachBefore calls for each of the given transformer
// the transformer, followed by the inner transformer
type eachBefore []music.Transformer

// Wrap wraps the given inner handler with the returned handler
func (e eachBefore) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		for _, tr := range e {
			evts = tr.Transform(evts...)
		}
		return inner.Transform(evts...)
	})
}

func EachBefore(t ...music.Transformer) eachBefore {
	return eachBefore(t)
}

type eachAfter []music.Transformer

// Wrap wraps the given inner handler with the returned handler
func (e eachAfter) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		evts = inner.Transform(evts...)
		for _, tr := range e {
			evts = tr.Transform(evts...)
		}
		return evts
	})
}

func EachAfter(t ...music.Transformer) eachAfter {
	return eachAfter(t)
}
