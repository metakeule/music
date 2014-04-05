package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestReverse(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)
	c := music.Note(4, 50)

	evts := Reverse.Transform(a, b, c)

	if len(evts) != 3 {
		t.Errorf("wrong number of returned events, should be 3, is: %v", len(evts))
	}

	if !evts[0].Equals(c) {
		t.Errorf("wrong event[0], should be c")
	}

	if !evts[1].Equals(b) {
		t.Errorf("wrong event[1], should be b")
	}

	if !evts[2].Equals(a) {
		t.Errorf("wrong event[2], should be a")
	}

}
