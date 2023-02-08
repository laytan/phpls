package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

// TODO: rename file.
func NewUses(classLikeName string) *Uses {
	return &Uses{
		Uses:          make([]*ir.Name, 0),
		classLikeName: classLikeName,
	}
}

// Uses implements ir.Visitor.
type Uses struct {
	Uses          []*ir.Name
	classLikeName string
}

func (u *Uses) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	// Only parse a class-like node if the name matches (for multiple classes in a file).
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		if symbol.GetIdentifier(node) == u.classLikeName {
			return true
		}

		return false

	case *ir.TraitUseStmt:
		for _, trait := range typedNode.Traits {
			if name, ok := trait.(*ir.Name); ok {
				u.Uses = append(u.Uses, name)
			}
		}
	}

	return !symbol.IsScope(ir.GetNodeKind(node))
}

func (u *Uses) LeaveNode(ir.Node) {}
