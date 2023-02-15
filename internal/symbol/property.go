package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

type Property struct {
	*modified
	*doxed

	node *ir.PropertyListStmt
}

func NewProperty(node *ir.PropertyListStmt) *Property {
	return &Property{
		node:     node,
		modified: newModifiedFromNode(node),
		doxed:    NewDoxed(node),
	}
}

func (p *Property) Name() string {
	return nodeident.Get(p.node)
}

func (p *Property) Node() *ir.PropertyListStmt {
	return p.node
}
