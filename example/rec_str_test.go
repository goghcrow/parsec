package example

import (
	. "github.com/goghcrow/parsec/charstate"
	"testing"

	. "github.com/goghcrow/parsec"
)

// -- >  expr    = term   `chainl1` addop
// -- >  term    = factor `chainl1` mulop
// -- >  factor  = parens expr <|> integer
// -- >
// -- >  mulop   =   do{ symbol "*"; return (*)   }
// -- >          <|> do{ symbol "/"; return (div) }
// -- >
// -- >  addop   =   do{ symbol "+"; return (+) }
// -- >          <|> do{ symbol "-"; return (-) }
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
		n, _ := parseInt(v.(string))
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
		v, err := ExpectEof(Expr).Parse(NewStrState(s))
		if err != nil {
			panic(err)
		}
		return v.(int64)
	}

	t.Log(calc("1 + 2 * ( 6 - 3 ) + 3"))
}
