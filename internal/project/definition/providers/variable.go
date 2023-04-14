package providers

import (
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/traversers"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

// VariableProvider resolves the definition of a simple variable.
// It looks for it being assigned, as a parameter, or a global.
type VariableProvider struct{}

func NewVariable() *VariableProvider {
	return &VariableProvider{}
}

func (p *VariableProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprVariable
}

// TODO: use DefineExpr.
func (p *VariableProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	t := traversers.NewAssignment(ctx.Current().(*ast.ExprVariable))
	tt := traverser.NewTraverser(t)
	ctx.Scope().Accept(tt)

	if t.Assignment == nil {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{{
		Path:       ctx.Start().Path,
		Position:   t.Assignment.Position,
		Identifier: nodeident.Get(t.Assignment),
	}}, nil
}
