package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

type Property struct {
	*modified
	*Doxed

	node *ir.PropertyListStmt
}

func NewProperty(node *ir.PropertyListStmt) *Property {
	return &Property{
		node:     node,
		modified: newModifiedFromNode(node),
		Doxed:    NewDoxed(node),
	}
}

func (p *Property) Name() string {
	return nodeident.Get(p.node)
}

func (p *Property) Node() *ir.PropertyListStmt {
	return p.node
}
