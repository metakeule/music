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
		Volume: 1.0,
		Rhythm: rhythm.Neutral,
	}

	ticker := NewTicker(0, p)
	track := ticker.Start(base)

	volDown := At(ticker, map[uint]Transformer{
		500:  AddVolume(-0.1),
		1500: AddVolume(-0.1),
		2000: AddVolume(-0.1),
		2500: AddVolume(-0.1),
		3000: AddVolume(-0.1),
	})

	track.Parallel(
		track.Serial(
			Each(
				Repeat(
					2,
					Group(Note(0, 100), Rest(100), Note(1, 100), Rest(100)),
				),
				Pipe(AddVolume(0.3), SetScale{chord.Dur(note.A)}),
			),
			track.Serial(Note(0, 800)),
		),
		track.Serial(
			Each(
				Repeat(
					2,
					Group(Rest(150), Note(2, 50), Rest(100), Note(3, 50)),
				),
				Pipe(volDown, SetScale{chord.Dur(note.A)}),
			),
		),
	).Transform()

	p.Play()
}
