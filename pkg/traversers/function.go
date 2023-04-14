package traversers

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/nodescopes"
)

func NewFunction(name string) *Function {
	return &Function{
		name:           name,
		currNodeIsRoot: true,
	}
}

type Function struct {
	visitor.Null
	name           string
	Function       *ast.StmtFunction
	currNodeIsRoot bool
}

func (f *Function) EnterNode(node ast.Vertex) bool {
	defer func() { f.currNodeIsRoot = false }()

	if f.Function != nil {
		return false
	}

	// If the scope of the traverser is a function, the first call will be a
	// function which we need to ignore.
	if f.currNodeIsRoot && node.GetType() == ast.TypeStmtFunction {
		return true
	}

	if fn, ok := node.(*ast.StmtFunction); ok && nodeident.Get(fn.Name) == f.name {
		f.Function = fn
	}

	if nodescopes.IsScope(node.GetType()) {
		return false
	}

	return true
}
