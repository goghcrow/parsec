package bytestate

import (
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

	LitFloat = Regex(
		"(?:[+-]?(?:0|[1-9][0-9]*)(?:[.][0-9]+)+(?:[eE][-+]?[0-9]+)?)" +
			"|" +
			"(?:[+-]?(?:0|[1-9][0-9]*)(?:[.][0-9]+)?(?:[eE][-+]?[0-9]+)+)",
	)
	LitInt = Regex("(?:[+-]?0b(?:0|1[0-1]*))" +
		"|" +
		"(?:[+-]?0x(?:0|[1-9a-fA-F][0-9a-fA-F]*))" +
		"|" +
		"(?:[+-]?0o(?:0|[1-7][0-7]*))" +
		"|" +
		"(?:[+-]?(?:0|[1-9][0-9]*))",
	)
	LitStr = Regex("(?:\"(?:[^\"\\\\]*|\\\\[\"\\\\trnbf\\/]|\\\\u[0-9a-fA-F]{4})*\")" +
		"|" +
		"(?:`[^`]*`)",
	)
	Ident = Regex("[a-zA-Z\\p{L}_][a-zA-Z0-9\\p{L}_]*") // 支持 unicode
)

func OneOf(bytes string) Parser  { return ByteSatisfy(oneOf(bytes), "one of '"+bytes+"'") }
func NoneOf(bytes string) Parser { return ByteSatisfy(noneOf(bytes), "none of '"+bytes+"'") }

func AnyChar() Parser    { return ByteSatisfy(func(b byte) bool { return true }, "any byte") }
func Char(b byte) Parser { return ByteSatisfy(equals(b), string(b)) }

func ByteSatisfy(pred func(byte) bool, expect string) Parser {
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*ByteState)
		loc := s.Save()
		r, ok := s.NextIf(pred)
		if ok {
			return r, nil
		}
		return nil, s.trapExpect(loc, expect, r)
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
	patten := regexp.MustCompile("^" + reg)
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*ByteState)
		loc := s.Save()
		found := patten.FindString(string(s.seq[s.Pos:]))
		if found == "" {
			return nil, Trap(loc, "expect pattern '%s'", reg)
		} else {
			for _, b := range []byte(found) {
				s.move(b)
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
