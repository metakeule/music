package transform

import (
	"sort"

	"github.com/metakeule/music"
)

type at struct {
	*music.Ticker
	last   uint
	trafos map[uint]music.Transformer
	sorted []int
}

func At(ticker *music.Ticker, trafos map[uint]music.Transformer) music.Transformer {
	a := &at{
		Ticker: ticker,
		last:   ticker.Current,
		trafos: trafos,
		sorted: []int{},
	}

	a.sort()
	return a
}

func (a *at) sort() {
	a.sorted = []int{}

	for pos := range a.trafos {
		a.sorted = append(a.sorted, int(pos))
	}

	sort.Ints(a.sorted)

}

func (a *at) Transform(events ...*music.Event) []*music.Event {
	// TODO
	// 1. get all transformers up to the current position of the ticker
	// 2. make a pipe of them
	// 3. use the pipe to transform the events
	// Pipe(...)

	trafos := []music.Transformer{}

	trafosIds := []uint{}
	current := a.Ticker.Current

	for _, i := range a.sorted {
		if uint(i) <= current {
			trafosIds = append(trafosIds, uint(i))
		}
	}

	for _, i := range trafosIds {
		trafos = append(trafos, a.trafos[i])
	}

	pipe := Pipe(trafos...)

	return pipe.Transform(music.Group(events...).Clone()...)
}
