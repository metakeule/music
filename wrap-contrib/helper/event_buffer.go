package helper

import "github.com/metakeule/music"

// EventBuffer is a EventWriter that may be used to spy on
// music.Handlers and keep what they have written.
// It may then be written to another (the real) EventWriter
type EventBuffer struct {
	music.EventWriter // necessary to allow "Unwrap from wrapstesting"
	events            []*music.Event
	changed           bool
}

// adds to the Events
func (f *EventBuffer) Write(ev ...*music.Event) {
	f.changed = true
	f.events = append(f.events, ev...)
}

// Reset set the EventBuffer to the defaults
func (f *EventBuffer) Reset() {
	f.events = []*music.Event{}
	f.changed = false
}

// WriteTo writes header, body and status code to the given EventWriter, if something changed
func (f *EventBuffer) WriteSerial(wr music.EventWriter) {
	if f.HasChanged() {
		for _, ev := range f.events {
			wr.Write(ev)
		}
	}
}

// WriteTo writes header, body and status code to the given EventWriter, if something changed
func (f *EventBuffer) WriteParallel(wr music.EventWriter) {
	if f.HasChanged() {
		wr.Write(f.events...)
	}
}

// Body returns the body as slice of bytes
func (f *EventBuffer) Events() []*music.Event {
	return f.events
}

// HasChanged returns true if something has been written to the EventBuffer
func (f *EventBuffer) HasChanged() bool { return f.changed }

// NewEventBuffer creates a new EventBuffer
func NewEventBuffer(w music.EventWriter) (f *EventBuffer) {
	f = &EventBuffer{}
	f.EventWriter = w
	return
}
