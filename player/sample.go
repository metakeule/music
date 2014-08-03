package player

import (
	"fmt"

	"github.com/metakeule/music"
)

type sample struct {
	// func (in *instrument) New(num int) []music.Voice {
	num int
	*instrument
	// offset int
}

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

var sampleLoader = `SynthDef("sample%d", { |gate=1,bufnum = 0,amp=1, out=0, pan=0| var z;
	z =  EnvGen.kr(Env.perc,gate) * PlayBuf.ar(%d, bufnum, BufRateScale.kr(bufnum));
	FreeSelfWhenDone.kr(z);
	Out.ar(out, Pan2.ar(z, pos: pan, level: amp));
} )`

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
