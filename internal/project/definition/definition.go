package definition

import (
	"errors"
	"fmt"

	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
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

func TrieNodeToDef(node *traversers.TrieNode) *Definition {
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

// func ResolveExprToScopeOfDefinition(ctx context.Context) (*expr.UpResolvement, bool) {
// 	path, left := expr.Resolve(ctx.Current(), ContextToScopes(ctx))
// 	if left > 1 {
// 		what.Happens("could not resolve the expression enough")
// 		return nil, false
// 	}
//
// 	what.Func()
// 	what.Is(path)
// 	what.Is(left)
//
// 	return path[len(path)-1], true
// }
