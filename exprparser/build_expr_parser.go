package exprparser

import . "github.com/goghcrow/parsec"

// BuildExpressionParser 从操作符表(结合性&优先级)构建一个表达式 parser
// 📢: 每一层的优先级相同(结合性可能不同), 层之间优先级降序
// 注意:
// 1. 相同优先级的前缀后缀操作符只能出现一次 (e.g. 如果 - 是代表负数, 则不允许 --2)
// 2. 相同优先级的前缀后缀操作符优先左关联 (e.g. 如果 ++ 是后缀自增, 则 -2++ 等 -1, 而不是 -3)
// 具体实例参见 example/buildexpr_test.go
func BuildExpressionParser(opers OperatorTable, term Parser) Parser {
	p := term
	for _, ops := range opers {
		p = makeParser(p, ops)
	}
	return p
}

//goland:noinspection SpellCheckingInspection
func makeParser(term Parser, ops []Operator) Parser {
	rassoc, lassoc, nassoc, prefix, postfix := groupByOpers(ops)

	rassocOp := Choice(rassoc...)
	lassocOp := Choice(lassoc...)
	nassocOp := Choice(nassoc...)
	prefixOp := Label(Choice(prefix...), "") // 前后缀可选, 取消错误信息
	postfixOp := Label(Choice(postfix...), "")

	ambiguous := func(assoc Assoc, op Parser) Parser {
		return Try(Bind(op, func(_ interface{}) Parser {
			return Fail("ambiguous use of a %s associative operator", assoc)
		}))
	}

	ambiguousRight := ambiguous(AssocRight, rassocOp)
	ambiguousLeft := ambiguous(AssocLeft, lassocOp)
	ambiguousNon := ambiguous(AssocNone, nassocOp)

	id := func(a interface{}) interface{} { return a }
	postfixP := Either(postfixOp, Return(id))
	prefixP := Either(prefixOp, Return(id))

	termP := Bind(prefixP, func(pre interface{}) Parser {
		return Bind(term, func(x interface{}) Parser {
			return Bind(postfixP, func(post interface{}) Parser {
				postFn := post.(func(interface{}) interface{})
				preFn := pre.(func(interface{}) interface{})
				// 📢: 前缀优先于后缀
				return Return(postFn(preFn(x)))
			})
		})
	})

	// 这里逻辑 跟 Chainr 一致, 只多了歧义处理
	var rassocP, rassocP1 func(x interface{}) Parser
	rassocP = func(x interface{}) Parser {
		return Alt(
			Bind(rassocOp, func(f interface{}) Parser {
				return Bind(Bind(termP, rassocP1), func(y interface{}) Parser {
					fn := f.(func(x, y interface{}) interface{})
					return Return(fn(x, y))
				})
			}),
			ambiguousLeft,
			ambiguousNon,
		)
	}
	rassocP1 = func(x interface{}) Parser {
		return Either(rassocP(x), Return(x))
	}

	// 这里逻辑 跟 Chainl.chainl1Rest 一致, 只多了歧义处理
	var lassocP, lassocP1 func(x interface{}) Parser
	lassocP = func(x interface{}) Parser {
		return Alt(
			Bind(lassocOp, func(f interface{}) Parser {
				return Bind(termP, func(y interface{}) Parser {
					fn := f.(func(x, y interface{}) interface{})
					return lassocP1(fn(x, y))
				})
			}),
			ambiguousRight,
			ambiguousNon,
		)
	}
	lassocP1 = func(x interface{}) Parser {
		return Either(lassocP(x), Return(x))
	}

	nassocP := func(x interface{}) Parser {
		return Bind(nassocOp, func(f interface{}) Parser {
			return Bind(termP, func(y interface{}) Parser {
				// 与左结合的区别是, 不继续匹配
				fn := f.(func(x, y interface{}) interface{})
				return Alt(
					ambiguousRight,
					ambiguousLeft,
					ambiguousNon,
					Return(fn(x, y)),
				)
			})
		})
	}

	return Bind(termP, func(x interface{}) Parser {
		return Label(Alt(
			rassocP(x),
			lassocP(x),
			nassocP(x),
			Return(x),
		), "expect `operator`")
	})
}

//goland:noinspection SpellCheckingInspection
func groupByOpers(ops []Operator) (rassoc, lassoc, nassoc, prefix, postfix []Parser) {
	for _, op := range ops {
		switch op.OperKind {
		case Infix:
			switch op.Assoc {
			case AssocNone:
				nassoc = append(nassoc, op)
			case AssocLeft:
				lassoc = append(lassoc, op)
			case AssocRight:
				rassoc = append(rassoc, op)
			default:
				panic("unreached")
			}
		case Prefix:
			prefix = append(prefix, op)
		case Postfix:
			postfix = append(postfix, op)
		default:
			panic("unreached")
		}
	}
	return
}
