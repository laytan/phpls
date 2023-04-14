package traversers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodeident"
)

type FunctionCall struct {
	visitor.Null
	name   string
	Result *ast.ExprFunctionCall
}

func NewFunctionCall(name string) *FunctionCall {
	return &FunctionCall{name: name}
}

func (v *FunctionCall) EnterNode(node ast.Vertex) bool {
	return v.Result == nil
}

func (v *FunctionCall) ExprFunctionCall(node *ast.ExprFunctionCall) {
	if v.name == nodeident.Get(node) {
		v.Result = node
	}
}
