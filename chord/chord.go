package chord

import (
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/scale"
)

// a chord is a scale but typically has less tones
func Dim(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		Steps:             []uint{3, 3, 6},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Aug(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		Steps:             []uint{4, 4, 4},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

// a chord is a scale but typically has less tones
func Dur(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		Steps:             []uint{4, 3, 5},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func DurMaj7(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		// C Cis D Dis E  F Fis G   Gis A Ais B C
		// 0   1 2 e   4+ 1 2   3+  1   2 3   4+
		Steps:             []uint{4, 3, 4, 1},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func DurMin7(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		// C Cis D Dis E  F Fis G   Gis A Ais B C
		// 0   1 2 e   4+ 1 2   3+  1   2 3   4+
		Steps:             []uint{4, 3, 3, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func Moll(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		Steps:             []uint{3, 4, 5},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func MollMaj7(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		// C Cis D Dis E  F Fis G   Gis A Ais B C
		// 0   1 2 e   4+ 1 2   3+  1   2 3   4+
		Steps:             []uint{3, 4, 4, 1},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}

func MollMin7(base note.Note) *scale.Periodic {
	return &scale.Periodic{
		// C Cis D Dis E  F Fis G   Gis A Ais B C
		// 0   1 2 e   4+ 1 2   3+  1   2 3   4+
		Steps:             []uint{3, 4, 3, 2},
		NumChromaticSteps: 12,
		BaseNote:          base,
	}
}
