package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
)

func NewClassLike(name string) *ClassLike {
	return &ClassLike{name: name}
}

// ClassLike implements ir.Visitor.
type ClassLike struct {
	visitor.Null
	name      string
	ClassLike ast.Vertex
}

func (c *ClassLike) EnterNode(node ast.Vertex) bool {
	if c.ClassLike != nil {
		return false
	}

	if nodescopes.IsClassLike(node.GetType()) && nodeident.Get(node) == c.name {
		c.ClassLike = node
	}

	return !nodescopes.IsScope(node.GetType())
}
