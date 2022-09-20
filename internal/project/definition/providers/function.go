package providers

import (
	"errors"
	"log"

	"appliedgo.net/what"
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

	if def := p.checkLocal(ctx, name); def != nil {
		return def, nil
	}

	if def := p.checkNamespaced(ctx, name); def != nil {
		return def, nil
	}

	if def := p.checkGlobal(name); def != nil {
		return def, nil
	}

	return nil, definition.ErrNoDefinitionFound
}

// Checks for local functions (defined inside other functions or constructs).
func (p *FunctionProvider) checkLocal(ctx context.Context, name string) *definition.Definition {
	if ir.GetNodeKind(ctx.Scope()) == ir.KindRoot {
		return nil
	}

	t := traversers.NewFunction(name)
	ctx.Scope().Walk(t)

	if t.Function == nil {
		return nil
	}

	return &definition.Definition{
		Path: ctx.Start().Path,
		Node: symbol.NewFunction(t.Function),
	}
}

// Check for other functions, defined in namespaces.
func (p *FunctionProvider) checkNamespaced(
	ctx context.Context,
	name string,
) *definition.Definition {
	def, ok := definition.FindFullyQualified(ctx.Root(), name, ir.KindFunctionStmt)
	if !ok {
		what.Happens("could not find namespaced function definition for %s", name)
		return nil
	}

	return def
}

// Check for global functions.
func (p *FunctionProvider) checkGlobal(name string) *definition.Definition {
	def, err := Index().Find(`\`+name, ir.KindFunctionStmt)
	if err != nil {
		what.Happens(err.Error())
		if !errors.Is(err, index.ErrNotFound) {
			log.Println(err)
		}

		return nil
	}

	return &definition.Definition{
		Path: def.Path,
		Node: def.Symbol,
	}
}
