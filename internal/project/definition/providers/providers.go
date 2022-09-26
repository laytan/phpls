package providers

import (
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/samber/do"
)

var (
	Index  = func() index.Index { return do.MustInvoke[index.Index](nil) }
	Typer  = func() typer.Typer { return do.MustInvoke[typer.Typer](nil) }
	Wrkspc = func() wrkspc.Wrkspc { return do.MustInvoke[wrkspc.Wrkspc](nil) }
)

func DefineExpr(ctx context.Context) (*definition.Definition, error) {
	if res, _, left := expr.Resolve(ctx.Current(), definition.ContextToScopes(ctx)); left == 0 &&
		res != nil {
		return &definition.Definition{
			Path: res.Path,
			Node: symbol.New(res.Node),
		}, nil
	}

	return nil, definition.ErrNoDefinitionFound
}
