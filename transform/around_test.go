package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestAround(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)
	c := music.Note(4, 300)
	d := music.Note(5, 400)

	evts := Around([]*music.Event{c}, []*music.Event{b}).Transform(d, a)

	if len(evts) != 4 {
		t.Errorf("wrong number of returned events, should be 4, is: %v", len(evts))
	}

	if !evts[0].Equals(c) {
		t.Errorf("wrong event[0], should be c")
	}

	if !evts[1].Equals(d) {
		t.Errorf("wrong event[1], should be d")
	}

	if !evts[2].Equals(a) {
		t.Errorf("wrong event[2], should be a")
	}

	if !evts[3].Equals(b) {
		t.Errorf("wrong event[3], should be b")
	}

}
