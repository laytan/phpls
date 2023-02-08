package index

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/symbol"
)

type IndexTraverser struct {
	index Index

	currentNamespace string
	currentPath      string
	nodes            chan<- *IndexNode
}

func NewIndexTraverser() *IndexTraverser {
	return &IndexTraverser{}
}

func (t *IndexTraverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			t.currentNamespace = "\\" + typedNode.NamespaceName.Value + "\\"
		}

		return true

	case *ir.FunctionStmt, *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		sym := symbol.New(node)
		fqn := fqn.New(t.currentNamespace + sym.Identifier())
		t.nodes <- NewIndexNode(fqn, t.currentPath, sym)

		return false

	case *ir.FunctionCallExpr:
		if fn, ok := typedNode.Function.(*ir.Name); ok && fn.Value == "define" {
			sym := symbol.NewGlobalConstant(typedNode)
			fqn := fqn.New("\\" + sym.Identifier())
			t.nodes <- NewIndexNode(fqn, t.currentPath, sym)
		}

		return false

	default:
		return true
	}
}

func (t *IndexTraverser) LeaveNode(node ir.Node) {
	if _, ok := node.(*ir.Root); ok {
		close(t.nodes)
	}
}

func (t *IndexTraverser) Reset(path string, nodes chan<- *IndexNode) {
	t.currentNamespace = "\\"
	t.currentPath = path
	t.nodes = nodes
}
