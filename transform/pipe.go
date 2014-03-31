package transform

// wendet die ergebnisse der Transformer nacheinander an

import "github.com/metakeule/music"

func Pipe(t ...music.Transformer) music.Transformer {
	return music.TransformerFunc(func(events ...*music.Event) []*music.Event {
		res := music.Events(events...).Clone()
		for _, tr := range t {
			res = tr.Transform(res...)
		}
		return res
	})
}
