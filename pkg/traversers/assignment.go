package traversers

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
)

func NewAssignment(variable *ir.SimpleVar) *Assignment {
	return &Assignment{
		variable: variable,
	}
}

// Assignment implements ir.Visitor.
type Assignment struct {
	variable   *ir.SimpleVar
	Assignment *ir.SimpleVar
	Scope      ir.Node
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
				a.Scope = typedNode
			}
		}

	case *ir.Parameter:
		if typedNode.Variable.Name == a.variable.Name {
			a.Assignment = typedNode.Variable
			a.Scope = typedNode
		}

	case *ir.GlobalStmt:
		for _, varNode := range typedNode.Vars {
			typedVar, ok := varNode.(*ir.SimpleVar)
			if !ok {
				continue
			}

			if typedVar.Name == a.variable.Name {
				a.Assignment = typedVar
				a.Scope = typedNode
			}
		}

	case *ir.ListExpr:
		for _, item := range typedNode.Items {
			typedVar, ok := item.Val.(*ir.SimpleVar)
			if !ok {
				log.Printf("Got type %T for an item in list expression, expected *ir.SimpleVar", item.Val)
				continue
			}

			if typedVar.Name == a.variable.Name {
				a.Assignment = typedVar
				a.Scope = typedNode
			}
		}

	}

	return true
}

func (a *Assignment) LeaveNode(ir.Node) {}
