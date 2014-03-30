package wraps

import (
	"github.com/metakeule/music/wrap"

	"github.com/metakeule/music"
)

type repeat struct {
	times uint
	tr    music.Transformer
}

func Repeat(times uint, trafo music.Transformer) wrap.Wrapper {
	return &repeat{times, trafo}
}

// Wrap wraps the given inner handler with the returned handler
func (r *repeat) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
		all := []*music.Event{}
		for i := 0; i < int(r.times); i++ {
			all = append(all, r.tr.Transform(music.Events(evts...).Clone()...)...)
		}
		return inner.Transform(all...)
	})
}
