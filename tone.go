package music

// eine melodie iste eine abfolge von Noten
// type Sequence []*Note

// ein Tone ist eine konkretisierung einer Note
type Tone struct {
	Instrument           string             // instrument, welches die Note spielt, wenn leerer string: pause
	Start                uint               // startpunkt des tones in Ticks (feinste auflösung des stücks)
	Duration             uint               // dauer des tones in ticks
	Frequency            float64            // frequenz, mit der das instrument angesteuert wird
	InstrumentParameters map[string]float64 // parameter, mit denen das instrument angesteuert wird, dazu zählen auch die
	// ausgabe kanäle
	Amplitude float32 // amplitude, mit der das instrument angesteuert wird
}
