package example

import (
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/charstate"
)

// 处理公共前缀 parser 的回溯
func Cache(p Parser) Parser {
	return NewParser(func(s State) (interface{}, error) {
		c := getCache(s)
		it := c.get(p, s)
		if it == nil {
			v, err := p.Parse(s)
			c.put(p, s, v, err)
			return v, err
		} else {
			s.Restore(it.rest)
			return it.result, it.err
		}
	})
}

func TestCommonPrefix(t *testing.T) {
	a_ := Cache(Str("a"))
	b_ := Cache(Str("b"))
	p := Alt(Rep(a_, 10), Seq(Rep(a_, 9), b_, func(xs, x interface{}) interface{} {
		return append(xs.([]interface{}), x)
	}))
	_, err := p.Parse(NewState("aaaaaaaaab"))
	if err != nil {
		panic(err)
	}
}

func BenchmarkCache(b *testing.B) {
	a_ := Cache(Str("a"))
	b_ := Cache(Str("b"))
	p := Alt(Rep(a_, 10), Seq(Rep(a_, 9), b_, func(xs, x interface{}) interface{} {
		return append(xs.([]interface{}), x)
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(NewState("aaaaaaaaab"))
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkWithoutCache(b *testing.B) {
	a_ := Str("a")
	b_ := Str("b")
	p := Alt(Rep(a_, 10), Seq(Rep(a_, 9), b_, func(xs, x interface{}) interface{} {
		return append(xs.([]interface{}), x)
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(NewState("aaaaaaaaab"))
		if err != nil {
			panic(err)
		}
	}
}
