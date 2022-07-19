package traversers

import "github.com/VKCOM/noverify/src/ir"

func NewAssignment(variable *ir.SimpleVar) *assignment {
	return &assignment{
		variable: variable,
	}
}

// assignment implements ir.Visitor.
type assignment struct {
	variable   *ir.SimpleVar
	Assignment *ir.SimpleVar
}

func (a *assignment) EnterNode(node ir.Node) bool {
	if a.Assignment != nil {
		return false
	}

	switch typedNode := node.(type) {
	case *ir.Assign:
		if assigned, ok := typedNode.Variable.(*ir.SimpleVar); ok {
			if assigned.Name == a.variable.Name {
				a.Assignment = assigned
			}
		}
	case *ir.Parameter:
		if typedNode.Variable.Name == a.variable.Name {
			a.Assignment = typedNode.Variable
		}
	}

	return true
}

func (a *assignment) LeaveNode(ir.Node) {}
