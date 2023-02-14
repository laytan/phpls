package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/phprivacy"
)

func NewProperty(name string, classLikeName string, privacy phprivacy.Privacy) *Property {
	return &Property{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
	}
}

// Property implements ir.Visitor.
type Property struct {
	Property      *ir.PropertyListStmt
	name          string
	classLikeName string
	privacy       phprivacy.Privacy
}

func (m *Property) EnterNode(node ir.Node) bool {
	if m.Property != nil {
		return false
	}

	switch typedNode := node.(type) {
	// Only parse a class-like node if the name matches (for multiple classes in a file).
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		return nodeident.Get(node) == m.classLikeName

	case *ir.PropertyListStmt:
		for _, property := range typedNode.Properties {
			stmt := property.(*ir.PropertyStmt)

			if stmt.Variable.Name != m.name {
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

				m.Property = typedNode
				return false
			}
		}
	}

	return !nodescopes.IsScope(ir.GetNodeKind(node))
}

func (m *Property) LeaveNode(ir.Node) {}
