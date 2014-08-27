package music

type Transformer interface {
	Transform(Tracker)
}

type TransformerFunc func(Tracker)

func (tf TransformerFunc) Transform(tr Tracker) {
	tf(tr)
}

// func Modify(pos string, v Voice, params ...map[string]float64) *mod {
type seqMod struct {
	v   Voice
	seq []map[string]float64
	Pos int
}

type seqModTrafo struct {
	*seqMod
	pos Measure
}

func (sm *seqMod) NextModify(pos string) Transformer {
	return &seqModTrafo{seqMod: sm, pos: M(pos)}
}

func (sm *seqModTrafo) Transform(tr Tracker) {
	tr.At(sm.pos, Change(sm.seqMod.v, Params(sm.seqMod.seq[sm.seqMod.Pos])))
	if sm.seqMod.Pos < len(sm.seqMod.seq)-1 {
		sm.seqMod.Pos++
	} else {
		sm.seqMod.Pos = 0
	}
}

func SeqeuenceModify(v Voice, paramSeq ...Parameter) *seqMod {
	seq := []map[string]float64{}

	for _, p := range paramSeq {
		seq = append(seq, p.Params())
	}

	return &seqMod{
		seq: seq,
		v:   v,
	}
}

type seqPlay struct {
	seq        []map[string]float64
	initParams map[string]float64
	v          Voice
	Pos        int
}

func (sp *seqPlay) NextPlayDur(pos, dur string) Transformer {
	return &seqPlayTrafo{seqPlay: sp, pos: M(pos), dur: M(dur)}
}

type seqPlayTrafo struct {
	*seqPlay
	pos Measure
	dur Measure
}

func SeqeuencePlay(v Voice, initParams Parameter, paramSeq ...Parameter) *seqPlay {
	seq := []map[string]float64{}

	for _, p := range paramSeq {
		seq = append(seq, p.Params())
	}

	s := &seqPlay{
		// initParams: nil,
		seq: seq,
		v:   v,
	}

	if initParams != nil {
		s.initParams = initParams.Params()
	}
	return s
}

func (spt *seqPlayTrafo) Params() (p map[string]float64) {
	p = map[string]float64{}

	for k, v := range spt.seqPlay.initParams {
		p[k] = v
	}

	for k, v := range spt.seqPlay.seq[spt.seqPlay.Pos] {
		p[k] = v
	}
	return
}

func (spt *seqPlayTrafo) Transform(tr Tracker) {
	tr.At(spt.pos, On(spt.seqPlay.v, spt))
	tr.At(spt.pos+spt.dur, Off(spt.seqPlay.v))
	if spt.seqPlay.Pos < len(spt.seqPlay.seq)-1 {
		spt.seqPlay.Pos++
	} else {
		spt.seqPlay.Pos = 0
	}
}

type play struct {
	pos Measure
	Voice
	Params map[string]float64
}

func Play(pos string, v Voice, params ...Parameter) *play {
	return &play{M(pos), v, MergeParams(params...)}
}

func (p *play) Transform(t Tracker) {
	// fmt.Printf("tempo at %s: %v BPM\n", p.pos, t.TempoAt(p.pos))
	t.At(p.pos, On(p.Voice, Params(p.Params)))
}

type playDur struct {
	pos Measure
	dur Measure
	Voice
	Params map[string]float64
}

func PlayDur(pos, dur string, v Voice, params ...Parameter) *playDur {
	return &playDur{M(pos), M(dur), v, MergeParams(params...)}
}

func (p *playDur) Transform(t Tracker) {
	// fmt.Printf("tempo at %s: %v BPM\n", p.pos, t.TempoAt(p.pos))
	t.At(p.pos, On(p.Voice, Params(p.Params)))
	t.At(p.pos+p.dur, Off(p.Voice))
}

type exec struct {
	pos   Measure
	fn    func(e *Event)
	voice Voice
	type_ string
}

func (e *exec) Transform(t Tracker) {
	ev := newEvent(e.voice, e.type_)
	ev.Runner = e.fn
	t.At(e.pos, ev)
}

func Exec(pos string, v Voice, type_ string, fn func(t *Event)) Transformer {
	return &exec{M(pos), fn, v, type_}
}

type stop struct {
	pos Measure
	Voice
}

func Stop(pos string, v Voice) *stop {
	return &stop{M(pos), v}
}

func (p *stop) Transform(t Tracker) {
	t.At(p.pos, Off(p.Voice))
}

// type end struct{}

func (e End) Transform(t Tracker) {
	t.At(M(string(e)), fin)
}

// var End = end{}
type End string

type Start string

func (s Start) Transform(t Tracker) {
	t.At(M(string(s)), start)
}

// var Start = begin{}

type stopAll struct {
	pos    Measure
	Voices []Voice
}

func StopAll(pos string, vs ...[]Voice) *stopAll {
	s := &stopAll{pos: M(pos)}

	for _, v := range vs {
		s.Voices = append(s.Voices, v...)
	}

	return s
}

func (p *stopAll) Transform(t Tracker) {
	for i := 0; i < len(p.Voices); i++ {
		t.At(p.pos, Off(p.Voices[i]))
	}
}

type setTempo struct {
	Tempo Tempo
	Pos   Measure
}

func SetTempo(at string, t Tempo) *setTempo {
	return &setTempo{t, M(at)}
}

func (s *setTempo) Transform(t Tracker) {
	t.SetTempo(s.Pos, s.Tempo)
}

type mod struct {
	pos Measure
	Voice
	Params map[string]float64
}

func Modify(pos string, v Voice, params ...Parameter) *mod {
	return &mod{M(pos), v, MergeParams(params...)}
}

func (p *mod) Transform(t Tracker) {
	t.At(p.pos, Change(p.Voice, Params(p.Params)))
}

type times struct {
	times int
	trafo Transformer
}

func (n *times) Transform(t Tracker) {
	for i := 0; i < n.times; i++ {
		n.trafo.Transform(t)
	}
}

func Times(num int, trafo Transformer) Transformer {
	return &times{times: num, trafo: trafo}
}

type tempoSpan struct {
	current  float64
	step     float64
	modifier func(current, step float64) float64
}

type tempoSpanTrafo struct {
	*tempoSpan
	pos string
}

func Add(current, step float64) float64 {
	return current + step
}

func Multiply(current, step float64) float64 {
	return current * step
}

func (ts *tempoSpan) NextTempoChange(pos string) Transformer {
	return &tempoSpanTrafo{ts, pos}
}

func (ts *tempoSpanTrafo) Transform(t Tracker) {
	var newtempo float64
	if ts.current == -1 {
		newtempo = ts.modifier(t.TempoAt(M(ts.pos)).BPM(), ts.step)
	} else {
		ts.current = ts.modifier(ts.current, ts.step)
		newtempo = ts.current
	}
	//rounded := RoundFloat(newtempo, 4)
	t.SetTempo(M(ts.pos), BPM(newtempo))
}

// for start = -1 takes the current tempo
func TempoSpan(start float64, step float64, modifier func(current, step float64) float64) *tempoSpan {
	return &tempoSpan{current: start, step: step, modifier: modifier}
}

type seqBool struct {
	seq   []bool
	pos   int
	trafo Transformer
}

func (s *seqBool) Transform(t Tracker) {
	if s.seq[s.pos] {
		s.trafo.Transform(t)
	}
	if s.pos < len(s.seq)-1 {
		s.pos++
	} else {
		s.pos = 0
	}
}

func SequenceBool(trafo Transformer, seq ...bool) Transformer {
	return &seqBool{seq: seq, pos: 0, trafo: trafo}
}

type sequence struct {
	Pos int
	seq []Transformer
}

func (s *sequence) Transform(t Tracker) {
	s.seq[s.Pos].Transform(t)
	if s.Pos < len(s.seq)-1 {
		s.Pos++
	} else {
		s.Pos = 0
	}
}

func Sequence(seq ...Transformer) Transformer {
	return &sequence{seq: seq}
}

type compose []Transformer

func (c compose) Transform(t Tracker) {
	for _, trafo := range c {
		trafo.Transform(t)
	}
}

func Compose(trafos ...Transformer) Transformer {
	return compose(trafos)
}

type linearDistribute struct {
	from, to float64
	steps    int
	dur      Measure
	key      string
	// from, to float64, steps int, dur Measure
}

// LinearDistribution creates a transformer that modifies the given parameter param
// from the value from to the value to in n steps in linear growth for a total duration dur
func LinearDistribution(param string, from, to float64, n int, dur Measure) *linearDistribute {
	return &linearDistribute{from, to, n, dur, param}
}

func (l *linearDistribute) ModifyDistributed(position string, v Voice) Transformer {
	return &linearDistributeTrafo{l, v, M(position)}
}

type linearDistributeTrafo struct {
	*linearDistribute
	v   Voice
	pos Measure
}

func (ld *linearDistributeTrafo) Transform(tr Tracker) {
	width, diff := LinearDistributedValues(ld.linearDistribute.from, ld.linearDistribute.to, ld.linearDistribute.steps, ld.linearDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	val := ld.linearDistribute.from
	for i := 0; i < ld.linearDistribute.steps; i++ {
		tr.At(pos, Change(ld.v, Params(map[string]float64{ld.linearDistribute.key: val})))
		pos += width
		val += diff
	}
}

// ---------------------------------------

// func ExponentialDistributedValues(from, to float64, steps int, dur Measure) (width Measure, diffs []float64) {

type expDistribute struct {
	from, to float64
	steps    int
	dur      Measure
	key      string
	// from, to float64, steps int, dur Measure
}

// ExponentialDistribution creates a transformer that modifies the given parameter param
// from the value from to the value to in n steps in exponential growth for a total duration dur
func ExponentialDistribution(param string, from, to float64, n int, dur Measure) *expDistribute {
	return &expDistribute{from, to, n, dur, param}
}

func (l *expDistribute) ModifyDistributed(position string, v Voice) Transformer {
	return &expDistributeTrafo{l, v, M(position)}
}

type expDistributeTrafo struct {
	*expDistribute
	v   Voice
	pos Measure
}

func (ld *expDistributeTrafo) Transform(tr Tracker) {
	width, diffs := ExponentialDistributedValues(ld.expDistribute.from, ld.expDistribute.to, ld.expDistribute.steps, ld.expDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	for i := 0; i < ld.expDistribute.steps; i++ {
		tr.At(pos, Change(ld.v, Params(map[string]float64{ld.expDistribute.key: diffs[i]})))
		pos += width
		//val += diff
	}
}
