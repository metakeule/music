package wraps

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/metakeule/music"
	"github.com/metakeule/music/bar"
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/rhythm"
	"github.com/metakeule/music/scale"
	"github.com/metakeule/music/wrap"
	// . "github.com/metakeule/music/wrap-contrib/helper"
)

type record struct {
	*bytes.Buffer
}

func newRecord() *record {
	return &record{&bytes.Buffer{}}
}

func (r *record) Rec(v *music.Tone) {
	if v.Instrument == "" {
		fmt.Fprintf(r.Buffer, "- @%v %v\n", v.Start, v.Duration)
	} else {
		fmt.Fprintf(r.Buffer, "[%s(%v)] @%v %v %0.2fhz amp %v\n", v.Instrument, v.InstrumentParameters, v.Start, v.Duration, v.Frequency, v.Amplitude)
	}
}

func NewEventWriter(r *record) music.EventWriter {
	start := &music.Event{}
	start.Bar = bar.Bar4To4
	start.Scale = scale.Dur(note.C)
	start.Instrument = "piano"
	start.InstrumentParams = map[string]int{"legato": 1}
	start.Tempo = music.Tempo(100)
	start.Volume = 0.5
	start.Rhythm = rhythm.NewPop(0.8, 0.6, 0.9, 0.3)

	tonewriter := music.ToneWriterFunc(r.Rec)
	return start.EventWriter(tonewriter)
}

type Rest uint

func (r Rest) Transform(w music.EventWriter, events []*music.Event) {
	w.Write(music.Rest(uint(r)))
}

func TestAfter(t *testing.T) {
	h := wrap.New(
		After(Rest(200)),
		wrap.Transformer(Rest(100)),
	)

	rec := newRecord()
	ew := NewEventWriter(rec)

	h.Transform(ew, nil)

	expected := "- @0 750\n- @750 1500\n"
	if rec.String() != expected {
		t.Errorf(`expected %#v, got %#v`, expected, rec.String())
	}
}
