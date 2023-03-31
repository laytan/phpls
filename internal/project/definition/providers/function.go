package providers

import (
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/php-parser/pkg/ast"
)

// FunctionProvider resolves the definition of a function call.
// It first looks for it in the current scope (a local function declaration).
// if it can't find it, the function will be looked for in the global scope.
type FunctionProvider struct{}

func NewFunction() *FunctionProvider {
	return &FunctionProvider{}
}

func (p *FunctionProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprFunctionCall
}

func (p *FunctionProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
