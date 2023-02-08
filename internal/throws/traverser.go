package throws

import (
	"github.com/VKCOM/noverify/src/ir"
)

type ThrowsTraverser struct {
	Result       []ir.Node
	visitedFirst bool
}

func NewThrowsTraverser() *ThrowsTraverser {
	return &ThrowsTraverser{
		Result: []ir.Node{},
	}
}

func (t *ThrowsTraverser) EnterNode(node ir.Node) bool {
	if !t.visitedFirst {
		t.visitedFirst = true
		return true
	}

	switch node.(type) {
	case *ir.TryStmt, *ir.FunctionCallExpr, *ir.CatchStmt, *ir.ThrowStmt:
		t.Result = append(t.Result, node)
		return false

		// PERF: we can probably get away with returning true in a couple of cases.
	default:
		return true
	}
}

func (t *ThrowsTraverser) LeaveNode(node ir.Node) {}
