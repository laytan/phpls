package nodescopes

import (
	"github.com/VKCOM/php-parser/pkg/ast"
)

// Scopes are node kinds that declare a scope
// where for example variables are scoped to.
var (
	Scopes = map[ast.Type]bool{
		ast.TypeStmtFunction:      true,
		ast.TypeExprClosure:       true,
		ast.TypeExprArrowFunction: true,

		ast.TypeStmtClass:       true,
		ast.TypeStmtClassMethod: true,

		ast.TypeStmtTrait:     true,
		ast.TypeStmtInterface: true,
	}

	ClassLikeScopes = map[ast.Type]bool{
		ast.TypeStmtClass:     true,
		ast.TypeStmtTrait:     true,
		ast.TypeStmtInterface: true,
	}

	NameScopes = map[ast.Type]bool{
		ast.TypeNameRelative:       true,
		ast.TypeNameFullyQualified: true,
		ast.TypeName:               true,
	}
)

func IsScope(kind ast.Type) bool {
	return Scopes[kind]
}

func IsClassLike(kind ast.Type) bool {
	return ClassLikeScopes[kind]
}

func IsNonClassLikeScope(kind ast.Type) bool {
	return !IsClassLike(kind)
}

func IsName(kind ast.Type) bool {
	return NameScopes[kind]
}
