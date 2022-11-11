package parsec

import "fmt"

type State interface {
	Next() (interface{}, bool)
	Save() Pos
	Restore(l Pos)
	Put(interface{})
	Get() interface{}
}

type Error struct {
	Pos
	Msg string
}

func (e Error) Error() string { return fmt.Sprintf("%s in %s", e.Msg, e.Pos) }

type Pos struct {
	Idx  int
	Col  int
	Line int
}

func (p Pos) String() string {
	return fmt.Sprintf("pos %d line %d col %d", p.Idx+1, p.Line+1, p.Col+1)
}
