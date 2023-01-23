package lisp

import (
	"fmt"
	"strconv"
	"strings"
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
