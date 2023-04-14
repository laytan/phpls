package traversers

import (
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/nodescopes"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
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
