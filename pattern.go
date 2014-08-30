package music

type Pattern interface {
	Pattern(*Track)
}

type PatternFunc func(*Track)

func (tf PatternFunc) Pattern(tr *Track) {
	tf(tr)
}

/*
// func Modify(pos string, v Voice, params ...map[string]float64) *mod {
type seqMod struct {
	v   Voice
	seq []map[string]float64
	Pos int
}
*/

type seqModTrafo struct {
	*seqPlay
	pos            Measure
	overrideParams Parameter
}

func (sm *seqModTrafo) Pattern(tr *Track) {
	tr.At(sm.pos, Change(sm.seqPlay.v, Params(sm.seqPlay.seq[sm.seqPlay.Pos], sm.overrideParams)))
	if sm.seqPlay.Pos < len(sm.seqPlay.seq)-1 {
		sm.seqPlay.Pos++
	} else {
		sm.seqPlay.Pos = 0
	}
}

/*
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
*/

type seqPlay struct {
	seq        []Parameter
	initParams Parameter
	v          *Voice
	Pos        int
}

func (sp *seqPlay) Modify(pos string, params ...Parameter) Pattern {
	return &seqModTrafo{seqPlay: sp, pos: M(pos), overrideParams: Params(params...)}
}

func (sp *seqPlay) PlayDur(pos, dur string, params ...Parameter) Pattern {
	return &seqPlayTrafo{seqPlay: sp, pos: M(pos), dur: M(dur), overrideParams: Params(params...)}
}

type seqPlayTrafo struct {
	*seqPlay
	pos            Measure
	dur            Measure
	overrideParams Parameter
}

func ParamSequence(v *Voice, initParams Parameter, paramSeq ...Parameter) *seqPlay {
	return &seqPlay{
		initParams: initParams,
		seq:        paramSeq,
		v:          v,
	}
}

func (spt *seqPlayTrafo) Params() (p map[string]float64) {
	return Params(spt.seqPlay.initParams, spt.seqPlay.seq[spt.seqPlay.Pos], spt.overrideParams).Params()
}

func (spt *seqPlayTrafo) Pattern(tr *Track) {
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
	*Voice
	Params Parameter
}

func Play(pos string, v *Voice, params ...Parameter) *play {
	return &play{M(pos), v, Params(params...)}
}

func (p *play) Pattern(t *Track) {
	// fmt.Printf("tempo at %s: %v BPM\n", p.pos, t.TempoAt(p.pos))
	t.At(p.pos, On(p.Voice, p.Params))
}

type playDur struct {
	pos Measure
	dur Measure
	*Voice
	Params Parameter
}

func PlayDur(pos, dur string, v *Voice, params ...Parameter) *playDur {
	return &playDur{M(pos), M(dur), v, Params(params...)}
}

func (p *playDur) Pattern(t *Track) {
	// fmt.Printf("tempo at %s: %v BPM\n", p.pos, t.TempoAt(p.pos))
	t.At(p.pos, On(p.Voice, p.Params))
	t.At(p.pos+p.dur, Off(p.Voice))
}

type exec_ struct {
	pos   Measure
	fn    func(e *Event)
	voice *Voice
	type_ string
}

func (e *exec_) Pattern(t *Track) {
	ev := newEvent(e.voice, e.type_)
	ev.Runner = e.fn
	t.At(e.pos, ev)
}

func Exec(pos string, v *Voice, type_ string, fn func(t *Event)) Pattern {
	return &exec_{M(pos), fn, v, type_}
}

type stop struct {
	pos Measure
	*Voice
}

func Stop(pos string, v *Voice) *stop {
	return &stop{M(pos), v}
}

func (p *stop) Pattern(t *Track) {
	t.At(p.pos, Off(p.Voice))
}

// type end struct{}

func (e End) Pattern(t *Track) {
	t.At(M(string(e)), fin)
}

// var End = end{}
type End string

type Start string

func (s Start) Pattern(t *Track) {
	t.At(M(string(s)), start)
}

// var Start = begin{}

type stopAll struct {
	pos    Measure
	Voices []*Voice
}

func StopAll(pos string, vs ...[]*Voice) *stopAll {
	s := &stopAll{pos: M(pos)}

	for _, v := range vs {
		s.Voices = append(s.Voices, v...)
	}

	return s
}

func (p *stopAll) Pattern(t *Track) {
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

func (s *setTempo) Pattern(t *Track) {
	t.SetTempo(s.Pos, s.Tempo)
}

type mod struct {
	pos Measure
	*Voice
	Params Parameter
}

func Modify(pos string, v *Voice, params ...Parameter) *mod {
	return &mod{M(pos), v, Params(params...)}
}

func (p *mod) Pattern(t *Track) {
	t.At(p.pos, Change(p.Voice, p.Params))
}

type times struct {
	times int
	trafo Pattern
}

func (n *times) Pattern(t *Track) {
	for i := 0; i < n.times; i++ {
		n.trafo.Pattern(t)
	}
}

func Times(num int, trafo Pattern) Pattern {
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

/*
func Add(current, step float64) float64 {
	return current + step
}

func Multiply(current, step float64) float64 {
	return current * step
}
*/

func (ts *tempoSpan) SetTempo(pos string) Pattern {
	return &tempoSpanTrafo{ts, pos}
}

func (ts *tempoSpanTrafo) Pattern(t *Track) {
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
func TempoSequence(start float64, step float64, modifier func(current, step float64) float64) *tempoSpan {
	return &tempoSpan{current: start, step: step, modifier: modifier}
}

type seqBool struct {
	seq   []bool
	pos   int
	trafo Pattern
}

func (s *seqBool) Pattern(t *Track) {
	if s.seq[s.pos] {
		s.trafo.Pattern(t)
	}
	if s.pos < len(s.seq)-1 {
		s.pos++
	} else {
		s.pos = 0
	}
}

func PatternOnOffSequence(trafo Pattern, seq ...bool) Pattern {
	return &seqBool{seq: seq, pos: 0, trafo: trafo}
}

type sequence struct {
	Pos int
	seq []Pattern
}

func (s *sequence) Pattern(t *Track) {
	s.seq[s.Pos].Pattern(t)
	if s.Pos < len(s.seq)-1 {
		s.Pos++
	} else {
		s.Pos = 0
	}
}

func PatternSequence(seq ...Pattern) Pattern {
	return &sequence{seq: seq}
}

type compose []Pattern

func (c compose) Pattern(t *Track) {
	for _, trafo := range c {
		trafo.Pattern(t)
	}
}

func Patterns(trafos ...Pattern) Pattern {
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

func (l *linearDistribute) ModifyDistributed(position string, v *Voice) Pattern {
	return &linearDistributeTrafo{l, v, M(position)}
}

type linearDistributeTrafo struct {
	*linearDistribute
	v   *Voice
	pos Measure
}

func (ld *linearDistributeTrafo) Pattern(tr *Track) {
	width, diff := LinearDistributedValues(ld.linearDistribute.from, ld.linearDistribute.to, ld.linearDistribute.steps, ld.linearDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	val := ld.linearDistribute.from
	for i := 0; i < ld.linearDistribute.steps; i++ {
		tr.At(pos, Change(ld.v, ParamsMap(map[string]float64{ld.linearDistribute.key: val})))
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

func (l *expDistribute) ModifyDistributed(position string, v *Voice) Pattern {
	return &expDistributeTrafo{l, v, M(position)}
}

type expDistributeTrafo struct {
	*expDistribute
	v   *Voice
	pos Measure
}

func (ld *expDistributeTrafo) Pattern(tr *Track) {
	width, diffs := ExponentialDistributedValues(ld.expDistribute.from, ld.expDistribute.to, ld.expDistribute.steps, ld.expDistribute.dur)
	// tr.At(ld.pos, Change(ld.v, ))
	pos := ld.pos
	for i := 0; i < ld.expDistribute.steps; i++ {
		tr.At(pos, Change(ld.v, ParamsMap(map[string]float64{ld.expDistribute.key: diffs[i]})))
		pos += width
		//val += diff
	}
}