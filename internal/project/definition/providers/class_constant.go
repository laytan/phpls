package providers

import (
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/php-parser/pkg/ast"
)

// ClassConstantProvider resolves the definition of a class constant, Foo::BAR.
type ClassConstantProvider struct{}

func NewClassConstant() *ClassConstantProvider {
	return &ClassConstantProvider{}
}

func (p *ClassConstantProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprClassConstFetch
}

func (p *ClassConstantProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
