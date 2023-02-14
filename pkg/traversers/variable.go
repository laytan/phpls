package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

// Variable implements ir.Visitor.
type Variable struct {
	name   string
	Result *ir.SimpleVar
}

func NewVariable(name string) *Variable {
	return &Variable{name: name}
}

func (v *Variable) EnterNode(node ir.Node) bool {
	if v.Result != nil {
		return false
	}

	switch typedNode := node.(type) {
	case *ir.SimpleVar:
		if v.name == nodeident.Get(node) {
			v.Result = typedNode
			return false
		}

		return true

	default:
		return true
	}
}

func (v *Variable) LeaveNode(ir.Node) {}
