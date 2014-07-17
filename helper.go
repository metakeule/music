package music

import "strconv"

// stolen from https://groups.google.com/forum/#!topic/golang-nuts/ITZV08gAugI
func RoundFloat(x float64, prec int) float64 {
	frep := strconv.FormatFloat(x, 'g', prec, 64)
	f, _ := strconv.ParseFloat(frep, 64)
	return f
}

type TempoSorted []tempoAt

func (t TempoSorted) Len() int      { return len(t) }
func (t TempoSorted) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t TempoSorted) Less(i, j int) bool {
	return t[i].AbsPos < t[j].AbsPos
}

type EventsSorted []*Event

func (t EventsSorted) Len() int      { return len(t) }
func (t EventsSorted) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t EventsSorted) Less(i, j int) bool {
	return t[i].AbsPosition < t[j].AbsPosition
}
