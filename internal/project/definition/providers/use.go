package providers

import (
	"fmt"
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
)

// UseProvider resolves the definition of 'use Foo/Bar/FooBar' statements.
// It is a separate provider than the NameProvider because the NameProvider
// will resolve any alias, in this case we want the source of the alias instead.
type UseProvider struct{}

func NewUse() *UseProvider {
	return &UseProvider{}
}

func (p *UseProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeName && ctx.DirectlyWrappedBy(ast.TypeStmtUse)
}

func (p *UseProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	key := fqn.New(fqn.PartSeperator + nodeident.Get(ctx.Current()))
	res, ok := index.Current.Find(key)
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
