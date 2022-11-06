package chars

import "github.com/goghcrow/parsec"

// todo https://hackage.haskell.org/package/parsec-3.1.15.1/docs/src/Text.Parsec.Token.html#local-6989586621679060596

// lexeme p
//        = do{ x <- p; whiteSpace; return x  }

func Lexeme(p parsec.Parser) parsec.Parser {
	return parsec.Bind(p, func(x interface{}) parsec.Parser {
		return parsec.Bind(WhiteSpace(), func(_ interface{}) parsec.Parser {
			return parsec.Return(x)
		})
	})
}

// whiteSpace
//        | noLine && noMulti  = skipMany (simpleSpace <?> "")
//        | noLine             = skipMany (simpleSpace <|> multiLineComment <?> "")
//        | noMulti            = skipMany (simpleSpace <|> oneLineComment <?> "")
//        | otherwise          = skipMany (simpleSpace <|> oneLineComment <|> multiLineComment <?> "")
//        where
//          noLine  = null (commentLine languageDef)
//          noMulti = null (commentStart languageDef)

func WhiteSpace() parsec.Parser {
	return Spaces // todo
}

// todo
func ReservedOp(name string) parsec.Parser {
	//       lexeme $ try $
	//        do{ _ <- string name
	//          ; notFollowedBy (opLetter languageDef) <?> ("end of " ++ show name)
	//          }
	// todo...
	return Lexeme(parsec.Try(parsec.Bind(Str(name), func(_ interface{}) parsec.Parser {
		return parsec.Label(parsec.NotFollowedBy(Regex(`\w\d+`)), "end of `%s`", name)
	})))

}

func Symbol(name string) parsec.Parser {
	return Lexeme(Str(name)) // todo
}

func Parens(p parsec.Parser) parsec.Parser {
	return parsec.Mid(Symbol("("), p, Symbol(")"))
}
