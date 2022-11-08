# parsec

Golang parsec library inspired by [haskell parsec](https://hackage.haskell.org/package/parsec).

## Combinator

Document of [Combinators](https://hackage.haskell.org/package/parsec-3.1.15.1/docs/Text-ParserCombinators-Parsec-Combinator.html)

Document of [BuildExpressionParser](https://hackage.haskell.org/package/parsec-3.1.15.1/docs/Text-Parsec-Expr.html)

```go
func Return(x interface{}) Parser
func Fail(f string, a ...interface{}) Parser
func Map(p Parser, f func(interface{}) interface{}) Parser
func Bind(p Parser, f func(interface{}) Parser) Parser
func Seq(front, rear Parser, mapper func(x, y interface{}) interface{}) Parser
func List(ps ...Parser) Parser
func Try(p Parser) Parser
func LookAhead(p Parser) Parser
func Either(a, b Parser) Parser
func Choice(xs ...Parser) Parser
func Count(p Parser, n int) Parser
func Between(open, close, p Parser) Parser
func Mid(start, p, end Parser) Parser
func Left(l, r Parser) Parser
func Right(l, r Parser) Parser
func Trim(p, cut Parser) Parser
func Option(p Parser, x interface{}) Parser
func Optional(p Parser) Parser
func SkipMany(p Parser) Parser
func SkipMany1(p Parser) Parser
func Many(p Parser) Parser
func Many1(p Parser) Parser
func SepBy(p, sep Parser) Parser
func SepBy1(p, sep Parser) Parser
func EndBy(end, sep Parser) Parser
func EndBy1(end, sep Parser) Parser
func SepEndBy(p, sep Parser) Parser
func SepEndBy1(p, sep Parser) Parser
func Chainl(p, op Parser, x interface{}) Parser
func Chainl1(p, op Parser) Parser
func Chainr(p, op Parser, x interface{}) Parser
func Chainr1(p, op Parser) Parser
func NotFollowedBy(p Parser) Parser
func ManyTill(p, end Parser) Parser
func ExpectEof(p Parser) Parser
func Label(p Parser, fmt string, a ...interface{}) Parser
func Trace(p Parser, trace func(error, interface{}, []interface{})) Parser

// alias
var (
	Unit    = Return
	FlatMap = Bind
	Apply   = Map
	Alt     = Choice
	Skip    = Optional
	Rep     = Count
)
```
## States

As parametric input stream, [Byte State](states/bytestate), [Rune State](states/charstate) or [Token State](states/tokstate) are builtin supporting.
And you can write your input state by implementing [`State`](state.go#L5) interface.

Although parsec can implement both lexer and parser, and can even directly calculate the results at once, 
it is still recommended to use token state and generate ast through `apply` function, which will have clearer responsibilities.


## Examples 

[An example of parser that eliminate left recursion.](example/rec_str_test.go)

```haskell
expr    = term   `chainl1` addop
term    = factor `chainl1` mulop
factor  = parens expr <|> integer
 
mulop   =   do{ symbol "*"; return (*)   }
        <|> do{ symbol "/"; return (div) }

addop   =   do{ symbol "+"; return (+) }
        <|> do{ symbol "-"; return (-) }
```

```go
package example

import (
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

func TestLRec(t *testing.T) {
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
		n, _ := strconv.ParseInt(v.(string), 10, 64)
		return n
	}

	// tokens
	tokLP := Trim(Char('('), Space)
	tokRP := Trim(Char(')'), Space)
	tokTimes := Trim(Char('*'), Space)
	tokDiv := Trim(Char('/'), Space)
	tokPlus := Trim(Char('+'), Space)
	tokSub := Trim(Char('-'), Space)
	tokInt := Trim(LitInt, Space)

	// syntax
	Expr := NewRule()
	Term := NewRule()
	Factor := NewRule()
	Mulop := NewRule()
	Addop := NewRule()

	Expr.Pattern = Chainr1(Term, Addop)
	Term.Pattern = Chainr1(Factor, Mulop)
	Factor.Pattern = Alt(
		Mid(tokLP, Expr, tokRP),
		tokInt.Map(applyInt),
	)
	Mulop.Pattern = Alt(
		tokTimes.Map(applyBinOp("*")),
		tokDiv.Map(applyBinOp("/")),
	)
	Addop.Pattern = Alt(
		tokPlus.Map(applyBinOp("+")),
		tokSub.Map(applyBinOp("-")),
	)

	calc := func(s string) int64 {
		v, err := ExpectEof(Expr).Parse(NewState(s))
		if err != nil {
			panic(err)
		}
		return v.(int64)
	}

	t.Log(calc("1 + 2 * ( 6 - 3 ) + 3"))
}
```

[An example of an expression parser that handles prefix signs, postfix increment and basic arithmetic.](example/buildexpr_tokenstate_test.go)

```haskell
expr    = buildExpressionParser table term
        <?> "expression"

term    =  parens expr
        <|> natural
        <?> "simple expression"

table   = [ [prefix "-" negate, prefix "+" id ]
          , [postfix "++" (+1)]
          , [binary "*" (*) AssocLeft, binary "/" (div) AssocLeft ]
          , [binary "+" (+) AssocLeft, binary "-" (-)   AssocLeft ]
          ]

binary  name fun assoc = Infix (do{ reservedOp name; return fun }) assoc
prefix  name fun       = Prefix (do{ reservedOp name; return fun })
postfix name fun       = Postfix (do{ reservedOp name; return fun })
```

```go
package example

import (
	"strconv"
	"testing"

	"github.com/goghcrow/go-parsec/lexer"
	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/exprparser"
	. "github.com/goghcrow/parsec/states/tokstate"
)

func TestBuildExpressionParser(t *testing.T) {
	const (
		Lp lexer.TokenKind = iota + 1
		Rp

		OpPlus
		OpSub
		OpTimes
		OpDiv

		OpIncr
		OpDecr

		Add
		Sub
		Mul
		Div
		Incr
		Decr

		Int
		Space
	)
	
	lex := lexer.BuildLexer(func(lex *lexer.Lexicon) {
		lex.Str(Lp, "(")
		lex.Str(Rp, ")")
		lex.Str(OpTimes, "*")
		lex.Str(OpDiv, "/")
		lex.Str(OpIncr, "++")
		lex.Str(OpPlus, "+")
		lex.Str(OpDecr, "--")
		lex.Str(OpSub, "-")

		lex.Keyword(Add, "add")
		lex.Keyword(Sub, "sub")
		lex.Keyword(Mul, "mul")
		lex.Keyword(Div, "div")
		lex.Keyword(Incr, "incr")
		lex.Keyword(Decr, "decr")

		lex.Regex(Int, `\d+`)
		lex.Regex(Space, `\s+`).Skip()
	})

	applyPrefix := func(op interface{}) interface{} {
		return func(v interface{}) interface{} {
			switch op.(*lexer.Token).TokenKind {
			case OpSub:
				return -v.(int64)
			case OpPlus:
				return v
			case OpDecr, Decr:
				return v.(int64) - 1
			case OpIncr, Incr:
				return v.(int64) + 1
			default:
				panic("unreached")
			}
		}
	}
	applyPostfix := func(op interface{}) interface{} {
		return func(v interface{}) interface{} {
			switch op.(*lexer.Token).TokenKind {
			case OpIncr:
				return v.(int64) + 1
			case OpDecr:
				return v.(int64) - 1
			default:
				panic("unreached")
			}
		}
	}
	applyBinary := func(op interface{}) interface{} {
		return func(x, y interface{}) interface{} {
			switch op.(*lexer.Token).TokenKind {
			case OpPlus, Add:
				return x.(int64) + y.(int64)
			case OpSub, Sub:
				return x.(int64) - y.(int64)
			case OpTimes, Mul:
				return x.(int64) * y.(int64)
			case OpDiv, Div:
				return x.(int64) / y.(int64)
			default:
				panic("unreached")
			}
		}
	}
	applyNature := func(v interface{}) interface{} {
		n, err := strconv.ParseInt(v.(*lexer.Token).Lexeme, 10, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	parens := func(p Parser) Parser { return Between(Str("("), Str(")"), p) }
	reservedOp := func(s string) Parser { return Str(s) }

	binary := func(name string, assoc Assoc) Operator {
		return Operator{
			OperKind: Infix,
			Assoc:    assoc,
			Parser:   reservedOp(name).Map(applyBinary),
		}
	}
	prefix := func(name string) Operator {
		return Operator{
			OperKind: Prefix,
			Parser:   reservedOp(name).Map(applyPrefix),
		}
	}
	postfix := func(name string) Operator {
		return Operator{
			OperKind: Postfix,
			Parser:   reservedOp(name).Map(applyPostfix),
		}
	}

	// 优先级倒序, 同组优先级相同
	table := [][]Operator{
		{
			prefix("-"),
			prefix("+"),
		},
		{
			prefix("--"),
			prefix("++"),
			postfix("++"),
			postfix("--"),

			prefix("incr"),
			prefix("decr"),
		},
		{
			binary("*", AssocLeft),
			binary("/", AssocLeft),

			binary("mul", AssocLeft),
			binary("div", AssocLeft),
		},
		{
			binary("+", AssocLeft),
			binary("-", AssocLeft),

			binary("add", AssocLeft),
			binary("sub", AssocLeft),
		},
	}

	Expr := NewRule()
	Term := NewRule()

	Expr.Pattern = Label(BuildExpressionParser(table, Term), "expect expression")

	Term.Pattern = Label(Alt(parens(Expr), Tok(Int, "int").Map(applyNature)), "expect simple expression")

	calc := func(s string) int64 {
		v, err := ExpectEof(Expr).Parse(NewState(lex.MustLex(s)))
		if err != nil {
			panic(err)
		}
		return v.(int64)
	}

	for _, tt := range []struct {
		s      string
		expect int64
	}{
		{"++1", 2},
		{"--1", 0},
		{"incr 1", 2},
		{"decr 1", 0},
		{"1++", 2},
		{"1--", 0},
		{"++1++", 3},
		{"--1--", -1},
		{"++1--", 1},
		{"--1++", 1},

		{"---2", -3},
		{"-2++", -1},

		{"1 + 2 * (6 - -3) / 3 - 3", 4},
		{"1 add 2 mul (6 sub -3) div 3 sub 3", 4},
	} {
		t.Run(tt.s, func(t *testing.T) {
			actual := calc(tt.s)
			if actual != tt.expect {
				t.Errorf("%s expect %d actual %d", tt.s, tt.expect, actual)
			}
		})
	}
}
```