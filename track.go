package music

import (
	"fmt"
	"io"
	"runtime/debug"
	"sort"
)

type Track struct {
	Bars   []Measure
	absPos Measure
	tempi  []tempoAt
	Events []*Event
	// loops  []Pattern
	loops    map[string]Pattern
	compiled bool
	started  bool
}

func newTrack(tempo Tempo, m Measure) *Track {
	return &Track{
		absPos: Measure(0),
		Bars:   []Measure{m},
		loops:  map[string]Pattern{},
		tempi:  []tempoAt{tempoAt{AbsPos: Measure(0), Tempo: tempo}},
	}
}

func (t *Track) SetLoop(name string, patterns ...Pattern) *Track {
	t.loops[name] = Patterns(patterns...)
	return t
}

// RemoveLoop removes all loops if name == ""
func (t *Track) RemoveLoop(name string) *Track {
	if name == "" {
		t.loops = map[string]Pattern{}
		return t
	}
	delete(t.loops, name)
	return t
}

func (t *Track) GetLoop(name string) Pattern {
	return t.loops[name]
}

func (t *Track) Start(patterns ...Pattern) *Track {
	for _, pt := range t.loops {
		t.Patterns(pt)
	}
	t.Patterns(patterns...)
	t.started = true
	return t
}

func (t *Track) nextBar() {
	if !t.started {
		panic("call Start() before Next()/Fill()")
	}
	t.Bars = append(t.Bars, t.Bars[len(t.Bars)-1])
	t.absPos = t.absPos + t.Bars[len(t.Bars)-1]
	for _, pt := range t.loops {
		t.Patterns(pt)
	}
}

func (t *Track) changeBar(newBar Measure) {
	if !t.started {
		panic("call Start() before Change()")
	}
	t.Bars = append(t.Bars, newBar)
	t.absPos = t.absPos + t.Bars[len(t.Bars)-1]
	for _, pt := range t.loops {
		t.Patterns(pt)
	}
}

// raster is, how many ticks will equal to 3 chars width
func (t *Track) print(tempo Tempo, unit string, wr io.Writer) {
	raster := int(tempo.MilliSecs(M(unit)))

	if !t.compiled {
		panic("track is not compiled")
	}
	// events must be sorted by tick
	// Voice => Tick => Event
	voiceLines := map[*Voice]map[int]*Event{}
	for _, ev := range t.Events {
		// ev.Runner(ev)
		// fmt.Printf("%v %v %v\n", ev.Tick, ev.Type, ev.Voice.Name())

		eventMap, has := voiceLines[ev.Voice]

		if !has {
			eventMap = map[int]*Event{}
		}

		eventMap[int(ev.tick)] = ev

		voiceLines[ev.Voice] = eventMap
	}

	for _, evts := range voiceLines {
		tick := int(0)
		placeholder := " "
		sortedTickKeys := []int{}

		for t := range evts {
			sortedTickKeys = append(sortedTickKeys, int(t))
		}

		sort.Ints(sortedTickKeys)

		for j, sk := range sortedTickKeys {
			// 1. print 3 times / raster _ for any tick between the last and the current
			diff := sk - tick

			d := diff / raster

			ev := evts[sk]

			if j == 0 {
				fmt.Fprint(wr, ev.Voice.instrument.Name()+": ")
				// fmt.Print(ev.Voice.Name() + " ")
			}
			if d > 0 {
				for i := 0; i < d; i++ {
					if placeholder == "-" && i == 0 {
						continue
					}
					//				if placeholder == " " {
					//				fmt.Fprint(wr, placeholder+placeholder+placeholder)
					//		} else {
					fmt.Fprint(wr, placeholder+placeholder)
					//	}
					// fmt.Print(placeholder)
				}

			}

			switch ev.Type {
			case "ON":
				//fmt.Print("<")
				fmt.Fprintf(wr, "%v-", ev.Params.Params()["note"])
				// fmt.Print(ev.FinalParams()["note"])
				placeholder = "-"
			case "OFF":
				// fmt.Fprint(wr, "|")
				// fmt.Print("|")
				placeholder = " "
			case "CHANGE":
				fmt.Fprint(wr, "##")
				fmt.Fprint(wr, placeholder)
				// fmt.Print("#")
			}

			tick = int(ev.tick)
			//ev.Voice
		}

		fmt.Fprint(wr, "\n")
		// fmt.Print("\n")
	}

}

// raster is, how many ticks will equal to 3 chars width
/*
func (t *Track) Debug() {
	if !t.IsCompiled() {
		panic("track is not compiled")
	}
	// events must be sorted by tick
	// Voice => Tick => Even
	voiceLines := map[Voice]map[int]*Event{}

	for _, ev := range t.Events {
		// ev.Runner(ev)

		eventMap, has := voiceLines[ev.Voice]

		if !has {
			eventMap = map[int]*Event{}
		}

		eventMap[int(ev.Tick)] = ev

		voiceLines[ev.Voice] = eventMap
	}

	for _, evts := range voiceLines {

		sortedTickKeys := []int{}

		for t := range evts {
			sortedTickKeys = append(sortedTickKeys, int(t))
		}

		sort.Ints(sortedTickKeys)

		for j, sk := range sortedTickKeys {

			ev := evts[sk]

			if j == 0 {
				fmt.Print(ev.Voice.Name() + ": ")
			}

			switch ev.Type {
			case "ON":
				//fmt.Print("<")
				fmt.Printf("@%v %v ", sk, ev.FinalParams()["note"])
			case "OFF":
				fmt.Printf("@%v |", sk)
			case "CHANGE":
				fmt.Printf("@%v #", sk)
			}

			//ev.Voice
		}

		fmt.Print("\n")
	}

}
*/

func (t *Track) CurrentBar() Measure {
	return t.Bars[len(t.Bars)-1]
}

func (t *Track) BarNum() int {
	return len(t.Bars)
}

//func (t *Track) TempoAt(pos string, tempo Tempo) {
func (t *Track) SetTempo(pos Measure, tempo Tempo) {
	num, posInLast := t.CurrentBar().Add(pos)
	abs := t.absPos + Measure(num)*t.CurrentBar() + posInLast
	tempAt := tempoAt{AbsPos: abs, Tempo: tempo}
	// fmt.Printf("set tempo to: %v at %v\n", tempo, tempAt.AbsPos)
	t.tempi = append(t.tempi, tempAt)
}

func (t *Track) At(pos Measure, events ...*Event) {
	num, posInLast := t.CurrentBar().Add(pos)
	abs := t.absPos + Measure(num)*t.CurrentBar() + posInLast

	for _, ev := range events {
		e := ev.Clone()
		// fmt.Println("absposition", abs)
		e.absPosition = abs
		t.Events = append(t.Events, e)
	}
}

// tempoAt returns the Tempo at the given barnum and barposition
// t.Tempi must be sorted before this method is called
// If a tempo was set at the given point in time, this tempo is returned.
// Otherwise the last tempo change before that position ist returned
func (t *Track) TempoAt(abspos Measure) Tempo {
	// fmt.Printf("search tempo at %v\n", abspos)
	// t.Tempi is expected to be sorted by BarNum and BarPosition (ascending)
	// Go through it in reverse order to get the latest tempo change that matches
	for i := len(t.tempi) - 1; i >= 0; i-- {
		ta := t.tempi[i]

		// if we have a tempochange at the exact position, return it
		if ta.AbsPos == abspos {
			// fmt.Printf("found %0.2f\n", ta.Tempo)
			return ta.Tempo
		}

		// return the last tempochange
		if ta.AbsPos < abspos {
			// fmt.Printf("found %0.2f\n", ta.Tempo)
			return ta.Tempo
		}
	}

	panic("no tempo found")
}

func (t *Track) Patterns(tf ...Pattern) *Track {
	for _, trafo := range tf {
		trafo.Pattern(t)
	}
	return t
}

func (t *Track) Next(tr ...Pattern) *Track {
	t.nextBar()
	t.Patterns(tr...)
	return t
}

func (t *Track) Change(bar string, tr ...Pattern) *Track {
	t.changeBar(M(bar))
	t.Patterns(tr...)
	return t
}

// fill with num bars, transformers are repeated each bar
func (t *Track) Fill(num int, tr ...Pattern) *Track {
	for i := 0; i < num; i++ {
		t.nextBar()
		t.Patterns(tr...)
	}
	return t
}

// calculates and sets the ticks for all events
func (t *Track) compile() {
	/*
		sort.Sort(TempoSorted(t.Tempi))
		sort.Sort(EventsSorted(t.Events))
	*/

	if len(t.Events) == 0 {
		t.compiled = true
		return
	}

	tempoChanges := map[Measure]Tempo{}

	for _, tm := range t.tempi {
		tempoChanges[tm.AbsPos] = tm.Tempo
	}

	events := map[Measure][]*Event{}

	for _, ev := range t.Events {
		//fmt.Println("AbsPosition", ev.AbsPosition)
		events[ev.absPosition] = append(events[ev.absPosition], ev)
	}

	//	prevTempoNum := 0
	var millisecs float64 = 0
	currentTempo := tempoChanges[Measure(0)]
	// fmt.Printf("start tempo: %v\n", currentTempo.MilliSecs(Measure(1)))

	// lastEventPos := t.Events[len(t.Events)-1].AbsPosition

	//fmt.Println("lastEventPos", int(lastEventPos))

	//for i := 0; i < int(lastEventPos)+1; i++ {
	i := 0
	for {
		if len(events) == 0 {
			break
		}
		tm, hasT := tempoChanges[Measure(i)]
		if hasT {
			// fmt.Printf("has tempo changes at position %v\n", Measure(i))
			currentTempo = tm
		}

		// fmt.Println("currentTempo", currentTempo)

		evts, hasE := events[Measure(i)]
		if hasE {
			// fmt.Println("millisecs", millisecs)
			//fmt.Printf("has events at position %v (%v), millisecs: %v\n", i, Measure(i), millisecs)
			for _, ev := range evts {
				ev.tick = uint(millisecs) //currentTempo.MilliSecs(ev.AbsPosition)
			}

			delete(events, Measure(i))
		}
		// fmt.Printf("adding %d\n", int(currentTempo.MilliSecs(Measure(1))))
		//millisecs += int(currentTempo.MilliSecs(Measure(1)))
		//fmt.Printf("adding %d (%0f)\n", int(RoundFloat(currentTempo.MilliSecs(Measure(1)), 0)), currentTempo.MilliSecs(Measure(1)))
		// fmt.Printf("adding %d\n", int(currentTempo.MilliSecs(Measure(1))))
		//millisecs += int(RoundFloat(currentTempo.MilliSecs(Measure(1)), 0))
		//millisecs += int(currentTempo.MilliSecs(Measure(1)))
		millisecs += currentTempo.MilliSecs(Measure(1))
		i++
	}
	t.compiled = true
	debug.FreeOSMemory()
}

/*
func (t *Track) IsCompiled() bool {
	return t.compiled
}
*/

type tempoAt struct {
	Tick   uint
	Tempo  Tempo
	AbsPos Measure
}

func (t *Track) Print(tempo Tempo, unit string, wr io.Writer) {
	t.compile()
	t.print(tempo, unit, wr)
}
