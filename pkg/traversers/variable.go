package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodeident"
)

type Variable struct {
	visitor.Null
	name   string
	Result *ast.ExprVariable
}

func NewVariable(name string) *Variable {
	return &Variable{name: name}
}

func (v *Variable) EnterNode(node ast.Vertex) bool {
	return v.Result == nil
}

func (v *Variable) ExprVariable(node *ast.ExprVariable) {
	if v.name == nodeident.Get(node) {
		v.Result = node
	}
}
