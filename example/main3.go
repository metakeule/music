package main

import (
	. "github.com/metakeule/music"
	"github.com/metakeule/music/bar"
	"github.com/metakeule/music/chord"
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/player"
	"github.com/metakeule/music/rhythm"
	. "github.com/metakeule/music/transform"
)

func main() {
	p := player.NewSclang()

	base := &Event{
		Bar: bar.Bar4To4,
		//Scale: scale.Moll(note.C),
		//Scale:      chord.Moll(note.C),
		Scale:      chord.Aug(note.C),
		Instrument: "default",
		//InstrumentParams: map[string]float64{"dur": 0.01, "sustain": 0, "legato": 0.2, "sendGate": 1},
		InstrumentParams: map[string]float64{"ar": 4, "dr": 4},
		// InstrumentParams: map[string]float64{},
		Tempo: Tempo(40),
		//Tempo:  Tempo(50),
		Volume: 0.8,
		Rhythm: rhythm.NewPop(0.8, 0.6, 0.9, 0.3),
	}

	ticker := NewTicker(0, p)
	track := ticker.Start(base)

	track.Parallel(
		track.Serial(
			Each(
				Each(
					Each(
						Repeat(
							2,
							Group(Note(0, 100), Note(2, 50), Note(1, 100), Note(4, 50), Note(3, 75)),
						),
						SetScale{chord.Dur(note.C)},
						SetScale{chord.Moll(note.F)},
						SetScale{chord.Aug(note.Gis)},
						SetScale{chord.DurMin7(note.Cis)},
						SetScale{chord.Dim(note.A)},
					),

					Pass,
					Reverse,
				),
				Pass,
				AddHeight(1),
				AddHeight(-1),
				AddHeight(-3),
			),

			track.Serial(Note(0, 800)),
		),

		/*
			track.Serial(
				Repeat(
					200,
					Group(
						Note(-12, 100),
						Note(-9, 200),
					),
				),
				// Repeat(20, AddTempo(4)),
			),
		*/
	).Transform()

	p.Play()
}
