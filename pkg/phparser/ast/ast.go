package ast

import (
	plexer "github.com/alecthomas/participle/v2/lexer"
)

type Node interface {
	Position() plexer.Position
}

type BaseNode struct {
	Pos           plexer.Position
	LineComments  []plexer.Token `parser:"@LineComment*"`
	BlockComments []plexer.Token `parser:"@BlockComment*"`
}

func (b *BaseNode) Position() plexer.Position {
	return b.Pos
}

type Program struct {
	BaseNode
	NonPHP     []plexer.Token `parser:"@NonPHP*"`
	Statements []Statement    `parser:"@@*"`
}

type Extends struct {
	BaseNode
	Name Name `parser:"Extends @@"`
}

type Implements struct {
	BaseNode
	Names []Name `parser:"Implements @@ ( Comma @@ )*"`
}

type Name struct {
	Value string `parser:"@Ident"`
}
