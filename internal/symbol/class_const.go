package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

type ClassConst struct {
	*modified

	node *ir.ClassConstListStmt
}

func NewClassConst(node *ir.ClassConstListStmt) *ClassConst {
	return &ClassConst{
		node:     node,
		modified: newModifiedFromNode(node),
	}
}

func (p *ClassConst) Name() string {
	return nodeident.Get(p.node)
}

func (p *ClassConst) Node() *ir.ClassConstListStmt {
	return p.node
}
