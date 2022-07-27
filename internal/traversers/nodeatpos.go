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

	if n.pos >= uint(pos.StartPos) && n.pos <= uint(pos.EndPos) {
		n.Nodes = append(n.Nodes, node)
		return true
	}

	return false
}

func (n *NodeAtPos) LeaveNode(ir.Node) {}
