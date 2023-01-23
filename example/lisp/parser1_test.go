package lisp

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/goghcrow/lexer"
	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/tokstate"
)

func TestPrintCons(t *testing.T) {
	fmt.Println(cons("hello", "world"))
	fmt.Println(cons(atom("hello"), atom("world")))
	fmt.Println(cons(1, 2))
	fmt.Println(cons(1, cons(2, 3)))
	fmt.Println(cons(1, cons(2, cons(3, null))))
}

func TestParser1(t *testing.T) {
	src := `
(define (fact n) 
	(if (= n 0)
		1
		( * n (fact(- n 1))))) ; fact
(fact 10)
; hello world
(display '('"hello\n" . '"world\t!"))
; comment eof
`
	v, err := sExprParser(src)
	if err != nil {
		panic(err)
	}
	if fmt.Sprint(v) != `[(define (fact n) (if (= n 0) 1 (* n (fact (- n 1))))) (fact 10) (display (quote ((quote "hello\n") quote "world\t!")))]` {
		panic(fmt.Sprint(v))
	}
}

type Token = *lexer.Token

var sExprParser = buildSExprParser1()

func buildSExprParser1() func(s string) (interface{}, error) {
	const (
		TLp lexer.TokenKind = iota + 1
		TRp
		TDot
		TAtom
		TNum
		TStr
		TQuote
		TComment
		TSpace
	)
	var (
		tLP    = Tok(TLp, "(")
		tRP    = Tok(TRp, ")")
		tDot   = Tok(TDot, ".")
		tAtom  = Tok(TAtom, "atom")
		tNum   = Tok(TNum, "number")
		tStr   = Tok(TStr, "string")
		tQuote = Tok(TQuote, "quote")
	)
	lex := lexer.BuildLexer(func(lex *lexer.Lexicon) {
		lex.Str(TLp, "(")
		lex.Str(TRp, ")")
		lex.Str(TDot, ".")
		lex.Regex(TNum, lexer.RegFloat+"|"+lexer.RegInt)
		lex.Regex(TStr, lexer.RegStr)
		lex.Str(TQuote, "'")
		lex.Regex(TAtom, `[^().;\s]+`)
		lex.Regex(TComment, `(?:;.*?\n)|(?:;.*?\z)`).Skip()
		lex.Regex(TSpace, `\s+`).Skip()
	})

	SExpr := NewRule()
	StrRule := NewRule()
	AtomRule := NewRule()
	NumRule := NewRule()
	QuoteRule := NewRule()
	DotListRule := NewRule()
	ListRule := NewRule()

	NumRule.Pattern = tNum.Map(toNum)
	StrRule.Pattern = tStr.Map(toStr)
	AtomRule.Pattern = tAtom.Map(toAtom)
	QuoteRule.Pattern = Right(tQuote, SExpr).Map(toQuote)
	DotListRule.Pattern = Mid(tLP, List(Many1(SExpr), tDot, SExpr), tRP).Map(toDotList)
	ListRule.Pattern = Mid(tLP, Many(SExpr), tRP).Map(toList)
	SExpr.Pattern = Label(Choice(NumRule, StrRule, AtomRule, QuoteRule, DotListRule, ListRule), "expect sexpr")

	pgrm := ExpectEof(Many1(SExpr))

	return func(s string) (interface{}, error) {
		toks, err := lex.Lex(s)
		if err != nil {
			return nil, err
		}
		v, err := pgrm.Parse(NewState(toks))
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func toList(v interface{}) interface{} {
	xs := v.([]interface{})
	if len(xs) == 0 {
		return null
	}
	return cons(xs[0], toList(xs[1:]))
}
func parseDotList_(xs []interface{}, tail interface{}) interface{} {
	if len(xs) == 0 {
		return tail
	}
	return cons(xs[0], parseDotList_(xs[1:], tail))
}
func toDotList(v interface{}) interface{} {
	xs := v.([]interface{})
	return parseDotList_(xs[0].([]interface{}), xs[2])
}
func toQuote(v interface{}) interface{} { return cons(atom("quote"), cons(v, null)) }
func toAtom(v interface{}) interface{}  { return atom(v.(Token).Lexeme) }
func toNum(v interface{}) interface{} {
	n, err := strconv.ParseFloat(v.(Token).Lexeme, 64)
	if err != nil {
		panic("invalid number: " + v.(Token).Lexeme)
	}
	return n
}
func toStr(v interface{}) interface{} {
	v, err := strconv.Unquote(v.(Token).Lexeme)
	if err != nil {
		panic("invalid string: " + v.(Token).Lexeme)
	}
	return v
}
