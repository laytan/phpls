package providers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/nodeident"
)

// ThisProvider resolves the definition of the current class scope for '$this'
// variables.
type ThisProvider struct{}

func NewThis() *ThisProvider {
	return &ThisProvider{}
}

func (p *ThisProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	if kind != ast.TypeExprVariable {
		return false
	}

	return nodeident.Get(ctx.Current()) == "$this"
}

// TODO: use DefineExpr.
// TODO: merge with variable provider.
func (p *ThisProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	cls := ctx.ClassScope()
	return []*definition.Definition{{
		Path:       ctx.Start().Path,
		Position:   cls.GetPosition(),
		Identifier: nodeident.Get(cls),
	}}, nil
}
