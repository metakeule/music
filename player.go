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

func millisecsToTick(ms int) int { return ms * 1000000 }

func tickToSeconds(tick int) float32 { return float32(tick) / float32(1000000000) }

type EventPlayer interface {
	PlayDur(pos, dur string, params ...Parameter) Transformer
	Play(pos string, params ...Parameter) Transformer
	Stop(pos string) Transformer
	Modify(pos string, params ...Parameter) Transformer
}
