package project

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

// Resolves the fully qualified name for the given name node.
//
// This resolves, use statements, aliassed use statements are resolved to the
// non-aliassed version.
func (p *Project) FQN(root *ir.Root, name *ir.Name) *traversers.FQN {
	if name.IsFullyQualified() {
		return traversers.NewFQN(name.Value)
	}

	traverser := traversers.NewFQNTraverser()
	root.Walk(traverser)

	return traverser.ResultFor(name)
}

func (p *Project) assignment(scope ir.Node, variable *ir.SimpleVar) *ir.SimpleVar {
	traverser := traversers.NewAssignment(variable)
	scope.Walk(traverser)

	return traverser.Assignment
}

func (p *Project) globalAssignment(root *ir.Root, globalVar *ir.SimpleVar) *ir.SimpleVar {
	// First search the current file for the assignment.
	traverser := traversers.NewGlobalAssignment(globalVar)
	root.Walk(traverser)

	// TODO: search the whole project if the global is not assigned here.

	return traverser.Assignment
}

func (p *Project) function(
	scope ir.Node,
	call *ir.FunctionCallExpr,
) (*symbol.FunctionStmtSymbol, string) {
	traverser, err := traversers.NewFunction(call)
	if err != nil {
		return nil, ""
	}

	scope.Walk(traverser)
	if traverser.Function != nil {
		return symbol.NewFunction(traverser.Function), ""
	}

	// No definition found locally, searching globally.
	name, ok := call.Function.(*ir.Name)
	if !ok {
		return nil, ""
	}

	results := p.symbolTrie.SearchExact(name.Value)

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if function, ok := res.Symbol.(*symbol.FunctionStmtSymbol); ok {
			return function, res.Path
		}
	}

	return nil, ""
}

func (p *Project) classLike(
	root *ir.Root,
	name *ir.Name,
) (*symbol.ClassLikeStmtSymbol, string) {
	fqn := p.FQN(root, name)

	results := p.symbolTrie.SearchExact(fqn.Name())

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if res.Namespace != fqn.Namespace() {
			continue
		}

		if symbol, ok := res.Symbol.(*symbol.ClassLikeStmtSymbol); ok {
			return symbol, res.Path
		}
	}

	return nil, ""
}

type rootRetriever struct {
	project *Project
}

func (r *rootRetriever) RetrieveRoot(n *resolvequeue.Node) (*ir.Root, error) {
	file := r.project.FindFileInTrie(n.FQN, n.Kind)
	if file == nil {
		return nil, fmt.Errorf("No file indexed for FQN %s", n.FQN)
	}

	ast := r.project.ParseFileCached(file)
	if ast == nil {
		return nil, fmt.Errorf("Error parsing file for FQN %s", n.FQN)
	}

	return ast, nil
}

func (p *Project) method(
	root *ir.Root,
	classLikeScope ir.Node,
	method string,
) (*ir.ClassMethodStmt, string, error) {
	fqn := p.FQN(root, &ir.Name{Value: symbol.GetIdentifier(classLikeScope)})
	fmt.Printf("Trying to get method %s of %s\n", method, fqn.String())
	resolveQueue := resolvequeue.New(
		&rootRetriever{project: p},
		&resolvequeue.Node{FQN: fqn, Kind: ir.GetNodeKind(classLikeScope)},
	)

	isCurr := true
	for res := resolveQueue.Queue.Dequeue(); res != nil; func() {
		res = resolveQueue.Queue.Dequeue()
		isCurr = false
	}() {
		file, symbol := p.FindFileAndSymbolInTrie(res.FQN, res.Kind)

		if file == nil {
			return nil, "", fmt.Errorf("Can't get file for FQN: %s", res.FQN.String())
		}

		var privacy phprivacy.Privacy
		switch symbol.Symbol.NodeKind() {
		case ir.KindClassStmt:
			// If first index (source file) search for any privacy,
			// if not, search for protected and public privacy.
			privacy = phprivacy.PrivacyProtected

			if isCurr {
				privacy = phprivacy.PrivacyPrivate
			}
		default:
			// Traits and interface members are available everywhere.
			privacy = phprivacy.PrivacyPrivate
		}

		ast := p.ParseFileCached(file)
		if ast == nil {
			return nil, "", fmt.Errorf("Error parsing ast for %s", file.path)
		}

		traverser := traversers.NewMethod(method, res.FQN.Name(), privacy)
		ast.Walk(traverser)

		if traverser.Method != nil {
			return traverser.Method, symbol.Path, nil
		}
	}

	return nil, "", nil
}
