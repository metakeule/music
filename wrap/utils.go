package wrap

import "github.com/metakeule/music"

func WrapperToTransformer(w Wrapper) music.Transformer {
	return w.Wrap(noop)
}

// WrapperFunc is a function that acts as Wrapper
type WrapperFunc func(music.Transformer) music.Transformer

// Wrap makes the WrapperFunc fullfill the Wrapper interface by calling itself.
func (wf WrapperFunc) Wrap(in music.Transformer) music.Transformer { return wf(in) }

// TransformerWrapper returns a Wrapper for a Transformer.
// The returned Wrapper simply runs the given handler and ignores the
// inner handler in the stack.
func Transformer(h music.Transformer) Wrapper {
	return TransformTransformerWrapperFunc(
		func(inner music.Transformer, w music.EventWriter, events []*music.Event) {
			h.Transform(w, events)
		},
	)
}

// HandlerFunc serves the same purpose as Handler but for a function of the type
// signature as TransformerFunc
func TransformerFunc(fn func(music.EventWriter, []*music.Event)) Wrapper {
	return TransformTransformerWrapperFunc(
		func(inner music.Transformer, w music.EventWriter, events []*music.Event) {
			fn(w, events)
		},
	)
}

// ServeHandler can serve the given request with the aid of the given handler
type Transform interface {
	// ServeHandler serves the given request with the aid of the given handler
	Transform(music.Transformer, music.EventWriter, []*music.Event)
}

// RunTransformationWrapper returns a Wrapper for a RunTransformation
func TransformWrapper(wh Transform) Wrapper {
	fn := func(inner music.Transformer, w music.EventWriter, events []*music.Event) {
		wh.Transform(inner, w, events)
	}
	return TransformTransformerWrapperFunc(fn)
}

// RunTransformationTransformer creates a Transformer by using the given RunTransformation
func TransformTransformer(wh Transform, inner music.Transformer) music.Transformer {
	return music.TransformerFunc(
		func(w music.EventWriter, events []*music.Event) {
			wh.Transform(inner, w, events)
		},
	)
}

// RunTransformationTransformerFunc serves the same purpose as RunTransformationTransformer but for a function of the type
// signature as RunTransformationModifyFunc
func TransformTransformerFunc(fn func(music.Transformer, music.EventWriter, []*music.Event), inner music.Transformer) music.Transformer {
	return music.TransformerFunc(
		func(w music.EventWriter, events []*music.Event) {
			fn(inner, w, events)
		},
	)
}

// ServeHandlerFunc is a function that handles the given request with the aid of the given handler
// and is a Wrapper
type TransformTransformerWrapperFunc func(inner music.Transformer, w music.EventWriter, events []*music.Event)

// Wrap makes the ServeHandlerFunc fullfill the Wrapper interface by calling itself.
func (f TransformTransformerWrapperFunc) Wrap(inner music.Transformer) music.Transformer {
	return music.TransformerFunc(
		func(w music.EventWriter, events []*music.Event) {
			f(inner, w, events)
		},
	)
}
