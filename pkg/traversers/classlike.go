package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
)

// TODO: accept a fqn and search right namespace.
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

	if nodescopes.IsClassLike(ir.GetNodeKind(node)) && nodeident.Get(node) == c.name {
		c.ClassLike = node
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (c *ClassLike) LeaveNode(ir.Node) {}
