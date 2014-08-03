package player

import (
	"bytes"
	"fmt"

	"github.com/metakeule/music"
)

/*
/g_new - create a new group

N *
int	new group ID
int	add action (0,1,2, 3 or 4 see below)
int	add target ID

1	add the new group to the the tail of the group specified by the add target ID.

fmt.Fprintf(v.instrument.sc.buffer, `, [\g_new, \%d, 1, \%d]`, v.instrument.name, v.instrNum, v.paramsStr(ev))


*/

type group struct {
	sc     *sc
	id     int
	name   string
	parent int
}

/*
func (g *group) Id(name string) int {
	groupno, ok := g.sc.groupsByName[name]
	if !ok {
		panic("unknown group " + name)
	}
	return groupno
}
*/

func (g *group) Id() int {
	return g.id
}

func (g *group) paramsStr(ev *music.Event) string {
	var buf bytes.Buffer

	for k, v := range ev.FinalParams() {
		if k[0] != '_' {
			fmt.Fprintf(&buf, `, \%s, %v`, k, float32(v))
		}
	}

	return buf.String()

}

func (g *group) Change(ev *music.Event) {
	fmt.Fprintf(g.sc.buffer, `, [\n_set, %d%s]`, g.id, g.paramsStr(ev))
}

func (g *group) Mute(*music.Event)   { panic("mute not allowed for group") }
func (g *group) UnMute(*music.Event) { panic("unmute not allowed for group") }
func (g *group) Name() string        { return g.name }
func (g *group) On(ev *music.Event)  { panic("on not allowed for group") }
func (g *group) Off(ev *music.Event) { panic("off not allowed for group") }
func (g *group) Offset() int         { return 0 }

func (s *sc) NewGroup(name string, parentGroup *group) *group {
	s.groupNumber++
	g := &group{
		id:     s.groupNumber,
		name:   name,
		sc:     s,
		parent: 1010,
	}
	if parentGroup != nil {
		g.parent = parentGroup.Id()
	}
	s.groups = append(s.groups, g)
	return g
}
