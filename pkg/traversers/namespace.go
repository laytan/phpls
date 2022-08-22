package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewNamespace(row uint) *Namespace {
	return &Namespace{
		row: row,
	}
}

// Namespace implements ir.Visitor.
type Namespace struct {
	row    uint
	Result *ir.NamespaceStmt
}

func (n *Namespace) EnterNode(node ir.Node) bool {
	// Stop after given row.
	if ir.GetPosition(node).StartLine >= int(n.row) {
		return false
	}

	if ns, ok := node.(*ir.NamespaceStmt); ok {
		n.Result = ns
	}

	// Don't go into scopes, namespace is always top level.
	return !symbol.IsScope(node)
}

func (n *Namespace) LeaveNode(ir.Node) {}
