package player

func millisecsToTick(ms int) int { return ms * 1000000 }

func tickToSeconds(tick int) float32 { return float32(tick) / float32(1000000000) }
