package providers

import (
	"fmt"
	"log"

	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/php-parser/pkg/ast"
)

// ConstantProvider resolves the definition of constant fetches.
type ConstantProvider struct{}

func NewConstant() *ConstantProvider {
	return &ConstantProvider{}
}

func (c *ConstantProvider) CanDefine(ctx *context.Ctx, kind ast.Type) bool {
	return kind == ast.TypeExprConstFetch
}

// TODO: return non-array.
func (c *ConstantProvider) Define(ctx *context.Ctx) ([]*definition.Definition, error) {
	key := fqn.New(fqn.PartSeperator + nodeident.Get(ctx.Current().(*ast.ExprConstFetch).Const))
	result, ok := index.Current.Find(key)
	if !ok {
		log.Println(
			fmt.Errorf("[providers.ConstantProvider.Define]: unable to find %s in index", key),
		)
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{definition.IndexNodeToDef(result)}, nil
}
