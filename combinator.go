package parsec

// ----------------------------------------------------------------
// Parser Combinators
// ----------------------------------------------------------------

// tips:
// 如果期望失败不消耗 state, 套个 Try
// 被 Optional 包装的 Parser 永远成功, Either 或者 Choice 失效
// SepBy SepEndBy Chainl Chainr Many, Many1, SkipMany, SkipMany1 等传入的 p 如果不消耗 state 会 stackoverflow

// alias
//
//goland:noinspection GoUnusedGlobalVariable
var (
	Unit    = Return
	FlatMap = Bind
	Apply   = Map
	Alt     = Choice
	Skip    = Optional
	Rep     = Count
)

func Return(x interface{}) Parser {
	return parser(func(s State) (interface{}, error) {
		return x, nil
	})
}

// Fail 不消耗 state, 总是失败
func Fail(f string, a ...interface{}) Parser {
	return parser(func(s State) (interface{}, error) {
		return nil, Trap(s.Save(), f, a...)
	})
}

func Map(p Parser, f func(interface{}) interface{}) Parser {
	return parser(func(s State) (interface{}, error) {
		v, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		return f(v), nil
	})
}

func Bind(p Parser, f func(interface{}) Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		v, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		return f(v).Parse(s)
	})
}

func Seq(front, rear Parser, mapper func(x, y interface{}) interface{}) Parser {
	return Bind(front, func(x interface{}) Parser {
		return Bind(rear, func(y interface{}) Parser {
			return Return(mapper(x, y))
		})
	})
}

func List(ps ...Parser) Parser {
	if len(ps) == 0 {
		return Return([]interface{}{})
	}
	return Bind(ps[0], func(x interface{}) Parser {
		return Bind(List(ps[1:]...), func(xs interface{}) Parser {
			return Return(Cons(x, xs))
		})
	})
}

// Try 支持 lookaheadN
// 错误发生时不消耗 state, 其他跟 p 一样
func Try(p Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		v, err := p.Parse(s)
		if err == nil {
			return v, nil
		}
		s.Restore(pos)
		return nil, err
	})
}

// LookAhead peek p 的值
// 如果失败会消费 state, 如果不期望消费可以 LookAhead(Try(p))
func LookAhead(p Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		v, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		s.Restore(pos)
		return v, err
	})
}

func Either(a, b Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		v, err := Try(a).Parse(s)
		if err == nil {
			return v, nil
		}
		return b.Parse(s)
	})
}

// Choice 按顺序尝试 ps 直到成功, 返回成功的 p 的返回值
// 注意: 传入 Choice 的 parser 都应该被 Try 包装
// 方便起见, 内部已经处理, xs 除了最后一个都会自动 Try
// foldr (<|>) mzero ps
func Choice(xs ...Parser) Parser {
	if len(xs) == 0 {
		return Fail("no choice")
	}
	if len(xs) == 1 {
		return xs[0]
	}
	return Either(xs[0], Choice(xs[1:]...))
}

// Count 应用 p n 次, 返回 []any
func Count(p Parser, n int) Parser {
	if n <= 0 {
		return Return([]interface{}{})
	}
	return Seq(p, Count(p, n-1), Cons)
}

// Between 依次 parse open p close, 返回 p 的返回值
// do{ _ <- open; x <- p; _ <- close; return x }
// e.g. braces  = between (symbol "{") (symbol "}")
func Between(open, close, p Parser) Parser { return Right(open, Left(p, close)) }

func Mid(start, p, end Parser) Parser { return Between(start, end, p) }

func Left(l, r Parser) Parser {
	return Bind(l, func(v interface{}) Parser { return Right(r, Return(v)) })
}

// Right do{ _ <- l; r }
func Right(l, r Parser) Parser { return Bind(l, func(v interface{}) Parser { return r }) }

func Trim(p, cut Parser) Parser { return Mid(Many(cut), p, Many(cut)) }

// Option 尝试 p,失败不消耗 state, 成功返回 p 的返回值, 失败返回默认值 v
// p <|> return x
func Option(p Parser, x interface{}) Parser { return Either(p, Return(x)) }

// Optional 尝试应用 p, 成功则消耗 state, 丢弃返回值
// do{ _ <- p; return ()} <|> return ()
func Optional(p Parser) Parser { return Option(Right(p, Nil), nil) }

// SkipMany 应用 p >= 0 次, 跳过结果
// do{ _ <- many p; return ()} <|> return ()
func SkipMany(p Parser) Parser { return Skip(Many(p)) }

// SkipMany1 应用 p >= 1 次, 跳过结果
// 注意 Skip(Many1(p)) != SkipMany1(p)
// do{ _ <- p; skipMany p }
func SkipMany1(p Parser) Parser { return Right(p, SkipMany(p)) }

// Many 应用 p >= 0 次, 返回 []any
func Many(p Parser) Parser { return Option(Many1(p), []interface{}{}) }

// Many1 应用 p >= 1 次, 返回 []any
// do{ x <- p; xs <- many p; return (x:xs) }
// e.g. word  = many1 letter
func Many1(p Parser) Parser {
	return Bind(p, func(x interface{}) Parser {
		return Bind(Many(p), func(xs interface{}) Parser {
			return Return(Cons(x, xs))
		})
	})
}

// SepBy parse 被 sep 分隔的 >=0 个 p, 不以 seq 结尾, 返回 []any
// sepBy1 p sep <|> return []
func SepBy(p, sep Parser) Parser { return Option(SepBy1(p, sep), []interface{}{}) }

// SepBy1 parse 被 sep 分隔的 >=1 个 p, 不以 seq 结尾, 返回 []any
// do{ x <- p; xs <- many (sep >> p); return (x:xs) }
func SepBy1(p, sep Parser) Parser { return Seq(p, Many(Right(sep, p)), Cons) }

// EndBy parse 被 sep 分隔的 >= 0 个 p, seq 结尾, 返回 []any
// many (do{ x <- p; _ <- sep; return x })
func EndBy(end, sep Parser) Parser { return Many(Left(end, sep)) }

// EndBy1 parse 被 sep 分隔的 >= 1 个 p, seq 结尾, 返回 []any
// many1 (do{ x <- p; _ <- sep; return x })
func EndBy1(end, sep Parser) Parser { return Many1(Left(end, sep)) }

// SepEndBy parse 被 sep 分隔的 >= 0 个 p, 结尾的 seq 可选, 返回 []any
// sepEndBy1 p sep <|> return []
func SepEndBy(p, sep Parser) Parser { return Option(SepEndBy1(p, sep), []interface{}{}) }

// SepEndBy1 parse 被 sep 分隔的 >= 1 个 p, 结尾的 seq 可选, 返回 []any
// do{ x <- p ; do{ _ <- sep ; xs <- sepEndBy p sep ; return (x:xs)  } <|> return [x] }
func SepEndBy1(p, sep Parser) Parser { return Seq(p, Left(Many(Right(sep, p)), Optional(sep)), Cons) }

// Chainl 构造左结合双目运算符解析, 可以用来处理左递归文法
// op 必须返回 func(l interface {}, r interface {}) interface {}
// parse >=0 次被 op 分隔的 p, 返回左结合调用 f 得到的值, 如果 0 次, 返回默认值 x
// chainr1 p op <|> return x
func Chainl(p, op Parser, x interface{}) Parser {
	return Option(Chainl1(p, op), x)
}

// Chainl1 构造左结合双目运算符解析, 可以用来处理左递归文法
// op 必须返回 func(l interface {}, r interface {}) interface {}
// parse >=1 次被 op 分隔的 p, 返回左结合调用 f 得到的值
// do { x <- p; rest x } where rest x = do{ f <- op ; y <- p ; rest (f x y) } <|> return x
func Chainl1(p, op Parser) Parser {
	var chainl1Rest func(lval interface{}) Parser
	chainl1Rest = func(lval interface{}) Parser {
		opv := Bind(op, func(f interface{}) Parser {
			// 左结合: 优先匹配 p(即 term), 然后递归的匹配 term op
			return Bind(p, func(rval interface{}) Parser {
				fn := f.(func(x, y interface{}) interface{})
				return chainl1Rest(fn(lval, rval))
			})
		})
		return Option(opv, lval)
	}
	return Bind(p, chainl1Rest)
}

// Chainr 构造右结合双目运算符解析
// op 必须返回 func(l interface {}, r interface {}) interface {}
// parse >=0 次被 op 分隔的 p, 返回右结合调用 f 得到的值, 如果 0 次, 返回默认值 x
// chainr1 p op <|> return x
func Chainr(p, op Parser, x interface{}) Parser {
	return Option(Chainr1(p, op), x)
}

// Chainr1 构造右结合双目运算符解析
// op 必须返回 func(l interface {}, r interface {}) interface {}
// parse >=1 次被 op 分隔的 p, 返回右结合调用 f 得到的值
// do{ x <- p; rest x } where rest x = do{ f <- op ; y <- scan ; return (f x y)  } <|> return x
func Chainr1(p, op Parser) Parser {
	return Bind(p, func(lval interface{}) Parser {
		seq := Bind(op, func(f interface{}) Parser {
			// 右结合就是自然地递归下降
			return Bind(Chainr1(p, op), func(rval interface{}) Parser {
				fn := f.(func(x, y interface{}) interface{})
				return Return(fn(lval, rval))
			})
		})
		return Option(seq, lval)
	})
}

// ===== Tricky Combinators =====

// NotFollowedBy 只有在 p 匹配失败时才成功, 不消耗 state, 可以用来实现最长匹配
// 例如，在识别 keywords（e.g. let）时，需要确保关键词后面没有合法的标识符(e.g. lets)
// 可以写成 let := Left(Str("let"), NotFollowedBy(Regex(`[\d\w]+`)))
// try (do{ c <- try p; unexpected (show c) } <|> return () )
func NotFollowedBy(p Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		c, err := p.Parse(s)
		if err == nil {
			return nil, Trap(pos, "unexpect `%s`", Show(c))
		}
		s.Restore(pos)
		return nil, nil
	})
}

// ManyTill 应用 p>=0 次, 直到 end 成功, 返回 p 匹配的列表[]any
// 可以用来实现注释: do{ string "<!--" ; manyTill anyChar (try (string "-->")) }
// do{ _ <- end; return [] } <|> do{ x <- p; xs <- manyTill; return (x:xs) }
func ManyTill(p, end Parser) Parser {
	return Either(
		Right(Try(end), Return([]interface{}{})),
		Bind(p, func(x interface{}) Parser {
			return Bind(ManyTill(p, end), func(xs interface{}) Parser {
				return Return(Cons(x, xs))
			})
		}),
	)
}

func ExpectEof(p Parser) Parser { return Left(p, Eof) }

// ===== Debug Combinators =====

// Label p 失败且未消费 state, 会用 msg 替换错误信息, 其他行为与 P 相同
// 通常用在一组 alternatives 最后, 展示更高层的信息, 而不是 alt 最后的错误信息
func Label(p Parser, fmt string, a ...interface{}) Parser {
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		v, err := p.Parse(s)
		if err != nil {
			if pos == s.Save() {
				return nil, Trap(pos, fmt, a...)
			} else {
				return nil, err
			}
		}
		return v, nil
	})
}

// Trace 可以用来调试 parser
// 回调函数参数: p error, p 返回值, 剩余的状态
func Trace(p Parser, trace func(error, interface{}, []interface{})) Parser {
	return parser(func(s State) (interface{}, error) {
		v, err := p.Parse(s)
		if err != nil {
			trace(err, v, nil)
			return nil, err
		}
		pos := s.Save()
		xs, _ := Left(Many(Any), Eof).Parse(s)
		trace(nil, v, xs.([]interface{}))
		s.Restore(pos)
		return v, nil
	})
}
