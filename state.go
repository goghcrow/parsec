package parsec

import "fmt"

type State interface {
	Next() (interface{}, bool)
	Save() Loc
	Restore(l Loc)
	Put(interface{})
	Get() interface{}
}

type Error struct {
	Loc
	Msg string
}

func (e Error) Error() string { return fmt.Sprintf("%s in %s", e.Msg, e.Loc) }

type Loc struct {
	Col  int
	Line int
	Pos  int
}

func (l Loc) String() string {
	return fmt.Sprintf("pos %d line %d col %d", l.Pos+1, l.Line+1, l.Col+1)
}
