package traversers

import (
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/phprivacy"
)

func NewClassLikeConstant(
	name string,
	classLikeName string,
	privacy phprivacy.Privacy,
) *ClassLikeConstant {
	return &ClassLikeConstant{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
	}
}

type ClassLikeConstant struct {
	visitor.Null
	ClassLikeConstant *ast.StmtClassConstList
	name              string
	classLikeName     string
	privacy           phprivacy.Privacy
}

func (m *ClassLikeConstant) EnterNode(node ast.Vertex) bool {
	if m.ClassLikeConstant != nil {
		return false
	}

	// Only parse a class-like node if the name matches (for multiple classes in a file).
	if nodescopes.IsClassLike(node.GetType()) {
		return nodeident.Get(node) == m.classLikeName
	}

	return !nodescopes.IsScope(node.GetType())
}

func (m *ClassLikeConstant) StmtClassConstList(constList *ast.StmtClassConstList) {
	hasConst := false
	for _, constNode := range constList.Consts {
		typedConstNode, ok := constNode.(*ast.StmtConstant)
		if !ok {
			log.Printf(
				"Unexpected node type %T for constant, expecting *ast.StmtConstant",
				constNode,
			)
			continue
		}

		if nodeident.Get(typedConstNode.Name) == m.name {
			hasConst = true
		}
	}

	if !hasConst {
		return
	}

	hasPrivacy := false
	for _, mod := range constList.Modifiers {
		privacy, err := phprivacy.FromString(nodeident.Get(mod))
		if err != nil {
			continue
		}

		hasPrivacy = true

		if !m.privacy.CanAccess(privacy) {
			continue
		}

		m.ClassLikeConstant = constList
		return
	}

	// No privacy, means public privacy, means accessible.
	if !hasPrivacy {
		m.ClassLikeConstant = constList
		return
	}
}
