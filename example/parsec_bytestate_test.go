package example

import (
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/states/bytestate"
)

func TestByteStateCombinators(t *testing.T) {
	for _, tt := range []struct {
		name   string
		p      Parser
		s      State
		expect string
		error  string
		pos    *Pos
	}{
		{
			name:   "nil",
			p:      Nil,
			s:      NewState(""),
			expect: "<nil>",
		},
		{
			name:   "any",
			p:      Any,
			s:      NewState("a"),
			expect: "a",
		},
		{
			name:  "an!",
			p:     Any,
			s:     NewState(""),
			error: "expect `any` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "eof",
			p:      Eof,
			s:      NewState(""),
			expect: "<nil>",
		},
		{
			name:  "eof!",
			p:     Eof,
			s:     NewState("a"),
			error: "expect end of input in pos 1 line 1 col 1",
		},
		{
			name: "satisfy!",
			p: Satisfy(func(i interface{}) bool {
				r, ok := i.(rune)
				if ok {
					return r == 'a'
				}
				b, ok := i.(byte)
				if ok {
					return b == 'a'
				}
				panic("not reached")
			}, "a"),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name: "satisfy!",
			p: Satisfy(func(i interface{}) bool {
				r, ok := i.(rune)
				if ok {
					return r == 'b'
				}
				b, ok := i.(byte)
				if ok {
					return b == 'b'
				}
				panic("not reached")
			}, "b"),
			s:     NewState("a"),
			error: "expect `b` actual `a` in pos 1 line 1 col 1",
		},
		{
			name:   "return",
			p:      Return("a"),
			s:      NewState(""),
			expect: "a",
		},
		{
			name:  "fail",
			p:     Fail("fail"),
			s:     NewState(""),
			error: "fail in pos 1 line 1 col 1",
		},
		{
			name: "map",
			p: Map(Regex("\\d+"), func(v interface{}) interface{} {
				i, _ := strconv.ParseInt(v.(string), 10, 64)
				return i
			}),
			s:      NewState("42"),
			expect: "42",
		},
		{
			name: "bind",
			p: Bind(Str("a"), func(a interface{}) Parser {
				return Bind(Str("b"), func(b interface{}) Parser {
					return Return(a.(string) + b.(string))
				})
			}),
			s:      NewState("ab"),
			expect: "ab",
		},
		{
			name: "bind!",
			p: Bind(Str("a"), func(a interface{}) Parser {
				return Bind(Str("b"), func(b interface{}) Parser {
					return Return(a.(string) + b.(string))
				})
			}),
			s:     NewState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name: "seq",
			p: Seq(Str("a"), Str("b"), func(a, b interface{}) interface{} {
				return a.(string) + b.(string)
			}),
			s:      NewState("ab"),
			expect: "ab",
		},
		{
			name: "seq!",
			p: Seq(Str("a"), Str("b"), func(a, b interface{}) interface{} {
				return a.(string) + b.(string)
			}),
			s:     NewState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "list",
			p:      List(),
			s:      NewState(""),
			expect: "[]",
		},
		{
			name:   "list",
			p:      List(Str("a"), Str("b"), Str("c")),
			s:      NewState("abc"),
			expect: "[a b c]",
		},
		{
			name:  "list!",
			p:     List(Str("a"), Str("b"), Str("c")),
			s:     NewState("abd"),
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:   "try",
			p:      Try(Str("a")),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name:  "try!",
			p:     Str("abc"),
			s:     NewState("abd"),
			pos:   &Pos{Idx: 2},
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:  "try!",
			p:     Try(Str("abc")),
			s:     NewState("abd"),
			pos:   &Pos{Idx: 0}, // Try 恢复状态
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:   "either",
			p:      Either(Str("a"), Str("b")),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name:   "either",
			p:      Either(Str("a"), Str("b")),
			s:      NewState("b"),
			expect: "b",
		},
		{
			name:  "either!",
			p:     Either(Str("a"), Str("b")),
			s:     NewState("c"),
			error: "expect `b` actual `c` in pos 1 line 1 col 1", // 错误信息不对
		},
		{
			name:  "either!",
			p:     Label(Either(Str("a"), Str("b")), "expect a or b"), // 用 label 替换错误
			s:     NewState("c"),
			error: "expect a or b in pos 1 line 1 col 1",
		},
		{
			name:  "choice!",
			p:     Choice(),
			s:     NewState(""),
			error: "no choice in pos 1 line 1 col 1",
		},
		{
			name:   "choice",
			p:      Choice(Str("a")),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name:   "choice",
			p:      Choice(Str("a"), Str("b")),
			s:      NewState("b"),
			expect: "b",
		},
		{
			name:  "choice!",
			p:     Label(Choice(Str("a"), Str("b")), "expect a or b"), // 用 label 修改错误修心
			s:     NewState("c"),
			error: "expect a or b in pos 1 line 1 col 1",
		},
		{
			name:   "count",
			p:      Count(Str("a"), -1),
			s:      NewState("b"),
			expect: "[]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 0),
			s:      NewState("b"),
			expect: "[]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 1),
			s:      NewState("a"),
			expect: "[a]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 3),
			s:      NewState("aaa"),
			expect: "[a a a]",
		},
		{
			name:  "count!",
			p:     Count(Str("a"), 1),
			s:     NewState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:  "count!",
			p:     Count(Str("a"), 2),
			s:     NewState("ab"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:   "between",
			p:      Between(Str("("), Str(")"), Str("a")),
			s:      NewState("(a)"),
			expect: "a",
		},
		{
			name:  "between!",
			p:     Between(Str("("), Str(")"), Str("a")),
			s:     NewState("(b)"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:  "between!",
			p:     Between(Str("("), Str(")"), Str("a")),
			s:     NewState("(a]"),
			error: "expect `)` actual `]` in pos 3 line 1 col 3",
		},
		{
			name:   "mid",
			p:      Mid(Str("("), Str("a"), Str(")")),
			s:      NewState("(a)"),
			expect: "a",
		},
		{
			name:  "mid!",
			p:     Mid(Str("("), Str("a"), Str(")")),
			s:     NewState("(b)"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:  "mid!",
			p:     Mid(Str("("), Str("a"), Str(")")),
			s:     NewState("(a]"),
			error: "expect `)` actual `]` in pos 3 line 1 col 3",
		},
		{
			name:   "left",
			p:      Left(Str("a"), Str("b")),
			s:      NewState("ab"),
			expect: "a",
		},
		{
			name:  "left!",
			p:     Left(Str("a"), Str("b")),
			s:     NewState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "right",
			p:      Right(Str("a"), Str("b")),
			s:      NewState("ab"),
			expect: "b",
		},
		{
			name:  "right!",
			p:     Right(Str("a"), Str("b")),
			s:     NewState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewState("ba"),
			pos:    &Pos{Idx: 2},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 2},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewState("bab"),
			pos:    &Pos{Idx: 3},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewState("bbabb"),
			pos:    &Pos{Idx: 5},
			expect: "a",
		},
		{
			name:  "trim!",
			p:     Trim(Str("a"), Str("b")),
			s:     NewState("ca"),
			pos:   &Pos{Idx: 0},
			error: "expect `a` actual `c` in pos 1 line 1 col 1",
		},
		{
			name:   "option",
			p:      Option(Str("a"), "x"),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name:   "option!",
			p:      Option(Str("a"), "x"),
			s:      NewState("b"),
			expect: "x",
		},
		{
			name:   "optional",
			p:      Optional(Str("a")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 1},
			expect: "<nil>",
		},
		{
			name:   "optional",
			p:      Optional(Str("a")),
			s:      NewState("b"),
			pos:    &Pos{Idx: 0},
			expect: "<nil>",
		},
		{
			name:   "optional",
			p:      Optional(Str("ab")),
			s:      NewState("ac"),
			pos:    &Pos{Idx: 0},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewState("b"),
			pos:    &Pos{Idx: 0},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 1},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewState("aab"),
			pos:    &Pos{Idx: 2},
			expect: "<nil>",
		},
		{
			name:  "skipMany1!",
			p:     SkipMany1(Str("a")),
			s:     NewState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:   "skipMany1",
			p:      SkipMany1(Str("a")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 1},
			expect: "<nil>",
		},
		{
			name:   "skipMany1",
			p:      SkipMany1(Str("a")),
			s:      NewState("aab"),
			pos:    &Pos{Idx: 2},
			expect: "<nil>",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewState("b"),
			expect: "[]",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewState("ab"),
			expect: "[a]",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewState("aab"),
			expect: "[a a]",
		},
		{
			name:  "many1!",
			p:     Many1(Str("a")),
			s:     NewState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:   "many1",
			p:      Many1(Str("a")),
			s:      NewState("ab"),
			expect: "[a]",
		},
		{
			name:   "many1",
			p:      Many1(Str("a")),
			s:      NewState("aab"),
			expect: "[a a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewState(""),
			expect: "[]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewState("a"),
			pos:    &Pos{Idx: 1},
			expect: "[a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 1}, // 剩余,
			expect: "[a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 3},
			expect: "[a a]",
		},
		{
			name:  "seqBy1!",
			p:     SepBy1(Str("a"), Str(",")),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewState("a"),
			pos:    &Pos{Idx: 1},
			expect: "[a]",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 1}, // 剩余,
			expect: "[a]",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 3},
			expect: "[a a]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewState(""),
			expect: "[]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewState("a"),
			pos:    &Pos{Idx: 0}, // 需要消耗 a,
			expect: "[]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:  "endBy1!",
			p:     EndBy1(Str("a"), Str(",")),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:  "endBy1!",
			p:     EndBy1(Str("a"), Str(",")),
			s:     NewState("a"),
			error: "expect `,` actual end of input in pos 2 line 1 col 2",
		},
		{
			name:   "endBy1",
			p:      EndBy1(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:   "endBy1",
			p:      EndBy1(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewState(""),
			expect: "[]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewState("a"),
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 3},
			expect: "[a a]",
		},
		{
			name:  "sepEndBy1!",
			p:     SepEndBy1(Str("a"), Str(",")),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewState("a"),
			expect: "[a]",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewState("a,"),
			pos:    &Pos{Idx: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewState("a,a"),
			pos:    &Pos{Idx: 3},
			expect: "[a a]",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState(""),
			expect: "x",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+"),
			pos:    &Pos{Idx: 1},
			expect: "a",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+a+a"),
			expect: "[[a a] a]",
		},
		{
			name: "chainl1!",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+"),
			pos:    &Pos{Idx: 1},
			expect: "a",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+a+a"),
			expect: "[[a a] a]",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState(""),
			expect: "x",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+"),
			pos:    &Pos{Idx: 1},
			expect: "a",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewState("a+a+a"),
			expect: "[a [a a]]",
		},
		{
			name: "chainr1!",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a"),
			expect: "a",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+"),
			pos:    &Pos{Idx: 1},
			expect: "a",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewState("a+a+a"),
			expect: "[a [a a]]",
		},
		{
			name:  "notFollowedBy!",
			p:     NotFollowedBy(Str("a")),
			s:     NewState("a"),
			pos:   &Pos{Idx: 1},
			error: "unexpect `a` in pos 1 line 1 col 1",
		},
		{
			name:   "notFollowedBy",
			p:      NotFollowedBy(Str("a")),
			s:      NewState("b"),
			pos:    &Pos{Idx: 0},
			expect: "<nil>",
		},
		{
			name:   "manyTill",
			p:      ManyTill(Str("a"), Str("b")),
			s:      NewState("b"),
			pos:    &Pos{Idx: 1}, // 消耗 b
			expect: "[]",
		},
		{
			name:   "manyTill",
			p:      ManyTill(Str("a"), Str("b")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 2}, // 消耗 b
			expect: "[a]",
		},
		{
			name:  "manyTill!",
			p:     ManyTill(Str("a"), Str("b")),
			s:     NewState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "lookAhead",
			p:      LookAhead(Str("a")),
			s:      NewState("ab"),
			pos:    &Pos{Idx: 0}, // 不消耗
			expect: "a",
		},
		{
			name:  "lookAhead!",
			p:     LookAhead(Str("ab")),
			s:     NewState("ac"),
			pos:   &Pos{Idx: 1}, // 失败仍旧消耗
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:  "lookAhead!",
			p:     LookAhead(Try(Str("ab"))),
			s:     NewState("ac"),
			pos:   &Pos{Idx: 0}, // 失败不消耗
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "expectEof",
			p:      ExpectEof(Str("a")),
			s:      NewState("a"),
			pos:    &Pos{Idx: 1},
			expect: "a",
		},
		{
			name: "expectEof!",
			p:    ExpectEof(Str("a")),
			s:    NewState("ab"),
			// pos:   &Pos{Idx: 1}, //2
			error: "expect end of input in pos 2 line 1 col 2",
		},
		{
			name:  "label!",
			p:     Label(Str("abc"), "expect x"),
			s:     NewState("abd"),
			pos:   &Pos{Idx: 2}, // 已经消费的不替换错误信息
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:  "label!",
			p:     Label(Try(Str("abc")), "expect x"), // 用 label 替换错误
			s:     NewState("abd"),
			pos:   &Pos{Idx: 0}, // 未消费的替换错误信息
			error: "expect x in pos 1 line 1 col 1",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.p.Parse(tt.s)
			if tt.pos != nil {
				actual := tt.s.Save().Idx
				if actual != tt.pos.Idx {
					t.Errorf("expect pos %d actual pos %d", tt.pos.Idx, actual)
				}
			}
			if err != nil {
				actual := err.Error()
				if actual != tt.error {
					t.Errorf("expect \"%s\" actual \"%s\"", tt.error, actual)
				}
			} else {
				actual := Show(v) // fmt.Sprintf("%v", v)
				if actual != tt.expect {
					t.Errorf("expect \"%s\" actual \"%s\"", tt.expect, actual)
				}
			}
		})
	}
}
