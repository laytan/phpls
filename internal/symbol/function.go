package symbol

import (
	"fmt"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/traversers"
)

type Function struct {
	*canReturn
	*doxed
	*parametized

	node *ast.StmtFunction
}

func NewFunction(root rooter, node *ast.StmtFunction) *Function {
	doxed := NewDoxed(node)

	return &Function{
		canReturn: &canReturn{
			doxed:  doxed,
			rooter: root,
			node:   node,
		},
		parametized: &parametized{
			rooter: root,
			doxed:  doxed,
			node:   node,
		},
		doxed: doxed,
		node:  node,
	}
}

func NewFunctionFromFQN(root rooter, qualified *fqn.FQN) (*Function, error) {
	ft := traversers.NewFunction(qualified.Name())
	tv := traverser.NewTraverser(ft)
	root.Root().Accept(tv)

	if ft.Function == nil {
		return nil, fmt.Errorf(
			"[symbol.NewFunctionFromFQN]: can't find %v in the given root",
			qualified,
		)
	}

	return NewFunction(root, ft.Function), nil
}

func (m *Function) Name() string {
	return nodeident.Get(m.node)
}

func (m *Function) Node() *ast.StmtFunction {
	return m.node
}
