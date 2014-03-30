package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// After returns the outer transformer after the inner
func After(h music.Transformer) wrap.Wrapper {
	//return AfterFunc(h.Transform)
	return wrap.WrapperFunc(func(inner music.Transformer) music.Transformer {
		return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
			evts = inner.Transform(evts...)
			return h.Transform(evts...)
		})
	})
}
