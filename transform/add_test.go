package transform

import (
	"testing"

	"github.com/metakeule/music"
)

func TestAddVolume(t *testing.T) {
	a := music.Note(2, 100)
	a.Volume = 0.1
	b := music.Note(3, 200)

	evts := AddVolume(0.4).Transform(a, b)

	if len(evts) != 2 {
		t.Errorf("wrong number of returned events, should be 2, is: %v", len(evts))
	}

	if evts[0].Volume != 0.5 {
		t.Errorf("wrong volume of event[0], is %v, should be %v", evts[0].Volume, 0.5)
	}

	if evts[1].Volume != 0.4 {
		t.Errorf("wrong volume of event[1], is %v, should be %v", evts[1].Volume, 0.4)
	}

	evts = AddVolume(-0.4).Transform(a, b)

	if evts[0].Volume != -0.3 {
		t.Errorf("wrong volume of event[0], is %v, should be %v", evts[0].Volume, -0.3)
	}

	if evts[1].Volume != -0.4 {
		t.Errorf("wrong volume of event[1], is %v, should be %v", evts[1].Volume, -0.4)
	}

}

func TestAddHeight(t *testing.T) {
	a := music.Note(2, 100)
	b := music.Note(3, 200)

	evts := AddHeight(2).Transform(a, b)

	if len(evts) != 2 {
		t.Errorf("wrong number of returned events, should be 2, is: %v", len(evts))
	}

	if evts[0].Height != 4 {
		t.Errorf("wrong height of event[0], is %v, should be %v", evts[0].Height, 4)
	}

	if evts[1].Height != 5 {
		t.Errorf("wrong height of event[1], is %v, should be %v", evts[1].Height, 5)
	}

	evts = AddHeight(-3).Transform(a, b)

	if evts[0].Height != -1 {
		t.Errorf("wrong height of event[0], is %v, should be %v", evts[0].Height, -1)
	}

	if evts[1].Height != 0 {
		t.Errorf("wrong height of event[1], is %v, should be %v", evts[1].Height, 0)
	}
}
