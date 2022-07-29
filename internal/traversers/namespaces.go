package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
)

func NewNamespaces() *Namespaces {
	return &Namespaces{
		Namespaces: make([]string, 0),
		Uses:       make([]*ir.UseStmt, 0),
	}
}

// Namespaces implements ir.Visitor.
type Namespaces struct {
	Namespaces []string
	Uses       []*ir.UseStmt
}

func (n *Namespaces) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {

	case *ir.Root:
		return true

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			n.Namespaces = append(n.Namespaces, typedNode.NamespaceName.Value)
		}

		return false

	case *ir.UseStmt:
		n.Uses = append(n.Uses, typedNode)

		return false

	default:
		return false
	}
}

func (n *Namespaces) LeaveNode(ir.Node) {}
