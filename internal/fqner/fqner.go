package fqner

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/fqn"
)

func FullyQualifyName(root *ir.Root, name *ir.Name) *fqn.FQN {
	if strings.HasPrefix(name.Value, `\`) {
		return fqn.New(name.Value)
	}

	t := fqn.NewTraverser()
	root.Walk(t)

	return t.ResultFor(name)
}

func FindFullyQualifiedName(root *ir.Root, name *ir.Name) (*index.INode, bool) {
	qualified := FullyQualifyName(root, name)
	return index.FromContainer().Find(qualified)
}

type rooter interface {
	Root() *ir.Root
}

type FullyQualifier struct {
	rooter rooter

	cached *fqn.FQN
	node   ir.Node
}

func New(rooter rooter, node ir.Node) *FullyQualifier {
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

	var name *ir.Name
	switch typedNode := f.node.(type) {
	case *ir.ClassStmt:
		name = &ir.Name{
			Value:    typedNode.ClassName.Value,
			Position: typedNode.ClassName.Position,
		}
	case *ir.InterfaceStmt:
		name = &ir.Name{
			Value:    typedNode.InterfaceName.Value,
			Position: typedNode.InterfaceName.Position,
		}
	case *ir.TraitStmt:
		name = &ir.Name{
			Value:    typedNode.TraitName.Value,
			Position: typedNode.TraitName.Position,
		}
	case *ir.Identifier:
		name = &ir.Name{
			Value:    typedNode.Value,
			Position: typedNode.Position,
		}
	case *ir.Name:
		name = typedNode
	default:
		return nil
	}

	f.cached = FullyQualifyName(f.rooter.Root(), name)
	return f.cached
}
