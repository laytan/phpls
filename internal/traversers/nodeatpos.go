package traversers

import "github.com/VKCOM/noverify/src/ir"

func NewNodeAtPos(pos uint) *nodeatpos {
	return &nodeatpos{
		pos:   pos,
		Nodes: make([]ir.Node, 0),
	}
}

// nodeatpos implements ir.Visitor and populates Nodes with the nodes spanning pos.
type nodeatpos struct {
	pos   uint
	Nodes []ir.Node
}

func (n *nodeatpos) EnterNode(node ir.Node) bool {
	pos := ir.GetPosition(node)
	if n.pos >= uint(pos.StartPos) && n.pos <= uint(pos.EndPos) {
		n.Nodes = append(n.Nodes, node)
		return true
	}

	return false
}

func (n *nodeatpos) LeaveNode(ir.Node) {}
