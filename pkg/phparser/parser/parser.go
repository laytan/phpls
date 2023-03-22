package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/laytan/elephp/pkg/phparser/ast"
	"github.com/laytan/elephp/pkg/phparser/lexer"
)

var Parser = participle.MustBuild[ast.Program](
	participle.Lexer(lexer.NewDef()),
	participle.Elide("NonPHP", "LineComment", "BlockComment", "PHPStart", "PHPEchoStart", "PHPEnd"),
	participle.Union(ast.StatementImpls...), // Top level statements.
	participle.Union(ast.ClassStatementImpls...), // Class level statements.
	participle.Union(ast.CallableStatementImpls...), // Callable level statements.
	participle.Union(ast.ExprImpls...), // Any expression.
	participle.Union(ast.ExprNoArrAccImpls...), // Any expression - `ast.ArrAccessGroup`
)
