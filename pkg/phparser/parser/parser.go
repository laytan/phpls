package parser

import (
	"github.com/alecthomas/participle/v2"
	"github.com/laytan/elephp/pkg/phparser/ast"
	"github.com/laytan/elephp/pkg/phparser/lexer"
)

var Parser = participle.MustBuild[ast.Program](
	participle.Lexer(lexer.NewDef()),

	// Tokens that can go everywhere.
	participle.Elide("BlockComment", "LineComment", "PHPStart", "PHPEchoStart", "PHPEnd", "NonPHP"),

	participle.Union(ast.StatementImpls...),
	participle.Union(ast.ClassStatementImpls...),
	participle.Union(ast.CallableStatementImpls...),
	participle.Union(ast.ComplexStrContentImps...),
	participle.Union(ast.ValueImpls...),
	participle.Union(ast.OperationImpls...),

	participle.UseLookahead(participle.MaxLookahead),
)
