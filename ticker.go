package music

type Ticker struct {
	Current uint
	ToneWriter
}

func NewTicker(startTick uint, w ToneWriter) *Ticker {
	return &Ticker{startTick, w}
}

func (t *Ticker) Start(base *Event) *Start {
	base.MustBeComplete()
	return &Start{t, t.ToneWriter, base}
}

type Start struct {
	Ticker *Ticker
	ToneWriter
	start *Event
}

type parallel struct {
	trs    []Transformer
	Ticker *Ticker
	ToneWriter
	start *Event
}

func (p *parallel) Transform(events ...*Event) []*Event {
	// first: take all transformer events
	allEvents := []*Event{}
	currentTicker := p.Ticker.Current

	for _, tr := range p.trs {
		p.Ticker.Current = currentTicker
		allEvents = append(allEvents, tr.Transform(Group(events...).Clone()...)...)
	}

	all := []*Tone{}
	for _, ev := range allEvents {
		var t *Tone
		_, _, t = ev.ApplyTo(p.start).Tone(currentTicker, 0)
		all = append(all, t)
	}
	p.ToneWriter.Write(all...)
	if len(allEvents) > 0 {
		p.Ticker.Current, _, _ = allEvents[0].ApplyTo(p.start).duration(currentTicker, 0)
	}
	return []*Event{}
}

func (p *parallel) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(events ...*Event) []*Event {
		p.Transform(events...)
		return inner.Transform()
	})
}

func (s *Start) Parallel(trs ...Transformer) *parallel {
	return &parallel{
		trs:        trs,
		Ticker:     s.Ticker,
		ToneWriter: s.ToneWriter,
		start:      s.start,
	}
}

type serial struct {
	trs    []Transformer
	Ticker *Ticker
	ToneWriter
	start *Event
}

func (s *Start) Serial(trs ...Transformer) *serial {
	return &serial{
		trs:        trs,
		Ticker:     s.Ticker,
		ToneWriter: s.ToneWriter,
		start:      s.start,
	}
}

func (s *serial) Transform(events ...*Event) []*Event {
	all := []*Tone{}
	currentPositionInBar := uint(0)

	for _, tr := range s.trs {
		for _, ev := range tr.Transform(Group(events...).Clone()...) {
			var t *Tone
			s.Ticker.Current, currentPositionInBar, t = ev.ApplyTo(s.start).Tone(s.Ticker.Current, currentPositionInBar)
			all = append(all, t)
		}

	}
	s.ToneWriter.Write(all...)
	return []*Event{}
}

func (s *serial) Wrap(inner Transformer) Transformer {
	return TransformerFunc(func(events ...*Event) []*Event {
		s.Transform(events...)
		return inner.Transform()
	})
}
