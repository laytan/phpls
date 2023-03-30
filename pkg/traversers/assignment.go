package traversers

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/nodevar"
)

func NewAssignment(variable *ast.ExprVariable) *Assignment {
	return &Assignment{
		variable: variable,
		isFirst:  true,
	}
}

type Assignment struct {
	visitor.Null
	variable   *ast.ExprVariable
	Assignment *ast.ExprVariable
	Scope      ast.Vertex
	isFirst    bool
}

func (a *Assignment) EnterNode(node ast.Vertex) bool {
	defer func() { a.isFirst = false }()

	// Only check the current scope.
	if !a.isFirst && nodescopes.IsScope(node.GetType()) {
		return false
	}

	// Only check assignments before our variable.
	if pos := node.GetPosition(); pos != nil {
		if pos.StartPos > a.variable.Position.StartPos {
			return false
		}
	}

	varName := nodeident.Get(a.variable)
	if nodevar.IsAssignment(node.GetType()) {
		assigned := nodevar.Assigned(node)
		for _, ass := range assigned {
			if nodeident.Get(ass) == varName {
				a.Assignment = ass
				a.Scope = node
				break
			}
		}
	}

	return true
}

func (a *Assignment) Parameter(param *ast.Parameter) {
	varName := nodeident.Get(a.variable)
	if varName == nodeident.Get(param.Var.(*ast.ExprVariable)) {
		a.Assignment = param.Var.(*ast.ExprVariable)
		a.Scope = param
	}
}
