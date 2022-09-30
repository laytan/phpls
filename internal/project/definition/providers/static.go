package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// StaticProvider provides definitions for static method calls like Foo::bar().
type StaticProvider struct{}

func NewStatic() *StaticProvider {
	return &StaticProvider{}
}

func (p *StaticProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindStaticCallExpr
}

func (p *StaticProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
