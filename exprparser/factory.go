package exprparser

import . "github.com/goghcrow/parsec"

//data Operator u t = Postfix (Parsec String u (t -> t))
//				  | InfixL (Parsec String u (t -> t -> t)) -- Left associative
//				  | InfixR (Parsec String u (t -> t -> t)) -- Right associative

// Prefix、PostFix 必须 返回 func(interface{}) interface{}
// Infix 必须 返回 func(l, r interface{}) interface{}

func NewPrefix(p Parser) Operator {
	return Operator{
		OperKind: Prefix,
		Parser:   p,
	}
}

func NewPostfix(p Parser) Operator {
	return Operator{
		OperKind: Postfix,
		Parser:   p,
	}
}

func NewInfixL(p Parser) Operator {
	return Operator{
		OperKind: Infix,
		Assoc:    AssocLeft,
		Parser:   p,
	}
}

func NewInfixR(p Parser) Operator {
	return Operator{
		OperKind: Infix,
		Assoc:    AssocRight,
		Parser:   p,
	}
}

func NewInfixN(p Parser) Operator {
	return Operator{
		OperKind: Infix,
		Assoc:    AssocNone,
		Parser:   p,
	}
}
