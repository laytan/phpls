package traversers

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/token"
	"github.com/laytan/php-parser/pkg/visitor"
)

func NewNodeAtPos(pos int) *NodeAtPos {
	return &NodeAtPos{
		pos:   pos,
		Nodes: []ast.Vertex{},
	}
}

type NodeAtPos struct {
	visitor.Null
	pos   int
	Nodes []ast.Vertex

	// If the cursor is inside a comment, this is set to that comment
	// and nodes are the nodes containing the comment.
	// The comment is always the last/most specific node.
	Comment *token.Token
}

func (n *NodeAtPos) EnterNode(node ast.Vertex) bool {
	if n.Comment != nil {
		return true
	}

	n.checkComment(node)

	if n.pos >= node.GetPosition().StartPos && n.pos <= node.GetPosition().EndPos {
		n.Nodes = append(n.Nodes, node)
		return true
	}

	return false
}

// TODO: cleaner abstraction, maybe generate this?
//
// Gets the first free floating list of a node (this is where any doc comments before the node would be added by the parser)
// and check it for comments.
func (n *NodeAtPos) checkComment(node ast.Vertex) {
	switch tn := node.(type) {
	case *ast.StmtFunction:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		n.checkTokens(node, tn.FunctionTkn.FreeFloating)
	case *ast.StmtTrait:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		n.checkTokens(node, tn.TraitTkn.FreeFloating)
	case *ast.StmtClass:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		if len(tn.Modifiers) > 0 {
			n.checkTokens(node, tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating)
			return
		}

		n.checkTokens(node, tn.ClassTkn.FreeFloating)
	case *ast.StmtInterface:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		n.checkTokens(node, tn.InterfaceTkn.FreeFloating)
	case *ast.StmtPropertyList:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		if len(tn.Modifiers) > 0 {
			n.checkTokens(node, tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating)
			return
		}

		log.Println("Warning: *ast.StmtPropertyList without any attributes or modifiers, is that possible?")
	case *ast.StmtClassMethod:
		if len(tn.AttrGroups) > 0 {
			n.checkTokens(node, tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating)
			return
		}

		if len(tn.Modifiers) > 0 {
			n.checkTokens(node, tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating)
			return
		}

		n.checkTokens(node, tn.FunctionTkn.FreeFloating)
	}
}

func (n *NodeAtPos) checkTokens(parent ast.Vertex, tokens []*token.Token) {
	for _, ff := range tokens {
		n.checkToken(parent, ff)
	}
}

func (n *NodeAtPos) checkToken(parent ast.Vertex, t *token.Token) {
	if t.ID != token.T_COMMENT && t.ID != token.T_DOC_COMMENT {
		return
	}

	if n.pos >= t.Position.StartPos && n.pos <= t.Position.EndPos {
		n.Comment = t
		n.Nodes = append(n.Nodes, parent)
	}
}
