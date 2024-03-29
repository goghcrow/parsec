package bytestate

import (
	"github.com/goghcrow/lexer"
	"regexp"
	"strings"

	. "github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Primitive Byte Parsers
// ----------------------------------------------------------------

//goland:noinspection GoUnusedGlobalVariable
var (
	Upper  = ByteSatisfy(func(b byte) bool { return b >= 'A' && b <= 'Z' }, "upper")
	Lower  = ByteSatisfy(func(b byte) bool { return b >= 'a' && b <= 'z' }, "lower")
	Letter = Alt(Upper, Lower)

	Digit  = ByteSatisfy(func(b byte) bool { return b >= '0' && b <= '9' }, "digit")
	Number = ByteSatisfy(func(b byte) bool { return b >= '0' && b <= '9' }, "number")

	Space  = OneOf(" \t\r\n")
	Spaces = SkipMany(Space)

	Letters  = Many1(Letter)
	Digits   = Many1(Digit)
	AlphaNum = Either(Letter, Digit)

	Numbers   = Many1(Number)
	AlphaNums = Many1(AlphaNum)

	NewLine   = OneOf("\n")
	Crlf      = Str("\r\n")
	EndOfLine = Either(NewLine, Crlf)

	Tab = Char('\t')

	LitFloat = Regex(lexer.RegFloat)
	LitInt   = Regex(lexer.RegInt)
	LitNum   = Either(LitFloat, LitInt)
	LitStr   = Regex(lexer.RegStr)
	Ident    = Regex(lexer.RegIdent) // 支持 unicode
)

func OneOf(bytes string) Parser  { return ByteSatisfy(oneOf(bytes), "one of '"+bytes+"'") }
func NoneOf(bytes string) Parser { return ByteSatisfy(noneOf(bytes), "none of '"+bytes+"'") }

func AnyChar() Parser    { return ByteSatisfy(func(b byte) bool { return true }, "any byte") }
func Char(b byte) Parser { return ByteSatisfy(equals(b), string(b)) }

func ByteSatisfy(pred func(byte) bool, expect string) Parser {
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*ByteState)
		pos := s.Save()
		r, ok := s.NextIf(pred)
		if ok {
			return r, nil
		}
		return nil, s.trapExpect(pos, expect, r)
	})
}

func Str(str string) Parser {
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*ByteState)
		for _, c := range []byte(str) {
			r, ok := s.NextIf(func(b byte) bool { return b == c })
			if !ok {
				return nil, s.trapExpect(s.Save(), string(c), r)
			}
		}
		return str, nil
	})
}

func Regex(reg string) Parser {
	patten := regexp.MustCompile("^(?:" + reg + ")")
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*ByteState)
		pos := s.Save()
		found := patten.FindString(string(s.seq[s.Idx:]))
		if found == "" {
			return nil, Trap(pos, "expect pattern '%s'", reg)
		} else {
			for _, b := range []byte(found) {
				s.forward(b)
			}
			return found, nil
		}
	})
}

// ----------------------------------------------------------------
// Util
// ----------------------------------------------------------------

type bytePred func(byte) bool

func indexOf(s string, b byte) int { return strings.IndexByte(s, b) }
func constTrue(b byte) bool        { return true }
func equals(a byte) bytePred       { return func(b byte) bool { return a == b } }
func oneOf(bytes string) bytePred  { return func(r byte) bool { return indexOf(bytes, r) >= 0 } }
func noneOf(bytes string) bytePred { return func(r byte) bool { return indexOf(bytes, r) < 0 } }
