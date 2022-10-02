package expr

import (
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/stack"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

type ExprType int //nolint:revive // Type is not really descriptive and a reserved word in lowercase.

const (
	ExprTypeProperty ExprType = iota
	ExprTypeMethod
	ExprTypeVariable
	ExprTypeName
	ExprTypeStaticMethod
	ExprTypeFunction
	ExprTypeNew
)

type Scopes struct {
	Path  string
	Root  *ir.Root
	Class ir.Node
	Block ir.Node
}

type DownResolvement struct {
	ExprType   ExprType
	Identifier string
}

type UpResolvement struct {
	ExprType   ExprType
	Identifier string
	Class      *phpdoxer.TypeClassLike
}

type Resolver interface {
	Down(node ir.Node) (resolvement *DownResolvement, next ir.Node, done bool)
}
type ClassResolver interface {
	Resolver
	Up(
		ctx *phpdoxer.TypeClassLike,
		privacy phprivacy.Privacy,
		toResolve *DownResolvement,
	) (result *Resolved, nextCtx *phpdoxer.TypeClassLike, done bool)
}

type StartResolver interface {
	Resolver
	Up(
		scopes Scopes,
		toResolve *DownResolvement,
	) (result *Resolved, nextCtx *phpdoxer.TypeClassLike, privacy phprivacy.Privacy, done bool)
}

func AllResolvers() *map[ExprType]Resolver {
	all := make(map[ExprType]Resolver, len(resolvers)+len(starters))
	for exprT, resolver := range resolvers {
		all[exprT] = resolver
	}

	for exprT, resolver := range starters {
		all[exprT] = resolver
	}

	return &all
}

type Resolved struct {
	Node ir.Node
	Path string
}

// TODO: accept scopes as a pointer.
func Resolve(
	node ir.Node,
	scopes Scopes,
) (result *Resolved, lastClass *phpdoxer.TypeClassLike, left int) {
	symbols := stack.New[*DownResolvement]()
	Down(AllResolvers(), symbols, node)

	what.Happens("Symbols: %v", symbols.String())

	if symbols.Peek() == nil {
		return nil, nil, 0
	}

	start := symbols.Pop()
	var res *Resolved
	var next *phpdoxer.TypeClassLike
	var privacy phprivacy.Privacy
	for _, starter := range starters {
		if r, n, p, ok := starter.Up(scopes, start); ok {
			res = r
			privacy = p
			next = n
			break
		}
	}

	if next == nil {
		if res != nil {
			// Run out the stack, to see how many were left to do.
			left = 0
			for curr := symbols.Pop(); curr != nil; curr = symbols.Pop() {
				left++
			}

			return res, nil, left
		}

		return nil, nil, 0
	}

	for curr := symbols.Pop(); curr != nil; curr = symbols.Pop() {
		resolver := resolvers[curr.ExprType]
		res, n, ok := resolver.Up(next, privacy, curr)
		if !ok && n != nil {
			// Run out the stack, to see how many were left to do.
			left = 1
			for curr = symbols.Pop(); curr != nil; curr = symbols.Pop() {
				left++
			}

			return res, n, left
		}

		next = n
		result = res
	}

	return result, next, 0
}

func Down(
	resolvers *map[ExprType]Resolver,
	symbols *stack.Stack[*DownResolvement],
	current ir.Node,
) {
	what.Happens("Down: %T", current)
	for _, resolver := range *resolvers {
		if resolvement, next, ok := resolver.Down(current); ok {
			symbols.Push(resolvement)

			if next == nil {
				return
			}

			Down(resolvers, symbols, next)
			break
		}
	}
}

func Up(symbols *stack.Stack[*DownResolvement], startClassScope, startScope ir.Node) {
}

func newResolveQueue(c *phpdoxer.TypeClassLike) (*resolvequeue.ResolveQueue, error) {
	sym, err := index.FromContainer().Find(c.Name, symbol.ClassLikeScopes...)
	if err != nil {
		return nil, fmt.Errorf("newResolveQueue(%s): %w", c, err)
	}

	return resolvequeue.New(rootRetriever, &resolvequeue.Node{
		FQN:  fqn.New(c.Name),
		Kind: sym.Symbol.NodeKind(),
	}), nil
}

// walkContext is the context of a current iteration in the walk of the resolve queue.
type walkContext struct {
	// The FQN of the current class.
	FQN *fqn.FQN

	// The definition of the current class.
	Curr *traversers.TrieNode

	// The root of the current class's file.
	Root *ir.Root
}

func walkResolveQueue(
	queue *resolvequeue.ResolveQueue,
	walker func(*walkContext) (done bool, err error),
) error {
	for res := queue.Queue.Dequeue(); res != nil; res = queue.Queue.Dequeue() {
		def, err := index.FromContainer().Find(res.FQN.String(), res.Kind)
		if err != nil {
			return fmt.Errorf("walkResolveQueue: index.Find(%s, %d): %w", res.FQN, res.Kind, err)
		}

		root, err := wrkspc.FromContainer().IROf(def.Path)
		if err != nil {
			return fmt.Errorf("walkResolveQueue: wrkspc.IROf(%s): %w", def.Path, err)
		}

		done, err := walker(
			&walkContext{
				FQN:  res.FQN,
				Curr: def,
				Root: root,
			},
		)
		if err != nil {
			return err
		}

		if done {
			break
		}
	}

	return nil
}

func createAndWalkResolveQueue(
	ctx *phpdoxer.TypeClassLike,
	walker func(*walkContext) (*Resolved, *phpdoxer.TypeClassLike),
) (*Resolved, *phpdoxer.TypeClassLike, bool) {
	queue, err := newResolveQueue(ctx)
	if err != nil {
		log.Println(err)
		return nil, nil, false
	}

	var resultNode *Resolved
	var resultClass *phpdoxer.TypeClassLike
	err = walkResolveQueue(queue, func(wc *walkContext) (done bool, err error) {
		rn, rc := walker(wc)
		if rn != nil {
			resultNode = rn
		}

		if rc != nil {
			resultClass = rc
		}

		return rn != nil || rc != nil, nil
	})
	if err != nil {
		log.Println(err)
		return nil, nil, false
	}

	if resultNode == nil && resultClass == nil {
		return nil, nil, false
	}

	return resultNode, resultClass, true
}

func rootRetriever(n *resolvequeue.Node) (*ir.Root, error) {
	res, err := index.FromContainer().Find(n.FQN.String(), n.Kind)
	if err != nil {
		return nil, fmt.Errorf("rootRetriever: index.Find(%s, %d): %w", n.FQN, n.Kind, err)
	}

	root, err := wrkspc.FromContainer().IROf(res.Path)
	if err != nil {
		return nil, fmt.Errorf("rootRetriever: wrkspc.IROf(%s): %w", res.Path, err)
	}

	return root, nil
}
