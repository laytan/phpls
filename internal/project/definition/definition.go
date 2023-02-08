package definition

import (
	"errors"
	"fmt"

	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/symbol"
)

const (
	ErrUnexpectedNodeFmt = "Unexpected node of type %T (expected %s) when retrieving symbol definition"
)

var (
	ErrUnexpectedNode    = fmt.Errorf(ErrUnexpectedNodeFmt, nil, "")
	ErrNoDefinitionFound = errors.New(
		"No definition found for symbol at given position",
	)
)

type Definition struct {
	Path string
	Node symbol.Symbol
}

func IndexNodeToDef(node *index.IndexNode) *Definition {
	return &Definition{
		Path: node.Path,
		Node: node.Symbol,
	}
}

func ContextToScopes(ctx context.Context) expr.Scopes {
	return expr.Scopes{
		Path:  ctx.Start().Path,
		Root:  ctx.Root(),
		Class: ctx.ClassScope(),
		Block: ctx.Scope(),
	}
}
