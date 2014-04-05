package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestPipe(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)
	c := music.Note(4, 50)

	evts := Pipe(AddLength(10), AddHeight(3)).Transform(a, b, c)

	if len(evts) != 3 {
		t.Errorf("wrong number of returned events, should be 3, is: %v", len(evts))
	}

	a_mod := music.Note(5, 110)
	b_mod := music.Note(6, 210)
	c_mod := music.Note(7, 60)

	if !evts[0].Equals(a_mod) {
		t.Errorf("wrong event[0], should be a_mod")
	}

	if !evts[1].Equals(b_mod) {
		t.Errorf("wrong event[1], should be b_mod")
	}

	if !evts[2].Equals(c_mod) {
		t.Errorf("wrong event[2], should be c_mod")
	}

}
