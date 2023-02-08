package symbol

import (
	"github.com/VKCOM/noverify/src/ir"
)

// Scopes are node kinds that declare a scope
// where for example variables are scoped to.
var (
	Scopes = map[ir.NodeKind]any{
		ir.KindFunctionStmt:      true,
		ir.KindClosureExpr:       true,
		ir.KindArrowFunctionExpr: true,

		ir.KindClassStmt:       true,
		ir.KindClassMethodStmt: true,

		ir.KindTraitStmt:     true,
		ir.KindInterfaceStmt: true,
	}

	ClassLikeScopes = []ir.NodeKind{
		ir.KindClassStmt,
		ir.KindTraitStmt,
		ir.KindInterfaceStmt,
	}
)

func IsScope(kind ir.NodeKind) bool {
	_, ok := Scopes[kind]
	return ok
}

func IsClassLike(kind ir.NodeKind) bool {
	return kind == ir.KindClassStmt ||
		kind == ir.KindInterfaceStmt ||
		kind == ir.KindTraitStmt
}

func IsNonClassLikeScope(kind ir.NodeKind) bool {
	return kind == ir.KindFunctionStmt || kind == ir.KindClassMethodStmt ||
		kind == ir.KindClosureExpr
}
