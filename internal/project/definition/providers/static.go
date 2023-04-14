package providers

import (
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/php-parser/pkg/ast"
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
