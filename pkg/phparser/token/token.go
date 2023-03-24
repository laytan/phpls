package token

import (
	"github.com/alecthomas/participle/v2/lexer"
)

//go:generate stringer -type TokenType
//go:generate go run lookup_ident_gen.go
type Type lexer.TokenType

const (
	EOF          = Type(lexer.EOF)
	Illegal Type = iota // Anything not able to be lexed.

	PHPStart     // `<?php`
	PHPEnd       // `?>`
	PHPEchoStart // `<?=`

	NonPHP // Any characters not in `PHPStart`|`PHPEchoStart` and `PHPEnd`.

	Ident // An identifier like a function name.

	KeywordsStart
	Function
	Return
	Abstract
	And
	As
	Break
	Callable
	Case
	Catch
	Class
	Clone
	Const
	Continue
	Default
	Die
	Exit
	Do
	Echo
	Else
	ElseIf // literal:"elseif"
	EndDeclare
	EndFor
	EndForEach
	EndIf
	EndSwitch
	EndWhile
	Extends
	Final
	Finally
	Fn
	For
	ForEach
	Global
	GoTo
	If
	Implements
	Include
	IncludeOnce // literal:"include_once"
	InstanceOf
	InsteadOf
	Interface
	KVar // The `var` keyword.
	Match
	Namespace
	New
	Or
	Print
	Private
	Protected
	Public
	Readonly
	Require
	RequireOnce // literal:"require_once"
	Static
	Switch
	Throw
	Trait
	Try
	Use
	While
	XOR
	Yield
	YieldFrom // literal:"yield from" // YieldFrom has a special case in the lexer (because it has a space).
	True      // TODO: this could be 'TRUE' or 'true', should think about case sensitivity in general.
	False
	KeywordsEnd

	Reference // An `&` ch.
	Variadic  // A `...` sequence.

	Var // A `$var` variable.

	Number // Any number value, not necessarily a legal/valid number.

	Not
	ErrorSuppress

	QuestionMark

	Concat

	Assign
	ConcatAssign

	Plus
	Minus
	Divide
	Times
	Comma
	Colon
	Semicolon
	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket

	BinaryOr // `|`.

	ClassAccess // `->`.

	LineComment // A `// comment`.
	// TODO: parse further (@var etc.).
	BlockComment // A `/* comment */`.

	SimpleString // A string using single quotes.

	StringStart   // A double quote to start a complex string.
	StringContent // Literal string text inside the complex string.
	StringEnd     // A double quote ending a complex string.

	Equals          // ==
	StrictEquals    // ===
	NotEquals       // !=
	StrictNotEquals // !==

	Count // Marks the end of the tokens, never lexed.
)
