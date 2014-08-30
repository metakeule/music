package music

type Group struct{}

func (g Group) Name() string {
	return "group"
}

func NewGroup(g Generator) *Voice {
	return &Voice{Generator: g, Instrument: Group{}, SCGroup: g.NewGroupId()}
}

/*
/g_new - create a new group

N *
int	new group ID
int	add action (0,1,2, 3 or 4 see below)
int	add target ID

1	add the new group to the the tail of the group specified by the add target ID.

fmt.Fprintf(v.instrument.sc.buffer, `, [\g_new, \%d, 1, \%d]`, v.instrument.name, v.instrNum, v.paramsStr(ev))
*/
