package providers

import (
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/php-parser/pkg/ast"
)

// MethodProvider resolves the definition of a method call.
// $this->test(), $this->test->test(), $foo->bar() etc. for example.
type MethodProvider struct{}

func NewMethod() *MethodProvider {
	return &MethodProvider{}
}

func (p *MethodProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprMethodCall
}

func (p *MethodProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
