package resolvequeue

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/queue"
)

type Node struct {
	FQN  *fqn.FQN
	Kind ir.NodeKind
}

type ResolveQueue struct {
	rootRetriever func(*Node) (*ir.Root, error)

	Queue *queue.Queue[*Node]

	// Order doesnt really matter so don't really nead a queue.
	Implements []*Node
}

func New(rootRetriever func(*Node) (*ir.Root, error), node *Node) *ResolveQueue {
	r := &ResolveQueue{
		rootRetriever: rootRetriever,
		Queue:         queue.New[*Node](),
		Implements:    []*Node{},
	}

	r.parse(node)

	return r
}

func (r *ResolveQueue) Resolve(
	node *Node,
) (uses []*Node, extends []*Node, implements []*Node) {
	root, err := r.rootRetriever(node)
	if err != nil {
		log.Println(fmt.Errorf("ResolveQueue.Resolve error during parsing: %w", err))
		return nil, nil, nil
	}

	traverser := NewResolver(node)
	root.Walk(traverser)

	uses = make([]*Node, 0, len(traverser.Uses))
	extends = make([]*Node, 0, len(traverser.Extends))
	implements = make([]*Node, 0, len(traverser.Implements))

	fqnTraverser := fqn.NewFQNTraverser()
	root.Walk(fqnTraverser)

	for _, use := range traverser.Uses {
		uses = append(uses, &Node{FQN: fqnTraverser.ResultFor(use), Kind: ir.KindTraitStmt})
	}

	for _, extend := range traverser.Extends {
		extends = append(
			extends,
			&Node{FQN: fqnTraverser.ResultFor(extend), Kind: ir.KindClassStmt},
		)
	}

	for _, implement := range traverser.Implements {
		implements = append(implements, &Node{
			FQN:  fqnTraverser.ResultFor(implement),
			Kind: ir.KindInterfaceStmt,
		})
	}

	return uses, extends, implements
}

func (r *ResolveQueue) parse(node *Node) {
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

func (r *ResolveQueue) insert(node *Node) {
	if node.Kind == ir.KindClassImplementsStmt {
		r.Implements = append(r.Implements, node)
		return
	}

	r.Queue.Enqueue(node)
}
