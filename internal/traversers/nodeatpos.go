package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
)

func NewNodeAtPos(pos uint) *NodeAtPos {
	return &NodeAtPos{
		pos:   pos,
		Nodes: make([]ir.Node, 0),
	}
}

// NodeAtPos implements ir.Visitor and populates Nodes with the nodes spanning pos.
type NodeAtPos struct {
	pos   uint
	Nodes []ir.Node
}

func (n *NodeAtPos) EnterNode(node ir.Node) bool {
	pos := ir.GetPosition(node)
	if pos == nil {
		return true
	}

	switch node.(type) {
	case *ir.ClassExtendsStmt:
		// Weird edge case where the position of this node is only the 'extends'
		// keyword, but it still has the class name (*ir.Name) child with a different position.
		// NOTE: we are not appending the node to Nodes because we don't know if it matches,
		// NOTE: so if we need the ClassExtendsStmt in the returned nodes in the future, this needs to change.
		return true
	default:
		break
	}

	if n.pos >= uint(pos.StartPos) && n.pos <= uint(pos.EndPos) {
		n.Nodes = append(n.Nodes, node)
		return true
	}

	return false
}

func (n *NodeAtPos) LeaveNode(ir.Node) {}
