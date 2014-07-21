package player

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metakeule/music"

	// "github.com/metakeule/sclang"
)

// TODO: perhaps use groups instead of instrNumbers

type sc struct {
	// osc         *sclang.OscClient

	buffer           *bytes.Buffer
	synthdefs        map[string][]byte
	sampleDir        string
	instrNumber      int
	sampleNumber     int
	sampleInstNumber int
	samples          map[string]int
	samplesChannels  map[string]int
	voicesToNum      map[string]int
	numToVoices      map[int]music.Voice
	AudioFile        string
	ScoreFile        string
	busses           map[string]int
	busNumber        int
}

func (s *sc) SetSampleDir(p string) {
	s.sampleDir = p
}

type instrument struct {
	name   string
	bus    bool
	sc     *sc
	offset int
}

func (s *sc) NewInstrument(name string, offset int) music.Instrument {
	return &instrument{name: name, sc: s, offset: offset}
}

func (s *sc) NewBusRoute(name string) music.Instrument {
	return &instrument{name: name, sc: s, bus: true}
}

type bus struct {
	sc *sc
}

func (s *sc) NewBusHub(names ...string) *bus {
	for _, name := range names {
		s.busNumber++
		s.busses[name] = s.busNumber
	}
	return &bus{s}
}

func (b *bus) BusId(name string) int {
	busno, ok := b.sc.busses[name]
	if !ok {
		panic("unknown bus " + name)
	}
	return busno
}

func (b *bus) Mute(*music.Event) {
	panic("mute not allowed for bus")
}

func (b *bus) UnMute(*music.Event) {
	panic("unmute not allowed for bus")
}

func (b *bus) Name() string {
	return "bushub"
}

func (b *bus) On(ev *music.Event) {
	panic("on not allowed for bus")
}

func (b *bus) Off(ev *music.Event) {
	panic("off not allowed for bus")
}

func (b *bus) Offset() int {
	return 0
}

func (b *bus) Change(ev *music.Event) {
	busses := ev.Params
	for name, val := range busses {
		busno, ok := b.sc.busses[name]
		if !ok {
			panic("unknown bus " + name)
		}
		fmt.Fprintf(b.sc.buffer, `, [\c_set, \%d, %v]`, busno, val)
	}
}

// /c_set busint, valfloat

// func (b *bus) New(num int) []music.Voice {
/*
	v := make([]music.Voice, num)
	for i := 0; i < num; i++ {
		in.sc.instrNumber++
		name := fmt.Sprintf("%s-%d", in.name, i)
		vc := &voice{
			name:       name,
			instrument: in,
			num:        i,
			instrNum:   in.sc.instrNumber,
		}
		v[i] = vc
		in.sc.voicesToNum[name] = in.sc.instrNumber
		in.sc.numToVoices[in.sc.instrNumber] = vc
	}
	return v
*/
// }

/*
(
var myPath;
myPath = PathName.new("MyDisk/SC 2.2.8 f/Sounds/FunkyChicken.aiff");
myPath.fileNameWithoutExtension.postln;
)
*/

/*
func (in *instrument) New(num int) []music.Voice {
	v := make([]music.Voice, num)
	for i := 0; i < num; i++ {
		in.sc.instrNumber++
		name := fmt.Sprintf("%s-%d", in.name, i)
		vc := &voice{
			name:       name,
			instrument: in,
			num:        i,
			instrNum:   in.sc.instrNumber,
		}
		v[i] = vc
		in.sc.voicesToNum[name] = in.sc.instrNumber
		in.sc.numToVoices[in.sc.instrNumber] = vc
	}
	return v
}
*/

/*
in.sc.instrNumber++
		name := fmt.Sprintf("%s-%d", in.name, i)
*/

type sample struct {
	// func (in *instrument) New(num int) []music.Voice {
	num int
	*instrument
	// offset int
}

func (s *sc) NewSample(name string, numChan int, offset int) *sample {
	if len(s.samples) == 0 {
		s.instrNumber++
	}
	s.sampleNumber++

	s.samples[name] = s.sampleNumber
	s.samplesChannels[name] = numChan
	fmt.Printf("sample %s has number %d\n", name, s.sampleNumber)
	// idx := strings.LastIndex(name, ".")
	//return &sample{s.sampleNumber, &instrument{"sample" + name[:idx], s}}
	return &sample{s.sampleNumber, &instrument{offset: offset, name: fmt.Sprintf("sample%d", numChan), sc: s}}
}

type sampleVoice struct {
	*voice
	*sample
	offset int
}

func (sv *sampleVoice) On(ev *music.Event) {
	sv.voice.instrument.sc.instrNumber++
	if sv.voice.instrNum > 2000 {
		fmt.Fprintf(sv.voice.instrument.sc.buffer, `, [\n_free, %d]`, sv.voice.instrNum)
	}

	sv.voice.instrNum = sv.voice.instrument.sc.instrNumber
	if sv.voice.mute {
		return
	}
	fmt.Fprintf(
		sv.voice.instrument.sc.buffer,
		//`, [\s_new, \%s, %d, 0, 0, \bufnum, b.sample%d%s]`,
		`, [\s_new, \%s, %d, 0, 0, \bufnum, %d%s]`,
		//`, [\s_new, \%s, -1, 0, 0, \bufnum, %d%s]`,
		sv.voice.instrument.name,
		sv.voice.instrNum,
		sv.sample.num,
		sv.voice.paramsStr(ev),
	)
}

func (sv *sampleVoice) Offset() int {
	return sv.offset
}

func (sv *sampleVoice) SetOffset(offset int) {
	sv.offset = offset
}

/*
func (sv *sampleVoice) Off(ev *music.Event) {
}
*/

/*
func (sv *sampleVoice) Off(ev *music.Event) {
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, 0]`, v.instrNum)
	fmt.Fprintf(sv.voice.instrument.sc.buffer, `, [\n_set, %d, \gate, -1]`, sv.voice.instrNum)
}
*/

// TODO: fix the issue: each sample needs to start with an upper voice, otherwise
// all samples using sample1 get on the first voice the same
func (s *sample) New(num int) []*sampleVoice {
	v := make([]*sampleVoice, num)
	for i := 0; i < num; i++ {
		//s.instrument.sc.instrNumber++
		s.instrument.sc.sampleNumber++
		name := fmt.Sprintf("%s-%d", s.instrument.name, i)
		vc := &voice{
			name:       name,
			instrument: s.instrument,
			num:        i,
			// instrNum:   s.instrument.sc.instrNumber,
			instrNum: s.instrument.sc.sampleNumber,
		}
		v[i] = &sampleVoice{vc, s, s.offset}
		//s.instrument.sc.voicesToNum[name] = s.instrument.sc.instrNumber
		s.instrument.sc.voicesToNum[name] = s.instrument.sc.sampleNumber
		// s.instrument.sc.numToVoices[s.instrument.sc.instrNumber] = vc
		s.instrument.sc.numToVoices[s.instrument.sc.sampleNumber] = vc
	}
	return v
}

type voice struct {
	instrument *instrument
	name       string
	// voice number
	num         int
	instrNum    int
	initialized bool
	mute        bool
}

func (v *voice) Mute(*music.Event) {
	v.mute = true
}

func (v *voice) UnMute(*music.Event) {
	v.mute = false
}

func (v *voice) Name() string {
	return v.name
}

func (v *voice) Offset() int {
	return v.instrument.offset
}

func (in *instrument) Name() string {
	return in.name
}

func (in *instrument) New(num int) []music.Voice {
	v := make([]music.Voice, num)
	for i := 0; i < num; i++ {
		in.sc.instrNumber++
		name := fmt.Sprintf("%s-%d", in.name, i)
		vc := &voice{
			name:       name,
			instrument: in,
			num:        i,
			instrNum:   in.sc.instrNumber,
		}
		v[i] = vc
		in.sc.voicesToNum[name] = in.sc.instrNumber
		in.sc.numToVoices[in.sc.instrNumber] = vc
	}
	return v
}

func NewSc() *sc {
	return &sc{
		// osc:         &sclang.OscClient{},
		instrNumber: 2000,
		// sampleNumber:    4000,
		sampleInstNumber: 4000,
		voicesToNum:      map[string]int{},
		numToVoices:      map[int]music.Voice{},
		synthdefs:        map[string][]byte{},
		samples:          map[string]int{},
		samplesChannels:  map[string]int{},
		AudioFile:        "",
		busses:           map[string]int{},
		busNumber:        16,
	}
}

func (v *voice) paramsStr(ev *music.Event) string {
	var buf bytes.Buffer

	for k, v := range ev.FinalParams() {
		fmt.Fprintf(&buf, `, \%s, %v`, k, float32(v))
	}

	return buf.String()

}

func (v *voice) On(ev *music.Event) {
	v.instrument.sc.instrNumber++
	if v.instrument.bus {
		fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, 1200%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
		return
	}
	if v.instrNum > 2000 {
		fmt.Fprintf(v.instrument.sc.buffer, `, [\n_free, %d]`, v.instrNum)
	}
	v.instrNum = v.instrument.sc.instrNumber
	/*
		v.instrument.sc.instrNumber++

		fmt.Fprintf(v.instrument.sc.buffer, `, [\n_free, %d]`, v.instrNum)
		v.instrNum = v.instrument.sc.instrNumber
	*/
	if v.mute {
		return
	}
	fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 1, 1010%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
}

func (v *voice) Off(ev *music.Event) {
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, 0]`, v.instrNum)
	fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d, \gate, -1]`, v.instrNum)
}

func (v *voice) Change(ev *music.Event) {
	fmt.Fprintf(v.instrument.sc.buffer, `, [\n_set, %d%s]`, v.instrNum, v.paramsStr(ev))
}

func (s *sc) LoadSynthDefPool() {
	home := os.Getenv("HOME")
	p := filepath.Join(home, ".local/share/SuperCollider/quarks/SynthDefPool/pool")
	s.LoadSynthDefs(p)
}

// p is the full path to a directory from which the synthdefs are loaded
func (s *sc) LoadSynthDefs(p string) {
	f, err := os.Open(p)
	if err != nil {
		fmt.Println("can't load synthdefs from ", p, ": ", err.Error())
		return
	}

	files, e := f.Readdir(-1)

	if e != nil {
		fmt.Println("can't load synthdefs from ", p, ": ", e.Error())
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			data, er := ioutil.ReadFile(filepath.Join(p, file.Name()))
			if er == nil {
				s.synthdefs[file.Name()] = data
			}
		}
	}

}

/*
//var soundLoader = `bufnum = Buffer.read(s, "%s");
var soundLoader = `SynthDef("%s", { |bufnum = 0|
    Out.ar( 1,
        PlayBuf.ar(2, bufnum, BufRateScale.kr(bufnum))
    )
})
`

var sampleLoaderOld = `SynthDef("sample%d", { |bufnum = 0|
    Out.ar( 0,
        PlayBuf.ar(%d, bufnum, BufRateScale.kr(bufnum))
    )
})
`
*/
var sampleLoader = `SynthDef("sample%d", { |gate=1,bufnum = 0,amp=1, out=0, pan=0| var z;
	z =  EnvGen.kr(Env.perc,gate) * PlayBuf.ar(%d, bufnum, BufRateScale.kr(bufnum));
	FreeSelfWhenDone.kr(z);
	Out.ar(out, Pan2.ar(z, pos: pan, level: amp));
} )`

/*
func (s *sc) LoadSamples(p string) {
	f, err := os.Open(p)
	if err != nil {
		fmt.Println("can't load samples from ", p, ": ", err.Error())
		return
	}

	files, e := f.Readdir(-1)

	if e != nil {
		fmt.Println("can't load samples from ", p, ": ", e.Error())
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			fullpath := filepath.Join(p, file.Name())
			idx := strings.LastIndex(file.Name(), ".")
			s.synthdefs[file.Name()] = []byte(fmt.Sprintf(soundLoader, fullpath, "sample"+file.Name()[:idx]))
		}
	}

}
*/

func millisecsToTick(ms int) int {
	return ms * 1000000
	//return 0
}

func tickToSeconds(tick int) float32 {
	return float32(tick) / float32(1000000000)
}

// startOffset is in milliseconds and must be positive
func (s *sc) Play(startOffset uint, evts ...*music.Event) {
	dir, err := ioutil.TempDir("/tmp", "go-sc-music-generator")
	if err != nil {
		panic(err.Error())
	}

	defer os.RemoveAll(dir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			os.RemoveAll(dir)
			os.Exit(2)
			_ = sig
		}
	}()

	sclangCodeFile := s.ScoreFile
	if sclangCodeFile == "" {
		sclangCodeFile = filepath.Join(dir, "sclang-code.scd")
	}
	oscCodeFile := filepath.Join(dir, "sclang-compiled.osc")
	audioFile := s.AudioFile
	if audioFile == "" {
		audioFile = filepath.Join(dir, "out.aiff")
	}
	libraryPath := "/usr/local/share/SuperCollider/SCClassLibrary"

	s.buffer = &bytes.Buffer{}

	tickMapped := map[int][]*music.Event{}
	ticksSorted := []int{}

	finTick := uint(0)
	startTick := uint(0)

	for _, ev := range evts {
		currTick := int(ev.Tick)
		if ev.Type == "ON" {
			// Todo calculation ms back to ticks
			currTick = millisecsToTick(ev.Voice.Offset()) + currTick
		}
		//tickMapped[int(ev.Tick)] = append(tickMapped[int(ev.Tick)], ev)
		tickMapped[int(currTick)] = append(tickMapped[int(currTick)], ev)
		if ev.Type == "fin" {
			if finTick == 0 || finTick > ev.Tick {
				finTick = ev.Tick
			}
		}
		if ev.Type == "start" {
			if startTick == 0 || startTick < ev.Tick {
				startTick = ev.Tick
			}
		}
	}

	var tickNegative int = 0

	for ti := range tickMapped {
		if ti < tickNegative {
			tickNegative = ti
		}
		ticksSorted = append(ticksSorted, ti)
	}

	sort.Ints(ticksSorted)
	fmt.Fprintf(s.buffer, "(\n")

	for _, sdef := range s.synthdefs {
		fmt.Fprintf(s.buffer, strings.TrimSpace(string(sdef))+".writeDefFile;")
	}

	//for sampleName, sampleId := range s.samples {
	for sampleName, _ := range s.samples {
		fullpath := filepath.Join(s.sampleDir, sampleName)

		ch, err := numChannels(fullpath)

		if err != nil {
			panic(fmt.Sprintf("can't open sample file %s, reason: %s", sampleName, err.Error()))
		}

		if ch != s.samplesChannels[sampleName] {
			panic(fmt.Sprintf("sample file %s has %d channels and not %d", sampleName, ch, s.samplesChannels[sampleName]))
		}

		fmt.Fprintf(s.buffer, strings.TrimSpace(fmt.Sprintf(sampleLoader, s.samplesChannels[sampleName], s.samplesChannels[sampleName]))+".writeDefFile;")
	}

	// lame --decode file.mp3 output.wav

	/*
		for sampleName, _ := range s.samples {
			// fullpath := filepath.Join(s.sampleDir, sampleName)
			idx := strings.LastIndex(sampleName, ".")
			fmt.Fprintf(s.buffer, strings.TrimSpace(fmt.Sprintf(soundLoader, "sample"+sampleName[:idx]))+".writeDefFile;")
			//fmt.Fprintf(s.buffer, fmt.Sprintf(`b.sample%d = Buffer.read(s, "%s", bufnum);`, sampleId, fullpath))
			// fmt.Fprintf(s.buffer, fmt.Sprintf(`Buffer.read(s, "%s", bufnum: %d);`, fullpath, sampleId))
		}
	*/
	// fmt.Fprintf(s.buffer, "0.5.wait;")

	/*
		bufnum = Buffer.read(s, "%s");
	*/

	fmt.Fprintf(s.buffer, "TempoClock.default.tempo = 1; \n")
	fmt.Fprintf(s.buffer, "x = [\n")

	fmt.Fprintf(s.buffer, `  [%0.6f, `, 0.0)

	// create the bus routing group
	fmt.Fprintf(s.buffer, fmt.Sprintf(`[\g_new, %d, 0, 0],`, 1200))
	// create the instruments group
	fmt.Fprintf(s.buffer, fmt.Sprintf(`[\g_new, %d, 0, 0],`, 1010))
	// /g_new

	first_sample := true
	// [\b_allocRead, 1, "/home/benny/Entwicklung/gopath/src/github.com/metakeule/music/example/samples/piano.aiff"],
	for sampleName, sampleId := range s.samples {
		fullpath := filepath.Join(s.sampleDir, sampleName)
		if !first_sample {
			fmt.Fprintf(s.buffer, ", ")
		}
		first_sample = false
		fmt.Fprintf(s.buffer, fmt.Sprintf(`[\b_allocRead, %d, "%s"]`, sampleId, fullpath))
	}

	fmt.Fprintf(s.buffer, "],\n")

	t := 0

	beginOffset := float32(startOffset) / float32(1000)

allEvents:
	for _, ti := range ticksSorted {
		if startTick != 0 && ti < int(startTick) {
			t = ti
			continue
		}
		if startTick != 0 {
			ti = ti - int(startTick)
		}
		if finTick != 0 && int(finTick) <= ti {
			t = int(finTick)
			break allEvents
		}
		// inSecs := float32(ti) / float32(1000000000)
		inSecs := tickToSeconds(ti+(tickNegative*(-1))) + 0.000001 + beginOffset
		fmt.Fprintf(s.buffer, `  [%0.6f`, inSecs)
		for _, ev := range tickMapped[ti] {
			// println(ev.Type)
			ev.Runner(ev)
		}
		t = ti
		fmt.Fprintf(s.buffer, "],\n")
	}

	//fmt.Fprintf(s.buffer, "  [%v, [\\c_set, 0, 0]]];\n", float32(t+10000000)/float32(1000000000))
	// free the default group and unset the controllers
	//fmt.Fprintf(s.buffer, "  [%v, [\\/clearSched], [\\g_deepFree, 1], [\\c_set, 0, 0]]];\n", float32(t)/float32(1000000000))

	// 0.16666667
	// stattdessen: 0.1666666
	// 0.16666667
	fmt.Fprintf(s.buffer, "  [%0.6f, [\\g_deepFree, 1], [\\c_set, 0, 0]]];\n", float32(t)/float32(1000000000))
	fmt.Fprintf(s.buffer, `Score.write(x, "`+oscCodeFile+`");`+"\n")
	fmt.Fprintf(s.buffer, "\n\n"+` "quitting".postln; 0.exit; )`)
	ioutil.WriteFile(sclangCodeFile, s.buffer.Bytes(), 0644)
	cmd := exec.Command("sclang", "-r", "-s", "-l", libraryPath, sclangCodeFile)
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("ERROR running sclang")
		fmt.Printf("%s\n", out)
		fmt.Println(err)
		return
	}

	// sample rate
	// channels
	// file format
	// bit depth

	//	cmd = exec.Command("scsynth", "-N", oscCodeFile, "_", audioFile, "44100", "AIFF", "int16", "-o", "2")
	// sampleFormat "int8", "int16", "int24", "int32", "mulaw", "alaw","float"
	// from http://doc.sccode.org/Classes/SoundFile.html#-sampleFormat
	// headerFormat
	// from http://doc.sccode.org/Classes/SoundFile.html#-headerFormat
	/*
	   "AIFF"	Apple/SGI AIFF format
	   "WAV","WAVE", "RIFF"	Microsoft WAV format
	   "Sun", "NeXT"	Sun/NeXT AU format
	   "SD2"	Sound Designer 2
	   "IRCAM"	Berkeley/IRCAM/CARL
	   "raw"	no header = raw data
	   "MAT4"	Matlab (tm) V4.2 / GNU Octave 2.0
	   "MAT5"	Matlab (tm) V5.0 / GNU Octave 2.1
	   "PAF"	Ensoniq PARIS file format
	   "SVX"	Amiga IFF / SVX8 / SV16 format
	   "NIST"	Sphere NIST format
	   "VOC"	VOC files
	   "W64"	Sonic Foundry's 64 bit RIFF/WAV
	   "PVF"	Portable Voice Format
	   "XI"	Fasttracker 2 Extended Instrument
	   "HTK"	HMM Tool Kit format
	   "SDS"	Midi Sample Dump Standard
	   "AVR"	Audio Visual Research
	   "FLAC"	FLAC lossless file format
	   "CAF"	Core Audio File format
	*/
	//cmd = exec.Command("scsynth", "-N", oscCodeFile, "_", audioFile, "48000", "AIFF", "int16", "-o", "2")
	//cmd = exec.Command("scsynth", "-N", oscCodeFile, "_", audioFile, "48000", "AIFF", "float", "-o", "2")
	cmd = exec.Command("scsynth", "-N", oscCodeFile, "_", audioFile, "96000", "AIFF", "int32", "-o", "2")
	if s.AudioFile == "" {
		cmd = exec.Command("scsynth", "-N", oscCodeFile, "_", audioFile, "96000", "AIFF", "float", "-o", "2")
	}
	out, err = cmd.CombinedOutput()

	_ = out
	// fmt.Printf("%s\n", out)
	if err != nil {
		fmt.Println("ERROR running scsynth")
		fmt.Println(err)
		return
	}

	if s.AudioFile == "" {
		// S16_BE
		// --channels=2 --file-type raw|au|voc|wav --rate=48000 --format=S16_BE
		//cmd = exec.Command("aplay", "--rate=48000", "-f", "cdr", audioFile)
		//cmd = exec.Command("aplay", "--rate=48000", "-f", "U24_BE", audioFile)
		//cmd = exec.Command("aplay", "-f", "S16_BE", "-c2", "--rate=48000", audioFile)
		// "--start-delay=1000"
		cmd = exec.Command("aplay", "-f", "FLOAT_BE", "-c2", "--rate=96000", audioFile)
		//cmd = exec.Command("aplay", "-f", "S32_BE", "-c2", "--rate=48000", audioFile)
		// -f S16_BE -c2 -f44100
		cmd.Run()
		if err != nil {
			fmt.Println("ERROR running aplay")
			fmt.Println(err)
			return
		}
	}
}
