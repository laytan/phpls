package definition

import (
	"errors"
	"fmt"

	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/php-parser/pkg/position"
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
	Path       string
	Position   *position.Position
	Identifier string
}

func IndexNodeToDef(node *index.INode) *Definition {
	return &Definition{
		Path:       node.Path,
		Position:   node.Position,
		Identifier: node.Identifier,
	}
}

func ContextToScopes(ctx *context.Ctx) *expr.Scopes {
	return &expr.Scopes{
		Path:  ctx.Start().Path,
		Root:  ctx.Root(),
		Class: ctx.ClassScope(),
		Block: ctx.Scope(),
	}
}
