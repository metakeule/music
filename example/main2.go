package main

import (
	. "github.com/metakeule/music"
	"github.com/metakeule/music/bar"
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/player"
	"github.com/metakeule/music/rhythm"
	"github.com/metakeule/music/scale"
	"github.com/metakeule/music/transform"
)

func main() {
	p := player.NewSclang()

	base := &Event{
		Bar:        bar.Bar4To4,
		Scale:      scale.Dur(note.Dis),
		Instrument: "default",
		//InstrumentParams: map[string]float64{"dur": 0.01, "sustain": 0, "legato": 0.2, "sendGate": 1},
		InstrumentParams: map[string]float64{"ar": 4, "dr": 4},
		// InstrumentParams: map[string]float64{},
		Tempo: Tempo(35),
		//Tempo:  Tempo(50),
		Volume: 0.8,
		Rhythm: rhythm.NewPop(0.8, 0.6, 0.9, 0.3),
	}

	ticker := NewTicker(0, p)
	track := ticker.Start(base)

	// A == B == C

	// TODO: check with different tracks
	track.Serial(

		transform.Repeat(3, track.Serial(Note(3, 200))),
		// A
		track.Serial(Note(1, 100)),
		track.Parallel(
			Group(
				Note(3, 100),
				Note(5, 100),
			),
		),
		track.Serial(Note(1, 100)),
		track.Parallel(
			Group(
				Note(3, 100),
				Note(5, 100),
			),
		),
		// END OF A

		// B
		track.Serial(
			Note(1, 100),
			track.Parallel(
				Group(
					Note(3, 100),
					Note(5, 100),
				),
			),
			Note(1, 100),
			track.Parallel(
				Group(
					Note(3, 100),
					Note(5, 100),
				),
			),
		),
		// END OF B

		// C
		track.Parallel(
			track.Serial(
				Events(
					Note(1, 100),
					Note(3, 100),
					Note(1, 100),
					Note(3, 100),
				),
			),
			track.Serial(
				Events(
					Rest(100),
					Note(5, 100),
					Rest(100),
					Note(5, 100),
				),
			),
		),
	// END OF C
	).Transform()

	p.Play()
}
