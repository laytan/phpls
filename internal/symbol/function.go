package symbol

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/traversers"
)

type Function struct {
	*canReturn
	*doxed
	*parametized

	node *ir.FunctionStmt
}

func NewFunction(root rooter, node *ir.FunctionStmt) *Function {
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
	root.Root().Walk(ft)

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

func (m *Function) Node() *ir.FunctionStmt {
	return m.node
}
