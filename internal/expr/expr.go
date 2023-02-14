package expr

import (
	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/stack"
)

type Type int

const (
	TypeProperty Type = iota
	TypeMethod
	TypeVariable
	TypeName
	TypeStaticMethod
	TypeFunction
	TypeNew
	TypeClassConstant
)

type Scopes struct {
	Path  string
	Root  *ir.Root
	Class ir.Node
	Block ir.Node
}

type DownResolvement struct {
	ExprType   Type
	Identifier string
	Position   *position.Position
}

type UpResolvement struct {
	ExprType   Type
	Identifier string
	Class      *phpdoxer.TypeClassLike
}

type Resolver interface {
	Down(node ir.Node) (resolvement *DownResolvement, next ir.Node, done bool)
}
type ClassResolver interface {
	Resolver
	Up(
		ctx *fqn.FQN,
		privacy phprivacy.Privacy,
		toResolve *DownResolvement,
	) (result *Resolved, nextCtx *fqn.FQN, done bool)
}

type StartResolver interface {
	Resolver
	Up(
		scopes *Scopes,
		toResolve *DownResolvement,
	) (result *Resolved, nextCtx *fqn.FQN, privacy phprivacy.Privacy, done bool)
}

func AllResolvers() *map[Type]Resolver {
	all := make(map[Type]Resolver, len(resolvers)+len(starters))
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

func Resolve(
	node ir.Node,
	scopes *Scopes,
) (result *Resolved, lastClass *fqn.FQN, left int) {
	symbols := stack.New[*DownResolvement]()
	Down(AllResolvers(), symbols, node)

	if symbols.Peek() == nil {
		return nil, nil, 0
	}

	start := symbols.Pop()
	var next *fqn.FQN
	var privacy phprivacy.Privacy
	for _, starter := range starters {
		what.Happens("Up: %T", starter)
		if r, n, p, ok := starter.Up(scopes, start); ok {
			result = r
			privacy = p
			next = n
			break
		}
	}

	if next == nil {
		if result != nil {
			// Run out the stack, to see how many were left to do.
			left = 0
			for curr := symbols.Pop(); curr != nil; curr = symbols.Pop() {
				left++
			}

			return result, nil, left
		}

		return nil, nil, 0
	}

	for curr := symbols.Pop(); curr != nil; curr = symbols.Pop() {
		if next == nil {
			return result, nil, symbols.Length() + 1
		}

		resolver := resolvers[curr.ExprType]
		what.Happens("Up: %T", resolver)
		res, n, ok := resolver.Up(next, privacy, curr)
		if !ok {
			return res, n, symbols.Length() + 1
		}

		result = res
		next = n
	}

	return result, next, 0
}

func Down(
	resolvers *map[Type]Resolver,
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
