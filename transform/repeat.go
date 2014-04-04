package transform

import (
	"github.com/metakeule/music"
)

type repeat struct {
	times uint
	tr    music.Transformer
}

func (r *repeat) Transform(evts ...*music.Event) []*music.Event {
	all := []*music.Event{}
	for i := 0; i < int(r.times); i++ {
		all = append(all, r.tr.Transform(music.Group(evts...).Clone()...)...)
	}
	return all
}

func Repeat(times uint, trafo music.Transformer) music.Transformer {
	return &repeat{times, trafo}
}
