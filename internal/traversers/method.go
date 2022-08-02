package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewMethod(name string, classLikeName string, privacy phprivacy.Privacy) *Method {
	return &Method{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
	}
}

// Method implements ir.Visitor.
type Method struct {
	Method        *ir.ClassMethodStmt
	name          string
	classLikeName string
	privacy       phprivacy.Privacy
}

func (m *Method) EnterNode(node ir.Node) bool {
	if m.Method != nil {
		return false
	}

	switch typedNode := node.(type) {
	// Only parse a class-like node if the name matches (for multiple classes in a file).
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		return symbol.GetIdentifier(node) == m.classLikeName

	case *ir.ClassMethodStmt:
		if typedNode.MethodName.Value != m.name {
			return false
		}

		for _, mod := range typedNode.Modifiers {
			privacy, err := phprivacy.FromString(mod.Value)
			if err != nil {
				continue
			}

			if !m.privacy.CanAccess(privacy) {
				continue
			}

			m.Method = typedNode
			return false
		}
	}

	return !symbol.IsScope(node)
}

func (m *Method) LeaveNode(ir.Node) {}
