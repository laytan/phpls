package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/laytan/elephp/pkg/phparser/token"
)

type Lexer struct {
	input *bufio.Reader

	cursor uint
	line   uint
	bol    uint // The beginning of the line.

	ch            rune
	lastTokenType token.TokenType

	peeked []rune // Peeked characters, peeked[0] (if set) is always the character directly after ch.

	inPHP bool

	// Keeping track of complex strings requires these state variables.
	inStr          bool
	inStrVar       bool
	inStrVarBraced bool
}

// New creates a new lexer to read from input.
func New(input io.Reader) *Lexer {
	l := &Lexer{input: bufio.NewReader(input)}
	l.read()
	l.cursor = 0
	return l
}

// NewStartInPHP returns a new lexer, which does not require an opening `<?php` tag.
func NewStartInPHP(input io.Reader) *Lexer {
	l := New(input)
	l.inPHP = true
	return l
}

// Next parses the next token and returns it.
// returns tokenType token.EOF when there are no more tokens.
func (l *Lexer) Next() (t token.Token) {
	defer func() {
		l.lastTokenType = t.Type
	}()

	// what.Happens("last token: %s", l.lastTokenType)
	// what.Happens("ch: %s", string(l.ch))
	// what.Happens(
	// 	"in str, in strvar, in strvarbraced: %v, %v, %v",
	// 	l.inStr,
	// 	l.inStrVar,
	// 	l.inStrVarBraced,
	// )

	if !l.inPHP {
		if l.peekSeqWithSpace('?', 'p', 'h', 'p') {
			l.inPHP = true
			l.readN(5)
			t.Type = token.PHPStart
			t.Literal = "<?php"
			return t
		} else if l.peekSeqWithSpace('?', '=') {
			l.inPHP = true
			l.readN(3)
			t.Type = token.PHPEchoStart
			t.Literal = "<?="
			return t
		} else if l.ch != 0 {
			t = token.Token{
				Row: l.line,
				Col: l.cursor - l.bol,
			}
            t.Type = token.NonPHP
			t.Literal = l.readNonPHP()
			return t
		}
	}

	if l.inStrVar && l.isStrVarEnd() {
		l.inStrVar = false
		l.inStrVarBraced = false
	}

	l.skipWhitespace()
	t = token.Token{
		Literal: string(l.ch),
		Row:     l.line,
		Col:     l.cursor - l.bol,
	}

	if l.ch == 0 {
		t.Literal = ""
		t.Type = token.EOF
		return t
	}

	if l.inStr {
		switch l.ch {
		case '"':
			t.Type = token.StringEnd
			l.inStr = false
			l.inStrVar = false
			l.inStrVarBraced = false
		default:
			if !l.inStrVar {
				if l.isStrVarStart() {
					l.inStrVar = true
					l.inStrVarBraced = false
				}

				if l.isBracedStrVarStart() {
					l.inStrVar = true
					l.inStrVarBraced = true
				}
			}

			if l.inStrVar {
				break
			}

			content := strings.Builder{}
			for found := false; !found; {
				read := l.readUntil('{', '$', '"')
				content.WriteString(read)

				found = true

				// If we got { or $ at the end of the string, add it as content.
				if l.peek() == '"' {
					content.WriteRune(l.ch)
					l.read()
					break
				}

				if l.ch == '{' {
					found = l.isBracedStrVarStart()
				}

				if l.ch == '$' {
					found = l.isStrVarStart()
				}
			}

			t.Type = token.StringContent
			t.Literal = content.String()
			return t
		}
	}

	// If the type is still illegal, check these cases.
	if t.Type == token.Illegal {
		switch l.ch {
		case '{':
			t.Type = token.LBrace
		case '}':
			t.Type = token.RBrace
		case '(':
			t.Type = token.LParen
		case ')':
			t.Type = token.RParen
		case '[':
			t.Type = token.LBracket
		case ']':
			t.Type = token.RBracket
		case ',':
			t.Type = token.Comma
		case '=':
			t.Type = token.Assign
		case ';':
			t.Type = token.Semicolon
		case '+':
			t.Type = token.Plus
		case '/':
			if !l.peekIs('/') {
				break
			}

			t.Type = token.LineComment
			t.Literal = l.readUntil('\n')
			return t
		case '\'':
			str, ok := l.readUntilIncl('\'')
			t.Literal = str
			if !ok {
				return t
			}

			t.Type = token.SimpleString
			return t
		case '"':
			t.Type = token.StringStart
			l.inStr = true
		case '-':
			if l.peekIs('>') {
				l.read()
				t.Type = token.ClassAccess
				t.Literal = "->"
			} else {
				t.Type = token.Minus
				t.Literal = "-"
			}
		case '$':
			l.read() // consume $
			ident := l.readIdent()
			if len(ident) == 0 {
				return t
			}

			t.Type = token.Var
			t.Literal = "$" + ident
			return t
		default:
			if l.isNumberStart() {
				t.Literal = l.readNumber()
				t.Type = token.Number
				return t
			}

			if isIdent(l.ch) {
				t.Literal = l.readIdent()
				t.Type = token.LookupIdent(t.Literal)
				return t
			}

			if l.ch == '?' && l.peekIs('>') {
				l.read()
				l.inPHP = false
				t.Type = token.PHPEnd
				t.Literal = "?>"
			}
		}
	}

	l.read()
	return t
}

// skipWhitespace reads until a non-whitespace character is found,
// incrementing l.line if a line break is encountered.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		if l.ch == '\n' {
			l.read()
			l.bol = l.cursor
			l.line++
			continue
		}

		l.read()
	}
}

// readNonPHP rads everything until either `<?=`, `<?php` or EOF.
func (l *Lexer) readNonPHP() string {
	res := strings.Builder{}
	for l.ch != 0 {
		res.WriteRune(l.ch)
		l.read()

		if l.ch == '<' {
			if l.peekSeqWithSpace('?', '=') || l.peekSeqWithSpace('?', 'p', 'h', 'p') {
				return res.String()
			}
		}
	}

	return res.String()
}
