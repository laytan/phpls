package fqner

import (
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor/traverser"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
)

func FullyQualifyName(root *ast.Root, name *ast.Name) *fqn.FQN {
	if strings.HasPrefix(nodeident.Get(name), `\`) {
		return fqn.New(nodeident.Get(name))
	}

	t := fqn.NewTraverser()
	tv := traverser.NewTraverser(t)
	root.Accept(tv)

	return t.ResultFor(name)
}

func FindFullyQualifiedName(root *ast.Root, name *ast.Name) (*index.INode, bool) {
	qualified := FullyQualifyName(root, name)
	return index.FromContainer().Find(qualified)
}

type rooter interface {
	Root() *ast.Root
}

type FullyQualifier struct {
	rooter rooter

	cached *fqn.FQN
	node   ast.Vertex
}

func New(rooter rooter, node ast.Vertex) *FullyQualifier {
	return &FullyQualifier{
		rooter: rooter,
		node:   node,
	}
}

func NewFromFQN(v *fqn.FQN) *FullyQualifier {
	return &FullyQualifier{
		cached: v,
	}
}

func (f *FullyQualifier) GetFQN() *fqn.FQN {
	if f.cached != nil {
		return f.cached
	}

	var name *ast.Name
	switch tn := f.node.(type) {
	case *ast.Name:
		name = tn
	default:
		name = &ast.Name{
			Position: f.node.GetPosition(),
			Parts: functional.Map(
				strings.Split(nodeident.Get(f.node), "\\"),
				func(s string) ast.Vertex { return &ast.NamePart{Value: []byte(s)} },
			),
		}
	}

	f.cached = FullyQualifyName(f.rooter.Root(), name)
	return f.cached
}
