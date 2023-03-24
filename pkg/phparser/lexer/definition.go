package lexer

import (
	"io"

	participle "github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

type Definition struct {
	symbols map[string]participle.TokenType
}

func NewDef() *Definition {
	ld := &Definition{
		symbols: map[string]participle.TokenType{},
	}

	for i := token.EOF; i < token.Count; i++ {
		ld.symbols[i.String()] = participle.TokenType(i)
	}

	return ld
}

var _ participle.Definition = &Definition{}

func (l *Definition) Symbols() map[string]participle.TokenType {
	return l.symbols
}

func (l *Definition) Lex(filename string, r io.Reader) (participle.Lexer, error) {
	return New(filename, r), nil
}
