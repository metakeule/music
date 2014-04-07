package music

// eine skala liefert zu einer scalen position eine frequenz
// bei z.B. C-Dur müssen alle möglichen positionen aller möglichen frequencen berücksichtigt werden
// 0 ist die referenz-position, z.b. bei C-Dur das eingestrichene C. -8 wäre dann eine Oktave darunter
type Scale interface {
	Frequency(scalePosition int) float64
}

type Bar struct {
	NumBeats uint // the number of base units that fits into a bar
	// der kehrwert davon ist die länge einer basiseinheit (die in der Note verwendet wird)
	// in einem 4/4 takt ist NumBeats 4 in einem 6/8 takt 6
	TempoBar uint // number of Bars (not beats!) per Minute
	// TempoBar * NumBeats = Beats per Minute
}

type Tempo uint // geschwindigkeit in Ticks per Minute

type Rhythm interface {
	// Amplitude returns an amplitude factor that is multiplied by the current volume and passed to the instrument
	// depending on the position in a Bar and the question if it has an accent
	Amplitude(bar *Bar, pos uint, accent bool) float32

	// verzögerung in % der basiseinheit des takes (für den groove)
	// positiv (laid back) oder negativ (vorgezogen)
	// in abhänigkeit vom takt und von der position des taktes
	// verändert die startposition
	Delay(bar *Bar, pos uint) int
}
