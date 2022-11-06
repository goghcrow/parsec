package toks

import (
	"github.com/goghcrow/go-parsec/lexer"
	"github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Primitive Token Parsers
// ----------------------------------------------------------------

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
