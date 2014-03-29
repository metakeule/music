package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/metakeule/music"
	"github.com/metakeule/music/bar"
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/rhythm"
	"github.com/metakeule/music/scale"
	"github.com/metakeule/music/wrap"
	"github.com/metakeule/music/wrap-contrib/wraps"
	"github.com/metakeule/sclang"
)

type player struct {
	tones []*music.Tone
}

func newPlayer() *player {
	return &player{}
}

func (p *player) Write(tone *music.Tone) {
	printTone(tone)
	p.tones = append(p.tones, tone)
}

type freesynths struct {
	*sync.RWMutex
	synths  []int
	pointer int
}

func (f *freesynths) add(id int) {
	f.Lock()

	defer f.Unlock()
	f.synths = append(f.synths, id)
}

func (f *freesynths) getList() []int {
	f.Lock()
	f.RLock()
	oldPointer := f.pointer
	newPointer := len(f.synths) - 1
	f.pointer = newPointer
	f.Unlock()
	defer f.Unlock()
	return f.synths[oldPointer:newPointer]
}

var freelist = &freesynths{&sync.RWMutex{}, []int{}, 0}

func (p *player) Play() {
	osc := &sclang.OscClient{}
	osc.Init()

	sc := sclang.NewSclang()
	sc.Quiet = true
	fn := func() {

		go func() {
			time.Sleep(1 * time.Second)
			list := freelist.getList()
			for _, l := range list {
				fmt.Printf("freeing %d\n", l)
				osc.Free(l)
			}
		}()

		for id, tone := range p.tones {
			// properties := map[string]interface{}{
			properties := map[string]float32{
				"freq": float32(tone.Frequency),
				"amp":  float32(tone.Amplitude),
				// "gate": 1,
			}

			for k, v := range tone.InstrumentParameters {
				properties[k] = float32(v)
			}
			osc.New(2, "default", 0, 0, map[string]interface{}{})
			osc.Set(2, properties)

			//osc.NewReplace(id, "default", 2, properties)
			//osc.NewTail(id, tone.Instrument, 0, properties)
			time.Sleep(time.Duration(int64(tone.Duration)) * time.Millisecond)
			/*
				osc.Set(id, map[string]float32{
					"amp": 0.0,
				})
			*/
			osc.Set(2, map[string]float32{
				"amp": 0.0,
				// "gate": 0,
			})

			_ = id
			// fmt.Printf("adding %d to freelist\n", id)
			// freelist.add(id)
			// time.Sleep(100 * time.Microsecond)
			// osc.Free(id)
			/*
				go func() {
					time.Sleep(10 * time.Microsecond)
					osc.Free(id)
				}()
			*/

			// time.Sleep(10 * time.Microsecond)
		}

		time.Sleep(2 * time.Second)
	}
	err := sc.BootAndExec(fn)

	if err != nil {
		fmt.Println(err)
	}

}

func toDots(nu uint) (s string) {
	n := int(nu / 180)
	for i := 0; i < n; i++ {
		s += "#"
	}
	return s
}

func printTone(v *music.Tone) {
	fmt.Printf("%v \t%0.2f\n", toDots(v.Duration), v.Frequency)
}

func startEvent() *music.Event {
	return &music.Event{
		Bar:        bar.Bar4To4,
		Scale:      scale.Dur(note.Dis),
		Instrument: "default",
		//InstrumentParams: map[string]float64{"dur": 0.01, "sustain": 0, "legato": 0.2, "sendGate": 1},
		InstrumentParams: map[string]float64{"ar": 4, "dr": 4},
		// InstrumentParams: map[string]float64{},
		//Tempo:  music.Tempo(35),
		Tempo:  music.Tempo(50),
		Volume: 0.8,
		Rhythm: rhythm.NewPop(0.8, 0.6, 0.9, 0.3),
	}
}

func main() {
	p := newPlayer()

	ende := wrap.New(
		wraps.After(&music.Event{Volume: 0.05, Height: 3, Length: 600}),
		wraps.Before(&music.Event{Height: 14}),
		wraps.EachBefore(
			&music.Event{Volume: 0.75},
			&music.Event{Volume: 0.50},
			&music.Event{Volume: 0.35},
			&music.Event{Volume: 0.11},
			&music.Event{Volume: 0.10},
		),
		wraps.Repeat(4),
		wraps.Before(music.Note(0, 40)),
		wraps.Before(music.Rest(20)),
		wraps.Before(music.Note(2, 40)),
		wraps.Before(music.Rest(20)),
		wraps.Before(music.Note(-2, 80)),
		wraps.Before(music.Rest(20)),
	)

	motiv := wrap.New(
		wraps.After(&music.Event{Height: -7}),
		wraps.Repeat(2),
		wraps.After(&music.Event{Height: 6}),
		wraps.Before(music.Note(-2, 60)),
		wraps.Before(music.Rest(40)),
		wraps.Repeat(3),
		wraps.After(&music.Event{Height: -5}),
		wraps.Repeat(3),
		wraps.Before(music.Note(1, 40)),
		wraps.Before(music.Rest(10)),
	)

	variation := wrap.New(
		wraps.EachBefore(
			&music.Event{Scale: scale.Dur(note.F)},
			&music.Event{Scale: scale.Moll(note.Gis)},
			&music.Event{Scale: scale.Dur(note.D), Height: -2},
			&music.Event{Scale: scale.Dur(note.Ais), Height: -5},
			&music.Event{Scale: scale.Dur(note.Cis)},
		),
		wraps.Before(motiv),
	)

	w := wrap.New(
		wraps.After(ende),
		// wraps.Repeat(4),
		// wraps.Before(&music.Event{Height: -4}),
		wraps.Before(variation),
	)

	ew := startEvent().EventWriter(p)

	_ = w

	//w.Transform(ew, nil)
	//motiv.Transform(ew, nil)
	variation.Transform(ew, nil)

	p.Play()
}
