package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/phprivacy"
)

func NewMethod(name string, classLikeName string, privacy phprivacy.Privacy) *Method {
	return &Method{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
	}
}

func NewMethodStatic(name, classLikeName string, privacy phprivacy.Privacy) *Method {
	return &Method{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
		static:        true,
	}
}

// Method implements ir.Visitor.
type Method struct {
	Method        *ir.ClassMethodStmt
	name          string
	classLikeName string
	privacy       phprivacy.Privacy
	static        bool
}

func (m *Method) EnterNode(node ir.Node) bool {
	if m.Method != nil {
		return false
	}

	switch typedNode := node.(type) {
	// Only parse a class-like node if the name matches (for multiple classes in a file).
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		return nodeident.Get(node) == m.classLikeName

	case *ir.ClassMethodStmt:
		if typedNode.MethodName.Value != m.name {
			return false
		}

		hasPrivacy := false
		for _, mod := range typedNode.Modifiers {
			if mod.Value == "static" && !m.static {
				continue
			}

			privacy, err := phprivacy.FromString(mod.Value)
			if err != nil {
				continue
			}

			hasPrivacy = true

			if !m.privacy.CanAccess(privacy) {
				continue
			}

			m.Method = typedNode
			return false
		}

		// No privacy, means public privacy, means accesible.
		if !hasPrivacy {
			m.Method = typedNode
			return false
		}
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (m *Method) LeaveNode(ir.Node) {}
