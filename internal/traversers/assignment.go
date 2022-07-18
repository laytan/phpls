package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
)

func NewAssignment(variable *ir.SimpleVar) *assignment {
	return &assignment{
		variable:     variable,
		variableName: "$" + variable.Name,
	}
}

type assignment struct {
	visitor.Null

	variable     *ir.SimpleVar
	variableName string

	Assignment *ast.ExprAssign
}

func (v *assignment) ExprAssign(n *ast.ExprAssign) {
	if v.Assignment != nil {
		return
	}

	// TODO: a variable can be assigned multiple times, what should we do?

	// TODO: can we cancel the whole visitor when we find the assignment?

	// TODO: what about scopes (function, class, global, etc.)?

	variable, ok := n.Var.(*ast.ExprVariable)
	if !ok {
		panic("not ok")
	}

	identifier, ok := variable.Name.(*ast.Identifier)
	if !ok {
		panic("not ok")
	}

	if string(identifier.IdentifierTkn.Value) == v.variableName {
		v.Assignment = n
	}
}
