package context

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
)

type Ctx struct {
	start      *position.Position
	nodes      []ir.Node
	comment    *token.Token
	curr       int
	scope      ir.Node
	classScope ir.Node
}

func New(pos *position.Position) (*Ctx, error) {
	ctx := &Ctx{
		start: pos,
	}

	err := ctx.init()

	return ctx, err
}

// Whether the current node is wrapped by the given kind.
func (c *Ctx) WrappedBy(kind ir.NodeKind) bool {
	for i := c.curr - 1; i >= 0; i-- {
		if ir.GetNodeKind(c.nodes[i]) == kind {
			return true
		}
	}

	return false
}

// Whether the current node wraps the given node kind.
func (c *Ctx) Wraps(kind ir.NodeKind) bool {
	for i := c.curr + 1; i < len(c.nodes); i++ {
		if ir.GetNodeKind(c.nodes[i]) == kind {
			return true
		}
	}

	return false
}

// Whether the current node directly is wrapped by the given kind.
func (c *Ctx) DirectlyWrappedBy(kind ir.NodeKind) bool {
	return c.curr != 0 && ir.GetNodeKind(c.nodes[c.curr-1]) == kind
}

// Whether the current node directly wraps the given node kind.
func (c *Ctx) DirectlyWraps(kind ir.NodeKind) bool {
	return c.curr != len(c.nodes)-2 && ir.GetNodeKind(c.nodes[c.curr+1]) == kind
}

type CommentPosition int

// Returns the token of the comment if the cursor is in a comment.
// Otherwise nil.
// The second return value is the index of the cursor into the comment.
func (c *Ctx) InComment() (*token.Token, CommentPosition) {
	if c.comment == nil {
		return nil, 0
	}

	content := wrkspc.FromContainer().FContentOf(c.start.Path)
	cursorPos := position.LocToPos(content, c.start.Row, c.start.Col)
	return c.comment, CommentPosition(int(cursorPos) - c.comment.Position.StartPos)
}

// Position returns the position of the cursor.
func (c *Ctx) Position() *position.Position {
	return c.start
}

// The next scope of the current node, (function, root, method, class etc.).
func (c *Ctx) Scope() ir.Node {
	return c.scope
}

// The next scope of the current node, but only the class-like nodes.
func (c *Ctx) ClassScope() ir.Node {
	return c.classScope
}

// Advances to the next wrapped node,
// returning whether there was a node to advance to.
func (c *Ctx) Advance() bool {
	if c.curr == 0 {
		return false
	}

	c.curr--
	c.setScopes()

	return true
}

// The current node and whether there is a current node.
func (c *Ctx) Current() ir.Node {
	return c.nodes[c.curr]
}

// Get the root/top most node.
func (c *Ctx) Root() *ir.Root {
	return c.nodes[0].(*ir.Root)
}

func (c *Ctx) Path() string {
	return c.Position().Path
}

func (c *Ctx) Start() *position.Position {
	return c.start
}

// Go up the wrappings, finding the first scope and first class scope.
// defaulting to root.
func (c *Ctx) setScopes() {
	c.scope = c.Root()
	c.classScope = c.Root()

	var foundScope, foundClassScope bool
	for i := c.curr - 1; i >= 0; i-- {
		if foundScope && foundClassScope {
			return
		}

		n := c.nodes[i]
		if !foundScope && nodescopes.IsScope(ir.GetNodeKind(n)) {
			c.scope = n
			foundScope = true
		}

		if !foundClassScope && nodescopes.IsClassLike(ir.GetNodeKind(n)) {
			c.classScope = n
			foundClassScope = true
		}
	}
}

func (c *Ctx) init() error {
	content, root, err := wrkspc.FromContainer().AllOf(c.start.Path)
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

	c.comment = nap.Comment

	return nil
}
