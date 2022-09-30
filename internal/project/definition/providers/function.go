package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// FunctionProvider resolves the definition of a function call.
// It first looks for it in the current scope (a local function declaration).
// if it can't find it, the function will be looked for in the global scope.
type FunctionProvider struct{}

func NewFunction() *FunctionProvider {
	return &FunctionProvider{}
}

func (p *FunctionProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	if kind != ir.KindFunctionCallExpr {
		return false
	}

	_, ok := ctx.Current().(*ir.FunctionCallExpr).Function.(*ir.Name)
	return ok
}

func (p *FunctionProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
