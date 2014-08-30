package music

/*

type Player []*Event

func (p Player) Play() {
	p.PlayWithOffset(0)
}

func (p Player) PlayWithOffset(startOffset uint) {
	// do the playing
}
*/

func millisecsToTick(ms float64) int {
	return int(RoundFloat(ms*1000000.0, 0))
}

func tickToSeconds(tick int) float32 { return float32(tick) / float32(1000000000) }

type EventPlayer interface {
	PlayDur(pos, dur string, params ...Parameter) Pattern
	Play(pos string, params ...Parameter) Pattern
	Stop(pos string) Pattern
	Modify(pos string, params ...Parameter) Pattern
}
