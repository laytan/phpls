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
		call: call,
		name: name,
	}, nil
}

// Function implements ir.Visitor.
type Function struct {
	call     *ir.FunctionCallExpr
	name     *ir.Name
	Function *ir.FunctionStmt
}

func (f *Function) EnterNode(node ir.Node) bool {
	if f.Function != nil {
		return false
	}

	if function, ok := node.(*ir.FunctionStmt); ok {
		if function.FunctionName.Value == f.name.Value {
			f.Function = function
		}

		return false
	}

	return true
}

func (f *Function) LeaveNode(ir.Node) {}
