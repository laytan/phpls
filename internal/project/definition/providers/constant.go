package providers

import (
	"fmt"
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
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
	result, ok := index.FromContainer().Find(key)
	if !ok {
		log.Println(
			fmt.Errorf("[providers.ConstantProvider.Define]: unable to find %s in index", key),
		)
		return nil, definition.ErrNoDefinitionFound
	}

	return []*definition.Definition{definition.IndexNodeToDef(result)}, nil
}
