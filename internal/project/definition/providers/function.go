package providers

import (
	"errors"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
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

func (p *FunctionProvider) Define(ctx context.Context) (*definition.Definition, error) {
	// No error checking needed, this is all validated in the CanDefine above.
	n := ctx.Current().(*ir.FunctionCallExpr)
	name := n.Function.(*ir.Name).Value

	if def := p.checkLocal(ctx.Scope(), ctx.Start().Path, name); def != nil {
		return def, nil
	}

	if def := p.checkGlobal(ctx, name); def != nil {
		return def, nil
	}

	return nil, definition.ErrNoDefinitionFound
}

func (p *FunctionProvider) checkLocal(
	scope ir.Node,
	path string,
	name string,
) *definition.Definition {
	t := traversers.NewFunction(name)
	scope.Walk(t)

	if t.Function == nil {
		return nil
	}

	return &definition.Definition{
		Path: path,
		Node: symbol.NewFunction(t.Function),
	}
}

func (p *FunctionProvider) checkGlobal(ctx context.Context, name string) *definition.Definition {
	n, err := ctx.Index().Find(`\`+name, ir.KindFunctionStmt)
	if err != nil {
		if !errors.Is(err, index.ErrNotFound) {
			log.Println(err)
		}

		return nil
	}

	return &definition.Definition{
		Node: n.Symbol,
		Path: n.Path,
	}
}
