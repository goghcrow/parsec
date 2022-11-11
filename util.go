package parsec

import (
	"fmt"

	"github.com/goghcrow/go-parsec/lexer"
)

func Cons(x, xs interface{}) interface{} {
	return append([]interface{}{x}, xs.([]interface{})...)
}

func Trap(pos Pos, f string, a ...interface{}) Error {
	return Error{pos, fmt.Sprintf(f, a...)}
}

func Show(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case rune:
		return string(v)
	case *lexer.Token:
		return v.String()
	case byte:
		return string([]byte{v})
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", i)
	}
}
