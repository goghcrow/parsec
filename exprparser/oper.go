package exprparser

import (
	"sort"

	. "github.com/goghcrow/parsec"
)

type Assoc int

const (
	AssocNone = iota
	AssocLeft
	AssocRight
)

func (a Assoc) String() string { return [...]string{"none", "left", "right"}[a] }

type OperKind int

const (
	Prefix = iota
	Postfix
	Infix
)

type Operator struct {
	OperKind

	// 只有 Infix 需要, 前后缀无结合性
	Assoc

	// Prefix、PostFix 必须 返回 func(interface{}) interface{}
	// Infix 必须 返回 func(l, r interface{}) interface{}
	Parser

	// 优先级, 使用 BuildOperatorTable 时需要, 使用字面量构造不需要
	Prec float32
}

// OperatorTable
// 📢: 每一层的优先级相同(结合性可能不同), 层之间优先级降序
type OperatorTable [][]Operator

func BuildOperatorTable(ops []Operator) OperatorTable {
	group := map[float32][]Operator{}
	var precs []float32
	for _, op := range ops {
		if group[op.Prec] == nil {
			group[op.Prec] = []Operator{}
			precs = append(precs, op.Prec)
		}
		group[op.Prec] = append(group[op.Prec], op)
	}
	sort.SliceStable(precs, func(i, j int) bool { return precs[i] > precs[j] })
	tbl := make([][]Operator, len(precs))
	for i, prec := range precs {
		tbl[i] = group[prec]
	}
	return tbl
}
