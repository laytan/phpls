package traversers

import (
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
)

func NewUses(classLikeName string) *Uses {
	return &Uses{
		Uses:          make([]*ast.Name, 0),
		classLikeName: classLikeName,
	}
}

type Uses struct {
	visitor.Null
	Uses          []*ast.Name
	classLikeName string
}

func (u *Uses) EnterNode(node ast.Vertex) bool {
	if nodescopes.IsClassLike(node.GetType()) {
		return nodeident.Get(node) == u.classLikeName
	}

	return !nodescopes.IsScope(node.GetType())
}

func (u *Uses) StmtTraitUse(node *ast.StmtTraitUse) {
	for _, trait := range node.Traits {
		if name, ok := trait.(*ast.Name); ok {
			u.Uses = append(u.Uses, name)
		} else {
			log.Printf("*ast.StmtTraitUse uses node that is not *ast.Name but %T", trait)
		}
	}
}
