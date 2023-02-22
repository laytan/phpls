package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

type Method struct {
	*modified
	*canReturn
	*doxed
	*parametized

	node *ir.ClassMethodStmt
}

func NewMethod(root rooter, node *ir.ClassMethodStmt) *Method {
	doxed := NewDoxed(node)

	return &Method{
		modified: newModifiedFromNode(node),
		canReturn: &canReturn{
			doxed:  doxed,
			rooter: root,
			node:   node,
		},
		doxed: doxed,
		parametized: &parametized{
			doxed:  doxed,
			rooter: root,
			node:   node,
		},
		node: node,
	}
}

func (m *Method) Name() string {
	return nodeident.Get(m.node)
}

func (m *Method) Node() *ir.ClassMethodStmt {
	return m.node
}
