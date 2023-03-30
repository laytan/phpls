package index

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
)

type INode struct {
	FQN        *fqn.FQN
	Path       string
	Position   *position.Position
	Identifier string
	Kind       ast.Type
}

func NewINode(fqns *fqn.FQN, path string, node ast.Vertex) *INode {
	return &INode{
		FQN:        fqns,
		Path:       path,
		Position:   node.GetPosition(),
		Identifier: nodeident.Get(node),
		Kind:       node.GetType(),
	}
}

func (i *INode) MatchesKind(kinds ...ast.Type) bool {
	if len(kinds) == 0 {
		return true
	}

	for _, kind := range kinds {
		if kind == ast.TypeRoot {
			return true
		}

		if kind == i.Kind {
			return true
		}
	}

	return false
}

func (i *INode) ToIRNode(root *ast.Root) ast.Vertex {
	t := &toNodeTraverser{
		Node: i,
	}
	tv := traverser.NewTraverser(t)
	root.Accept(tv)
	return t.Result
}

type toNodeTraverser struct {
	visitor.Null
	Result ast.Vertex
	Node   *INode
}

func (t *toNodeTraverser) EnterNode(node ast.Vertex) bool {
	if t.Result != nil {
		return false
	}

	if node.GetType() != t.Node.Kind {
		return true
	}

	if nodeident.Get(node) != t.Node.Identifier {
		return true
	}

	nPos := node.GetPosition()
	sPos := t.Node.Position
	if sPos.StartPos != nPos.StartPos || sPos.EndPos != nPos.EndPos ||
		sPos.StartLine != nPos.StartLine ||
		sPos.EndLine != nPos.EndLine {
		return true
	}

	t.Result = node
	return false
}
