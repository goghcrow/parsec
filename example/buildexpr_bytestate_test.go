package example

import (
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/exprparser"
	. "github.com/goghcrow/parsec/states/bytestate"
)

// 解析前缀正负 后缀自增 和数学运算
//
//	expr    = buildExpressionParser table term
//	        <?> "expression"
//
//	term    =  parens expr
//	        <|> natural
//	        <?> "simple expression"
//
//	table   = [ [prefix "-" negate, prefix "+" id ]
//	          , [postfix "++" (+1)]
//	          , [binary "*" (*) AssocLeft, binary "/" (div) AssocLeft ]
//	          , [binary "+" (+) AssocLeft, binary "-" (-)   AssocLeft ]
//	          ]
//
//	binary  name fun assoc = Infix (do{ reservedOp name; return fun }) assoc
//	prefix  name fun       = Prefix (do{ reservedOp name; return fun })
//	postfix name fun       = Postfix (do{ reservedOp name; return fun })
func TestBuildExpressionParser_ByteState(t *testing.T) {
	applyPrefix := func(op interface{}) interface{} {
		return func(v interface{}) interface{} {
			switch op {
			case "-":
				return -v.(int64)
			case "+":
				return v
			case "--", "decr":
				return v.(int64) - 1
			case "++", "incr":
				return v.(int64) + 1
			default:
				panic("unreached")
			}
		}
	}
	applyPostfix := func(op interface{}) interface{} {
		return func(v interface{}) interface{} {
			switch op {
			case "++":
				return v.(int64) + 1
			case "--":
				return v.(int64) - 1
			default:
				panic("unreached")
			}
		}
	}
	applyBinary := func(op interface{}) interface{} {
		return func(x, y interface{}) interface{} {
			switch op {
			case "+", "add":
				return x.(int64) + y.(int64)
			case "-", "sub":
				return x.(int64) - y.(int64)
			case "*", "mul":
				return x.(int64) * y.(int64)
			case "/", "div":
				return x.(int64) / y.(int64)
			default:
				panic("unreached")
			}
		}
	}
	applyNature := func(v interface{}) interface{} {
		n, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	Int := Trim(Regex("\\d+"), Space)
	LP := Trim(Str("("), Space)
	RP := Trim(Str(")"), Space)

	parens := func(p Parser) Parser { return Between(LP, Trim(RP, Space), p) }
	reservedOp := func(s string) Parser { return Trim(Str(s), Space) }

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

	Term.Pattern = Label(Alt(parens(Expr), Int.Map(applyNature)), "expect simple expression")

	calc := func(s string) int64 {
		v, err := ExpectEof(Expr).Parse(NewState(s))
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
