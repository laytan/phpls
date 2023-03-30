package providers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/project/definition"
)

// NameProvider resolves the definition of a class-like name.
// This defines the name part of 'new Name()', 'Name::method()', 'extends Name' etc.
type NameProvider struct{}

func NewName() *NameProvider {
	return &NameProvider{}
}

func (p *NameProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	// TODO: ast.fullyqualified name, ast.relativename
	if kind != ast.TypeName {
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
		ctx.Current().(*ast.Name),
	)
	if !ok {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{definition.IndexNodeToDef(tdef)}, nil
}
