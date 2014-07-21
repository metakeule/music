package player

import "testing"

func TestMillisecsToTick(t *testing.T) {
	corpus := []int{-5, -1, 0, 1, 5, 1000}

	for _, offset := range corpus {
		tick := millisecsToTick(offset)
		secs := tickToSeconds(tick)
		millisecs := int(secs * 1000)
		if millisecs != offset {
			t.Errorf("expected %d, got: %d", offset, millisecs)
		}
	}

}
