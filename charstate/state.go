package chars

import (
	. "github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Rune State
// ----------------------------------------------------------------

func NewStrState(s string) State {
	return &StrState{seq: []rune(s)}
}

const eof rune = -1

type StrState struct {
	seq []rune
	Loc
}

func (s *StrState) Save() Loc                 { return s.Loc }
func (s *StrState) Restore(l Loc)             { s.Loc = l }
func (s *StrState) Next() (interface{}, bool) { return s.NextIf(constTrue) }
func (s *StrState) NextIf(pred func(rune) bool) (rune, bool) {
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
func (s *StrState) move(r rune) {
	s.Pos++
	if r == '\n' {
		s.Line++
		s.Col = 0
	} else {
		s.Col++
	}
}
func (s *StrState) trapExpect(loc Loc, expect string, actual rune) error {
	if actual == eof {
		return Trap(loc, "expect `%s` actual end of input", expect)
	} else {
		return Trap(loc, "expect `%s` actual `%s`", expect, string(actual))
	}
}
