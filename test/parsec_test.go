package test

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/goghcrow/parsec"
	. "github.com/goghcrow/parsec/charstate"
)

func TestComment(t *testing.T) {
	comment := Right(Str("<!--"), ManyTill(Regex(`[\w\s]+`), Str("-->")))
	v, err := comment.Parse(NewStrState("<!--hello world-->"))
	if err != nil {
		panic(err)
	}
	expect := "[hello world]"
	actual := fmt.Sprintf("%s", v)
	if expect != actual {
		t.Errorf("expect %s actual %s", expect, actual)
	}
}

func TestKeyword(t *testing.T) {
	let := Left(Str("let"), NotFollowedBy(Regex(`[\d\w]+`)))
	parse, err := let.Parse(NewStrState("let a = 1"))
	if err != nil {
		panic(err)
	}
	if parse.(string) != "let" {
		t.Errorf("expect let actual %s", parse)
	}

	parse, err = let.Parse(NewStrState("lets go"))
	if parse != nil {
		t.Errorf("expect error actual %s", parse)
	}
	expect := "unexpect `s` in pos 4 line 1 col 4"
	actual := err.Error()
	if actual != expect {
		t.Errorf("expect error %s actual %s", expect, actual)
	}
}

func TestCombinators(t *testing.T) {
	for _, tt := range []struct {
		name   string
		p      Parser
		s      State
		expect string
		error  string
		loc    *Loc
	}{
		{
			name:   "nil",
			p:      Nil,
			s:      NewStrState(""),
			expect: "<nil>",
		},
		{
			name:   "any",
			p:      Any,
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name:  "an!",
			p:     Any,
			s:     NewStrState(""),
			error: "expect `any` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "eof",
			p:      Eof,
			s:      NewStrState(""),
			expect: "<nil>",
		},
		{
			name:  "eof!",
			p:     Eof,
			s:     NewStrState("a"),
			error: "expect end of input in pos 1 line 1 col 1",
		},
		{
			name: "satisfy!",
			p: Satisfy(func(i interface{}) bool {
				return i.(rune) == 'a'
			}, "a"),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name: "satisfy!",
			p: Satisfy(func(i interface{}) bool {
				return i.(rune) == 'b'
			}, "b"),
			s:     NewStrState("a"),
			error: "expect `b` actual `a` in pos 1 line 1 col 1",
		},
		{
			name:   "return",
			p:      Return("a"),
			s:      NewStrState(""),
			expect: "a",
		},
		{
			name:  "fail",
			p:     Fail("fail"),
			s:     NewStrState(""),
			error: "fail in pos 1 line 1 col 1",
		},
		{
			name: "map",
			p: Map(Regex("\\d+"), func(v interface{}) interface{} {
				i, _ := strconv.ParseInt(v.(string), 10, 64)
				return i
			}),
			s:      NewStrState("42"),
			expect: "42",
		},
		{
			name: "bind",
			p: Bind(Str("a"), func(a interface{}) Parser {
				return Bind(Str("b"), func(b interface{}) Parser {
					return Return(a.(string) + b.(string))
				})
			}),
			s:      NewStrState("ab"),
			expect: "ab",
		},
		{
			name: "bind!",
			p: Bind(Str("a"), func(a interface{}) Parser {
				return Bind(Str("b"), func(b interface{}) Parser {
					return Return(a.(string) + b.(string))
				})
			}),
			s:     NewStrState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name: "seq",
			p: Seq(Str("a"), Str("b"), func(a, b interface{}) interface{} {
				return a.(string) + b.(string)
			}),
			s:      NewStrState("ab"),
			expect: "ab",
		},
		{
			name: "seq!",
			p: Seq(Str("a"), Str("b"), func(a, b interface{}) interface{} {
				return a.(string) + b.(string)
			}),
			s:     NewStrState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "list",
			p:      List(),
			s:      NewStrState(""),
			expect: "[]",
		},
		{
			name:   "list",
			p:      List(Str("a"), Str("b"), Str("c")),
			s:      NewStrState("abc"),
			expect: "[a b c]",
		},
		{
			name:  "list!",
			p:     List(Str("a"), Str("b"), Str("c")),
			s:     NewStrState("abd"),
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:   "try",
			p:      Try(Str("a")),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name:  "try!",
			p:     Str("abc"),
			s:     NewStrState("abd"),
			loc:   &Loc{Pos: 2},
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:  "try!",
			p:     Try(Str("abc")),
			s:     NewStrState("abd"),
			loc:   &Loc{Pos: 0}, // Try 恢复状态
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:   "either",
			p:      Either(Str("a"), Str("b")),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name:   "either",
			p:      Either(Str("a"), Str("b")),
			s:      NewStrState("b"),
			expect: "b",
		},
		{
			name:  "either!",
			p:     Either(Str("a"), Str("b")),
			s:     NewStrState("c"),
			error: "expect `b` actual `c` in pos 1 line 1 col 1", // 错误信息不对
		},
		{
			name:  "either!",
			p:     Label(Either(Str("a"), Str("b")), "expect a or b"), // 用 label 替换错误
			s:     NewStrState("c"),
			error: "expect a or b in pos 1 line 1 col 1",
		},
		{
			name:  "choice!",
			p:     Choice(),
			s:     NewStrState(""),
			error: "no choice in pos 1 line 1 col 1",
		},
		{
			name:   "choice",
			p:      Choice(Str("a")),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name:   "choice",
			p:      Choice(Str("a"), Str("b")),
			s:      NewStrState("b"),
			expect: "b",
		},
		{
			name:  "choice!",
			p:     Label(Choice(Str("a"), Str("b")), "expect a or b"), // 用 label 修改错误修心
			s:     NewStrState("c"),
			error: "expect a or b in pos 1 line 1 col 1",
		},
		{
			name:   "count",
			p:      Count(Str("a"), -1),
			s:      NewStrState("b"),
			expect: "[]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 0),
			s:      NewStrState("b"),
			expect: "[]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 1),
			s:      NewStrState("a"),
			expect: "[a]",
		},
		{
			name:   "count",
			p:      Count(Str("a"), 3),
			s:      NewStrState("aaa"),
			expect: "[a a a]",
		},
		{
			name:  "count!",
			p:     Count(Str("a"), 1),
			s:     NewStrState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:  "count!",
			p:     Count(Str("a"), 2),
			s:     NewStrState("ab"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:   "between",
			p:      Between(Str("("), Str(")"), Str("a")),
			s:      NewStrState("(a)"),
			expect: "a",
		},
		{
			name:  "between!",
			p:     Between(Str("("), Str(")"), Str("a")),
			s:     NewStrState("(b)"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:  "between!",
			p:     Between(Str("("), Str(")"), Str("a")),
			s:     NewStrState("(a]"),
			error: "expect `)` actual `]` in pos 3 line 1 col 3",
		},
		{
			name:   "mid",
			p:      Mid(Str("("), Str("a"), Str(")")),
			s:      NewStrState("(a)"),
			expect: "a",
		},
		{
			name:  "mid!",
			p:     Mid(Str("("), Str("a"), Str(")")),
			s:     NewStrState("(b)"),
			error: "expect `a` actual `b` in pos 2 line 1 col 2",
		},
		{
			name:  "mid!",
			p:     Mid(Str("("), Str("a"), Str(")")),
			s:     NewStrState("(a]"),
			error: "expect `)` actual `]` in pos 3 line 1 col 3",
		},
		{
			name:   "left",
			p:      Left(Str("a"), Str("b")),
			s:      NewStrState("ab"),
			expect: "a",
		},
		{
			name:  "left!",
			p:     Left(Str("a"), Str("b")),
			s:     NewStrState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "right",
			p:      Right(Str("a"), Str("b")),
			s:      NewStrState("ab"),
			expect: "b",
		},
		{
			name:  "right!",
			p:     Right(Str("a"), Str("b")),
			s:     NewStrState("ac"),
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewStrState("ba"),
			loc:    &Loc{Pos: 2},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 2},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewStrState("bab"),
			loc:    &Loc{Pos: 3},
			expect: "a",
		},
		{
			name:   "trim",
			p:      Trim(Str("a"), Str("b")),
			s:      NewStrState("bbabb"),
			loc:    &Loc{Pos: 5},
			expect: "a",
		},
		{
			name:  "trim!",
			p:     Trim(Str("a"), Str("b")),
			s:     NewStrState("ca"),
			loc:   &Loc{Pos: 0},
			error: "expect `a` actual `c` in pos 1 line 1 col 1",
		},
		{
			name:   "option",
			p:      Option(Str("a"), "x"),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name:   "option!",
			p:      Option(Str("a"), "x"),
			s:      NewStrState("b"),
			expect: "x",
		},
		{
			name:   "optional",
			p:      Optional(Str("a")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 1},
			expect: "<nil>",
		},
		{
			name:   "optional",
			p:      Optional(Str("a")),
			s:      NewStrState("b"),
			loc:    &Loc{Pos: 0},
			expect: "<nil>",
		},
		{
			name:   "optional",
			p:      Optional(Str("ab")),
			s:      NewStrState("ac"),
			loc:    &Loc{Pos: 0},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewStrState("b"),
			loc:    &Loc{Pos: 0},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 1},
			expect: "<nil>",
		},
		{
			name:   "skipMany",
			p:      SkipMany(Str("a")),
			s:      NewStrState("aab"),
			loc:    &Loc{Pos: 2},
			expect: "<nil>",
		},
		{
			name:  "skipMany1!",
			p:     SkipMany1(Str("a")),
			s:     NewStrState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:   "skipMany1",
			p:      SkipMany1(Str("a")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 1},
			expect: "<nil>",
		},
		{
			name:   "skipMany1",
			p:      SkipMany1(Str("a")),
			s:      NewStrState("aab"),
			loc:    &Loc{Pos: 2},
			expect: "<nil>",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewStrState("b"),
			expect: "[]",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewStrState("ab"),
			expect: "[a]",
		},
		{
			name:   "many",
			p:      Many(Str("a")),
			s:      NewStrState("aab"),
			expect: "[a a]",
		},
		{
			name:  "many1!",
			p:     Many1(Str("a")),
			s:     NewStrState("b"),
			error: "expect `a` actual `b` in pos 1 line 1 col 1",
		},
		{
			name:   "many1",
			p:      Many1(Str("a")),
			s:      NewStrState("ab"),
			expect: "[a]",
		},
		{
			name:   "many1",
			p:      Many1(Str("a")),
			s:      NewStrState("aab"),
			expect: "[a a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewStrState(""),
			expect: "[]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewStrState("a"),
			loc:    &Loc{Pos: 1},
			expect: "[a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 1}, // 剩余,
			expect: "[a]",
		},
		{
			name:   "seqBy",
			p:      SepBy(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 3},
			expect: "[a a]",
		},
		{
			name:  "seqBy1!",
			p:     SepBy1(Str("a"), Str(",")),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewStrState("a"),
			loc:    &Loc{Pos: 1},
			expect: "[a]",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 1}, // 剩余,
			expect: "[a]",
		},
		{
			name:   "seqBy1",
			p:      SepBy1(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 3},
			expect: "[a a]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewStrState(""),
			expect: "[]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewStrState("a"),
			loc:    &Loc{Pos: 0}, // 需要消耗 a,
			expect: "[]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:   "endBy",
			p:      EndBy(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:  "endBy1!",
			p:     EndBy1(Str("a"), Str(",")),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:  "endBy1!",
			p:     EndBy1(Str("a"), Str(",")),
			s:     NewStrState("a"),
			error: "expect `,` actual end of input in pos 2 line 1 col 2",
		},
		{
			name:   "endBy1",
			p:      EndBy1(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:   "endBy1",
			p:      EndBy1(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewStrState(""),
			expect: "[]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewStrState("a"),
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy",
			p:      SepEndBy(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 3},
			expect: "[a a]",
		},
		{
			name:  "sepEndBy1!",
			p:     SepEndBy1(Str("a"), Str(",")),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewStrState("a"),
			expect: "[a]",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewStrState("a,"),
			loc:    &Loc{Pos: 2},
			expect: "[a]",
		},
		{
			name:   "sepEndBy1",
			p:      SepEndBy1(Str("a"), Str(",")),
			s:      NewStrState("a,a"),
			loc:    &Loc{Pos: 3},
			expect: "[a a]",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState(""),
			expect: "x",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+"),
			loc:    &Loc{Pos: 1},
			expect: "a",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainl",
			p: Chainl(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+a+a"),
			expect: "[[a a] a]",
		},
		{
			name: "chainl1!",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+"),
			loc:    &Loc{Pos: 1},
			expect: "a",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainl1",
			p: Chainl1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+a+a"),
			expect: "[[a a] a]",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState(""),
			expect: "x",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+"),
			loc:    &Loc{Pos: 1},
			expect: "a",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainr",
			p: Chainr(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			}), "x"),
			s:      NewStrState("a+a+a"),
			expect: "[a [a a]]",
		},
		{
			name: "chainr1!",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a"),
			expect: "a",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+"),
			loc:    &Loc{Pos: 1},
			expect: "a",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+a"),
			expect: "[a a]",
		},
		{
			name: "chainr1",
			p: Chainr1(Str("a"), Str("+").Map(func(_ interface{}) interface{} {
				return func(l, r interface{}) interface{} {
					return []interface{}{l, r}
				}
			})),
			s:      NewStrState("a+a+a"),
			expect: "[a [a a]]",
		},
		{
			name:  "notFollowedBy!",
			p:     NotFollowedBy(Str("a")),
			s:     NewStrState("a"),
			loc:   &Loc{Pos: 1},
			error: "unexpect `a` in pos 1 line 1 col 1",
		},
		{
			name:   "notFollowedBy",
			p:      NotFollowedBy(Str("a")),
			s:      NewStrState("b"),
			loc:    &Loc{Pos: 0},
			expect: "<nil>",
		},
		{
			name:   "manyTill",
			p:      ManyTill(Str("a"), Str("b")),
			s:      NewStrState("b"),
			loc:    &Loc{Pos: 1}, // 消耗 b
			expect: "[]",
		},
		{
			name:   "manyTill",
			p:      ManyTill(Str("a"), Str("b")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 2}, // 消耗 b
			expect: "[a]",
		},
		{
			name:  "manyTill!",
			p:     ManyTill(Str("a"), Str("b")),
			s:     NewStrState(""),
			error: "expect `a` actual end of input in pos 1 line 1 col 1",
		},
		{
			name:   "lookAhead",
			p:      LookAhead(Str("a")),
			s:      NewStrState("ab"),
			loc:    &Loc{Pos: 0}, // 不消耗
			expect: "a",
		},
		{
			name:  "lookAhead!",
			p:     LookAhead(Str("ab")),
			s:     NewStrState("ac"),
			loc:   &Loc{Pos: 1}, // 失败仍旧消耗
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:  "lookAhead!",
			p:     LookAhead(Try(Str("ab"))),
			s:     NewStrState("ac"),
			loc:   &Loc{Pos: 0}, // 失败不消耗
			error: "expect `b` actual `c` in pos 2 line 1 col 2",
		},
		{
			name:   "expectEof",
			p:      ExpectEof(Str("a")),
			s:      NewStrState("a"),
			loc:    &Loc{Pos: 1},
			expect: "a",
		},
		{
			name: "expectEof!",
			p:    ExpectEof(Str("a")),
			s:    NewStrState("ab"),
			// loc:   &Loc{Pos: 1}, //2
			error: "expect end of input in pos 2 line 1 col 2",
		},
		{
			name:  "label!",
			p:     Label(Str("abc"), "expect x"),
			s:     NewStrState("abd"),
			loc:   &Loc{Pos: 2}, // 已经消费的不替换错误信息
			error: "expect `c` actual `d` in pos 3 line 1 col 3",
		},
		{
			name:  "label!",
			p:     Label(Try(Str("abc")), "expect x"), // 用 label 替换错误
			s:     NewStrState("abd"),
			loc:   &Loc{Pos: 0}, // 未消费的替换错误信息
			error: "expect x in pos 1 line 1 col 1",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.p.Parse(tt.s)
			if tt.loc != nil {
				actual := tt.s.Save().Pos
				if actual != tt.loc.Pos {
					t.Errorf("expect pos %d actual pos %d", tt.loc.Pos, actual)
				}
			}
			if err != nil {
				actual := err.Error()
				if actual != tt.error {
					t.Errorf("expect \"%s\" actual \"%s\"", tt.error, actual)
				}
			} else {
				r, ok := v.(rune)
				actual := fmt.Sprintf("%v", v)
				if ok {
					actual = string(r)
				}
				if actual != tt.expect {
					t.Errorf("expect \"%s\" actual \"%s\"", tt.expect, actual)
				}
			}
		})
	}
}
