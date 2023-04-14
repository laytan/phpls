package throws

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/expr"
	"github.com/laytan/phpls/internal/fqner"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/symbol"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/ie"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/phpdoxer"
	"github.com/laytan/phpls/pkg/set"
	"github.com/laytan/phpls/pkg/traversers"
	"golang.org/x/exp/slices"
)

type rooter interface {
	Root() *ast.Root
	Path() string
}

type doxFinder interface {
	FindThrows() []*phpdoxer.NodeThrows
}

type Throws struct {
	rooter

	doxed doxFinder

	node ast.Vertex

	ignoreFirstFuncDoc bool
}

func NewResolver(root rooter, node ast.Vertex) *Throws {
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
	Node   ast.Vertex
	Throws []*fqn.FQN

	message string
}

func (v *Violation) Message() string {
	if len(v.message) > 0 {
		return v.message
	}

	typeStr := ie.IfElse(v.Node.GetType() == ast.TypeStmtFunction, "function", "method")
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
	return v.Node.GetPosition().StartLine
}

// Diagnose finds all the functions or methods in the given file that throw
// exceptions that are not added to the PHPDoc with an @throws tag or caught in
// that method.
func Diagnose(root rooter) (results []*Violation) {
	// NOTE: might be a good idea to hold a map in the index, from path to fqns.
	// That way we can easily get all the functions and methods of the file.

	nodes := make(chan ast.Vertex)
	violations := make(chan *Violation)

	go func() {
		defer close(violations)

		wg := sync.WaitGroup{}
		for throwing := range nodes {
			wg.Add(1)
			go func(throwing ast.Vertex) {
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
		tv := traverser.NewTraverser(t)
		root.Root().Accept(tv)
	}()

	for violation := range violations {
		results = append(results, violation)
	}

	return results
}

func (t *Throws) Throws() []*fqn.FQN {
	return functional.Map(t.throws(set.New[string]()).Slice(), fqn.New)
}

func (t *Throws) seenHash() string {
	pos := t.node.GetPosition()
	return fmt.Sprintf("%d%d%d%d", pos.StartLine, pos.StartPos, pos.EndLine, pos.EndPos)
}

// PERF: we should have an incremental cache,
// something like a map from a function or method to what it throws.
// if we then have 2 different calls to a method we already have its throws.
// might be able to keep the cache and invalidate when the file changes.
func (t *Throws) throws(seen *set.Set[string]) *set.Set[string] {
	thrownSet := set.New[string]()

	hash := t.seenHash()
	if seen.Has(hash) {
		return thrownSet
	}

	seen.Add(t.seenHash())

	firstCall := seen.Size() == 0

	if !firstCall || !t.ignoreFirstFuncDoc {
		switch t.node.(type) {
		case *ast.StmtFunction, *ast.StmtClassMethod:
			for _, throw := range t.phpDocThrows() {
				thrownSet.Add(throw.String())
			}
		}
	}

	tv := newThrowsTraverser()
	tvv := traverser.NewTraverser(tv)
	t.node.Accept(tvv)

	for _, result := range tv.Result {
		switch typedRes := result.(type) {
		case *ast.StmtTry:
			tryThrows := set.New[string]()

			blockThrows := &Throws{
				rooter: t.rooter,
				doxed:  symbol.NewDoxed(result),
				node:   result,
			}
			tryThrows = tryThrows.Union(blockThrows.throws(seen))

			// Go through each catch, first remove all the things
			// that are caught by the types.
			// Then add all the things that are thrown in the catch, to the tryThrows.
			// At the end, add all that is thrown in the finally to tryThrows.

			for i := range typedRes.Catches {
				catch := typedRes.Catches[i].(*ast.StmtCatch)

				fqnt := fqn.NewTraverser()
				fqntt := traverser.NewTraverser(fqnt)
				t.Root().Accept(fqntt)

				var toRemove []string
				for tryThrow := range tryThrows.Iterator() {
					checker := t.catches(fqn.New(tryThrow))

					for j := range catch.Types {
						catchedType := catch.Types[j].(*ast.Name)
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
				tryThrows = tryThrows.Union(catchThrows.throws(seen))
			}

			if typedRes.Finally != nil {
				finallyThrows := &Throws{
					rooter: t.rooter,
					doxed:  symbol.NewDoxed(typedRes.Finally),
					node:   typedRes.Finally,
				}
				tryThrows = tryThrows.Union(finallyThrows.throws(seen))
			}

			thrownSet = thrownSet.Union(tryThrows)

		case *ast.StmtThrow: // *ast.ExprThrow, what is that?
			resolvedRoot, resolvement, err := t.resolve(typedRes.Expr)
			if err != nil {
				log.Println(fmt.Errorf("[throws.Throws]: %w", err))
				continue
			}

			key := fqner.FullyQualifyName(resolvedRoot, resolvement.Node.(*ast.Name))
			thrownSet.Add(key.String())

		case *ast.ExprFunctionCall, *ast.ExprMethodCall, *ast.ExprStaticCall:
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
			thrownSet = thrownSet.Union(blockThrows.throws(seen))
		}
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
				fqntt := traverser.NewTraverser(fqnt)
				t.Root().Accept(fqntt)
			}

			resFqn := fqnt.ResultFor(&ast.Name{
				Parts:    nameParts(typedRes.Name),
				Position: t.node.GetPosition(),
			})
			throws = append(throws, resFqn)

		default:
			log.Printf("[throws.PhpDocThrows]: Detected @throws tag with unexpected type: %v, expected a class like", result)
		}
	}

	return throws
}

func (t *Throws) resolve(node ast.Vertex) (*ast.Root, *expr.Resolved, error) {
	scopes := traversers.NewScopesTraverser(node)
	scopest := traverser.NewTraverser(scopes)
	t.Root().Accept(scopest)
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

	resolvedRoot := wrkspc.Current.FIROf(resolvement.Path)

	return resolvedRoot, resolvement, nil
}

func (t *Throws) catches(thrown *fqn.FQN) func(catch *fqn.FQN) bool {
	defaultChecker := func(_ *fqn.FQN) bool { return false }

	throwNode, ok := index.Current.Find(thrown)
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
	visitor.Null
	Result       []ast.Vertex
	visitedFirst bool
}

func newThrowsTraverser() *throwsTraverser {
	return &throwsTraverser{
		Result: []ast.Vertex{},
	}
}

func (t *throwsTraverser) EnterNode(node ast.Vertex) bool {
	if !t.visitedFirst {
		t.visitedFirst = true
		return true
	}

	switch node.(type) {
	case *ast.StmtTry, *ast.ExprFunctionCall, *ast.StmtThrow, *ast.ExprMethodCall, *ast.ExprStaticCall:
		t.Result = append(t.Result, node)
		return false

		// PERF: we can probably get away with returning false in a couple of cases.
	default:
		return true
	}
}

func newThrowingSymbolTraverser(nodes chan<- ast.Vertex) *throwingSymbolsTraverser {
	return &throwingSymbolsTraverser{nodes: nodes}
}

type throwingSymbolsTraverser struct {
	visitor.Null
	nodes            chan<- ast.Vertex
	currentNamespace string
}

func (t *throwingSymbolsTraverser) EnterNode(node ast.Vertex) bool {
	switch typedNode := node.(type) {
	case *ast.StmtNamespace:
		if typedNode.Name != nil {
			t.currentNamespace = "\\" + nodeident.Get(typedNode.Name) + "\\"
		}

		return true

	case *ast.StmtFunction, *ast.StmtClassMethod:
		t.nodes <- node

		return false

	case *ast.Root, *ast.StmtClass, *ast.StmtTrait:
		return true

	default:
		return false
	}
}

func (t *throwingSymbolsTraverser) LeaveNode(node ast.Vertex) {
	if _, ok := node.(*ast.Root); ok {
		close(t.nodes)
	}
}

func nameParts(name string) []ast.Vertex {
	return functional.Map(
		strings.Split(name, "\\"),
		func(s string) ast.Vertex { return &ast.NamePart{Value: []byte(s)} },
	)
}
