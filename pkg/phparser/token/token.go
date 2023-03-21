package token

//go:generate stringer -type TokenType
//go:generate go run lookup_ident_gen.go
type TokenType uint8

const (
	Illegal TokenType = iota // Anything not able to be lexed.
	EOF                      // The end.

    PHPStart // `<?php`
    PHPEnd // `?>`
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
    KeywordsEnd

	Var // A `$var` variable.

	Number // Any number value, not necessarily a legal/valid number.

	Assign
	Plus
    Minus
	Comma
	Semicolon
	LParen
	RParen
	LBrace
	RBrace
    LBracket
    RBracket

    ClassAccess // `->`.

	LineComment // A `// comment`.

	SimpleString // A string using single quotes.

	StringStart   // A double quote to start a complex string.
	StringContent // Literal string text inside the complex string.
	StringEnd     // A double quote ending a complex string.
)

type Token struct {
	Type    TokenType
	Literal string

	Row uint
	Col uint
}

