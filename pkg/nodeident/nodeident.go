package nodeident

import (
	"log"
	"strings"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/pkg/functional"
	"github.com/laytan/phpls/pkg/nodevar"
)

// Returns the identifier/value of a node.
//
// NOTE: To be extended when needed.
// TODO: return bytes.
func Get(n ast.Vertex) string {
	if n == nil {
		log.Println("Warning: nodeident.Get with nil node")
		return ""
	}

	if nodevar.IsAssignment(n.GetType()) {
		assignment := nodevar.Assigned(n)
		if len(assignment) > 1 {
			log.Printf("Warning: nodeident.Get with assignment to multiple variables: %v", n)
		}

		return Get(assignment[0])
	}

	switch n := n.(type) {
	case *ast.Argument:
		return Get(n.Name)

	case *ast.ExprClassConstFetch:
		return Get(n.Const)

	case *ast.StmtClassMethod:
		return Get(n.Name)

	case *ast.StmtClass:
		return Get(n.Name)

	case *ast.ExprConstFetch:
		return Get(n.Const)

	case *ast.StmtConstant:
		return Get(n.Name)

	case *ast.ExprFunctionCall:
		return Get(n.Function)

	case *ast.StmtFunction:
		return Get(n.Name)

	case *ast.StmtInterface:
		return Get(n.Name)

	case *ast.Name:
		return strings.Join(functional.Map(n.Parts, Get), "\\")

	case *ast.NameFullyQualified:
		return "\\" + strings.Join(functional.Map(n.Parts, Get), "\\")

	case *ast.NameRelative:
		return "\\" + strings.Join(functional.Map(n.Parts, Get), "\\")

	case *ast.StmtNamespace:
		if n.Name == nil {
			return "\\"
		}

		return "\\" + Get(n.Name)

	case *ast.Parameter:
		return Get(n.Var)

	case *ast.StmtProperty:
		return Get(n.Var)

	case *ast.StmtPropertyList:
		if len(n.Props) > 1 || len(n.Props) == 0 {
			log.Panicf("PropertyListStmt has > 1 || 0 properties, how does this work?")
		}

		return Get(n.Props[0])

	case *ast.StmtClassConstList:
		if len(n.Consts) > 1 || len(n.Consts) == 0 {
			log.Printf("Trying to get identifier of *ir.ClassConstListStmt but there are %d possibilities", len(n.Consts))
		}

		return Get(n.Consts[0])

	case *ast.ExprVariable:
		return Get(n.Name)

	case *ast.StmtStaticVar:
		return Get(n.Var)

	case *ast.StmtTrait:
		return Get(n.Name)

	case *ast.StmtTraitUseAlias:
		return Get(n.Alias)

	case *ast.StmtUse:
		return Get(n.Use)

	case *ast.ExprMethodCall:
		return Get(n.Method)

	case *ast.ExprNullsafeMethodCall:
		return Get(n.Method)

	case *ast.ExprPropertyFetch:
		return Get(n.Var)

	case *ast.ExprNullsafePropertyFetch:
		return Get(n.Var)

	case *ast.ExprStaticCall:
		return Get(n.Call)

	case *ast.ScalarString:
		return string(n.Value)

	case *ast.ScalarMagicConstant:
		return string(n.Value)

	case *ast.Identifier:
		return string(n.Value)

	case *ast.NamePart:
		return string(n.Value)

	default:
		log.Printf("Warning: unimplemented nodeident case for node type %T", n)
		return ""
	}
}
