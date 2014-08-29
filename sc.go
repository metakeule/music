package music

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var audioFile = flag.String("out", "", "file to put the audio to (must be .aiff)")
var scoreFile = flag.String("score", "", "file to put the score to (must be .scd)")
var writeSynthDefs = flag.Bool("write-synthdefs", false, "write synthdefs in scserver mode")
var loadSamples = flag.Bool("load-samples", false, "load samples in scserver mode")

func New() *sc {
	s := &sc{
		// osc:         &sclang.OscClient{},
		instrNumber: 2000,
		// sampleNumber:    4000,
		sampleInstNumber: 4000,
		voicesToNum:      map[string]int{},
		numToVoices:      map[int]Voice{},
		synthdefs:        map[string][]byte{},
		samples:          map[string]int{},
		samplesChannels:  map[string]int{},
		AudioFile:        "",
		busses:           map[string]int{},
		busNumber:        16,
		groupNumber:      1012,
		groups:           []*group{},
		usedSamples:      map[string]struct{}{},
		usedInstruments:  map[string]struct{}{},
		// groupsByName:     map[string]int{},
	}
	s.scForInstr = &scForInstrument{sc: s}
	s.Bus = &bus{s.scForInstr}
	flag.Parse()
	if audioFile != nil {
		s.AudioFile = *audioFile
	}
	if scoreFile != nil {
		s.ScoreFile = *scoreFile
	}
	if writeSynthDefs != nil {
		s.WriteSynthDefs = *writeSynthDefs
	}
	if loadSamples != nil {
		s.LoadSamples = *loadSamples
	}
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
	numToVoices      map[int]Voice
	AudioFile        string
	ScoreFile        string
	busses           map[string]int
	busNumber        int
	Bus              *bus
	WriteSynthDefs   bool
	LoadSamples      bool
	groupNumber      int
	groups           []*group
	scServerOnline   bool
	tracks           []*Track
	scForInstr       *scForInstrument
	usedSamples      map[string]struct{}
	usedInstruments  map[string]struct{}
	synthDefDirs     []string
	// groupsByName     map[string]int
}

/*
bar string, tempo Tempo, tr ...Transformer) *Track {
	//t := NewTrack(BPM(120), M(bar))
*/
func (s *sc) Track(bar string, tempo Tempo, patterns ...Pattern) *Track {
	tr := newTrack(tempo, M(bar))
	tr.Patterns(patterns...)
	s.tracks = append(s.tracks, tr)
	return tr
}

func (s *sc) SetSampleDir(p string) {
	s.sampleDir = p
}

type scForInstrument struct {
	sc          *sc
	eventBuffer bytes.Buffer
}

func (s *scForInstrument) UseSample(name string) {
	s.sc.usedSamples[name] = struct{}{}
}

func (s *scForInstrument) UseInstrument(name string) {
	s.sc.usedInstruments[name] = struct{}{}
}

func (s *scForInstrument) IncrInstrNumber() int {
	s.sc.instrNumber++
	return s.sc.instrNumber
}

func (s *scForInstrument) IncrSampleNumber() int {
	s.sc.sampleNumber++
	return s.sc.sampleNumber
}

func (s *scForInstrument) SetNumToVoices(num int, v Voice) {
	s.sc.numToVoices[num] = v
}

func (s *scForInstrument) SetVoiceToNum(name string, num int) {
	s.sc.voicesToNum[name] = num
}

type meta struct {
	Offset float64
	MaxAmp float64
}

func (s *scForInstrument) GetSampleOffset(name string) int {
	data, err := ioutil.ReadFile(name + ".meta")
	if err != nil {
		fmt.Printf("file not found: " + name + ".meta, using offset of 0")
		return 0
	}

	var m meta
	err = json.Unmarshal(data, &m)
	if err != nil {
		panic("invalid json format for " + name + ".meta")
	}
	return int(RoundFloat(m.Offset*1000.0, 0)) * -1
}

func (s *scForInstrument) Write(b []byte) (num int, err error) {
	// return s.sc.buffer.Write(b)
	return s.eventBuffer.Write(b)
}

func (s *scForInstrument) GetBus(name string) int {
	busno, ok := s.sc.busses[name]
	if !ok {
		return -1
	}
	return busno
}

func (s *sc) Instrument(name string, offset int) Instrument {
	return &instrument{name: name, sc: s.scForInstr, offset: offset}
}

func (s *sc) Route(name string) Instrument {
	return &instrument{name: name, sc: s.scForInstr, bus: true}
}

func (s *sc) Sample(name string, numChan int, offset int) *sample {
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
	return &sample{s.sampleNumber, &instrument{offset: offset, name: fmt.Sprintf("sample%d", numChan), sc: s.scForInstr}, name}
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
	s.synthDefDirs = append(s.synthDefDirs, p)
	return

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

func (s *sc) loadSynthDefInDir(dir string, synthdef string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(dir, synthdef+".scd"))
}

func (s *sc) loadSynthDef(synthdef string) []byte {
	for _, d := range s.synthDefDirs {
		if data, err := s.loadSynthDefInDir(d, synthdef); err == nil {
			return data
		}
	}

	panic("could not find " + synthdef + ".scd in\n" + strings.Join(s.synthDefDirs, ",\n") + "\n")
}

func (s *sc) writeSynthDefs(w io.Writer) {

	for instr := range s.usedInstruments {
		sdef := s.loadSynthDef(instr)
		fmt.Fprintf(w, strings.TrimSpace(string(sdef))+".writeDefFile;")
	}

	/*
		for name, sdef := range s.synthdefs {
			_, has := s.usedInstruments[name]
			if !has {
				// fmt.Printf("instrument not used: %s\n", name)
				continue
			}
			fmt.Fprintf(w, strings.TrimSpace(string(sdef))+".writeDefFile;")
		}
	*/
}

func (s *sc) writeLoadSamples(w io.Writer) {
	// fmt.Printf("used samples: %#v\n", s.usedSamples)
	//for sampleName, sampleId := range s.samples {
	channelPlayers := map[int]struct{}{}

	for sampleName, _ := range s.samples {
		fullpath := filepath.Join(s.sampleDir, sampleName)

		_, has := s.usedSamples[sampleName]
		if !has {
			// fmt.Printf("sample not used: %s\n", sampleName)
			continue
		}

		ch, err := numChannels(fullpath)

		if err != nil {
			panic(fmt.Sprintf("can't open sample file %s, reason: %s", sampleName, err.Error()))
		}

		if ch != s.samplesChannels[sampleName] {
			panic(fmt.Sprintf("sample file %s has %d channels and not %d", sampleName, ch, s.samplesChannels[sampleName]))
		}

		if s.WriteSynthDefs {
			if _, has := channelPlayers[s.samplesChannels[sampleName]]; !has {
				channelPlayers[s.samplesChannels[sampleName]] = struct{}{}
				fmt.Fprintf(w, strings.TrimSpace(fmt.Sprintf(sampleLoader, s.samplesChannels[sampleName], s.samplesChannels[sampleName]))+".writeDefFile;")
			}
		}
	}
}

type eventWriterOptions struct {
	startOffset  uint
	startTick    uint
	tickNegative int
	ticksSorted  []int
	finTick      uint
	tickMapped   map[int][]*Event
}

func (s *sc) writeEvents(w io.Writer, opts eventWriterOptions) (skipSecs float32) {
	t := 0
	// withStartTick := 1.0
	skipSecs = float32(0.0)

	beginOffset := float32(opts.startOffset) / float32(1000)

	if opts.startTick != 0 {
		skipSecs = getSeconds(int(opts.startTick), opts.tickNegative, beginOffset)
		// skipSecs = tickToSeconds(int(startTick)+(tickNegative*(-1))) + 0.000001 + beginOffset
	}

	_ = skipSecs

	for _, ti := range opts.ticksSorted {
		if opts.finTick != 0 && int(opts.finTick) <= ti {
			t = int(opts.finTick)
			break
		}
		inSecs := getSeconds(ti, opts.tickNegative, beginOffset)
		// inSecs := tickToSeconds(ti+(tickNegative*(-1))) + 0.000001 + beginOffset
		fmt.Fprintf(w, `  [%0.6f`, inSecs)
		for _, ev := range opts.tickMapped[ti] {
			ev.Runner(ev)
		}
		t = ti
		fmt.Fprintf(w, "],\n")
	}

	fmt.Fprintf(w, "  [%0.6f, [\\g_deepFree, 1], [\\c_set, 0, 0]]];\n", float32(t)/float32(1000000000))
	return
}

func (s *sc) writeAtPosZero(w io.Writer) {
	// fmt.Printf("used samples: %#v\n", s.usedSamples)
	fmt.Fprintf(w, `  [%0.6f, `, 0.0)

	// create the bus routing group
	fmt.Fprintf(w, fmt.Sprintf(`[\g_new, %d, 0, 0],`, 1200))
	// create the instruments group
	fmt.Fprintf(w, fmt.Sprintf(`[\g_new, %d, 0, 0],`, 1010))

	for _, gr := range s.groups {
		fmt.Fprintf(w, `[\g_new, %d, 1, %d], `, gr.Id(), gr.parent)
	}

	// /g_new

	first_sample := true
	// [\b_allocRead, 1, "/home/benny/Entwicklung/gopath/src/github.com/metakeule/music/example/samples/piano.aiff"],
	for sampleName, sampleId := range s.samples {
		_, has := s.usedSamples[sampleName]
		if !has {
			// fmt.Printf("sample not used: %s\n", sampleName)
			continue
		}
		fullpath := filepath.Join(s.sampleDir, sampleName)
		if !first_sample {
			fmt.Fprintf(w, ", ")
		}
		first_sample = false
		fmt.Fprintf(w, fmt.Sprintf(`[\b_allocRead, %d, "%s"]`, sampleId, fullpath))
	}

	fmt.Fprintf(w, "],\n")

}

// startOffset is in milliseconds and must be positive
func (s *sc) Play(startOffset uint) {

	evts := []*Event{}

	for _, tr := range s.tracks {
		tr.compile()
		evts = append(evts, tr.Events...)
	}

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

	tickMapped := map[int][]*Event{}
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

	opts := eventWriterOptions{}
	opts.startOffset = startOffset
	opts.startTick = startTick
	opts.tickNegative = tickNegative
	opts.ticksSorted = ticksSorted
	opts.finTick = finTick
	opts.tickMapped = tickMapped

	// println("before write events")
	// we have to do the write events before the synthdefs and loadsamples because
	// we need to run everything to know which samples and synthdefs are needed
	skipSecs := s.writeEvents(&s.scForInstr.eventBuffer, opts)
	// println("after write events")

	s.buffer = &bytes.Buffer{}
	fmt.Fprintf(s.buffer, "(\n")

	s.checkForScServer()

	// TODO: make flags to upload samples and write definition files when in server mode
	if s.WriteSynthDefs {
		s.writeSynthDefs(s.buffer)
	}

	if !s.scServerOnline || s.LoadSamples {
		s.writeLoadSamples(s.buffer)
	}

	// eventBuffer := &bytes.Buffer{}
	// skipSecs := s.writeEvents(eventBuffer, opts)

	fmt.Fprintf(s.buffer, "TempoClock.default.tempo = 1; \n")
	fmt.Fprintf(s.buffer, "x = [\n")
	s.writeAtPosZero(s.buffer)
	// s.scForInstr.eventBuffer = bytes.Buffer{}

	// println("before copy")
	// s.buffer.Write(s.scForInstr.eventBuffer.Bytes())
	io.Copy(s.buffer, &s.scForInstr.eventBuffer)
	// println("after copy")

	// skipSecs := s.writeEvents(s.buffer, opts)

	// now := time.Now()

	// TODO change the generating code, so that the online server is reused
	if s.scServerOnline {
		println("server is online")
		fmt.Fprintf(s.buffer, "\n\nScore.play(x); )")
		err := s.runBulkScServerCode(s.buffer.String())
		if err != nil {
			panic(err)
			// println("server online")
		}
		// println("sent")
		//time.Sleep(time.Millisecond * 500)
		// time.Sleep(time.Second * 2)
		return
	}

	println("server is NOT online")
	fmt.Fprintf(s.buffer, `Score.write(x, "`+oscCodeFile+`");`+"\n")
	fmt.Fprintf(s.buffer, "\n\n"+` "quitting".postln; 0.exit; )`)
	err = ioutil.WriteFile(sclangCodeFile, s.buffer.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	// println("osc file written")
	fmt.Printf("tempfile %s written\n", sclangCodeFile)
	now := time.Now()
	if !s.mkOSCFile(libraryPath, sclangCodeFile) {
		println("could not write osc file")
		return
	}

	// fileWriteTime := time.Since(now)

	SclangTime := time.Since(now)
	now = time.Now()
	var exportFloat bool

	if s.AudioFile == "" {
		exportFloat = true
	}

	if !s.mkAudiofile(oscCodeFile, audioFile, exportFloat) {
		println("could not write audio file")
		return
	}

	println("audio file written")
	ScsynthTime := time.Since(now)

	// fmt.Printf("Time:\nwrite file: %s\nsclang: %s\nScsynth: %s\n", fileWriteTime, SclangTime, ScsynthTime)
	fmt.Printf("Time:\nsclang: %s\nScsynth: %s\n", SclangTime, ScsynthTime)

	if s.AudioFile == "" {
		playFile(audioFile, skipSecs)
	}
}

func getSeconds(tick int, negativeOffset int, offset float32) float32 {
	return tickToSeconds(tick+(negativeOffset*(-1))) + 0.000001 + offset
}

func (s *sc) runBulkScServerCode(code string) error {
	/*
		strArr := strings.Split(code, "\n")

		for _, str := range strArr {
			err := s.runScServerCode(str)
			if err != nil {
				return err
			}
		}

		return nil
	*/
	// return s.runScServerCode(code)
	// println(strings.Replace(code, "\n", "", -1))
	return s.runScServerCode(strings.Replace(code, "\n", "", -1))
}

func (s *sc) runScServerCode(code string) error {
	res, err := http.Post("http://localhost:9999/run", "application/octet-stream", strings.NewReader(code))
	if err == nil {
		defer res.Body.Close()
		b, err2 := ioutil.ReadAll(res.Body)
		if err2 == nil {
			if string(b) == "ok" {
				return nil
			} else {
				return fmt.Errorf(string(b))
			}
		} else {
			return err2
		}
	} else {
		return err
	}
}

func (s *sc) checkForScServer() {
	if s.runScServerCode(`"Go music script".postln;`) == nil {
		s.scServerOnline = true
	}
}

func (s *sc) mkOSCFile(libraryPath, sclangCodeFile string) (ok bool) {
	cmd := exec.Command(
		"sclang",
		"-r",
		"-s",
		"-l",
		libraryPath,
		sclangCodeFile,
	)
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("ERROR running sclang")
		fmt.Printf("%s\n", out)
		fmt.Println(err)
		return false
	}
	return true
}

func (s *sc) mkAudiofile(oscCodeFile, audioFile string, exportFloat bool) (ok bool) {
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

	format := "int32"
	if exportFloat {
		format = "float"
	}

	cmd := exec.Command(
		"scsynth",
		"-N",
		oscCodeFile,
		"_",
		audioFile,
		"96000",
		"AIFF",
		format,
		"-o",
		"2",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ERROR running scsynth")
		fmt.Println(err)
		fmt.Printf("%s\n", out)
		return false
	}
	return true
}

func playFile(audioFile string, skipSecs float32) (ok bool) {
	// S16_BE
	// --channels=2 --file-type raw|au|voc|wav --rate=48000 --format=S16_BE
	//cmd = exec.Command("aplay", "--rate=48000", "-f", "cdr", audioFile)
	//cmd = exec.Command("aplay", "--rate=48000", "-f", "U24_BE", audioFile)
	//cmd = exec.Command("aplay", "-f", "S16_BE", "-c2", "--rate=48000", audioFile)
	// "--start-delay=1000"

	// cmd = exec.Command("aplay", "-f", "FLOAT_BE", "-c2", "--rate=96000", audioFile)
	//cmd = exec.Command("aplay", "-f", "S32_BE", "-c2", "--rate=48000", audioFile)
	// -f S16_BE -c2 -f44100

	cmd := exec.Command(
		"play",
		"-q",
		audioFile,
		"trim",
		fmt.Sprintf(`%0.6f`, skipSecs),
	)

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR running play")
		fmt.Println(err)
		return false
	}
	return true
}
