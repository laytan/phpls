package lexer

import (
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

// isStrVarEnd determines if this is the end of the variable expression in a string.
// according to PHP docs, one layer of array access, or one layer of property access
// is allowed (in simplified form, braces can take any expression).
func (l *Lexer) isStrVarEnd() bool {
	// Braces take everything until '}'.
	if l.inStrVarBraced {
		return l.lastTokenType == lexer.TokenType(token.RBrace)
	}

	if l.lastTokenType == lexer.TokenType(token.RBracket) {
		return true
	}

	// Early return for the next statement if we are on a class fetch.
	if l.ch == '-' && l.peekIs('>') {
		return false
	}

	// If not inside of [], we are at the end, if the ch is not an identifier character.
	if l.ch != ']' && l.peekUntil('[', ']', '"') != ']' {
		return !isIdent(l.ch)
	}

	return false
}

func (l *Lexer) isBracedStrVarStart() bool {
	return l.ch == '{' && l.peekIs('$') && l.peekUntil('}', '"') == '}'
}

func (l *Lexer) isStrVarStart() bool {
	return l.ch == '$' && isIdent(l.peek())
}
