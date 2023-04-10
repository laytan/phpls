package fqner

import (
	"strings"

	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

func FullyQualifyName(root *ast.Root, name ast.Vertex) *fqn.FQN {
	if !nodescopes.IsName(name.GetType()) {
		panic(
			"FullyQualifyName can only be called with *ast.Name, *ast.NameFullyQualified or *ast.NameRelative",
		)
	}

	if strings.HasPrefix(nodeident.Get(name), `\`) {
		return fqn.New(nodeident.Get(name))
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
	content, root := wrkspc.Current.FAllOf(pos.Path)
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
