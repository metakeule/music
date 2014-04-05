package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestPick(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	evts := Pick([]uint{2, 3, 5}).Transform(a, b, a, a, b, b)

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

}

func TestPickEvery(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	evts := PickEvery(2).Transform(a, b, a, a, b, b)

	if len(evts) != 3 {
		t.Errorf("wrong number of returned events, should be 3, is: %v", len(evts))
	}

	if !evts[0].Equals(b) {
		t.Errorf("wrong event[0], should be b")
	}

	if !evts[1].Equals(a) {
		t.Errorf("wrong event[1], should be a")
	}

	if !evts[2].Equals(b) {
		t.Errorf("wrong event[2], should be b")
	}

	evts = PickEvery(3).Transform(a, b, a, a, b, b)

	if len(evts) != 2 {
		t.Errorf("wrong number of returned events, should be 2, is: %v", len(evts))
	}

	if !evts[0].Equals(a) {
		t.Errorf("wrong event[0], should be a")
	}

	if !evts[1].Equals(b) {
		t.Errorf("wrong event[1], should be b")
	}

}
