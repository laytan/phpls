package symbol

import (
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/php-parser/pkg/ast"
)

type Method struct {
	*modified
	*canReturn
	*doxed
	*parametized

	node *ast.StmtClassMethod
}

func NewMethod(root rooter, node *ast.StmtClassMethod) *Method {
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

func (m *Method) Node() *ast.StmtClassMethod {
	return m.node
}

func (m *Method) Vertex() ast.Vertex {
	return m.node
}
