package bytestate

import (
	. "github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Byte State
// ----------------------------------------------------------------

func NewState(s string) State {
	return &ByteState{seq: []byte(s)}
}

const eof byte = 0

type ByteState struct {
	seq []byte
	Loc
	ud interface{}
}

func (s *ByteState) Save() Loc                 { return s.Loc }
func (s *ByteState) Restore(l Loc)             { s.Loc = l }
func (s *ByteState) Next() (interface{}, bool) { return s.NextIf(constTrue) }
func (s *ByteState) NextIf(pred func(byte) bool) (byte, bool) {
	if s.Pos >= len(s.seq) {
		return eof, false
	}
	b := s.seq[s.Pos]
	if pred(b) {
		s.move(b)
		return b, true
	} else {
		return b, false
	}
}
func (s *ByteState) move(b byte) {
	s.Pos++
	if b == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}
func (s *ByteState) trapExpect(loc Loc, expect string, actual byte) error {
	if actual == eof {
		return Trap(loc, "expect `%s` actual end of input", expect)
	} else {
		return Trap(loc, "expect `%s` actual `%s`", expect, string(actual))
	}
}
func (s *ByteState) Put(ud interface{}) { s.ud = ud }
func (s *ByteState) Get() interface{}   { return s.ud }
