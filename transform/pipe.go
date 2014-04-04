package transform

// wendet die ergebnisse der Transformer nacheinander an

import "github.com/metakeule/music"

func Pipe(t ...music.Transformer) music.Transformer {
	return music.TransformerFunc(func(events ...*music.Event) []*music.Event {
		res := music.Group(events...).Clone()
		for _, tr := range t {
			res = tr.Transform(res...)
		}
		return res
	})
}

func Each(source music.Transformer, t ...music.Transformer) music.Transformer {
	return music.TransformerFunc(func(events ...*music.Event) []*music.Event {
		src := music.Group(source.Transform(events...)...)
		all := []*music.Event{}
		for _, tr := range t {
			all = append(all, tr.Transform(src.Clone()...)...)
		}
		return all
	})
}

// returns the events
type pass struct{}

func (p pass) Transform(events ...*music.Event) []*music.Event {
	return music.Group(events...).Clone()
}

var Pass = pass{}
