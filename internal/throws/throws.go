package throws

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/internal/expr"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/set"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

func ThrowsFromIndex(node *index.IndexNode) []*fqn.FQN {
	root, irNode, err := common.SymbolToNode(node.Path, node.Symbol)
	if err != nil {
		log.Println(fmt.Errorf("[throws.Throws]: %w", err))
		return nil
	}

	return common.Map(
		throws(root, irNode, node.Path).Slice(),
		func(val string) *fqn.FQN { return fqn.New(val) },
	)
}

func Throws(root *ir.Root, node ir.Node, path string) []*fqn.FQN {
	return common.Map(
		throws(root, node, path).Slice(),
		func(val string) *fqn.FQN { return fqn.New(val) },
	)
}

func throws(root *ir.Root, node ir.Node, path string) *set.Set[string] {
	thrownSet := set.New[string]()
	catchedSet := set.New[string]()

	switch node.(type) {
	case *ir.FunctionStmt, *ir.ClassMethodStmt:
		for _, throw := range PhpDocThrows(root, node) {
			thrownSet.Add(throw.String())
		}
	}

	traverser := NewThrowsTraverser()
	node.Walk(traverser)

	for _, result := range traverser.Result {
		switch typedRes := result.(type) {
		case *ir.TryStmt:
			thrownSet = thrownSet.Union(throws(root, result, path))

		case *ir.CatchStmt:
			fqnt := fqn.NewTraverser()
			root.Walk(fqnt)

			for _, catch := range typedRes.Types {
				switch typed := catch.(type) {
				case *ir.Name:
					catchedSet.Add(fqnt.ResultFor(typed).String())
				default:
					log.Printf("[throws.Throws]: Catch statement type has unexpected type: %v", typed)
				}
			}

		case *ir.ThrowStmt:
			resolvedRoot, resolvement, err := resolve(root, typedRes.Expr, path)
			if err != nil {
				log.Println(fmt.Errorf("[throws.Throws]: %w", err))
				continue
			}

			key := common.FullyQualify(resolvedRoot, symbol.GetIdentifier(resolvement.Node))
			thrownSet.Add(key.String())

		case *ir.FunctionCallExpr, *ir.MethodCallExpr:
			resolvedRoot, resolvement, err := resolve(root, result, path)
			if err != nil {
				log.Println(fmt.Errorf("[throws.Throws]: %w", err))
				continue
			}

			thrownSet = thrownSet.Union(throws(resolvedRoot, resolvement.Node, resolvement.Path))
		}
	}

	// Delete every thrown class that is caught by a catch.
	var toRemove []string
	for throw := range thrownSet.Iterator() {
		checker := Catches(fqn.New(throw))

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

func Catches(thrown *fqn.FQN) func(catch *fqn.FQN) bool {
	ind := index.FromContainer()
	wsp := wrkspc.FromContainer()

	defaultChecker := func(_ *fqn.FQN) bool { return false }

	throwNode, ok := ind.Find(thrown)
	if !ok {
		log.Println(
			fmt.Errorf(
				"[throws.Catches]: can't find node %s in index",
				thrown,
			),
		)
		return defaultChecker
	}

	rNode := &resolvequeue.Node{
		FQN:  throwNode.FQN,
		Kind: throwNode.Symbol.NodeKind(),
	}

	rQueue := resolvequeue.New(func(n *resolvequeue.Node) (*ir.Root, error) {
		iNode, ok := ind.Find(n.FQN)
		if !ok {
			return nil, fmt.Errorf(
				"[throws.Catches(%s)]: can't find node for %s in index",
				thrown,
				n.FQN,
			)
		}

		root, err := wsp.IROf(iNode.Path)
		if err != nil {
			return nil, fmt.Errorf(
				"Catches(%s): can't find root for %s in index: %w",
				thrown,
				n.FQN,
				err,
			)
		}

		return root, nil
	}, rNode)

	return func(catch *fqn.FQN) bool {
		defer rQueue.Queue.Reset()

		for currCls := rQueue.Queue.Dequeue(); currCls != nil; currCls = rQueue.Queue.Dequeue() {
			if currCls.FQN.String() == catch.String() {
				return true
			}
		}

		for _, currCls := range rQueue.Implements {
			if currCls.FQN.String() == catch.String() {
				return true
			}
		}

		return false
	}
}

func PhpDocThrows(root *ir.Root, node ir.Node) (throws []*fqn.FQN) {
	var fqnt *fqn.Traverser

	results := typer.FromContainer().Throws(node)
	for _, result := range results {
		switch typedRes := result.(type) {
		case *phpdoxer.TypeClassLike:
			if typedRes.FullyQualified {
				throws = append(throws, fqn.New(typedRes.Name))
				continue
			}

			if fqnt == nil {
				fqnt = fqn.NewTraverser()
				root.Walk(fqnt)
			}

			resFqn := fqnt.ResultFor(&ir.Name{Value: typedRes.Name})
			throws = append(throws, resFqn)

		default:
			log.Printf("[throws.PhpDocThrows]: Detected @throws tag with unexpected type: %v, expected a class like", result)
		}
	}

	return throws
}

func resolve(root *ir.Root, node ir.Node, path string) (*ir.Root, *expr.Resolved, error) {
	scopes := traversers.NewScopesTraverser(node)
	root.Walk(scopes)
	if !scopes.Done {
		return nil, nil, fmt.Errorf("[throws.resolve]: could not find the node in the root node")
	}

	resolvement, _, left := expr.Resolve(node, expr.Scopes{
		Path:  path,
		Root:  root,
		Class: scopes.Class,
		Block: scopes.Block,
	})
	if left != 0 {
		return nil, nil, fmt.Errorf(
			"[throws.resolve]: could not resolve the throw expression to its type",
		)
	}

	resolvedRoot, err := wrkspc.FromContainer().IROf(resolvement.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("[throws.resolve]: %w", err)
	}

	return resolvedRoot, resolvement, nil
}
