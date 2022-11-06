package parsec

// ----------------------------------------------------------------
// Factory
// ----------------------------------------------------------------

func NewRule() *SyntaxRule                                  { return &SyntaxRule{} }
func NewParser(p func(s State) (interface{}, error)) Parser { return parser(p) }

// ----------------------------------------------------------------
// Parser
// ----------------------------------------------------------------

type Parser interface {
	Parse(s State) (interface{}, error)
	Map(mapper func(v interface{}) interface{}) Parser
	FlatMap(f func(interface{}) Parser) Parser
}

type SyntaxRule struct {
	Pattern Parser
}

func (r *SyntaxRule) Parse(s State) (interface{}, error)           { return r.Pattern.Parse(s) }
func (r *SyntaxRule) Map(f func(v interface{}) interface{}) Parser { return Map(r, f) }
func (r *SyntaxRule) FlatMap(f func(v interface{}) Parser) Parser  { return FlatMap(r, f) }

// ----------------------------------------------------------------
// Parser Impl
// ----------------------------------------------------------------

type parser func(s State) (interface{}, error)

func (p parser) Parse(s State) (interface{}, error)           { return p(s) }
func (p parser) Map(f func(v interface{}) interface{}) Parser { return Map(p, f) }
func (p parser) FlatMap(f func(v interface{}) Parser) Parser  { return FlatMap(p, f) }
