package traversers

import "github.com/VKCOM/noverify/src/ir"

func NewAssignment(variable *ir.SimpleVar) *Assignment {
	return &Assignment{
		variable: variable,
	}
}

// Assignment implements ir.Visitor.
type Assignment struct {
	variable   *ir.SimpleVar
	Assignment *ir.SimpleVar
}

func (a *Assignment) EnterNode(node ir.Node) bool {
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
	case *ir.GlobalStmt:
		for _, varNode := range typedNode.Vars {
			typedVar, ok := varNode.(*ir.SimpleVar)
			if !ok {
				continue
			}

			if typedVar.Name == a.variable.Name {
				a.Assignment = typedVar
			}
		}
	}

	return true
}

func (a *Assignment) LeaveNode(ir.Node) {}
