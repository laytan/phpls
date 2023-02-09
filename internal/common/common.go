package common

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

func FullyQualify(root *ir.Root, name string) *fqn.FQN {
	if strings.HasPrefix(name, `\`) {
		return fqn.New(name)
	}

	t := fqn.NewTraverser()
	root.Walk(t)

	return t.ResultFor(&ir.Name{Value: name})
}

func FindFullyQualified(
	root *ir.Root,
	name string,
	kinds ...ir.NodeKind,
) (*index.INode, bool) {
	FQN := FullyQualify(root, name)
	return index.FromContainer().Find(FQN)
}

// MapFilter applies the mapper function to each entry in the slice and returns a new
// slice with the results.
//
// If mapper returns the default value for the R type, the result is not added
// to the returned slice.
func MapFilter[V comparable, R comparable](slice []V, mapper func(entry V) R) []R {
	res := make([]R, 0, len(slice))
	var defaultVal R

	for _, v := range slice {
		if mapped := mapper(v); mapped != defaultVal {
			res = append(res, mapped)
		}
	}

	return res
}

// Map applies the mapper function to each entry in the slice and returns a new
// slice with the results.
func Map[V any, R any](slice []V, mapper func(entry V) R) []R {
	res := make([]R, 0, len(slice))
	for _, v := range slice {
		res = append(res, mapper(v))
	}

	return res
}

func SymbolToNode(path string, sym symbol.Symbol) (*ir.Root, ir.Node, error) {
	root, err := wrkspc.FromContainer().IROf(path)
	if err != nil {
		return nil, nil, fmt.Errorf("[common.SymbolToNode]: %w", err)
	}

	symToNode := traversers.NewSymbolToNode(sym)
	root.Walk(symToNode)

	if symToNode.Result == nil {
		return nil, nil, fmt.Errorf(
			"[common.SymbolToNode]: Unable to find node matching symbol %v in path %s",
			sym,
			path,
		)
	}

	return root, symToNode.Result, nil
}
