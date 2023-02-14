package providers

import (
	"github.com/VKCOM/noverify/src/ir"
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

func (p *NameProvider) CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool {
	if kind != ir.KindName {
		return false
	}

	// If in a function, don't define it.
	if ctx.DirectlyWrappedBy(ir.KindFunctionCallExpr) {
		return false
	}

	// If a constant fetch, don't define it.
	if ctx.DirectlyWrappedBy(ir.KindConstFetchExpr) {
		return false
	}

	return true
}

// TODO: use DefineExpr.
func (p *NameProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	tdef, ok := fqner.FindFullyQualifiedName(
		ctx.Root(),
		ctx.Current().(*ir.Name),
	)
	if !ok {
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{definition.IndexNodeToDef(tdef)}, nil
}
