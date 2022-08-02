package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewClassLike(name string) *ClassLike {
	return &ClassLike{name: name}
}

// ClassLike implements ir.Visitor.
type ClassLike struct {
	name      string
	ClassLike ir.Node
}

func (c *ClassLike) EnterNode(node ir.Node) bool {
	if c.ClassLike != nil {
		return false
	}

	if symbol.IsClassLike(node) && symbol.GetIdentifier(node) == c.name {
		c.ClassLike = node
	}

	return !symbol.IsScope(node)
}

func (c *ClassLike) LeaveNode(ir.Node) {}
