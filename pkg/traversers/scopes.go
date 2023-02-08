package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

type ScopesTraverser struct {
	subject     ir.Node
	Block       ir.Node
	Class       ir.Node
	rootVisited bool
	Done        bool
}

func NewScopesTraverser(subject ir.Node) *ScopesTraverser {
	return &ScopesTraverser{
		subject: subject,
	}
}

func (s *ScopesTraverser) EnterNode(node ir.Node) bool {
	if !s.rootVisited {
		if _, ok := node.(*ir.Root); !ok {
			panic("ScopesTraverser only works on the root")
		}

		s.Block = node
		s.Class = node

		s.rootVisited = true
		return true
	}

	if s.Done {
		return false
	}

	if node == s.subject {
		s.Done = true
		return false
	}

	if symbol.IsClassLike(ir.GetNodeKind(node)) {
		s.Class = node
	}

	if symbol.IsScope(ir.GetNodeKind(node)) {
		s.Block = node
	}

	return true
}

func (s *ScopesTraverser) LeaveNode(node ir.Node) {}
