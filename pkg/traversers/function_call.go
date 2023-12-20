package traversers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodeident"
)

type FunctionCall struct {
	visitor.Null
	name     string
	Result   []*ast.ExprFunctionCall
	multiple bool
}

func NewFunctionCall(name string, multiple bool) *FunctionCall {
	return &FunctionCall{name: name, multiple: multiple}
}

func (v *FunctionCall) EnterNode(node ast.Vertex) bool {
	return v.multiple || v.Result == nil
}

func (v *FunctionCall) ExprFunctionCall(node *ast.ExprFunctionCall) {
	if v.name == nodeident.Get(node) {
		v.Result = append(v.Result, node)
	}
}
