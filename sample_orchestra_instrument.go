package music

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type SCSample struct {
	*SCInstrument
	*Sample
	Freq float64
}

func NewSCSampleFreq(g Generator, path string, freq float64, numVoices int) []*Voice {
	vs := NewSCSample(g, path, numVoices)
	vs[0].Instrument.(*SCSample).Freq = freq
	return vs
}

func NewSCSample(g Generator, path string, numVoices int) []*Voice {
	sample := NewSample(path)
	sample.SCBuffer = g.NewSampleBuffer()
	instr := &SCInstrument{
		name: fmt.Sprintf("sample%d", sample.Channels),
		Path: "",
	}
	// instr := NewSCInstrument(g, fmt.Sprintf("sample%d", sample.Channels), "")
	i := &SCSample{SCInstrument: instr, Sample: sample}
	return Voices(numVoices, g, i, -1)
}

/*
func (s *SCSample) LoadCode() []byte {
	return []byte(fmt.Sprintf(`SynthDef("sample%d", { |gate=1,bufnum = 0,amp=1, out=0, pan=0, rate=1| var z;
	z =  EnvGen.kr(Env.perc,gate) * PlayBuf.ar(%d, bufnum, BufRateScale.kr(bufnum) * rate);
	FreeSelfWhenDone.kr(z);
	Out.ar(out, Pan2.ar(z, pos: pan, level: amp));
} ).writeDefFile;
`, s.Sample.Channels, s.Sample.Channels))
}
*/

type SampleLibrary interface {
	SamplePath(instrument string, params map[string]float64) string
	Channels() []int // channel variants
}

type SCSampleInstrument struct {
	// *SCInstrument
	SampleLibrary
	// path to *Sample
	Samples    map[string]*Sample
	instrument string
	g          Generator
}

func (s *SCSampleInstrument) Name() string {
	return "samplelibrary"
}

func NewSCSampleInstrument(g Generator, instrument string, sampleLib SampleLibrary, numVoices int) []*Voice {
	/*
		instr := NewSCInstrument(g, fmt.Sprintf("sample%d", s.Sample.Channels), "")
		instr := &SCInstrument{
			name: fmt.Sprintf("sample%d", s.Sample.Channels),
			Path: path,
		}
	*/

	//i := &SCSampleInstrument{instr, sampleLib, map[string]*Sample{}, instrument, g}
	i := &SCSampleInstrument{sampleLib, map[string]*Sample{}, instrument, g}
	return Voices(numVoices, g, i, -1)
}

func (s *SCSampleInstrument) Sample(params map[string]float64) *Sample {
	samplePath := s.SampleLibrary.SamplePath(s.instrument, params)
	sample, has := s.Samples[samplePath]
	if !has {
		sample = NewSample(samplePath)
		s.Samples[samplePath] = sample
		sample.SCBuffer = s.g.NewSampleBuffer()
	}
	return sample
}

/*
func (s *SCSampleInstrument) LoadCode() []byte {
	var bf bytes.Buffer

	for _, ch := range s.SampleLibrary.Channels() {
		fmt.Fprintf(&bf, `SynthDef("sample%d", { |gate=1,bufnum = 0,amp=1, out=0, pan=0, rate=1| var z;
	z =  EnvGen.kr(Env.perc,gate) * PlayBuf.ar(%d, bufnum, BufRateScale.kr(bufnum) * rate);
	FreeSelfWhenDone.kr(z);
	Out.ar(out, Pan2.ar(z, pos: pan, level: amp));
} ).writeDefFile;
`, ch, ch)
	}

	return bf.Bytes()
}
*/

type Sample struct {
	// Name       string  // name of the sample to reference it [a-zA-Z0-9_]
	Path         string  // the path of the sample
	Offset       float64 // offset in milliseconds until max amplitude must be positiv
	MaxAmp       float64 // max amplitude value, must be between 0 and 1
	SCBuffer     int     // the sc buffer number
	Channels     uint    // number of channels
	NumFrames    int     // number of frames
	SampleRate   int     // e.g. 44100
	SampleFormat string  // e.g. int16
	Duration     float64 // duration in seconds
	HeaderFormat string  // e.g. WAV
}

/*
	{
		"offset": 0.01124716553288,
		"maxAmp": 0.8631591796875,
		"numFrames": 64637,"sampleRate": 44100,"channels": 2,"sampleFormat": "int16","duration": 1.4656916099773,"headerFormat": "WAV"}
*/

func NewSample(path string) *Sample {
	s := &Sample{Path: path}
	s.loadMeta()
	return s
}

func (s *Sample) loadMeta() {
	s.MaxAmp = 1
	s.Offset = 0
	s.Channels = 1
	metapath := s.Path + ".meta"
	data, err := ioutil.ReadFile(metapath)
	if err != nil {
		fmt.Printf("file not found: " + metapath + ", using defaults")
		return
	}

	err = json.Unmarshal(data, &s)
	if err != nil {
		panic("invalid json format for " + metapath)
	}
	s.Offset = s.Offset * 1000.0
}

/*
import (
	"fmt"
)

type sampleOrchestraInstrument struct {
	// func (in *instrument) New(num int) []Voice {
	num int
	*instrument
	name            string
	sampleOrchestra SampleOrchestra
	// offset int
}

// TODO: fix the issue: each sample needs to start with an upper voice, otherwise
// all samples using sample1 get on the first voice the same
func (s *sampleOrchestraInstrument) Voices(num int) []*sampleOrchestraVoice {
	// offset := s.instrument.sc.GetSampleOffset(s.name)
	v := make([]*sampleOrchestraVoice, num)
	for i := 0; i < num; i++ {
		samplenum := s.instrument.sc.IncrSampleNumber()
		name := fmt.Sprintf("%s-%d", s.instrument.name, i)
		vc := &voice{
			name:       name,
			instrument: s.instrument,
			num:        i,
			instrNum:   samplenum,
		}
		v[i] = &sampleOrchestraVoice{vc, s}
		s.instrument.sc.SetVoiceToNum(name, samplenum)
		s.instrument.sc.SetNumToVoices(samplenum, vc)
	}
	return v
}

type sampleOrchestraVoice struct {
	*voice
	*sampleOrchestraInstrument
}

func (v *sampleOrchestraVoice) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return PlayDur(pos, dur, v, params...)
}

func (v *sampleOrchestraVoice) Play(pos string, params ...Parameter) Pattern {
	return Play(pos, v, params...)
}

func (v *sampleOrchestraVoice) Stop(pos string) Pattern {
	return Stop(pos, v)
}

func (v *sampleOrchestraVoice) Modify(pos string, params ...Parameter) Pattern {
	return Modify(pos, v, params...)
}

// TODO it would be nice to be able to load the sample (track that its loaded, so it does not have to be loaded again)
// and then immediatly use it. the question is, if that works out in NRT and if it works when its loaded and used in the
// same command - trial and error
func (sv *sampleOrchestraVoice) On(ev *Event) {
	samplePath := sv.sampleOrchestra.SampleForParams(sv.sampleOrchestraInstrument.instrument.Name(), ev.Params)
	// sampleNum := sv.voice.instrument.sc.UseSample(samplePath)
	sv.voice.instrument.sc.UseSample(samplePath)
	sampleNum := 0

	//sv.voice.instrument.sc.instrNumber++
	instrnum := sv.voice.instrument.sc.IncrInstrNumber()
	if sv.voice.instrNum > 2000 {
		//fmt.Fprintf(sv.voice.instrument.sc.buffer, `, [\n_free, %d]`, sv.voice.instrNum)
		fmt.Fprintf(sv.voice.instrument.sc, `, [\n_free, %d]`, sv.voice.instrNum)
	}

	//sv.voice.instrNum = sv.voice.instrument.sc.instrNumber
	sv.voice.instrNum = instrnum
	if sv.voice.mute {
		return
	}
	fmt.Fprintf(
		//sv.voice.instrument.sc.buffer,
		sv.voice.instrument.sc,
		//`, [\s_new, \%s, %d, 0, 0, \bufnum, b.sample%d%s]`,
		`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
		//`, [\s_new, \%s, -1, 0, 0, \bufnum, %d%s]`,
		sv.voice.instrument.name,
		sv.voice.instrNum,
		// sv.sample.num,
		sampleNum,
		sv.voice.paramsStr(ev),
	)
}

func (sv *sampleOrchestraVoice) Change(ev *Event) {
	// change the params
	sv.sampleOrchestra.SampleForParams(sv.sampleOrchestraInstrument.instrument.Name(), ev.Params)
	sv.voice.Change(ev)
}

func (sv *sampleOrchestraVoice) Offset() int {
	// return sv.offset
	return 0
}

func (sv *sampleOrchestraVoice) SetOffset(offset int) {
	// sv.offset = offset
}
*/
/*
func (sv *sampleVoice) Off(ev *Event) {
}
*/

/*
func (sv *sampleVoice) Off(ev *Event) {
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, 0]`, v.instrNum)
	fmt.Fprintf(sv.voice.instrument.sc.buffer, `, [\n_set, %d, \gate, -1]`, sv.voice.instrNum)
}
*/
