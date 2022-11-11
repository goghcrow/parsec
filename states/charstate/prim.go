package charstate

import (
	"regexp"
	"strings"
	"unicode"

	. "github.com/goghcrow/parsec"
)

// ----------------------------------------------------------------
// Primitive Rune Parsers
// ----------------------------------------------------------------

//goland:noinspection GoUnusedGlobalVariable
var (
	Upper  = CharSatisfy(unicode.IsUpper, "upper")
	Lower  = CharSatisfy(unicode.IsLower, "lower")
	Letter = CharSatisfy(unicode.IsLetter, "letter")
	Digit  = CharSatisfy(unicode.IsDigit, "digit")
	Number = CharSatisfy(unicode.IsNumber, "number")

	Space  = CharSatisfy(unicode.IsSpace, "space")
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

func OneOf(runes string) Parser  { return CharSatisfy(oneOf(runes), "one of '"+runes+"'") }
func NoneOf(runes string) Parser { return CharSatisfy(noneOf(runes), "none of '"+runes+"'") }

func AnyChar() Parser    { return CharSatisfy(func(r rune) bool { return true }, "any rune") }
func Char(r rune) Parser { return CharSatisfy(equals(r), string(r)) }

func CharSatisfy(pred func(rune) bool, expect string) Parser {
	return NewParser(func(s_ State) (interface{}, error) {
		s := s_.(*CharState)
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
		s := s_.(*CharState)
		for _, c := range str {
			r, ok := s.NextIf(func(r rune) bool { return r == c })
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
		s := s_.(*CharState)
		pos := s.Save()
		found := patten.FindString(string(s.seq[s.Idx:]))
		if found == "" {
			return nil, Trap(pos, "expect pattern '%s'", reg)
		} else {
			for _, r := range found {
				s.move(r)
			}
			return found, nil
		}
	})
}

// ----------------------------------------------------------------
// Util
// ----------------------------------------------------------------

type runePred func(rune) bool

func indexOf(s string, r rune) int { return strings.IndexRune(s, r) }
func constTrue(r rune) bool        { return true }
func equals(a rune) runePred       { return func(b rune) bool { return a == b } }
func oneOf(runes string) runePred  { return func(r rune) bool { return indexOf(runes, r) >= 0 } }
func noneOf(runes string) runePred { return func(r rune) bool { return indexOf(runes, r) < 0 } }
