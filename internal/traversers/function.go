package traversers

import (
	"errors"

	"github.com/VKCOM/noverify/src/ir"
)

func NewFunction(call *ir.FunctionCallExpr) (*Function, error) {
	name, ok := call.Function.(*ir.Name)
	if !ok {
		return nil, errors.New(
			"Can't get function definition for given node because it has no name",
		)
	}

	return &Function{
		call:           call,
		name:           name,
		currNodeIsRoot: true,
	}, nil
}

// Function implements ir.Visitor.
type Function struct {
	call           *ir.FunctionCallExpr
	name           *ir.Name
	Function       *ir.FunctionStmt
	currNodeIsRoot bool
}

func (f *Function) EnterNode(node ir.Node) bool {
	defer func() { f.currNodeIsRoot = false }()

	if f.Function != nil {
		return false
	}

	if function, ok := node.(*ir.FunctionStmt); ok {
		if function.FunctionName.Value == f.name.Value {
			f.Function = function
		}

		// If the root node is a function, we need to return true so we
		// check the nodes inside it.
		if f.currNodeIsRoot {
			return true
		}

		return false
	}

	return true
}

func (f *Function) LeaveNode(ir.Node) {}
