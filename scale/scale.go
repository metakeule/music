package scale

import (
	"github.com/metakeule/music/note"
)

/*
	TODO: make cord scales
*/

type Chromatic struct {
	BaseNote note.Note
}

// a1 = 440hz = 69 (midi note)
/*
  international Name     Midinote
	[C-1]                  0
	[A-1]                  9
	C0                     12
	A0                     21
	C1                     24
	A1                     33
	C2                     36
	A2                     45
	C3                     48
	A3                     57
	C4 (c')                60
	A4 (a')                69
	C5 (c'')               72
	A5 (a'')               81
	C6 (c''')              84
	A6 (a''')              93
	C7 (c'''')             96
	A7 (a'''')            105
	C8 (c''''')           108
	A8 (a''''')           117
*/

func (s *Chromatic) Frequency(scalePosition uint) float64 {
	return note.MidiCps(float64(uint(s.BaseNote) + scalePosition))
}

type Periodic struct {
	Steps             []uint // each steps in factor of chromatic steps, begins with the second step (first is basetone)
	NumChromaticSteps uint   // the number of chromatic steps that correspond to one scale periodicy
	BaseNote          note.Note
}

// TODO: test it
func (s *Periodic) Frequency(scalePosition int) float64 {
	// we need to calculate the position in terms of the chromatic scale
	// and then we return the frequency via MidiCps
	num := len(s.Steps)

	posInScale := scalePosition % num
	cycle := scalePosition / num

	temp := int(s.BaseNote) + (cycle * int(s.NumChromaticSteps))

	if posInScale == 0 {
		return note.MidiCps(float64(temp))
	}

	if posInScale < 0 {
		for i := posInScale; i < 0; i++ {
			temp -= int(s.Steps[num+i])
		}
		return note.MidiCps(float64(temp))
	}

	for i := 0; i < posInScale; i++ {
		temp += int(s.Steps[i])
	}
	return note.MidiCps(float64(temp))
}

func Dur(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 2, 1, 2, 2, 2, 1},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Moll(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 1, 2, 2, 1, 2, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}
