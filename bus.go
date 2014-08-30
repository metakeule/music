package music

type Bus int

func (b Bus) Name() string {
	return "bushub"
}

var bushub = Bus(0)
var busses = map[string]int{}

func NewBus(g Generator, name string) *Voice {
	if _, has := busses[name]; has {
		panic("bus with name " + name + " already defined")
	}
	busses[name] = g.NewBusId()
	return &Voice{Generator: g, Instrument: bushub, Bus: busses[name]}
}
