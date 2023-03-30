package providers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// StaticProvider provides definitions for static method calls like Foo::bar().
type StaticProvider struct{}

func NewStatic() *StaticProvider {
	return &StaticProvider{}
}

func (p *StaticProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprStaticCall
}

func (p *StaticProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
