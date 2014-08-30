package music

/*
import (
	"fmt"
)

type bus struct {
	sc scer
}

func (b *bus) Id(name string) int {
	busno := b.sc.GetBus(name)
	// busno, ok := b.sc.busses[name]
	//if !ok {
	if busno == -1 {
		panic("unknown bus " + name)
	}
	return busno
}

func (b *bus) Change(ev *Event) {
	busses := ev.Params
	for name, val := range busses {
		//busno, ok := b.sc.busses[name]
		busno := b.sc.GetBus(name)
		// if !ok {
		if busno == -1 {
			panic("unknown bus " + name)
		}
		//fmt.Fprintf(b.sc.buffer, `, [\c_set, \%d, %v]`, busno, val)
		fmt.Fprintf(b.sc, `, [\c_set, \%d, %v]`, busno, val)
	}
}

func (b *bus) Mute(*Event)   { panic("mute not allowed for bus") }
func (b *bus) UnMute(*Event) { panic("unmute not allowed for bus") }
func (b *bus) Name() string  { return "bushub" }
func (b *bus) On(ev *Event)  { panic("on not allowed for bus") }
func (b *bus) Off(ev *Event) { panic("off not allowed for bus") }
func (b *bus) Offset() int   { return 0 }
*/
