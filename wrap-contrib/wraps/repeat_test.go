package wraps

import (
	"testing"

	"github.com/metakeule/music/wrap"
	// . "github.com/go-on/wrap-contrib/helper"
)

func TestRepeat(t *testing.T) {
	h := wrap.New(
		Repeat(3),
		wrap.Transformer(Rest(200)),
	)
	rec := newRecord()
	ew := NewEventWriter(rec)

	h.Transform(ew, nil)
	expected := "- @0 1500\n- @1500 1500\n- @3000 1500\n"
	if rec.String() != expected {
		t.Errorf(`expected %#v, got %#v`, expected, rec.String())
	}
}
