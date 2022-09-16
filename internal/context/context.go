package context

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/samber/do"
)

var Wrkspc = func() wrkspc.Wrkspc { return do.MustInvoke[wrkspc.Wrkspc](nil) }

type Context interface {
	// Whether the current node is wrapped by the given kind.
	WrappedBy(kind ir.NodeKind) bool
	// Whether the current node wraps the given node kind.
	Wraps(kind ir.NodeKind) bool

	// Whether the current node directly is wrapped by the given kind.
	DirectlyWrappedBy(kind ir.NodeKind) bool
	// Whether the current node directly wraps the given node kind.
	DirectlyWraps(kind ir.NodeKind) bool

	// The next scope of the current node, (function, root, method, class etc.).
	Scope() ir.Node
	// The next scope of the current node, but only the class-like nodes.
	ClassScope() ir.Node

	// Advances to the next wrapped node,
	// returning whether there was a node to advance to.
	Advance() bool

	// The current node and whether there is a current node.
	Current() ir.Node

	// Get the root/top most node.
	Root() *ir.Root

	Start() *position.Position
}

type context struct {
	start      *position.Position
	nodes      []ir.Node
	curr       int
	scope      ir.Node
	classScope ir.Node
}

func (c *context) WrappedBy(kind ir.NodeKind) bool {
	for i := c.curr - 1; i >= 0; i-- {
		if ir.GetNodeKind(c.nodes[i]) == kind {
			return true
		}
	}

	return false
}

func (c *context) Wraps(kind ir.NodeKind) bool {
	for i := c.curr + 1; i < len(c.nodes); i++ {
		if ir.GetNodeKind(c.nodes[i]) == kind {
			return true
		}
	}

	return false
}

func (c *context) DirectlyWrappedBy(kind ir.NodeKind) bool {
	return c.curr != 0 && ir.GetNodeKind(c.nodes[c.curr-1]) == kind
}

func (c *context) DirectlyWraps(kind ir.NodeKind) bool {
	return c.curr != len(c.nodes)-2 && ir.GetNodeKind(c.nodes[c.curr+1]) == kind
}

func (c *context) Scope() ir.Node {
	return c.scope
}

func (c *context) ClassScope() ir.Node {
	return c.classScope
}

func (c *context) Advance() bool {
	if c.curr == 0 {
		return false
	}

	c.curr--
	c.setScopes()

	return true
}

func (c *context) Current() ir.Node {
	return c.nodes[c.curr]
}

func (c *context) Root() *ir.Root {
	return c.nodes[0].(*ir.Root)
}

func (c *context) Start() *position.Position {
	return c.start
}

func (c *context) init() error {
	content, root, err := Wrkspc().AllOf(c.start.Path)
	if err != nil {
		return fmt.Errorf(
			"Unable to parse context of %s: %w",
			c.start.Path,
			err,
		)
	}

	apos := position.LocToPos(content, c.start.Row, c.start.Col)
	nap := traversers.NewNodeAtPos(apos)
	root.Walk(nap)

	c.nodes = nap.Nodes
	c.curr = len(nap.Nodes)
	c.Advance()

	return nil
}

// Go up the wrappings, finding the first scope and first class scope.
// defaulting to root.
func (c *context) setScopes() {
	c.scope = c.Root()
	c.classScope = c.Root()

	var foundScope, foundClassScope bool
	for i := c.curr - 1; i >= 0; i-- {
		if foundScope && foundClassScope {
			return
		}

		n := c.nodes[i]
		if !foundScope && symbol.IsScope(n) {
			c.scope = n
			foundScope = true
		}

		if !foundClassScope && symbol.IsClassLike(n) {
			c.classScope = n
			foundClassScope = true
		}
	}
}

func New(pos *position.Position) (Context, error) {
	ctx := &context{
		start: pos,
	}

	err := ctx.init()

	return ctx, err
}
