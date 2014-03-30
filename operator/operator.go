package operator

import "github.com/metakeule/music"

// creates num new events by applying the transformer to the input
func Repeat(num int, t music.Transformer) music.Transformer {
	return music.TransformerFunc(func(events ...*music.Event) []*music.Event {
		res := []*music.Event{}
		currents := music.Events(events...).Clone()
		for i := 0; i < num; i++ {
			currents = t.Transform(music.Events(currents...).Clone()...)
			res = append(res, currents...)
		}
		// fmt.Printf("%#v\n", res)
		/*
			for _, e := range res {
				fmt.Printf("%d\n", e.Height)
			}
		*/
		return res
	})
}

// wendet die ergebnisse der Transformer nacheinander an
func Pipe(t ...music.Transformer) music.Transformer {
	return music.TransformerFunc(func(events ...*music.Event) []*music.Event {
		res := music.Events(events...).Clone()
		for _, tr := range t {
			res = tr.Transform(res...)
		}
		return res
	})
}
