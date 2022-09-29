package providers

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/symbol"
)

// UseProvider resolves the definition of 'use Foo/Bar/FooBar' statements.
// It is a separate provider than the NameProvider because the NameProvider
// will resolve any alias, in this case we want the source of the alias instead.
type UseProvider struct{}

func NewUse() *UseProvider {
	return &UseProvider{}
}

func (p *UseProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindName && ctx.DirectlyWrappedBy(ir.KindUseStmt)
}

func (p *UseProvider) Define(ctx context.Context) (*definition.Definition, error) {
	fqn := `\` + ctx.Current().(*ir.Name).Value
	res, err := index.FromContainer().Find(fqn, symbol.ClassLikeScopes...)
	if err != nil {
		log.Println(err)
		return nil, definition.ErrNoDefinitionFound
	}

	return &definition.Definition{
		Path: res.Path,
		Node: res.Symbol,
	}, nil
}
