package throws

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/ie"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/set"
	"github.com/laytan/elephp/pkg/traversers"
	"golang.org/x/exp/slices"
)

type rooter interface {
	Root() *ir.Root
	Path() string
}

type doxFinder interface {
	FindThrows() []*phpdoxer.NodeThrows
}

type Throws struct {
	rooter

	doxed doxFinder

	node ir.Node

	ignoreFirstFuncDoc bool
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
	return NewResolver(rooter, n.ToIRNode(rooter.Root()))
}

type Violation struct {
	// Either a function or class method statement.
	Node   ir.Node
	Throws []*fqn.FQN

	message string
}

func (v *Violation) Message() string {
	if len(v.message) > 0 {
		return v.message
	}

	typeStr := ie.IfElse(ir.GetNodeKind(v.Node) == ir.KindFunctionStmt, "function", "method")
	throwsStr := strings.Join(functional.Map(v.Throws, functional.ToString[*fqn.FQN]), ", ")

	v.message = fmt.Sprintf(
		"The %s %s throws %s but these exceptions are not caught or added to the PHPDoc.",
		typeStr,
		nodeident.Get(v.Node),
		throwsStr,
	)
	return v.message
}

func (v *Violation) Code() string {
	return "uncaught"
}

func (v *Violation) Line() int {
	return ir.GetPosition(v.Node).StartLine
}

// Diagnose finds all the functions or methods in the given file that throw
// exceptions that are not added to the PHPDoc with an @throws tag or caught in
// that method.
func Diagnose(root rooter) (results []*Violation) {
	// NOTE: might be a good idea to hold a map in the index, from path to fqns.
	// That way we can easily get all the functions and methods of the file.

	nodes := make(chan ir.Node)
	violations := make(chan *Violation)

	go func() {
		defer close(violations)

		wg := sync.WaitGroup{}
		for throwing := range nodes {
			wg.Add(1)
			go func(throwing ir.Node) {
				defer wg.Done()

				r := &Throws{
					rooter:             root,
					doxed:              symbol.NewDoxed(throwing),
					node:               throwing,
					ignoreFirstFuncDoc: true,
				}
				throws := r.Throws()
				if len(throws) == 0 {
					return
				}

				// Remove any returned throw that is documented to be thrown.
				doxed := r.phpDocThrows()
				if len(doxed) > 0 {
					for i := len(throws) - 1; i >= 0; i-- {
						checker := r.catches(throws[i])

						for _, d := range doxed {
							if checker(d) {
								throws = slices.Delete(throws, i, i+1)
								if len(throws) == 0 {
									return
								}
							}
						}
					}
				}

				violations <- &Violation{
					Node:   throwing,
					Throws: throws,
				}
			}(throwing)
		}

		wg.Wait()
	}()

	go func() {
		t := newThrowingSymbolTraverser(nodes)
		root.Root().Walk(t)
	}()

	for violation := range violations {
		results = append(results, violation)
	}

	return results
}

func (t *Throws) Throws() []*fqn.FQN {
	return functional.Map(t.throws(true).Slice(), fqn.New)
}

func (t *Throws) throws(firstCall bool) *set.Set[string] {
	thrownSet := set.New[string]()
	catchedSet := set.New[string]()

	if !firstCall || !t.ignoreFirstFuncDoc {
		switch t.node.(type) {
		case *ir.FunctionStmt, *ir.ClassMethodStmt:
			for _, throw := range t.phpDocThrows() {
				thrownSet.Add(throw.String())
			}
		}
	}

	traverser := newThrowsTraverser()
	t.node.Walk(traverser)

	for _, result := range traverser.Result {
		switch typedRes := result.(type) {
		case *ir.TryStmt:
			tryThrows := set.New[string]()

			blockThrows := &Throws{
				rooter: t.rooter,
				doxed:  symbol.NewDoxed(result),
				node:   result,
			}
			tryThrows = tryThrows.Union(blockThrows.throws(false))

			// Go through each catch, first remove all the things
			// that are caught by the types.
			// Then add all the things that are thrown in the catch, to the tryThrows.
			// At the end, add all that is thrown in the finally to tryThrows.

			for i := range typedRes.Catches {
				catch := typedRes.Catches[i].(*ir.CatchStmt)

				fqnt := fqn.NewTraverser()
				t.Root().Walk(fqnt)

				var toRemove []string
				for tryThrow := range tryThrows.Iterator() {
					checker := t.catches(fqn.New(tryThrow))

					for j := range catch.Types {
						catchedType := catch.Types[j].(*ir.Name)
						qualified := fqnt.ResultFor(catchedType)

						if checker(qualified) {
							toRemove = append(toRemove, tryThrow)
						}
					}
				}
				for _, rm := range toRemove {
					tryThrows.Remove(rm)
				}

				catchThrows := &Throws{
					rooter: t.rooter,
					doxed:  symbol.NewDoxed(catch),
					node:   catch,
				}
				tryThrows = tryThrows.Union(catchThrows.throws(false))
			}

			if typedRes.Finally != nil {
				finallyThrows := &Throws{
					rooter: t.rooter,
					doxed:  symbol.NewDoxed(typedRes.Finally),
					node:   typedRes.Finally,
				}
				tryThrows = tryThrows.Union(finallyThrows.throws(false))
			}

			thrownSet = thrownSet.Union(tryThrows)

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
			thrownSet = thrownSet.Union(blockThrows.throws(false))
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

	results := t.doxed.FindThrows()
	for _, result := range results {
		switch typedRes := result.Type.(type) {
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

var _ ir.Visitor = &throwsTraverser{}

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
	case *ir.TryStmt, *ir.FunctionCallExpr, *ir.ThrowStmt:
		t.Result = append(t.Result, node)
		return false

		// PERF: we can probably get away with returning true in a couple of cases.
	default:
		return true
	}
}

func (t *throwsTraverser) LeaveNode(node ir.Node) {}

func newThrowingSymbolTraverser(nodes chan<- ir.Node) *throwingSymbolsTraverser {
	return &throwingSymbolsTraverser{nodes: nodes}
}

type throwingSymbolsTraverser struct {
	nodes            chan<- ir.Node
	currentNamespace string
}

func (t *throwingSymbolsTraverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.NamespaceStmt:
		if typedNode.NamespaceName != nil {
			t.currentNamespace = "\\" + typedNode.NamespaceName.Value + "\\"
		}

		return true

	case *ir.FunctionStmt, *ir.ClassMethodStmt:
		t.nodes <- node

		return false

	case *ir.Root, *ir.ClassStmt, *ir.TraitStmt:
		return true

	default:
		return false
	}
}

func (t *throwingSymbolsTraverser) LeaveNode(node ir.Node) {
	if _, ok := node.(*ir.Root); ok {
		close(t.nodes)
	}
}
