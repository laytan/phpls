package common

import (
	"errors"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/samber/do"
)

var Index = func() index.Index { return do.MustInvoke[index.Index](nil) }

func FullyQualify(root *ir.Root, name string) *typer.FQN {
	if strings.HasPrefix(name, `\`) {
		return typer.NewFQN(name)
	}

	t := typer.NewFQNTraverser()
	root.Walk(t)

	return t.ResultFor(&ir.Name{Value: name})
}

func FindFullyQualified(
	root *ir.Root,
	name string,
	kinds ...ir.NodeKind,
) (*traversers.TrieNode, bool) {
	FQN := FullyQualify(root, name)
	node, err := Index().Find(FQN.String(), kinds...)
	if err != nil {
		if !errors.Is(err, index.ErrNotFound) {
			log.Println(err)
		}

		return nil, false
	}

	return node, true
}
