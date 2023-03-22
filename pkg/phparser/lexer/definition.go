package lexer

import (
	"io"

	participle "github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

type LexerDef struct {
	symbols map[string]participle.TokenType
}

func NewDef() *LexerDef {
	ld := &LexerDef{
		symbols: map[string]participle.TokenType{},
	}

	for i := token.EOF; i < token.Count; i++ {
		ld.symbols[i.String()] = participle.TokenType(i)
	}

	return ld
}

var _ participle.Definition = &LexerDef{}

func (l *LexerDef) Symbols() map[string]participle.TokenType {
	return l.symbols
}

func (l *LexerDef) Lex(filename string, r io.Reader) (participle.Lexer, error) {
	return New(filename, r), nil
}
