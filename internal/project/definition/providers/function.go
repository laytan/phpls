package providers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/nodeident"
)

// FunctionProvider resolves the definition of a function call.
// It first looks for it in the current scope (a local function declaration).
// if it can't find it, the function will be looked for in the global scope.
type FunctionProvider struct{}

func NewFunction() *FunctionProvider {
	return &FunctionProvider{}
}

func (p *FunctionProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprFunctionCall || kind == ast.TypeStmtFunction
}

func (p *FunctionProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	// If stmt function, it is the current function, so just return that one.
	if ctx.Current().GetType() == ast.TypeStmtFunction {
		return []*definition.Definition{{
			Path:       ctx.Path(),
			Position:   ctx.Current().GetPosition(),
			Identifier: nodeident.Get(ctx.Current()),
		}}, nil
	}

	return DefineExpr(ctx)
}
