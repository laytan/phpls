package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

// FunctionCall implements ir.Visitor.
type FunctionCall struct {
	name   string
	Result *ir.FunctionCallExpr
}

func NewFunctionCall(name string) *FunctionCall {
	return &FunctionCall{name: name}
}

func (v *FunctionCall) EnterNode(node ir.Node) bool {
	if v.Result != nil {
		return false
	}

	switch typedNode := node.(type) {
	case *ir.FunctionCallExpr:
		if v.name == symbol.GetIdentifier(node) {
			v.Result = typedNode
			return false
		}

		return true

	default:
		return true
	}
}

func (v *FunctionCall) LeaveNode(ir.Node) {}
