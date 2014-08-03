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
)

func NewSc() *sc {
	s := &sc{
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
		groupNumber:      1012,
		groups:           []*group{},
		// groupsByName:     map[string]int{},
	}
	s.Bus = &bus{s}
	return s
}

type sc struct {
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
	Bus              *bus
	groupNumber      int
	groups           []*group
	// groupsByName     map[string]int
}

func (s *sc) SetSampleDir(p string) {
	s.sampleDir = p
}

func (s *sc) NewInstrument(name string, offset int) music.Instrument {
	return &instrument{name: name, sc: s, offset: offset}
}

func (s *sc) NewRoute(name string) music.Instrument {
	return &instrument{name: name, sc: s, bus: true}
}

func (s *sc) NewSample(name string, numChan int, offset int) *sample {
	if len(s.samples) == 0 {
		s.instrNumber++
	}
	s.sampleNumber++

	if _, exists := s.samples[name]; exists {
		panic("sample " + name + " already exists")
	}
	s.samples[name] = s.sampleNumber
	s.samplesChannels[name] = numChan
	// fmt.Printf("sample %s has number %d\n", name, s.sampleNumber)
	// idx := strings.LastIndex(name, ".")
	//return &sample{s.sampleNumber, &instrument{"sample" + name[:idx], s}}
	return &sample{s.sampleNumber, &instrument{offset: offset, name: fmt.Sprintf("sample%d", numChan), sc: s}}
}

func (s *sc) NewBus(name string, numchannels int) int {
	no := s.busNumber + 1
	s.busNumber += numchannels
	_, exists := s.busses[name]
	if exists {
		panic("bus " + name + " already defined")
	}
	s.busses[name] = no
	return no
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

	for _, gr := range s.groups {
		fmt.Fprintf(s.buffer, `[\g_new, %d, 1, %d], `, gr.Id(), gr.parent)
	}

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
		inSecs := tickToSeconds(ti+(tickNegative*(-1))) + 0.000001 + beginOffset
		fmt.Fprintf(s.buffer, `  [%0.6f`, inSecs)
		for _, ev := range tickMapped[ti] {
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
