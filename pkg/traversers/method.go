package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
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

type Method struct {
	visitor.Null
	Method        *ast.StmtClassMethod
	name          string
	classLikeName string
	privacy       phprivacy.Privacy
	static        bool
}

func (m *Method) EnterNode(node ast.Vertex) bool {
	if m.Method != nil {
		return false
	}

	// Only parse a class-like node if the name matches (for multiple classes in a file).
	if nodescopes.IsClassLike(node.GetType()) {
		return nodeident.Get(node) == m.classLikeName
	}

	return !nodescopes.IsScope(node.GetType())
}

func (m *Method) StmtClassMethod(node *ast.StmtClassMethod) {
	if nodeident.Get(node) != m.name {
		return
	}

	hasPrivacy := false
	for _, mod := range node.Modifiers {
		modStr := nodeident.Get(mod)
		if modStr == "static" && !m.static {
			continue
		}

		privacy, err := phprivacy.FromString(modStr)
		if err != nil {
			continue
		}

		hasPrivacy = true

		if !m.privacy.CanAccess(privacy) {
			continue
		}

		m.Method = node
		return
	}

	// No privacy, means public privacy, means accesible.
	if !hasPrivacy {
		m.Method = node
		return
	}
}
