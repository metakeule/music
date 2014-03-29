package wraps

import (
	"testing"

	"github.com/metakeule/music/wrap"
	// . "github.com/go-on/wrap-contrib/helper"
)

func TestBefore(t *testing.T) {
	h := wrap.New(
		Before(Rest(200)),
		wrap.Transformer(Rest(100)),
	)
	rec := newRecord()
	ew := NewEventWriter(rec)

	h.Transform(ew, nil)
	expected := "- @0 1500\n- @1500 750\n"
	if rec.String() != expected {
		t.Errorf(`expected %#v, got %#v`, expected, rec.String())
	}
}
