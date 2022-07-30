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
