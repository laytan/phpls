package index

import (
	"log"

	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
)

type INodeTraverser struct {
	visitor.Null
	currentNamespace string
	currentPath      string
	nodes            chan<- *INode
}

func NewIndexTraverser() *INodeTraverser {
	return &INodeTraverser{}
}

func (t *INodeTraverser) EnterNode(node ast.Vertex) bool {
	switch typedNode := node.(type) {
	case *ast.StmtNamespace:
		t.currentNamespace = nodeident.Get(typedNode)
		if t.currentNamespace != "\\" {
			t.currentNamespace += "\\"
		}

		return true

	case *ast.StmtFunction, *ast.StmtClass, *ast.StmtInterface, *ast.StmtTrait:
		fqn := fqn.New(t.currentNamespace + nodeident.Get(node))
		t.nodes <- NewINode(fqn, t.currentPath, node)

		return false

	case *ast.ExprFunctionCall:
		// Index a function call to define() as a constant.
		if nodeident.Get(typedNode) == "define" {
			if len(typedNode.Args) == 0 {
				return false
			}

			firstArg, ok := typedNode.Args[0].(*ast.Argument)
			if !ok {
				return false
			}

			ident, ok := firstArg.Expr.(*ast.ScalarString)
			if !ok {
				log.Println("found define call without a string argument")
				return false
			}

			id := string(ident.Value[1 : len(ident.Value)-1])
			fqn := fqn.New("\\" + id)
			t.nodes <- &INode{
				FQN:        fqn,
				Path:       t.currentPath,
				Position:   typedNode.Position,
				Identifier: id,
				Kind:       ast.TypeStmtConstant,
			}
		}

		return false

	default:
		return true
	}
}

func (t *INodeTraverser) LeaveNode(node ast.Vertex) {
	if _, ok := node.(*ast.Root); ok {
		close(t.nodes)
	}
}

func (t *INodeTraverser) Reset(path string, nodes chan<- *INode) {
	t.currentNamespace = "\\"
	t.currentPath = path
	t.nodes = nodes
}
