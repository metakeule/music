package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestSkip(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	// a b b
	evts := Skip([]uint{2, 3, 5}).Transform(a, b, a, a, b, b)

	if len(evts) != 3 {
		t.Errorf("wrong number of returned events, should be 3, is: %v", len(evts))
	}

	if !evts[0].Equals(a) {
		t.Errorf("wrong event[0], should be a")
	}

	if !evts[1].Equals(b) {
		t.Errorf("wrong event[1], should be b")
	}

	if !evts[2].Equals(b) {
		t.Errorf("wrong event[2], should be b")
	}

}

func TestSkipEvery(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	// a a b
	evts := SkipEvery(2).Transform(a, b, a, a, b, b)

	if len(evts) != 3 {
		t.Errorf("wrong number of returned events, should be 3, is: %v", len(evts))
	}

	if !evts[0].Equals(a) {
		t.Errorf("wrong event[0], should be a")
	}

	if !evts[1].Equals(a) {
		t.Errorf("wrong event[1], should be a")
	}

	if !evts[2].Equals(b) {
		t.Errorf("wrong event[2], should be b")
	}

	// a b a b
	evts = SkipEvery(3).Transform(a, b, a, a, b, b)

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
