package providers

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
)

// ConstantProvider resolves the definition of constant fetches.
type ConstantProvider struct{}

func NewConstant() *ConstantProvider {
	return &ConstantProvider{}
}

func (c *ConstantProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindConstFetchExpr
}

func (c *ConstantProvider) Define(ctx context.Context) (*definition.Definition, error) {
	result, err := index.
		FromContainer().
		Find(`\`+ctx.Current().(*ir.ConstFetchExpr).Constant.Value, ir.KindConstantStmt)
	if err != nil {
		log.Println(err)
		return nil, definition.ErrNoDefinitionFound
	}

	return definition.TrieNodeToDef(result), nil
}
