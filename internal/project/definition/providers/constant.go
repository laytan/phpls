package providers

import (
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/fqn"
)

// ConstantProvider resolves the definition of constant fetches.
type ConstantProvider struct{}

func NewConstant() *ConstantProvider {
	return &ConstantProvider{}
}

func (c *ConstantProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindConstFetchExpr
}

// TODO: return non-array
func (c *ConstantProvider) Define(ctx context.Context) ([]*definition.Definition, error) {
	key := fqn.New(fqn.PartSeperator + ctx.Current().(*ir.ConstFetchExpr).Constant.Value)
	result, ok := index.FromContainer().Find(key)
	if !ok {
		log.Println(
			fmt.Errorf("[providers.ConstantProvider.Define]: unable to find %s in index", key),
		)
		return nil, definition.ErrNoDefinitionFound
	}

	what.Is(result)

	return []*definition.Definition{definition.IndexNodeToDef(result)}, nil
}
