package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// ClassConstantProvider resolves the definition of a class constant, Foo::BAR.
type ClassConstantProvider struct{}

func NewClassConstant() *ClassConstantProvider {
	return &ClassConstantProvider{}
}

func (p *ClassConstantProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindClassConstFetchExpr
}

func (p *ClassConstantProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
