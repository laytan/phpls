package nodeident

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
)

// Returns the identifier/value of a node.
//
// NOTE: To be extended when needed.
func Get(n ir.Node) string {
	switch n := n.(type) {
	case *ir.Argument:
		return n.Name.Value

	case *ir.ClassConstFetchExpr:
		return n.ConstantName.Value

	case *ir.ClassExtendsStmt:
		return n.ClassName.Value

	case *ir.ClassMethodStmt:
		return n.MethodName.Value

	case *ir.ClassStmt:
		return n.ClassName.Value

	case *ir.ConstFetchExpr:
		return n.Constant.Value

	case *ir.ConstantStmt:
		return n.ConstantName.Value

	case *ir.FunctionCallExpr:
		name, ok := n.Function.(*ir.Name)
		if !ok {
			return ""
		}

		return name.Value

	case *ir.FunctionStmt:
		return n.FunctionName.Value

	case *ir.Identifier:
		return n.Value

	case *ir.InterfaceStmt:
		return n.InterfaceName.Value

	case *ir.MagicConstant:
		return n.Value

	case *ir.Name:
		return n.Value

	case *ir.NamespaceStmt:
		if n.NamespaceName != nil {
			return n.NamespaceName.Value
		}

		return ""

	case *ir.Parameter:
		return n.Variable.Name

	case *ir.PropertyStmt:
		return n.Variable.Name

	case *ir.PropertyListStmt:
		if len(n.Properties) > 1 || len(n.Properties) == 0 {
			log.Panicf("PropertyListStmt has > 1 || 0 properties, how does this work?")
		}

		return Get(n.Properties[0])

	case *ir.ClassConstListStmt:
		if len(n.Consts) > 1 || len(n.Consts) == 0 {
			log.Printf("Trying to get identifier of *ir.ClassConstListStmt but there are %d possibilities", len(n.Consts))
		}

		return Get(n.Consts[0])

	case *ir.SimpleVar:
		return n.Name

	case *ir.StaticVarStmt:
		return n.Variable.Name

	case *ir.String:
		return n.Value

	case *ir.TraitMethodRefStmt:
		return n.Method.Value

	case *ir.TraitStmt:
		return n.TraitName.Value

	case *ir.TraitUseAliasStmt:
		return n.Alias.Value

	case *ir.UseStmt:
		return n.Use.Value

	case *ir.Assign:
		if a, ok := n.Variable.(*ir.SimpleVar); ok {
			return a.Name
		}

		return ""

	case *ir.MethodCallExpr:
		if i, ok := n.Method.(*ir.Identifier); ok {
			return i.Value
		}

		return ""

	case *ir.StaticCallExpr:
		if c, ok := n.Call.(*ir.Identifier); ok {
			return c.Value
		}

		return ""

	default:
		return ""
	}
}
