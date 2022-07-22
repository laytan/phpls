package traversers

import "github.com/VKCOM/noverify/src/ir"

func NewGlobalAssignment(globalVar *ir.SimpleVar) *GlobalAssignment {
	return &GlobalAssignment{
		globalVar: globalVar,
	}
}

// Assignment implements ir.Visitor.
type GlobalAssignment struct {
	globalVar  *ir.SimpleVar
	Assignment *ir.SimpleVar
}

func (a *GlobalAssignment) EnterNode(node ir.Node) bool {
	if a.Assignment != nil {
		return false
	}

	// TODO: Handle globals that are assigned only within a scope.

	// TODO: It might make sense to return multiple assignments for everytime the global var is assigned.

	switch typedNode := node.(type) {
	// If we get within a function it is not global anymore.
	// OPTIM: add other scopes like class, etc. so we don't visit those nodes.
	case *ir.FunctionStmt:
		return false

	case *ir.SimpleVar:
		if typedNode.Name == a.globalVar.Name {
			a.Assignment = typedNode
		}
	}

	return true
}

func (a *GlobalAssignment) LeaveNode(ir.Node) {}
