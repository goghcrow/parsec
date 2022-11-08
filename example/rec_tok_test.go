package example

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/goghcrow/go-parsec/lexer"
	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/tokstate"
)

// expr    = term   `chainl1` addop
// term    = factor `chainl1` mulop
// factor  = parens expr <|> integer
//
// mulop   =   do{ symbol "*"; return (*)   }
//
//	<|> do{ symbol "/"; return (div) }
//
// addop   =   do{ symbol "+"; return (+) }
//
//	<|> do{ symbol "-"; return (-) }
func TestLRecTokState(t *testing.T) {
	const (
		Lp lexer.TokenKind = iota + 1
		Rp
		Times
		Div
		Plus
		Sub
		Int
		Space
	)
	lex := lexer.BuildLexer(func(lex *lexer.Lexicon) {
		lex.Str(Lp, "(")
		lex.Str(Rp, ")")
		lex.Str(Times, "*")
		lex.Str(Div, "/")
		lex.Str(Plus, "+")
		lex.Str(Sub, "-")
		lex.Regex(Int, `\d+`)
		lex.Regex(Space, `\s+`).Skip()
	})

	applyBinOp := func(op string) func(v interface{}) interface{} {
		return func(v interface{}) interface{} {
			// op 需要返回函数, 这里直接计算, 或者返回 ast 结点
			return func(x, y interface{}) interface{} {
				switch op {
				case "+":
					return x.(int64) + y.(int64)
				case "-":
					return x.(int64) - y.(int64)
				case "*":
					return x.(int64) * y.(int64)
				case "/":
					return x.(int64) / y.(int64)
				default:
					panic("invalid oper")
				}
			}
		}
	}

	applyInt := func(v interface{}) interface{} {
		n, err := strconv.ParseInt(v.(*lexer.Token).Lexeme, 10, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	// syntax
	Expr := NewRule()
	Term := NewRule()
	Factor := NewRule()
	Mulop := NewRule()
	Addop := NewRule()

	Expr.Pattern = Chainr1(Term, Addop)
	Term.Pattern = Chainr1(Factor, Mulop)
	Factor.Pattern = Alt(
		Mid(Str("("), Expr, Str(")")),
		Tok(Int, "Int").Map(applyInt),
	)
	Mulop.Pattern = Alt(
		Str("*").Map(applyBinOp("*")),
		Str("/").Map(applyBinOp("/")),
	)
	Addop.Pattern = Alt(
		Str("+").Map(applyBinOp("+")),
		Str("-").Map(applyBinOp("-")),
	)

	// debug
	{
		label := func(label string) func(error, interface{}, []interface{}) {
			return func(err error, v interface{}, xs []interface{}) {
				if err == nil {
					str := make([]string, len(xs))
					for i, x := range xs {
						str[i] = x.(*lexer.Token).Lexeme
					}
					fmt.Printf("%s: %v -> %s\n", label, v, strings.Join(str, ""))
				} else {
					fmt.Printf("%s: %s", label, err.Error())
				}
			}
		}
		Expr.Pattern = Trace(Expr.Pattern, label("expr"))
		Term.Pattern = Trace(Term.Pattern, label("term"))
		Factor.Pattern = Trace(Factor.Pattern, label("factor"))
		Mulop.Pattern = Trace(Mulop.Pattern, label("mulop"))
		Addop.Pattern = Trace(Addop.Pattern, label("addop"))
	}

	calc := func(s string) int64 {
		v, err := ExpectEof(Expr).Parse(NewState(lex.MustLex(s)))
		if err != nil {
			panic(err)
		}
		return v.(int64)
	}

	t.Log(calc("1 + 2 * ( 6 - 3 ) + 3"))
}
