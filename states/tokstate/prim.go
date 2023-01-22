package tokstate

import (
	"github.com/goghcrow/lexer"
	"github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Primitive
// ----------------------------------------------------------------

func Tok(k lexer.TokenKind, name string) parsec.Parser {
	return parsec.Satisfy(func(v interface{}) bool {
		return k == v.(*lexer.Token).TokenKind
	}, name)
}

func Str(s string) parsec.Parser {
	return parsec.Satisfy(func(v interface{}) bool {
		return s == v.(*lexer.Token).Lexeme
	}, s)
}
