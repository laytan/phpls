package traversers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodescopes"
)

type ScopesTraverser struct {
	visitor.Null
	subject     ast.Vertex
	Block       ast.Vertex
	Class       ast.Vertex
	rootVisited bool
	Done        bool
}

func NewScopesTraverser(subject ast.Vertex) *ScopesTraverser {
	return &ScopesTraverser{
		subject: subject,
	}
}

func (s *ScopesTraverser) EnterNode(node ast.Vertex) bool {
	if !s.rootVisited {
		if _, ok := node.(*ast.Root); !ok {
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

	if nodescopes.IsClassLike(node.GetType()) {
		s.Class = node
	}

	if nodescopes.IsScope(node.GetType()) {
		s.Block = node
	}

	return true
}
