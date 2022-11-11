//go:build parsec_iter
// +build parsec_iter

package legacy

import . "github.com/goghcrow/parsec"

// 一些展开或者迭代的版本, 性能会好点

var Eof = parser(func(s State) (interface{}, error) {
	pos := s.Save()
	r, ok := s.Next()
	if ok {
		return r, Trap(pos, "expect end of input")
	}
	return nil, nil
})

func List(ps ...Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		xs := make([]interface{}, len(ps))
		for i, p := range ps {
			v, err := p.Parse(s)
			if err != nil {
				return xs, err
			}
			xs[i] = v
		}
		return xs, nil
	})
}

func Choice(ps ...Parser) Parser {
	if len(ps) == 0 {
		return Fail("no choice")
	}
	if len(ps) == 1 {
		return ps[0]
	}
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		for _, p := range ps {
			v, err := Try(p).Parse(s)
			if err == nil {
				return v, nil
			}
		}
		return nil, Trap(pos, "no choice")
	})
}

func Count(p Parser, n int) Parser {
	if n <= 0 {
		return Return([]interface{}{})
	}
	return parser(func(s State) (interface{}, error) {
		xs := make([]interface{}, n)
		for i := 0; i < n; i++ {
			v, err := p.Parse(s)
			if err != nil {
				return xs, err
			}
			xs[i] = v
		}
		return xs, nil
	})
}

func Right(skip, p Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		_, err := skip.Parse(s)
		if err != nil {
			return nil, err
		}
		return p.Parse(s)
	})
}

func Optional(p Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		_, _ = Try(p).Parse(s)
		return nil, nil
	})
}

//func Many(p Parser) Parser {
//	thunk := parser(func(s State) (interface{}, error) { return Many_(p).Parse(s) })
//	return Option(Seq_(p, thunk, Cons), []interface{}{})
//}
//func Many1(p Parser) Parser { return Seq2(p, Many(p), Cons) }

func Many(p Parser) Parser {
	try := Try(p)
	return parser(func(s State) (interface{}, error) {
		var xs []interface{}
		for {
			x, err := try.Parse(s)
			if err != nil {
				return xs, nil
			}
			xs = append(xs, x)
		}
	})
}

func Many1(p Parser) Parser {
	try := Try(p)
	return parser(func(s State) (interface{}, error) {
		x, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		xs := []interface{}{x}
		for {
			x, err = try.Parse(s)
			if err != nil {
				return xs, nil
			}
			xs = append(xs, x)
		}
	})
}

func SepBy(p, sep Parser) Parser {
	sepp := Try(Right(sep, p))
	return parser(func(s State) (interface{}, error) {
		var xs []interface{}
		x, err := p.Parse(s)
		if err != nil {
			return xs, nil
		}
		xs = append(xs, x)
		for {
			x, err = sepp.Parse(s)
			if err != nil {
				return xs, nil
			}
			xs = append(xs, x)
		}
	})
}

func SepBy1(p, sep Parser) Parser {
	sepp := Try(Right(sep, p))
	return parser(func(s State) (interface{}, error) {
		var xs []interface{}
		x, err := p.Parse(s)
		if err != nil {
			return nil, err
		}
		xs = append(xs, x)
		for {
			x, err = sepp.Parse(s)
			if err != nil {
				return xs, nil
			}
			xs = append(xs, x)
		}
	})
}

//func SepBy1(p, sep Parser) Parser {
//	return Bind(p, func(x interface{}) Parser {
//		return Bind(Many(Right(sep, p)), func(xs interface{}) Parser {
//			return Return(Cons(x, xs))
//		})
//	})
//}

//func Skip1(p Parser) Parser { return p.Map(func(v interface{}) interface{} { return nil }) }

func ManyTill(p, end Parser) Parser {
	return parser(func(s State) (interface{}, error) {
		pos := s.Save()
		var xs []interface{}
		for {
			_, err := end.Parse(s)
			if err == nil {
				return xs, nil
			}

			x, err := p.Parse(s)
			if err != nil {
				s.Restore(pos)
				return nil, err
			}
			xs = append(xs, x)
		}
	})
}

// 这个需要 rune 和 byte 两个版本
//func Str(str string) Parser {
//	return parser(func(s_ State) (interface{}, error) {
//		s := s_.(*StrState)
//		pos := s.Save()
//		cnt := utf8.RuneCountInString(str)
//		if len(s.seq) < s.Idx+cnt {
//			return nil, Trap(pos, "expect `%s` actual end of input", str)
//		}
//		if str == string(s.seq[s.Idx:s.Idx+cnt]) {
//			for _, r := range []rune(str) {
//				s.move(r)
//			}
//			return str, nil
//		} else {
//			return nil, Trap(pos, "expect `%s`", str)
//		}
//	})
//}
