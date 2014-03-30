package wraps

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/wrap"
)

// Before returns the outer transformer before the inner
func Before(h music.Transformer) wrap.Wrapper {
	//return BeforeFunc(h.Transform)
	return wrap.WrapperFunc(func(inner music.Transformer) music.Transformer {
		return music.TransformerFunc(func(evts ...*music.Event) []*music.Event {
			evts = h.Transform(evts...)
			return inner.Transform(evts...)
		})
	})
}
