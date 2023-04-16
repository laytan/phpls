package fqner

import (
	"strings"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/position"
)

func FullyQualifyName(root *ast.Root, name ast.Vertex) *fqn.FQN {
	ident := nodeident.Get(name)
	if strings.HasPrefix(ident, `\`) {
		return fqn.New(ident)
	}

	t := fqn.NewTraverser()
	tv := traverser.NewTraverser(t)
	root.Accept(tv)

	return t.ResultFor(name)
}

func FindFullyQualifiedName(root *ast.Root, name ast.Vertex) (*index.INode, bool) {
	qualified := FullyQualifyName(root, name)
	return index.Current.Find(qualified)
}

// Returns whether the file at given pos needs a use statement for the given fqn.
func NeedsUseStmtFor(pos *position.Position, name *fqn.FQN) bool {
	content, root := wrkspc.Current.AllF(pos.Path)
	parts := strings.Split(name.String(), `\`)
	className := parts[len(parts)-1]

	// Get how it would be resolved in the current file state.
	actFQN := FullyQualifyName(root, &ast.Name{
		Position: pos.ToIRPosition(content),
		Parts:    []ast.Vertex{&ast.NamePart{Value: []byte(className)}},
	})

	// If the resolvement in current state equals the wanted fqn, no use stmt is needed.
	return actFQN.String() != name.String()
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
