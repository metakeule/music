package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestRepeat(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	evts := Repeat(2, Pass).Transform(a, b)

	if len(evts) != 4 {
		t.Errorf("wrong number of returned events, should be 4, is: %v", len(evts))
	}

	if !evts[0].Equals(a) {
		t.Errorf("wrong event[0], should be a")
	}

	if !evts[1].Equals(b) {
		t.Errorf("wrong event[1], should be b")
	}

	if !evts[2].Equals(a) {
		t.Errorf("wrong event[2], should be a")
	}

	if !evts[3].Equals(b) {
		t.Errorf("wrong event[3], should be b")
	}

}
