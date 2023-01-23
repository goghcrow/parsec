package lisp

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

func TestParser(t *testing.T) {
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
	v, err := parse(src)
	if err != nil {
		panic(err)
	}
	if fmt.Sprint(v) != `[(define (fact n) (if (= n 0) 1 (* n (fact (- n 1))))) (fact 10) (display (quote ((quote "hello\n") quote "world\t!")))]` {
		panic(fmt.Sprint(v))
	}
}

var pgrm = buildSExprParser()

func parse(s string) (interface{}, error) {
	return pgrm.Parse(NewState(s))
}

func buildSExprParser() Parser {
	SExpr := NewRule()
	StrRule := NewRule()
	AtomRule := NewRule()
	NumRule := NewRule()
	QuoteRule := NewRule()
	DotListRule := NewRule()
	ListRule := NewRule()
	CommentRule := NewRule()

	tokLp := Trim(Char('('), Space)
	tokRp := Trim(Char(')'), Space)
	tokDot := Trim(Char('.'), Space)
	tokAtom := Many1(NoneOf("().; \t\r\n\f"))

	s := s1{}

	NumRule.Pattern = LitNum.Map(s.parseNum)
	StrRule.Pattern = LitStr.Map(s.parseStr)
	AtomRule.Pattern = tokAtom.Map(s.parseAtom)
	QuoteRule.Pattern = Right(Char('\''), SExpr).Map(s.parseQuote)
	DotListRule.Pattern = Mid(tokLp, List(SepBy1(SExpr, Spaces), tokDot, SExpr), tokRp).Map(s.parseDotList)
	ListRule.Pattern = Mid(tokLp, SepBy(SExpr, Spaces), tokRp).Map(s.parseList)
	CommentRule.Pattern = List(Str(";"), ManyTill(AnyChar(), Either(NewLine, Eof)))
	SExpr.Pattern = Label(Choice(NumRule, StrRule, QuoteRule, CommentRule, ListRule, DotListRule, AtomRule), "expect sexpr")

	sep := SkipMany(Either(CommentRule, Space))
	return ExpectEof(Right(Try(sep), SepEndBy1(SExpr, sep)))
}

type s1 struct{}

func (s s1) parseList(v interface{}) interface{} {
	xs := v.([]interface{})
	if len(xs) == 0 {
		return null
	}
	return cons(xs[0], s.parseList(xs[1:]))
}
func (s s1) parseDotList_(xs []interface{}, tail interface{}) interface{} {
	if len(xs) == 0 {
		return tail
	}
	return cons(xs[0], s.parseDotList_(xs[1:], tail))
}
func (s s1) parseDotList(v interface{}) interface{} {
	xs := v.([]interface{})
	return s.parseDotList_(xs[0].([]interface{}), xs[2])
}
func (s s1) parseQuote(v interface{}) interface{} { return cons(atom("quote"), cons(v, null)) }
func (s s1) parseAtom(v interface{}) interface{}  { return atom(xs2str(v)) }
func (s s1) parseNum(v interface{}) interface{} {
	n, err := strconv.ParseFloat(v.(string), 64)
	if err != nil {
		panic("invalid number: " + v.(string))
	}
	return n
}
func (s s1) parseStr(v interface{}) interface{} {
	v, err := strconv.Unquote(v.(string))
	if err != nil {
		panic("invalid string: " + v.(string))
	}
	return v
}

func xs2str(v interface{}) string {
	xs := v.([]interface{})
	rs := make([]rune, len(xs))
	for i, x := range xs {
		rs[i] = x.(rune)
	}
	return string(rs)
}
