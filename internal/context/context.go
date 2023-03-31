package context

import (
	"fmt"

	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/token"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

type Ctx struct {
	start      *position.Position
	nodes      []ast.Vertex
	comment    *token.Token
	curr       int
	scope      ast.Vertex
	classScope ast.Vertex
}

func New(pos *position.Position) (*Ctx, error) {
	ctx := &Ctx{
		start: pos,
	}

	err := ctx.init()

	return ctx, err
}

// Whether the current node is wrapped by the given kind.
func (c *Ctx) WrappedBy(kind ast.Type) bool {
	for i := c.curr - 1; i >= 0; i-- {
		if c.nodes[i].GetType() == kind {
			return true
		}
	}

	return false
}

// Whether the current node wraps the given node kind.
func (c *Ctx) Wraps(kind ast.Type) bool {
	for i := c.curr + 1; i < len(c.nodes); i++ {
		if c.nodes[i].GetType() == kind {
			return true
		}
	}

	return false
}

// Whether the current node directly is wrapped by the given kind.
func (c *Ctx) DirectlyWrappedBy(kind ast.Type) bool {
	return c.curr != 0 && c.nodes[c.curr-1].GetType() == kind
}

// Whether the current node directly wraps the given node kind.
func (c *Ctx) DirectlyWraps(kind ast.Type) bool {
	return c.curr != len(c.nodes)-2 && c.nodes[c.curr+1].GetType() == kind
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
func (c *Ctx) Scope() ast.Vertex {
	return c.scope
}

// The next scope of the current node, but only the class-like nodes.
func (c *Ctx) ClassScope() ast.Vertex {
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
func (c *Ctx) Current() ast.Vertex {
	return c.nodes[c.curr]
}

// Get the root/top most node.
func (c *Ctx) Root() *ast.Root {
	return c.nodes[0].(*ast.Root)
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
		if !foundScope && nodescopes.IsScope(n.GetType()) {
			c.scope = n
			foundScope = true
		}

		if !foundClassScope && nodescopes.IsClassLike(n.GetType()) {
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
	nap := traversers.NewNodeAtPos(int(apos))
	tv := traverser.NewTraverser(nap)
	root.Accept(tv)

	c.nodes = nap.Nodes
	c.curr = len(nap.Nodes)
	c.Advance()

	c.comment = nap.Comment

	return nil
}
