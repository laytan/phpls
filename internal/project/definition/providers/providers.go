package providers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/nodeident"
)

func DefineExpr(ctx *context.Ctx) ([]*definition.Definition, error) {
	if res, _, left := expr.Resolve(ctx.Current(), definition.ContextToScopes(ctx)); left == 0 &&
		res != nil {
		return []*definition.Definition{{
			Path:       res.Path,
			Position:   ir.GetPosition(res.Node),
			Identifier: nodeident.Get(res.Node),
		}}, nil
	}

	return nil, definition.ErrNoDefinitionFound
}
