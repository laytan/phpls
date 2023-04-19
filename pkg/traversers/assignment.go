package traversers

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/nodescopes"
	"github.com/laytan/phpls/pkg/nodevar"
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
	if a.Assignment != nil {
		return false
	}

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

func (a *Assignment) ExprClosureUse(node *ast.ExprClosureUse) {
	switch tn := node.Var.(type) {
	case *ast.ExprVariable:
		if nodeident.Get(tn) == nodeident.Get(a.variable) {
			a.Assignment = tn
			a.Scope = node
		}
	default:
		log.Panicf("unexpected use var type %T", node.Var)
	}
}

func (a *Assignment) Parameter(param *ast.Parameter) {
	varName := nodeident.Get(a.variable)
	if varName == nodeident.Get(param.Var.(*ast.ExprVariable)) {
		a.Assignment = param.Var.(*ast.ExprVariable)
		a.Scope = param
	}
}
