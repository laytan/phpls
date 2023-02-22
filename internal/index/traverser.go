package index

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
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
		fqn := fqn.New(t.currentNamespace + nodeident.Get(node))
		t.nodes <- NewINode(fqn, t.currentPath, node)

		return false

	case *ir.FunctionCallExpr:
		// Index a function call to define() as a constant.

		if fn, ok := typedNode.Function.(*ir.Name); ok && fn.Value == "define" {
			if len(typedNode.Args) == 0 {
				return false
			}

			firstArg, ok := typedNode.Args[0].(*ir.Argument)
			if !ok {
				return false
			}

			ident, ok := firstArg.Expr.(*ir.String)
			if !ok {
				log.Println("found define call without a string argument")
				return false
			}

			fqn := fqn.New("\\" + ident.Value)
			t.nodes <- &INode{
				FQN:        fqn,
				Path:       t.currentPath,
				Position:   typedNode.Position,
				Identifier: ident.Value,
				Kind:       ir.KindConstantStmt,
			}
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
