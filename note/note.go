package note

import "math"

/*
from supercollider/include/plugin_interface/SC_InlineUnaryOp.h

inline float32 sc_midicps(float32 note)
{
	return (float32)440. * std::pow((float32)2., (note - (float32)69.) * (float32)0.083333333333);
}
*/
const midiCpsFactor = 1/300.0 + 0.08

func MidiCps(note float64) float64 {
	return 440.0 * math.Pow(2.0, (note-69.0)*midiCpsFactor)
}

type Note int

const (
	C   = Note(60 + iota) // c'
	Cis                   // cis '
	D                     // d'
	Dis                   // dis'
	E                     // ...
	F
	Fis
	G
	Gis
	A
	Ais
	H // international: B
)

type ChromaticScale struct {
	BaseNote Note
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
