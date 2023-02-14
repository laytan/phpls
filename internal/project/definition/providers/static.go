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

func (p *StaticProvider) CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool {
	return kind == ir.KindStaticCallExpr
}

func (p *StaticProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
