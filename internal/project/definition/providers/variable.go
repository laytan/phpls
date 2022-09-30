package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

// VariableProvider resolves the definition of a simple variable.
// It looks for it being assigned, as a parameter, or a global.
type VariableProvider struct{}

func NewVariable() *VariableProvider {
	return &VariableProvider{}
}

func (p *VariableProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindSimpleVar
}

// TODO: use DefineExpr.
func (p *VariableProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	t := traversers.NewAssignment(ctx.Current().(*ir.SimpleVar))
	ctx.Scope().Walk(t)

	if t.Assignment == nil {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{{
		Path: ctx.Start().Path,
		Node: symbol.NewAssignment(t.Assignment),
	}}, nil
}
