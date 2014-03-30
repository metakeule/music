package wrap

import "github.com/metakeule/music"

// WrapperFunc is a function that acts as Wrapper
type WrapperFunc func(music.Transformer) music.Transformer

// Wrap makes the WrapperFunc fullfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(in music.Transformer) music.Transformer { return wf(in) }
