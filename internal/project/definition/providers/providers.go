package providers

import (
	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/expr"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/pkg/nodeident"
)

func DefineExpr(ctx *context.Ctx) ([]*definition.Definition, error) {
	if res, _, left := expr.Resolve(ctx.Current(), definition.ContextToScopes(ctx)); left == 0 &&
		res != nil {
		return []*definition.Definition{{
			Path:       res.Path,
			Position:   res.Node.GetPosition(),
			Identifier: nodeident.Get(res.Node),
		}}, nil
	}

	return nil, definition.ErrNoDefinitionFound
}
