package charstate

import (
	. "github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Rune State
// ----------------------------------------------------------------

func NewState(s string) State {
	return &CharState{seq: []rune(s)}
}

const eof rune = -1

type CharState struct {
	seq []rune
	Pos
	ud interface{}
}

func (s *CharState) Save() Pos                 { return s.Pos }
func (s *CharState) Restore(l Pos)             { s.Pos = l }
func (s *CharState) Next() (interface{}, bool) { return s.NextIf(constTrue) }
func (s *CharState) NextIf(pred func(rune) bool) (rune, bool) {
	if s.Idx >= len(s.seq) {
		return eof, false
	}
	r := s.seq[s.Idx]
	if pred(r) {
		s.forward(r)
		return r, true
	} else {
		return r, false
	}
}
func (s *CharState) forward(r rune) {
	s.Idx++
	if r == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}
func (s *CharState) trapExpect(pos Pos, expect string, actual rune) error {
	if actual == eof {
		return Trap(pos, "expect `%s` actual end of input", expect)
	} else {
		return Trap(pos, "expect `%s` actual `%s`", expect, string(actual))
	}
}
func (s *CharState) Put(ud interface{}) { s.ud = ud }
func (s *CharState) Get() interface{}   { return s.ud }
