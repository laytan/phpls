package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
)

// Scopes are node kinds that declare a scope
// where for example variables are scoped to.
var (
	Scopes = map[ir.NodeKind]any{
		ir.KindFunctionStmt: true,

		ir.KindClassStmt:       true,
		ir.KindClassMethodStmt: true,

		ir.KindTraitStmt:     true,
		ir.KindInterfaceStmt: true,
	}
)

func IsScope(node ir.Node) bool {
	_, ok := Scopes[ir.GetNodeKind(node)]
	return ok
}

func IsClassLike(node ir.Node) bool {
	kind := ir.GetNodeKind(node)
	return kind == ir.KindClassStmt ||
		kind == ir.KindInterfaceStmt ||
		kind == ir.KindTraitStmt
}

func IsNonClassLikeScope(node ir.Node) bool {
	kind := ir.GetNodeKind(node)
	return kind == ir.KindFunctionStmt || kind == ir.KindClassMethodStmt
}
