package scale

import (
	"github.com/metakeule/music"
	"github.com/metakeule/music/note"
)

/*
	TODO: make cord scales
*/

type Scale interface {
	Degree(degree int) music.Parameter
}

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

func (s *Chromatic) Degree(scalePosition int) music.Parameter {
	return note.Note(float64(s.BaseNote) + float64(scalePosition))
	//return note.MidiCps(float64(s.BaseNote) + float64(scalePosition))
}

type Periodic struct {
	Steps             []uint // each steps in factor of chromatic steps, begins with the second step (first is basetone)
	NumChromaticSteps uint   // the number of chromatic steps that correspond to one scale periodicy
	BaseNote          note.Note
}

// TODO: test it
func (s *Periodic) Degree(scalePosition int) music.Parameter {
	// we need to calculate the position in terms of the chromatic scale
	// and then we return the frequency via MidiCps
	num := len(s.Steps)

	posInScale := scalePosition % num
	cycle := scalePosition / num

	temp := int(s.BaseNote) + (cycle * int(s.NumChromaticSteps))

	if posInScale == 0 {
		return note.Note(float64(temp))
		//return note.MidiCps(float64(temp))
	}

	if posInScale < 0 {
		for i := posInScale; i < 0; i++ {
			temp -= int(s.Steps[num+i])
		}
		return note.Note(float64(temp))
		// return note.MidiCps(float64(temp))
	}

	for i := 0; i < posInScale; i++ {
		temp += int(s.Steps[i])
	}
	return note.Note(float64(temp)) //  note.MidiCps(float64(temp))
}

func Ionian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 2, 1, 2, 2, 2, 1},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Dorian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 1, 2, 2, 2, 1, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Phrygian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{1, 2, 2, 2, 1, 2, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Lydian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 2, 2, 1, 2, 2, 1},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Mixolydian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 2, 1, 2, 2, 1, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Aeolian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{2, 1, 2, 2, 1, 2, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Locrian(base note.Note) *Periodic {
	return &Periodic{
		Steps:             []uint{1, 2, 2, 1, 2, 2, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Hypolydian(base note.Note) *Periodic {
	return Ionian(base)
}

func Hypomixolydian(base note.Note) *Periodic {
	return Dorian(base)
}

func Dur(base note.Note) *Periodic {
	return Ionian(base)
}

func Major(base note.Note) *Periodic {
	return Dur(base)
}

func Moll(base note.Note) *Periodic {
	return Aeolian(base)
}

func Minor(base note.Note) *Periodic {
	return Moll(base)
}

func Hypodorian(base note.Note) *Periodic {
	return Aeolian(base)
}

func Hypophrygian(base note.Note) *Periodic {
	return Locrian(base)
}

var Mood = map[string]func(base note.Note) *Periodic{
	"serious":  Dorian,
	"sad":      Hypodorian,
	"vehement": Phrygian,
	"tender":   Hypophrygian,
	"happy":    Lydian,
	"pious":    Hypolydian,
	"youthful": Mixolydian,
}
