package playernew

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
	buffer      *bytes.Buffer
	synthdefs   map[string][]byte
	instrNumber int
	voicesToNum map[string]int
	numToVoices map[int]music.Voice
	AudioFile   string
	ScoreFile   string
}

type instrument struct {
	name string
	sc   *sc
}

func (s *sc) NewInstrument(name string) music.Instrument {
	return &instrument{name, s}
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
		voicesToNum: map[string]int{},
		numToVoices: map[int]music.Voice{},
		synthdefs:   map[string][]byte{},
		AudioFile:   "",
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
	fmt.Fprintf(v.instrument.sc.buffer, `, [\n_free, %d]`, v.instrNum)
	v.instrNum = v.instrument.sc.instrNumber
	if v.mute {
		return
	}
	fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 0, 0%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
	//v.instrNum = v.instrument.sc.instrNumber
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 4, %d%s]`, v.instrument.name, v.instrNum, v.instrNum, v.paramsStr(ev))
	//fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, 0, 4, %d%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
	/*
		if !v.initialized {
			v.initialized = true
			v.instrNum = v.instrument.sc.instrNumber
			fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 0, 0%s]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
		} else {
			old := v.instrNum
			v.instrNum = v.instrument.sc.instrNumber
			fmt.Fprintf(v.instrument.sc.buffer, `, [\s_new, \%s, %d, 4, %d%s]`, v.instrument.name, v.instrNum, old, v.paramsStr(ev))
		}
	*/
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

func (s *sc) Play(evts ...*music.Event) {
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
		tickMapped[int(ev.Tick)] = append(tickMapped[int(ev.Tick)], ev)
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

	for ti := range tickMapped {
		ticksSorted = append(ticksSorted, ti)
	}

	sort.Ints(ticksSorted)
	fmt.Fprintf(s.buffer, "(\n")

	for _, sdef := range s.synthdefs {
		fmt.Fprintf(s.buffer, strings.TrimSpace(string(sdef))+".writeDefFile;")
	}

	fmt.Fprintf(s.buffer, "TempoClock.default.tempo = 1; \n")
	fmt.Fprintf(s.buffer, "x = [\n")

	t := 0

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
		inSecs := float32(ti) / float32(1000000000)
		fmt.Fprintf(s.buffer, `  [%v`, inSecs)
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
	fmt.Fprintf(s.buffer, "  [%v, [\\g_deepFree, 1], [\\c_set, 0, 0]]];\n", float32(t)/float32(1000000000))
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
