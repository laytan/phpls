package index

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
)

type INode struct {
	FQN        *fqn.FQN
	Path       string
	Position   *position.Position
	Identifier string
	Kind       ir.NodeKind
}

func NewINode(fqns *fqn.FQN, path string, node ir.Node) *INode {
	return &INode{
		FQN:        fqns,
		Path:       path,
		Position:   ir.GetPosition(node),
		Identifier: nodeident.Get(node),
		Kind:       ir.GetNodeKind(node),
	}
}

func (i *INode) MatchesKind(kinds ...ir.NodeKind) bool {
	if len(kinds) == 0 {
		return true
	}

	for _, kind := range kinds {
		if kind == ir.KindRoot {
			return true
		}

		if kind == i.Kind {
			return true
		}
	}

	return false
}

func (i *INode) ToIRNode(root ir.Node) ir.Node {
	t := &toNodeTraverser{
		Node: i,
	}
	root.Walk(t)

	return t.Result
}

type toNodeTraverser struct {
	Result ir.Node
	Node   *INode
}

func (t *toNodeTraverser) EnterNode(node ir.Node) bool {
	if t.Result != nil {
		return false
	}

	if ir.GetNodeKind(node) != t.Node.Kind {
		return true
	}

	if nodeident.Get(node) != t.Node.Identifier {
		return true
	}

	nPos := ir.GetPosition(node)
	sPos := t.Node.Position
	if sPos.StartPos != nPos.StartPos || sPos.EndPos != nPos.EndPos ||
		sPos.StartLine != nPos.StartLine ||
		sPos.EndLine != nPos.EndLine {
		return true
	}

	t.Result = node
	return false
}

func (t *toNodeTraverser) LeaveNode(ir.Node) {}
