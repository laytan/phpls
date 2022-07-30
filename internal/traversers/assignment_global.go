package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

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

// TODO: Handle globals that are assigned only within a scope.
func (a *GlobalAssignment) EnterNode(node ir.Node) bool {
	if a.Assignment != nil {
		return false
	}

	// Only care about globals, so when we come across a scope, don't go deeper.
	if symbol.IsScope(node) {
		return false
	}

	if variable, ok := node.(*ir.SimpleVar); ok && variable.Name == a.globalVar.Name {
		a.Assignment = variable
		return false
	}

	return true
}

func (a *GlobalAssignment) LeaveNode(ir.Node) {}
