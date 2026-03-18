package parse

import "github.com/expr-lang/expr/vm"

type Node interface {
	node()
}

type TextNode struct {
	Text string
}

func (TextNode) node() {}

type IfNode struct {
	RawExpr string
	Program *vm.Program
	Body    []Node
}

func (*IfNode) node() {}

type WhereNode struct {
	Body []Node
}

func (*WhereNode) node() {}

type SetNode struct {
	Body []Node
}

func (*SetNode) node() {}
