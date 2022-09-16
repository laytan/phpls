package resolvequeue

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewResolver(target *Node) *Resolver {
	return &Resolver{
		Uses:       []*ir.Name{},
		Extends:    []*ir.Name{},
		Implements: []*ir.Name{},
		target:     target,
	}
}

// Resolver implements ir.Visitor.
// Resolver retrieves the Trait Uses, Extended and implemented names of the target class.
type Resolver struct {
	Uses       []*ir.Name
	Extends    []*ir.Name
	Implements []*ir.Name

	target        *Node
	currNamespace string
}

func (r *Resolver) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.Root:
		return true

	case *ir.NamespaceStmt:
		r.currNamespace = typedNode.NamespaceName.Value
		return r.currNamespace == r.target.FQN.Namespace()

	default:
		// Don't go into scopes that are not necessary.
		if symbol.IsNonClassLikeScope(node) {
			return false
		}

		// Don't go into classes that don't match target.
		if symbol.IsClassLike(node) {
			if ir.GetNodeKind(node) != r.target.Kind {
				return false
			}

			if symbol.GetIdentifier(node) != r.target.FQN.Name() {
				return false
			}
		}

		switch typedNode := node.(type) {
		case *ir.TraitUseStmt:
			r.Uses = append(r.Uses, r.toNames(typedNode.Traits)...)
			return false

		case *ir.ClassExtendsStmt:
			r.Extends = append(r.Extends, typedNode.ClassName)
			return false

		case *ir.ClassImplementsStmt:
			r.Implements = append(r.Implements, r.toNames(typedNode.InterfaceNames)...)
			return false

		case *ir.InterfaceExtendsStmt:
			r.Implements = append(r.Implements, r.toNames(typedNode.InterfaceNames)...)
			return false

		default:
			return true
		}
	}
}

func (r *Resolver) LeaveNode(_ ir.Node) {
}

func (r *Resolver) toNames(nodes []ir.Node) []*ir.Name {
	names := make([]*ir.Name, 0, len(nodes))

	for _, trait := range nodes {
		name, ok := trait.(*ir.Name)
		if !ok {
			log.Printf("Resolver.toNames: expected type %T to be ir.Name\n", trait)
			continue
		}

		names = append(names, name)
	}

	return names
}
