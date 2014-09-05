package music

type Loop struct {
	Pattern
	// Number of Bars that correspond to the length of the loop
	// must be > 0
	NumBars uint
}
