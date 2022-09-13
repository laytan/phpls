package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
)

// PropertyProvider resolves the definition of a property accessed like $a->property.
// Where $a can also be $this, $a->foo->bar()->property etc.
type PropertyProvider struct{}

func NewProperty() *PropertyProvider {
	return &PropertyProvider{}
}

func (p *PropertyProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindPropertyFetchExpr
}

// Recurse all the way down the .Variable, until we get to a *ir.SimpleVar,
// get the type of that variable,
// go back up resolving the types as we go.
func (p *PropertyProvider) Define(ctx context.Context) (*definition.Definition, error) {
	n := ctx.Current().(*ir.PropertyFetchExpr)

	return definition.WalkToProperty(ctx, n)
}
