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
	Loc
	ud interface{}
}

func (s *CharState) Save() Loc                 { return s.Loc }
func (s *CharState) Restore(l Loc)             { s.Loc = l }
func (s *CharState) Next() (interface{}, bool) { return s.NextIf(constTrue) }
func (s *CharState) NextIf(pred func(rune) bool) (rune, bool) {
	if s.Pos >= len(s.seq) {
		return eof, false
	}
	r := s.seq[s.Pos]
	if pred(r) {
		s.move(r)
		return r, true
	} else {
		return r, false
	}
}
func (s *CharState) move(r rune) {
	s.Pos++
	if r == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}
func (s *CharState) trapExpect(loc Loc, expect string, actual rune) error {
	if actual == eof {
		return Trap(loc, "expect `%s` actual end of input", expect)
	} else {
		return Trap(loc, "expect `%s` actual `%s`", expect, string(actual))
	}
}
func (s *CharState) Put(ud interface{}) { s.ud = ud }
func (s *CharState) Get() interface{}   { return s.ud }
