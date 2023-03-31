package nodevar

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
)

func IsAssignment(t ast.Type) bool {
	return t == ast.TypeStmtGlobal || t >= ast.TypeExprAssign && t <= ast.TypeExprAssignShiftRight
}

// Assigned returns the variable being assigned to, should call IsAssignment first.
func Assigned(n ast.Vertex) []*ast.ExprVariable {
	switch nt := n.(type) {
	case *ast.ExprAssign:
		return Assigned(nt.Var)
	case *ast.ExprAssignReference:
		return Assigned(nt.Var)
	case *ast.ExprAssignBitwiseAnd:
		return Assigned(nt.Var)
	case *ast.ExprAssignBitwiseOr:
		return Assigned(nt.Var)
	case *ast.ExprAssignBitwiseXor:
		return Assigned(nt.Var)
	case *ast.ExprAssignCoalesce:
		return Assigned(nt.Var)
	case *ast.ExprAssignConcat:
		return Assigned(nt.Var)
	case *ast.ExprAssignDiv:
		return Assigned(nt.Var)
	case *ast.ExprAssignMinus:
		return Assigned(nt.Var)
	case *ast.ExprAssignMod:
		return Assigned(nt.Var)
	case *ast.ExprAssignMul:
		return Assigned(nt.Var)
	case *ast.ExprAssignPlus:
		return Assigned(nt.Var)
	case *ast.ExprAssignPow:
		return Assigned(nt.Var)
	case *ast.ExprAssignShiftLeft:
		return Assigned(nt.Var)
	case *ast.ExprAssignShiftRight:
		return Assigned(nt.Var)
	case *ast.StmtGlobal:
		res := make([]*ast.ExprVariable, 0, len(nt.Vars))
		for _, item := range nt.Vars {
			res = append(res, Assigned(item)...)
		}

		return res
	case *ast.ExprList: // List can be a assigned to for destructuring etc.
		res := make([]*ast.ExprVariable, 0, len(nt.Items))
		for _, item := range nt.Items {
			res = append(res, Assigned(item)...)
		}

		return res
	case *ast.ExprArrayItem: // Item inside list.
		return Assigned(nt.Val)
	case *ast.ExprVariable:
		return []*ast.ExprVariable{nt}
	default:
		log.Printf("Warning: nodevar.Assigned called with unsupported node of type %T", n)
		return nil
	}
}

func AssignmentExpr(n ast.Vertex) ast.Vertex {
	switch nt := n.(type) {
	case *ast.ExprAssign:
		return nt.Expr
	case *ast.ExprAssignReference:
		return nt.Expr
	case *ast.ExprAssignBitwiseAnd:
		return nt.Expr
	case *ast.ExprAssignBitwiseOr:
		return nt.Expr
	case *ast.ExprAssignBitwiseXor:
		return nt.Expr
	case *ast.ExprAssignCoalesce:
		return nt.Expr
	case *ast.ExprAssignConcat:
		return nt.Expr
	case *ast.ExprAssignDiv:
		return nt.Expr
	case *ast.ExprAssignMinus:
		return nt.Expr
	case *ast.ExprAssignMod:
		return nt.Expr
	case *ast.ExprAssignMul:
		return nt.Expr
	case *ast.ExprAssignPlus:
		return nt.Expr
	case *ast.ExprAssignPow:
		return nt.Expr
	case *ast.ExprAssignShiftLeft:
		return nt.Expr
	case *ast.ExprAssignShiftRight:
		return nt.Expr
	default:
		log.Printf("Warning: nodevar.AssigmentExpr called with unsupported node of type %T", n)
		return nil
	}
}
