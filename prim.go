package parsec

// ----------------------------------------------------------------
// Primitive General Parsers
// ----------------------------------------------------------------

//goland:noinspection GoUnusedGlobalVariable
var (
	Nil = Return(nil)
	Any = Satisfy(func(v interface{}) bool { return true }, "any")
	Eof = Label(Try(NotFollowedBy(Any)), "expect end of input")
)

func Satisfy(f func(interface{}) bool, expect string) Parser {
	return parser(func(s State) (interface{}, error) {
		loc := s.Save()
		nxt, ok := s.Next()
		if !ok {
			return nxt, Trap(loc, "expect `%s` actual end of input", expect)
		}
		if !f(nxt) {
			return nxt, Trap(loc, "expect `%s` actual `%s`", expect, Show(nxt))
		}
		return nxt, nil
	})
}
