package providers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/fqner"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/nodescopes"
)

// NameProvider resolves the definition of a class-like name.
// This defines the name part of 'new Name()', 'Name::method()', 'extends Name' etc.
type NameProvider struct{}

func NewName() *NameProvider {
	return &NameProvider{}
}

func (p *NameProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	if kind == ast.TypeStmtClass || kind == ast.TypeStmtInterface || kind == ast.TypeStmtTrait {
		return true
	}

	if !nodescopes.IsName(kind) {
		return false
	}

	// If in a function, don't define it.
	if ctx.DirectlyWrappedBy(ast.TypeExprFunctionCall) {
		return false
	}

	// If a constant fetch, don't define it.
	if ctx.DirectlyWrappedBy(ast.TypeExprConstFetch) {
		return false
	}

	return true
}

// TODO: use DefineExpr.
func (p *NameProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	tdef, ok := fqner.FindFullyQualifiedName(
		ctx.Root(),
		ctx.Current(),
	)
	if !ok {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{definition.IndexNodeToDef(tdef)}, nil
}
