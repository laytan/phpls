package expr

import (
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/stack"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/samber/do"
)

type ExprType int

const (
	ExprTypeProperty ExprType = iota
	ExprTypeMethod
	ExprTypeVariable
	ExprTypeName
	ExprTypeStaticMethod
	ExprTypeFunction
)

var (
	Wrkspc = func() wrkspc.Wrkspc { return do.MustInvoke[wrkspc.Wrkspc](nil) }
	Index  = func() index.Index { return do.MustInvoke[index.Index](nil) }
	Typer  = func() typer.Typer { return do.MustInvoke[typer.Typer](nil) }
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

type Resolver interface {
	Down(node ir.Node) (resolvemenet *DownResolvement, next ir.Node, done bool)
}
type ClassResolver interface {
	Resolver
	Up(
		ctx *phpdoxer.TypeClassLike,
		toResolve *DownResolvement,
	) (nextCtx *phpdoxer.TypeClassLike, done bool)
}

type StartResolver interface {
	Resolver
	Up(
		scoes Scopes,
		toResolve *DownResolvement,
	) (nextCtx *phpdoxer.TypeClassLike, done bool)
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

func Resolve(ctx context.Context) *phpdoxer.TypeClassLike {
	symbols := stack.New[*DownResolvement]()
	Down(AllResolvers(), symbols, ctx.Current())

	what.Happens("Symbols: %v", symbols.String())

	// We don't give the ctx to the resolvers as the Current() is misleading,
	// because it points to the end node when this is the start node.
	scopes := Scopes{
		Path:  ctx.Start().Path,
		Root:  ctx.Root(),
		Class: ctx.ClassScope(),
		Block: ctx.Scope(),
	}

	start := symbols.Pop()
	what.Is(start)
	var next *phpdoxer.TypeClassLike
	for _, starter := range starters {
		if n, ok := starter.Up(scopes, start); ok {
			next = n
			break
		}
	}

	if next == nil {
		return nil
	}

	what.Is(next)

	for curr := symbols.Pop(); curr != nil; curr = symbols.Pop() {
		resolver := resolvers[curr.ExprType]
		n, ok := resolver.Up(next, curr)
		if !ok {
			return nil
		}

		next = n
	}

	return next
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
	// TODO: this does not work when you have a trait and class of the same name/namespace.
	sym, err := Index().Find(c.Name, symbol.ClassLikeScopes...)
	if err != nil {
		return nil, err
	}

	return resolvequeue.New(func(n *resolvequeue.Node) (*ir.Root, error) {
		res, err := Index().Find(n.FQN.String(), n.Kind)
		if err != nil {
			return nil, err
		}

		return Wrkspc().IROf(res.Path)
	}, &resolvequeue.Node{
		FQN:  typer.NewFQN(c.Name),
		Kind: sym.Symbol.NodeKind(),
	}), nil
}

// walkContext is the context of a current iteration in the walk of the resolve queue.
type walkContext struct {
	// The FQN of the current class.
	FQN *typer.FQN

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
		def, err := Index().Find(res.FQN.String(), res.Kind)
		if err != nil {
			return err
		}

		root, err := Wrkspc().IROf(def.Path)
		if err != nil {
			return err
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
	walker func(*walkContext) *phpdoxer.TypeClassLike,
) (*phpdoxer.TypeClassLike, bool) {
	queue, err := newResolveQueue(ctx)
	if err != nil {
		log.Println(err)
		return nil, false
	}

	var result *phpdoxer.TypeClassLike
	err = walkResolveQueue(queue, func(wc *walkContext) (done bool, err error) {
		res := walker(wc)
		if res != nil {
			result = res
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		log.Println(err)
		return nil, false
	}

	if result == nil {
		return nil, false
	}

	return result, true
}
