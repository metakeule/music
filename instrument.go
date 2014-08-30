package music

import "io/ioutil"

type Generator interface {
	NewNodeId() int
	NewBusId() int
	NewGroupId() int
	NewSampleBuffer() int
}

type Instrument interface {
	Name() string
}

type SCInstrument struct {
	name   string  // name of the instrument must not have characters other than [a-zA-Z0-9_]
	Path   string  // path of the instrument
	Offset float64 // offset applied to every instance of the instrument
	used   bool
}

// the outer invoker may use the first voices instrument to query loadcode etc
func NewSCInstrument(g Generator, name, path string, numVoices int) []*Voice {
	i := &SCInstrument{
		name: name,
		Path: path,
	}
	return Voices(numVoices, g, i, -1)
}

func (i *SCInstrument) Name() string {
	return i.name
}

func (i *SCInstrument) IsUsed() bool {
	return i.used
}

func (i *SCInstrument) Use() {
	i.used = true
}

func (i *SCInstrument) LoadCode() []byte {
	data, err := ioutil.ReadFile(i.Path)
	if err != nil {
		panic("can't read instrument " + i.Path)
	}
	return data
}

func Voices(num int, g Generator, instr Instrument, groupid int) []*Voice {
	voices := make([]*Voice, num)

	for i := 0; i < num; i++ {
		voices[i] = &Voice{Generator: g, Instrument: instr}
		if groupid > -1 {
			voices[i].SCGroup = groupid
		}
	}

	return voices
}
