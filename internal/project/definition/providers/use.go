package providers

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/fqn"
)

// UseProvider resolves the definition of 'use Foo/Bar/FooBar' statements.
// It is a separate provider than the NameProvider because the NameProvider
// will resolve any alias, in this case we want the source of the alias instead.
type UseProvider struct{}

func NewUse() *UseProvider {
	return &UseProvider{}
}

func (p *UseProvider) CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool {
	return kind == ir.KindName && ctx.DirectlyWrappedBy(ir.KindUseStmt)
}

func (p *UseProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	key := fqn.New(fqn.PartSeperator + ctx.Current().(*ir.Name).Value)
	res, ok := index.FromContainer().Find(key)
	if !ok {
		log.Println(fmt.Errorf("[providers.UseProvider.Define]: can't find %s in index", key))
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{{
		Path:       res.Path,
		Position:   res.Position,
		Identifier: res.Identifier,
	}}, nil
}
