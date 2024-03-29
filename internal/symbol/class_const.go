package symbol

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/pkg/nodeident"
)

type ClassConst struct {
	*modified

	node *ast.StmtClassConstList
}

func NewClassConst(node *ast.StmtClassConstList) *ClassConst {
	return &ClassConst{
		node:     node,
		modified: newModifiedFromNode(node),
	}
}

func (p *ClassConst) Name() string {
	return nodeident.Get(p.node)
}

func (p *ClassConst) Node() *ast.StmtClassConstList {
	return p.node
}
