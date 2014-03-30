package main

import (
	"fmt"
	"sort"
	"time"
	"github.com/metakeule/music/wrap"

	"github.com/metakeule/music"
	"github.com/metakeule/music/bar"
	"github.com/metakeule/music/note"
	"github.com/metakeule/music/rhythm"
	"github.com/metakeule/music/scale"

	// "github.com/metakeule/music_1/wrap-contrib/wraps"
	"github.com/metakeule/sclang"
)

type action interface {
	Play(*sclang.OscClient) (sleepUntilNext uint)
}

type funcAction struct {
	fn             func(*sclang.OscClient)
	Info           string
	sleepUntilNext uint
}

func (f *funcAction) Play(osc *sclang.OscClient) (sleepUntilNext uint) {
	f.fn(osc)
	return f.sleepUntilNext
}

func newPlayAction(performanceId int, tone *music.Tone) *funcAction {
	fn := func(osc *sclang.OscClient) {
		properties := map[string]interface{}{
			"freq": float64(tone.Frequency),
			"amp":  float64(tone.Amplitude),
		}

		for k, v := range tone.InstrumentParameters {
			properties[k] = float64(v)
		}
		osc.New(performanceId, tone.Instrument, 0, 0, properties)
	}

	// sleepUntilNext will be set, when we have the
	// complete actions including the stop actions, to the time diff to the next action
	return &funcAction{fn: fn, Info: fmt.Sprintf("play %0.2f (id: %d)", tone.Frequency, performanceId)}
}

// TODO: check how to correctly free them (ohne knacksen)
func newStopAction(performanceId int) *funcAction {
	fn := func(osc *sclang.OscClient) {
		// TODO not freeing them leaks synths
		// osc.Free(performanceId)
		osc.Set(performanceId, map[string]float32{
			"amp": 0,
		})
	}

	// sleepUntilNext will be set, when we have the
	// complete actions including the stop actions, to the time diff to the next action
	return &funcAction{fn: fn, Info: fmt.Sprintf("stop: %d", performanceId)}
}

type player struct {
	tones   map[uint][]*music.Tone
	actions []action
}

func newPlayer() *player {
	return &player{tones: map[uint][]*music.Tone{}}
}

func (p *player) Write(tones ...*music.Tone) {
	for _, tone := range tones {
		p.tones[tone.Start] = append(p.tones[tone.Start], tone)
	}
}

func (p *player) printActions() {
	for _, a := range p.actions {
		fa := a.(*funcAction)
		fmt.Printf("%s....%d\n", fa.Info, fa.sleepUntilNext)
	}
}

func (p *player) prepareActions() {
	/*
			   0. make a map of stopping points to performance ids
			   1. get the order of the starting points (uint)
			   2. go through tones in the order of the starting points and have a performance id for each tone

		        create a play action for the tone and save the stopping point in the stopping map
		        save the play action in a play map[uint(startingpoint)][]action

		     3. take the starting points from the play map and the stopping points from the stopping map
		        and put them in a unique action starting points array. sort it

		     4. iterate through the sorted action starting points array, checking, if there are play actions
		        and / or stopping actions at that point save the resulting actions in the actions array.
		        the sleeping time being the difference between the starting points
	*/

	// maps starting points for stopping actions to performance ids of instruments
	var stoppingPoints = map[int][]int{}

	var toneStartingpointsSorted = []int{}

	for start := range p.tones {
		toneStartingpointsSorted = append(toneStartingpointsSorted, int(start))
	}

	sort.Ints(toneStartingpointsSorted)

	var player = map[int][]*funcAction{}
	var startingPointsUniq = map[int]bool{}

	performanceId := 10

	for _, startingPoint := range toneStartingpointsSorted {
		for _, tone := range p.tones[uint(startingPoint)] {
			player[startingPoint] = append(player[startingPoint], newPlayAction(performanceId, tone))
			stoppingAt := startingPoint + int(tone.Duration)
			stoppingPoints[stoppingAt] = append(stoppingPoints[stoppingAt], performanceId)
			startingPointsUniq[startingPoint] = true
			startingPointsUniq[stoppingAt] = true
			performanceId++
		}
	}

	var allStartingPointsSorted = []int{}

	for start := range startingPointsUniq {
		allStartingPointsSorted = append(allStartingPointsSorted, start)
	}

	sort.Ints(allStartingPointsSorted)

	lastStartingPoint := 0
	var lastFuncAction *funcAction

	// this loop should flatten all starting and stopping events into a series of
	// actions with sleeping time in between the actions if necessary
	// to get the sleeping time, the last starting point and the last action is tracked
	// and the last action gets the sleeping time as difference between the current starting
	// point and the last one, when a new starting point is handled
	for _, startingPoint := range allStartingPointsSorted {

		// now we know the sleeping time for the last function or the last round
		if lastFuncAction != nil {
			lastFuncAction.sleepUntilNext = uint(startingPoint - lastStartingPoint)
		}

		for _, pa := range player[startingPoint] {
			p.actions = append(p.actions, pa)
			// each action could be the last in the round
			lastFuncAction = pa
		}

		for _, performanceId := range stoppingPoints[startingPoint] {
			pa := newStopAction(performanceId)
			p.actions = append(p.actions, pa)
			// each action could be the last in the round
			lastFuncAction = pa
		}

		lastStartingPoint = startingPoint
	}
}

func (p *player) Play() {
	p.prepareActions()
	p.printActions()

	osc := &sclang.OscClient{}
	osc.Init()

	sc := sclang.NewSclang()
	sc.Quiet = true
	fn := func() {
		for _, action := range p.actions {
			sleepingTime := action.Play(osc)
			if sleepingTime > 0 {
				time.Sleep(time.Duration(int64(sleepingTime)) * time.Millisecond)
			}
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

func printTone(t []*music.Tone) {
	fmt.Println("---------")
	for _, v := range t {
		fmt.Printf("%v \t%0.2f\n", toDots(v.Duration), v.Frequency)
	}

}

func main() {
	p := newPlayer()

	start := &music.Event{
		Bar:        bar.Bar4To4,
		Scale:      scale.Dur(note.Dis),
		Instrument: "default",
		//InstrumentParams: map[string]float64{"dur": 0.01, "sustain": 0, "legato": 0.2, "sendGate": 1},
		InstrumentParams: map[string]float64{"ar": 4, "dr": 4},
		// InstrumentParams: map[string]float64{},
		Tempo: music.Tempo(35),
		//Tempo:  music.Tempo(50),
		Volume: 0.8,
		Rhythm: rhythm.NewPop(0.8, 0.6, 0.9, 0.3),
	}

	//seq := start.Sequencer(p)

	// TODO: sollte genau umgekehrt sein:
	// jeder neue eintrag sollte nacheinander kommen und
	// stattdessen wraps.Parallel() nötig sein, um sachen parallel zu machen

	ticker := music.NewTicker(0, p)
	next := ticker.Serial(start)
	// parallel := ticker.Parallel(start)

	// _ = parallel
	_ = next
	//ticker.Serial().Transform(start, music.Note(1, 100), music.Note(1, 100))

	wrap.New(
		next.Serial(
			music.Note(1, 100),
			next.Serial(music.Note(1, 100)),
			next.Parallel(
				music.Events(
					music.Note(3, 100),
					music.Note(5, 100),
				),
			),
			music.Note(1, 100),
			// next.Serial(music.Note(1, 100)),
			next.Parallel(
				music.Events(
					music.Note(3, 100),
					music.Note(5, 100),
				),
			),
		),
		// wraps.Repeat(2,
		/*
			next.Parallel(
				next.Serial(
					music.Events(
						music.Note(1, 100),
						music.Note(3, 100),
						music.Note(1, 100),
						music.Note(3, 100),
					),
				),
				next.Serial(
					music.Events(
						music.Rest(100),
						music.Note(5, 100),
						music.Rest(100),
						music.Note(5, 100),
					),
				),
			),
		*/
		// ),
		/*
			next.Serial(
				music.Events(
					music.Note(2, 200),
					music.Note(4, 200),
					music.Rest(100),
				),
			),
		*/
		/*
			next.Parallel(
				music.Events(
					music.Rest(200),
					music.Note(3, 100),
					music.Note(5, 100),
				),
			),
			wraps.Repeat(2,
				next.Serial(
					music.Events(
						music.Note(4, 200),
						music.Note(2, 200),
						music.Rest(100),
					),
				),
			),
		*/
	).Transform()

	/*
		wrap.New(
			wrap.Transformer(music.Note(1, 100)),
			wraps.DeepSerial(
				operator.Repeat(
					8,
					operator.Pipe(
						transform.AddLength(10),
						transform.AddHeight(1),
					),
				),
			),
			wraps.Serial(
				music.Events(music.Note(1, 100), music.Note(4, 200)),
				music.Note(-2, 50),
				music.Events(music.Note(3, 100), music.Note(6, 100)),
				music.Note(-2, 50),
			),
			wrap.EventWriter(seq),
		).Write()
	*/
	p.Play()
}
