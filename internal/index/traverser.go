package index

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/symbol"
)

type INodeTraverser struct {
	currentNamespace string
	currentPath      string
	nodes            chan<- *INode
}

func NewIndexTraverser() *INodeTraverser {
	return &INodeTraverser{}
}

func (t *INodeTraverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			t.currentNamespace = "\\" + typedNode.NamespaceName.Value + "\\"
		}

		return true

	case *ir.FunctionStmt, *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		sym := symbol.New(node)
		fqn := fqn.New(t.currentNamespace + sym.Identifier())
		t.nodes <- NewINode(fqn, t.currentPath, sym)

		return false

	case *ir.FunctionCallExpr:
		if fn, ok := typedNode.Function.(*ir.Name); ok && fn.Value == "define" {
			sym := symbol.NewGlobalConstant(typedNode)
			fqn := fqn.New("\\" + sym.Identifier())
			t.nodes <- NewINode(fqn, t.currentPath, sym)
		}

		return false

	default:
		return true
	}
}

func (t *INodeTraverser) LeaveNode(node ir.Node) {
	if _, ok := node.(*ir.Root); ok {
		close(t.nodes)
	}
}

func (t *INodeTraverser) Reset(path string, nodes chan<- *INode) {
	t.currentNamespace = "\\"
	t.currentPath = path
	t.nodes = nodes
}
