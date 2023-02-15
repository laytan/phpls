package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
)

func NewNodeAtPos(pos uint) *NodeAtPos {
	return &NodeAtPos{
		pos:   pos,
		Nodes: []ir.Node{},
	}
}

// NodeAtPos implements ir.Visitor and populates Nodes with the nodes spanning pos.
type NodeAtPos struct {
	pos   uint
	Nodes []ir.Node

	// If the cursor is inside a comment, this is set to that comment
	// and nodes are the nodes containing the comment.
	// The comment is always the last/most specific node.
	Comment *token.Token
}

func (n *NodeAtPos) EnterNode(node ir.Node) bool {
	if n.Comment != nil {
		return true
	}

	pos := ir.GetPosition(node)
	if pos == nil {
		return true
	}

	switch typedNode := node.(type) {
	case *ir.ClassExtendsStmt:
		// Weird edge case where the position of this node is only the 'extends'
		// keyword, but it still has the class name (*ir.Name) child with a different position.
		// NOTE: we are not appending the node to Nodes because we don't know if it matches,
		// NOTE: so if we need the ClassExtendsStmt in the returned nodes in the future, this needs to change.
		return true
	case *ir.FunctionStmt, *ir.TraitStmt, *ir.InterfaceStmt:
		n.checkComment(node, node, nil)
	case *ir.PropertyListStmt:
		n.checkComment(node, node, typedNode.Modifiers)
	case *ir.ClassMethodStmt:
		n.checkComment(typedNode, typedNode, typedNode.Modifiers)
	case *ir.ClassStmt:
		n.checkComment(typedNode, typedNode, typedNode.Modifiers)

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

func (n *NodeAtPos) checkComment(node ir.Node, toCheck ir.Node, modifiers []*ir.Identifier) {
	if len(modifiers) > 0 {
		n.checkComment(node, modifiers[0], nil)
		return
	}

	toCheck.IterateTokens(func(t *token.Token) bool {
		if t.ID != token.T_COMMENT && t.ID != token.T_DOC_COMMENT {
			return true
		}

		if n.pos >= uint(t.Position.StartPos) && n.pos <= uint(t.Position.EndPos) {
			n.Comment = t
			n.Nodes = append(n.Nodes, node)
		}

		return true
	})
}
