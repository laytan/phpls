package throws

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/set"
	oldsym "github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

type rooter interface {
	Root() *ir.Root
	Path() string
}

type Throws struct {
	rooter

	doxed *symbol.Doxed

	node ir.Node
}

func NewResolver(root rooter, node ir.Node) *Throws {
	return &Throws{
		rooter: root,
		doxed:  symbol.NewDoxed(node),
		node:   node,
	}
}

func NewResolverFromIndex(n *index.INode) *Throws {
	rooter := wrkspc.NewRooter(n.Path)
	return NewResolver(rooter, oldsym.ToNode(rooter.Root(), n.Symbol))
}

func (t *Throws) Throws() []*fqn.FQN {
	return functional.Map(t.throws().Slice(), fqn.New)
}

func (t *Throws) throws() *set.Set[string] {
	thrownSet := set.New[string]()
	catchedSet := set.New[string]()

	switch t.node.(type) {
	case *ir.FunctionStmt, *ir.ClassMethodStmt:
		for _, throw := range t.phpDocThrows() {
			thrownSet.Add(throw.String())
		}
	}

	traverser := newThrowsTraverser()
	t.node.Walk(traverser)

	for _, result := range traverser.Result {
		switch typedRes := result.(type) {
		case *ir.TryStmt:
			blockThrows := &Throws{
				rooter: t.rooter,
				doxed:  symbol.NewDoxed(result),
				node:   result,
			}
			thrownSet = thrownSet.Union(blockThrows.throws())

		case *ir.CatchStmt:
			fqnt := fqn.NewTraverser()
			t.Root().Walk(fqnt)

			for _, catch := range typedRes.Types {
				switch typed := catch.(type) {
				case *ir.Name:
					catchedSet.Add(fqnt.ResultFor(typed).String())
				default:
					log.Printf("[throws.Throws]: Catch statement type has unexpected type: %v", typed)
				}
			}

		case *ir.ThrowStmt:
			resolvedRoot, resolvement, err := t.resolve(typedRes.Expr)
			if err != nil {
				log.Println(fmt.Errorf("[throws.Throws]: %w", err))
				continue
			}

			key := fqner.FullyQualifyName(resolvedRoot, &ir.Name{
				Position: ir.GetPosition(resolvement.Node),
				Value:    nodeident.Get(resolvement.Node),
			})
			thrownSet.Add(key.String())

		case *ir.FunctionCallExpr, *ir.MethodCallExpr:
			resolvedRoot, resolvement, err := t.resolve(result)
			if err != nil {
				log.Println(fmt.Errorf("[throws.throws]: %w", err))
				continue
			}

			blockThrows := &Throws{
				rooter: wrkspc.NewRooter(resolvement.Path, resolvedRoot),
				doxed:  symbol.NewDoxed(resolvement.Node),
				node:   resolvement.Node,
			}
			thrownSet = thrownSet.Union(blockThrows.throws())
		}
	}

	// Delete every thrown class that is caught by a catch.
	var toRemove []string
	for throw := range thrownSet.Iterator() {
		checker := t.catches(fqn.New(throw))

		for catch := range catchedSet.Iterator() {
			if checker(fqn.New(catch)) {
				// Can't remove directly because we are iterating over it.
				toRemove = append(toRemove, throw)
			}
		}
	}

	for _, v := range toRemove {
		thrownSet.Remove(v)
	}

	return thrownSet
}

func (t *Throws) phpDocThrows() (throws []*fqn.FQN) {
	var fqnt *fqn.Traverser

	results := t.doxed.FindAllDocs(symbol.FilterDocKind(phpdoxer.KindThrows))
	for _, result := range results {
		resultType := result.(*phpdoxer.NodeThrows).Type

		switch typedRes := resultType.(type) {
		case *phpdoxer.TypeClassLike:
			if typedRes.FullyQualified {
				throws = append(throws, fqn.New(typedRes.Name))
				continue
			}

			if fqnt == nil {
				fqnt = fqn.NewTraverser()
				t.Root().Walk(fqnt)
			}

			resFqn := fqnt.ResultFor(&ir.Name{
				Value:    typedRes.Name,
				Position: ir.GetPosition(t.node),
			})
			throws = append(throws, resFqn)

		default:
			log.Printf("[throws.PhpDocThrows]: Detected @throws tag with unexpected type: %v, expected a class like", result)
		}
	}

	return throws
}

func (t *Throws) resolve(node ir.Node) (*ir.Root, *expr.Resolved, error) {
	scopes := traversers.NewScopesTraverser(node)
	t.Root().Walk(scopes)
	if !scopes.Done {
		return nil, nil, fmt.Errorf("[throws.resolve]: could not find the node in the root node")
	}

	resolvement, _, left := expr.Resolve(node, &expr.Scopes{
		Path:  t.Path(),
		Root:  t.Root(),
		Class: scopes.Class,
		Block: scopes.Block,
	})
	if left != 0 {
		return nil, nil, fmt.Errorf(
			"[throws.resolve]: could not resolve the throw expression to its type",
		)
	}

	resolvedRoot := wrkspc.FromContainer().FIROf(resolvement.Path)

	return resolvedRoot, resolvement, nil
}

func (t *Throws) catches(thrown *fqn.FQN) func(catch *fqn.FQN) bool {
	defaultChecker := func(_ *fqn.FQN) bool { return false }

	throwNode, ok := index.FromContainer().Find(thrown)
	if !ok {
		log.Println(
			fmt.Errorf(
				"[throws.catches]: can't find node %s in index",
				thrown,
			),
		)
		return defaultChecker
	}

	cls, err := symbol.NewClassLikeFromFQN(wrkspc.NewRooter(throwNode.Path), throwNode.FQN)
	if err != nil {
		log.Println(fmt.Errorf("[throws.catches]: can't new class: %w", err))
		return defaultChecker
	}

	// Cache loaded classes for next calls.
	classes := []*symbol.ClassLike{cls}
	return func(catch *fqn.FQN) bool {
		// Go through cached classes first.
		for _, c := range classes {
			if c.GetFQN().String() == catch.String() {
				return true
			}
		}

		// Got through all cached classes, start iterating from the last cached class.
		iter := classes[len(classes)-1].InheritsIter()
		for inhCls, done, err := iter(); !done; inhCls, done, err = iter() {
			if err != nil {
				log.Println(fmt.Errorf("[throws.catches]: inherit err: %w", err))
				continue
			}

			classes = append(classes, inhCls)

			if inhCls.GetFQN().String() == catch.String() {
				return true
			}
		}

		return false
	}
}

type throwsTraverser struct {
	Result       []ir.Node
	visitedFirst bool
}

func newThrowsTraverser() *throwsTraverser {
	return &throwsTraverser{
		Result: []ir.Node{},
	}
}

func (t *throwsTraverser) EnterNode(node ir.Node) bool {
	if !t.visitedFirst {
		t.visitedFirst = true
		return true
	}

	switch node.(type) {
	case *ir.TryStmt, *ir.FunctionCallExpr, *ir.CatchStmt, *ir.ThrowStmt:
		t.Result = append(t.Result, node)
		return false

		// PERF: we can probably get away with returning true in a couple of cases.
	default:
		return true
	}
}

func (t *throwsTraverser) LeaveNode(node ir.Node) {}
