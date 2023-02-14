package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
)

func ToNode(root *ir.Root, sym Symbol) ir.Node {
	t := &symNodeTraverser{sym: sym}
	root.Walk(t)

	return t.Result
}

// symNodeTraverser implements ir.Visitor.
type symNodeTraverser struct {
	sym    Symbol
	Result ir.Node
}

func (stn *symNodeTraverser) EnterNode(node ir.Node) bool {
	if stn.Result != nil {
		return false
	}

	if ir.GetNodeKind(node) != stn.sym.NodeKind() {
		return true
	}

	if nodeident.Get(node) != stn.sym.Identifier() {
		return true
	}

	nPos := ir.GetPosition(node)
	sPos := stn.sym.Position()
	if sPos.StartPos != nPos.StartPos || sPos.EndPos != nPos.EndPos ||
		sPos.StartLine != nPos.StartLine ||
		sPos.EndLine != nPos.EndLine {
		return true
	}

	stn.Result = node
	return false
}

func (stn *symNodeTraverser) LeaveNode(ir.Node) {}
