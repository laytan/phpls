package project

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/symbol"
)

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
	sourceFile *File,
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

// Resolves the fully qualified name for the given name node.
//
// This resolves, use statements, aliassed use statements are resolved to the
// non-aliassed version.
func (p *Project) FQN(root *ir.Root, name *ir.Name) *traversers.FQN {
	if name.IsFullyQualified() {
		return traversers.NewFQN(name.Value)
	}

	traverser := traversers.NewFQNTraverser(name)
	root.Walk(traverser)

	return traverser.Result()
}
