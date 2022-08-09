package resolvequeue

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/queue"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/typer"
	log "github.com/sirupsen/logrus"
)

type RootRetriever interface {
	RetrieveRoot(*Node) (*ir.Root, error)
}

type Node struct {
	FQN  *typer.FQN
	Kind ir.NodeKind
}

func New(rootRetriever RootRetriever, node *Node) *resolveQueue {
	r := &resolveQueue{
		rootRetriever: rootRetriever,

		// OPTIM: migth want to create a basic queue ourself, we are now pulling in this package, only for a queue.
		Queue:      queue.New[*Node](),
		Implements: []*Node{},
	}

	r.parse(node)

	return r
}

type resolveQueue struct {
	rootRetriever RootRetriever

	Queue *queue.Queue[*Node]

	// Order doesnt really matter so don't really nead a queue.
	Implements []*Node
}

func (r *resolveQueue) parse(node *Node) {
	r.insert(node)

	uses, extends, implements := r.Resolve(node)

	for _, use := range uses {
		r.parse(use)
	}

	for _, extend := range extends {
		r.parse(extend)
	}

	for _, implement := range implements {
		r.parse(implement)
	}
}

func (r *resolveQueue) insert(node *Node) {
	if node.Kind == ir.KindClassImplementsStmt {
		r.Implements = append(r.Implements, node)
		return
	}

	r.Queue.Enqueue(node)
}

func (r *resolveQueue) Resolve(
	node *Node,
) ([]*Node, []*Node, []*Node) {
	root, err := r.rootRetriever.RetrieveRoot(node)
	if err != nil {
		log.Error(fmt.Errorf("ResolveQueue.Resolve error during parsing: %w", err))
		return nil, nil, nil
	}

	traverser := NewResolver(node)
	root.Walk(traverser)

	var (
		uses       = make([]*Node, len(traverser.Uses))
		extends    = make([]*Node, len(traverser.Extends))
		implements = make([]*Node, len(traverser.Implements))
	)

	fqnTraverser := typer.NewFQNTraverser()
	root.Walk(fqnTraverser)

	for i, use := range traverser.Uses {
		uses[i] = &Node{FQN: fqnTraverser.ResultFor(use), Kind: ir.KindTraitStmt}
	}

	for i, extend := range traverser.Extends {
		extends[i] = &Node{FQN: fqnTraverser.ResultFor(extend), Kind: ir.KindClassStmt}
	}

	for i, implement := range traverser.Implements {
		implements[i] = &Node{
			FQN: fqnTraverser.ResultFor(implement),
			// NOTE: this can also be ir.KindInterfaceExtendsStmt but we do the same thing with them.
			Kind: ir.KindClassImplementsStmt,
		}
	}

	return uses, extends, implements
}

func NewResolver(target *Node) *resolver {
	return &resolver{
		Uses:       []*ir.Name{},
		Extends:    []*ir.Name{},
		Implements: []*ir.Name{},
		target:     target,
	}
}

// resolver implements ir.Visitor.
// resolver retrieves the Trait Uses, Extended and implemented names of the target class.
type resolver struct {
	Uses       []*ir.Name
	Extends    []*ir.Name
	Implements []*ir.Name

	target        *Node
	currNamespace string
}

func (r *resolver) EnterNode(node ir.Node) bool {
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

func (r *resolver) LeaveNode(_ ir.Node) {
}

func (r *resolver) toNames(nodes []ir.Node) []*ir.Name {
	names := make([]*ir.Name, len(nodes))

	for i, trait := range nodes {
		name, ok := trait.(*ir.Name)
		if !ok {
			log.Errorf("Resolver.toNames: expected type %T to be ir.Name\n", trait)
			continue
		}

		names[i] = name
	}

	return names
}
