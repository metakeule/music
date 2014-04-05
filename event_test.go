package music

import "testing"

func TestEventEquals(t *testing.T) {

	a := Note(2, 100)
	b := Note(2, 100)
	c := Note(3, 100)

	if !a.Equals(b) {
		t.Errorf("a should be equal to b")
	}

	if !a.Equals(a.Clone()) {
		t.Errorf("a should be equal to its clone")
	}

	if a.Equals(c) {
		t.Errorf("a should not be equal to c")
	}

}
