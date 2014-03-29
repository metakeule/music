package helper

import "github.com/metakeule/music"

// EventBuffer is a EventWriter that may be used to spy on
// music.Handlers and keep what they have written.
// It may then be written to another (the real) EventWriter
type EventBuffer struct {
	music.EventWriter // necessary to allow "Unwrap from wrapstesting"
	events            []*music.Event
	changed           bool
	globals           music.Globals
}

// Header returns the music.Header
func (f *EventBuffer) Globals() music.Globals {
	f.changed = true
	return f.globals
}

// forks from the underlying
func (f *EventBuffer) Fork() music.EventWriter {
	return f.EventWriter.Fork()
}

// adds to the Events
func (f *EventBuffer) Write(e *music.Event) {
	f.changed = true
	f.events = append(f.events, e)
}

// Reset set the EventBuffer to the defaults
func (f *EventBuffer) Reset() {
	f.events = []*music.Event{}
	f.changed = false
	f.globals = f.EventWriter.Globals()
}

// WriteTo writes header, body and status code to the given EventWriter, if something changed
func (f *EventBuffer) WriteTo(wr music.EventWriter) {
	if f.HasChanged() {
		f.WriteGlobalsTo(wr)

		for _, ev := range f.events {
			wr.Write(ev)
		}
	}
}

// Body returns the body as slice of bytes
func (f *EventBuffer) Events() []*music.Event {
	return f.events
}

// HasChanged returns true if something has been written to the EventBuffer
func (f *EventBuffer) HasChanged() bool { return f.changed }

// WriteHeadersTo adds the headers to the given EventWriter
func (f *EventBuffer) WriteGlobalsTo(w music.EventWriter) {
	globals := w.Globals()
	for k, v := range f.globals {
		globals.Set(k, v)
	}
}

// NewEventBuffer creates a new EventBuffer
func NewEventBuffer(w music.EventWriter) (f *EventBuffer) {
	f = &EventBuffer{}
	f.EventWriter = w
	f.globals = music.NewGlobals()
	return
}
