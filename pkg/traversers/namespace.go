package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodescopes"
)

func NewNamespace(row int) *Namespace {
	return &Namespace{
		row: row,
	}
}

func NewNamespaceFromNode(node ast.Vertex) *Namespace {
	return NewNamespace(node.GetPosition().StartLine)
}

type Namespace struct {
	visitor.Null
	row    int
	Result *ast.StmtNamespace
}

func (n *Namespace) EnterNode(node ast.Vertex) bool {
	// Stop after given row.
	if node.GetPosition().StartLine >= int(n.row) {
		return false
	}

	// Don't go into scopes, namespace is always top level.
	return !nodescopes.IsScope(node.GetType())
}

func (n *Namespace) StmtNamespace(node *ast.StmtNamespace) {
	n.Result = node
}
