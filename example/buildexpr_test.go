package example

import (
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/charstate"
	. "github.com/goghcrow/parsec/expr"
)

// 解析前缀正负 后缀自增 和数学运算
// -- >  expr    = buildExpressionParser table term
// -- >          <?> "expression"
// -- >
// -- >  term    =  parens expr
// -- >          <|> natural
// -- >          <?> "simple expression"
// -- >
// -- >  table   = [ [prefix "-" negate, prefix "+" id ]
// -- >            , [postfix "++" (+1)]
// -- >            , [binary "*" (*) AssocLeft, binary "/" (div) AssocLeft ]
// -- >            , [binary "+" (+) AssocLeft, binary "-" (-)   AssocLeft ]
// -- >            ]
// -- >
// -- >  binary  name fun assoc = Infix (do{ reservedOp name; return fun }) assoc
// -- >  prefix  name fun       = Prefix (do{ reservedOp name; return fun })
// -- >  postfix name fun       = Postfix (do{ reservedOp name; return fun })
func TestBuildExpressionParser(t *testing.T) {
	applyPrefix := func(op interface{}) interface{} {
		return func(v interface{}) interface{} {
			switch op {
			case "-":
				return -v.(int64)
			case "+":
				return v
			//case nil:
			//	return v
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
			default:
				panic("unreached")
			}
		}
	}
	applyBinary := func(op interface{}) interface{} {
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

	reservedOp := func(s string) Parser {
		return Trim(Str(s), Space)
	}

	binary := func(name string, assoc Assoc) Operator {
		return Operator{
			OperKind: Binary,
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

	table := [][]Operator{
		{
			prefix("-"),
			prefix("+"),
		},
		{
			postfix("++"),
		},
		{
			binary("*", AssocLeft),
			binary("/", AssocLeft),
		},
		{
			binary("+", AssocLeft),
			binary("-", AssocLeft),
		},
	}

	Expr := NewRule()
	Term := NewRule()

	Expr.Pattern = Label(BuildExpressionParser(table, Term), "expect expression")
	Term.Pattern = Label(Alt(Parens(Expr), Regex("\\d+").Map(applyNature)), "expect simple expression")

	calc := func(s string) int64 {
		v, err := Expr.Parse(NewStrState(s))
		if err != nil {
			panic(err)
		}
		return v.(int64)
	}
	t.Log(calc("-2++"))
	t.Log(calc("1+2*(6--3)/3"))
	// t.Log(calc("1 + 2 * ( 6 - -3 ) + 3")) // todo ReservedOp
}
