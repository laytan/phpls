package providers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// PropertyProvider resolves the definition of a property accessed like $a->property.
// Where $a can also be $this, $a->foo->bar()->property etc.
type PropertyProvider struct{}

func NewProperty() *PropertyProvider {
	return &PropertyProvider{}
}

func (p *PropertyProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprPropertyFetch
}

func (p *PropertyProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	return DefineExpr(ctx)
}
