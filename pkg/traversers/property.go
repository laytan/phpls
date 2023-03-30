package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
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

type Property struct {
	visitor.Null
	Property      *ast.StmtPropertyList
	name          string
	classLikeName string
	privacy       phprivacy.Privacy
}

func (m *Property) EnterNode(node ast.Vertex) bool {
	if m.Property != nil {
		return false
	}

	if nodescopes.IsClassLike(node.GetType()) {
		return nodeident.Get(node) == m.classLikeName
	}

	return !nodescopes.IsScope(node.GetType())
}

func (m *Property) StmtPropertyList(node *ast.StmtPropertyList) {
	for _, property := range node.Props {
		stmt := property.(*ast.StmtProperty)

		if nodeident.Get(stmt.Var) != m.name {
			continue
		}

		for _, mod := range node.Modifiers {
			privacy, err := phprivacy.FromString(nodeident.Get(mod))
			if err != nil {
				continue
			}

			if !m.privacy.CanAccess(privacy) {
				continue
			}

			m.Property = node
		}
	}
}
