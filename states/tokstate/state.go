package tokstate

import (
	"github.com/goghcrow/go-parsec/lexer"
	"github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Token State
// ----------------------------------------------------------------

func NewState(toks []*lexer.Token) parsec.State { return &TokState{seq: toks} }

type TokState struct {
	seq []*lexer.Token
	parsec.Loc
	ud interface{}
}

func (t *TokState) Save() parsec.Loc     { return t.Loc }
func (t *TokState) Restore(l parsec.Loc) { t.Loc = l }
func (t *TokState) Next() (interface{}, bool) {
	if t.Pos >= len(t.seq) {
		return nil, false
	}
	return t.move(), true
}
func (t *TokState) move() *lexer.Token {
	tok := t.seq[t.Pos]
	t.Pos++
	t.Col = tok.Col
	t.Line = tok.Line
	return tok
}
func (t *TokState) Put(ud interface{}) { t.ud = ud }
func (t *TokState) Get() interface{}   { return t.ud }
