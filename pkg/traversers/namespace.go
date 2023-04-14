package traversers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodescopes"
)

func NewNamespace(row int) *Namespace {
	return &Namespace{
		row: row,
	}
}

func NewNamespaceFirstResult() *Namespace {
	return &Namespace{
		firstResult: true,
	}
}

type Namespace struct {
	visitor.Null
	row         int
	firstResult bool
	Result      *ast.StmtNamespace
}

func (n *Namespace) EnterNode(node ast.Vertex) bool {
	if n.Result != nil && n.firstResult {
		return false
	}

	// Stop after given row.
	if !n.firstResult && node.GetPosition().StartLine >= n.row {
		return false
	}

	// Don't go into scopes, namespace is always top level.
	return !nodescopes.IsScope(node.GetType())
}

func (n *Namespace) StmtNamespace(node *ast.StmtNamespace) {
	n.Result = node
}
