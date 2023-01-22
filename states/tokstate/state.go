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
	parsec.Pos
	ud interface{}
}

func (t *TokState) Save() parsec.Pos     { return t.Pos }
func (t *TokState) Restore(l parsec.Pos) { t.Pos = l }
func (t *TokState) Next() (interface{}, bool) {
	if t.Idx >= len(t.seq) {
		return nil, false
	}
	return t.forward(), true
}
func (t *TokState) forward() *lexer.Token {
	tok := t.seq[t.Idx]
	t.Idx++
	t.Col = tok.Col
	t.Line = tok.Line
	return tok
}
func (t *TokState) Put(ud interface{}) { t.ud = ud }
func (t *TokState) Get() interface{}   { return t.ud }
