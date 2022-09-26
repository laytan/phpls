package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/symbol"
)

// NameProvider resolves the definition of the a class-like name.
// This defines the name part of 'new Name()', 'Name::method()', 'extends Name' etc.
type NameProvider struct{}

func NewName() *NameProvider {
	return &NameProvider{}
}

func (p *NameProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	if kind != ir.KindName {
		return false
	}

	// If in a function, don't define it.
	if ctx.DirectlyWrappedBy(ir.KindFunctionCallExpr) {
		return false
	}

	return true
}

// TODO: use DefineExpr.
func (p *NameProvider) Define(ctx context.Context) (*definition.Definition, error) {
	tdef, ok := common.FindFullyQualified(
		ctx.Root(),
		ctx.Current().(*ir.Name).Value,
		symbol.ClassLikeScopes...)
	if !ok {
		return nil, definition.ErrNoDefinitionFound
	}

	return definition.TrieNodeToDef(tdef), nil
}
