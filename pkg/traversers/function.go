package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewFunction(name string) *Function {
	return &Function{
		name:           name,
		currNodeIsRoot: true,
	}
}

// Function implements ir.Visitor.
type Function struct {
	name           string
	Function       *ir.FunctionStmt
	currNodeIsRoot bool
}

func (f *Function) EnterNode(node ir.Node) bool {
	defer func() { f.currNodeIsRoot = false }()

	if f.Function != nil {
		return false
	}

	// If the scope of the traverser is a function, the first call will be a
	// function which we need to ignore.
	kind := ir.GetNodeKind(node)
	if f.currNodeIsRoot && kind == ir.KindFunctionStmt {
		return true
	}

	if function, ok := node.(*ir.FunctionStmt); ok {
		if function.FunctionName.Value == f.name {
			f.Function = function
		}
	}

	if symbol.IsScope(node) {
		return false
	}

	return true
}

func (f *Function) LeaveNode(ir.Node) {}
