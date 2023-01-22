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
	Pos
	ud interface{}
}

func (s *ByteState) Save() Pos                 { return s.Pos }
func (s *ByteState) Restore(l Pos)             { s.Pos = l }
func (s *ByteState) Next() (interface{}, bool) { return s.NextIf(constTrue) }
func (s *ByteState) NextIf(pred func(byte) bool) (byte, bool) {
	if s.Idx >= len(s.seq) {
		return eof, false
	}
	b := s.seq[s.Idx]
	if pred(b) {
		s.forward(b)
		return b, true
	} else {
		return b, false
	}
}
func (s *ByteState) forward(b byte) {
	s.Idx++
	if b == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}
func (s *ByteState) trapExpect(pos Pos, expect string, actual byte) error {
	if actual == eof {
		return Trap(pos, "expect `%s` actual end of input", expect)
	} else {
		return Trap(pos, "expect `%s` actual `%s`", expect, string(actual))
	}
}
func (s *ByteState) Put(ud interface{}) { s.ud = ud }
func (s *ByteState) Get() interface{}   { return s.ud }
