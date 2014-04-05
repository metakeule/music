package music

type Transformer interface {
	// Transform transforms a Slice of musical events
	Transform(events ...*Event) []*Event
}

type TransformerFunc func(events ...*Event) []*Event

func (t TransformerFunc) Transform(events ...*Event) []*Event {
	return t(events...)
}
