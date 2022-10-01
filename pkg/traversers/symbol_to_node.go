package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

// SymbolToNode implements ir.Visitor.
type SymbolToNode struct {
	sym    symbol.Symbol
	Result ir.Node
}

func NewSymbolToNode(sym symbol.Symbol) *SymbolToNode {
	return &SymbolToNode{sym: sym}
}

func (stn *SymbolToNode) EnterNode(node ir.Node) bool {
	if stn.Result != nil {
		return false
	}

	if ir.GetNodeKind(node) != stn.sym.NodeKind() {
		return true
	}

	if symbol.GetIdentifier(node) != stn.sym.Identifier() {
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

func (stn *SymbolToNode) LeaveNode(ir.Node) {}
