package example

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

var null = &pair{} // empty list

type pair struct {
	car, cdr interface{}
}
type symbol struct {
	name string
}

func cons(car, cdr interface{}) *pair { return &pair{car, cdr} }
func atom(s string) *symbol           { return &symbol{s} }
func (s *symbol) String() string      { return s.name }
func str(v interface{}) string {
	if s, ok := v.(string); ok {
		return strconv.Quote(s)
	} else {
		return fmt.Sprint(v)
	}
}
func (p *pair) String() string {
	if p == null {
		return "()"
	}
	var b strings.Builder
	b.WriteString("(")
	var isCons bool
	for {
		b.WriteString(str(p.car))
		cdr := p.cdr
		if cdr == null {
			break
		}
		p, isCons = cdr.(*pair)
		if !isCons {
			b.WriteString(" . ")
			b.WriteString(str(cdr))
			break
		}
		b.WriteString(" ")
	}
	b.WriteString(")")
	return b.String()
}

func TestCons(t *testing.T) {
	fmt.Println(cons("hello", "world"))
	fmt.Println(cons(atom("hello"), atom("world")))
	fmt.Println(cons(1, 2))
	fmt.Println(cons(1, cons(2, 3)))
	fmt.Println(cons(1, cons(2, cons(3, null))))
}

func TestParser(t *testing.T) {
	src := `
(define (fact n) 
	(if (= n 0)
		1
		( * n (fact(- n 1))))) ; fact
(fact 10)
; hello world
'('"hello" . '"world")
"hello\nworld\t!"
`
	i, err := parse(src)
	if err != nil {
		panic(err)
	}
	fmt.Println(i)
}

func parse(s string) (interface{}, error) {
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
	tokAtom := Many1(NoneOf(" ().;\t\r\n"))

	NumRule.Pattern = LitFloat.Map(parseNum)
	StrRule.Pattern = LitStr.Map(parseStr)
	AtomRule.Pattern = tokAtom.Map(parseAtom)
	QuoteRule.Pattern = Right(Char('\''), SExpr).Map(parseQuote)
	DotListRule.Pattern = Mid(tokLp, List(SepBy1(SExpr, Spaces), tokDot, SExpr), tokRp).Map(parseDotList)
	ListRule.Pattern = Mid(tokLp, SepBy(SExpr, Spaces), tokRp).Map(parseList)
	CommentRule.Pattern = List(Str(";"), ManyTill(AnyChar(), NewLine))
	SExpr.Pattern = Label(Choice(NumRule, StrRule, QuoteRule, CommentRule, ListRule, DotListRule, AtomRule), "expect sexpr")

	sep := SkipMany(Either(CommentRule, Space))
	pgrm := Right(Try(sep), SepEndBy1(SExpr, sep))
	return ExpectEof(pgrm).Parse(NewState(s + "\n")) // \n 统一单行注释tillNewLine
}

func parseList(v interface{}) interface{} {
	xs := v.([]interface{})
	if len(xs) == 0 {
		return null
	}
	return cons(xs[0], parseList(xs[1:]))
}
func parseDotList_(xs []interface{}, tail interface{}) interface{} {
	if len(xs) == 0 {
		return tail
	}
	return cons(xs[0], parseDotList_(xs[1:], tail))
}
func parseDotList(v interface{}) interface{} {
	xs := v.([]interface{})
	return parseDotList_(xs[0].([]interface{}), xs[2])
}
func parseQuote(v interface{}) interface{} { return cons(atom("quote"), cons(v, null)) }
func parseAtom(v interface{}) interface{}  { return atom(xs2str(v)) }
func parseNum(v interface{}) interface{} {
	n, err := strconv.ParseFloat(v.(string), 64)
	if err != nil {
		panic("invalid number: " + v.(string))
	}
	return n
}
func parseStr(v interface{}) interface{} {
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
