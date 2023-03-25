package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/laytan/elephp/pkg/phparser/token"
)

// TODO: unicode might just be overhead, I don't think PHP supports this anyway, should see if byte-per-byte is faster than rune-per-rune (I think so).
type Lexer struct {
	name  string
	input *bufio.Reader

	cursor int
	line   int
	bol    int // The beginning of the line.

	ch            rune
	lastTokenType lexer.TokenType

	peeked []rune // Peeked characters, peeked[0] (if set) is always the character directly after ch.

	inPHP bool

	// Keeping track of complex strings requires these state variables.
	inStr          bool
	inStrVar       bool
	inStrVarBraced bool
}

var _ lexer.Lexer = &Lexer{}

// New creates a new lexer to read from input.
func New(name string, input io.Reader) *Lexer {
	l := &Lexer{name: name, input: bufio.NewReader(input)}
	l.read()
	l.cursor = 0
	return l
}

// NewStartInPHP returns a new lexer, which does not require an opening `<?php` tag.
func NewStartInPHP(name string, input io.Reader) *Lexer {
	l := New(name, input)
	l.inPHP = true
	return l
}

// Next parses the next token and returns it.
// returns tokenType token.EOF when there are no more tokens.
func (l *Lexer) Next() (t lexer.Token, err error) {
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
		t = lexer.Token{
			Pos: l.pos(),
		}

		switch {
		case l.peekSeqWithSpace('?', 'p', 'h', 'p'):
			l.inPHP = true
			l.readN(5)
			t.Type = lexer.TokenType(token.PHPStart)
			t.Value = "<?php"
		case l.peekSeqWithSpace('?', '='):
			l.inPHP = true
			l.readN(3)
			t.Type = lexer.TokenType(token.PHPEchoStart)
			t.Value = "<?="
		case l.ch != 0:
			t.Type = lexer.TokenType(token.NonPHP)
			t.Value = l.readNonPHP()
		}
		return
	}

	if l.inStrVar && l.isStrVarEnd() {
		l.inStrVar = false
		l.inStrVarBraced = false
	}

	l.skipWhitespace()
	t = lexer.Token{
		Type:  lexer.TokenType(token.Illegal),
		Value: string(l.ch),
		Pos:   l.pos(),
	}

	if l.ch == 0 {
		t.Value = ""
		t.Type = lexer.EOF
		return
	}

	if l.inStr {
		switch l.ch {
		case '"':
			t.Type = lexer.TokenType(token.StringEnd)
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
				_, _ = content.WriteString(read)

				found = true

				// If we got { or $ at the end of the string, add it as content.
				if l.peek() == '"' {
					_, _ = content.WriteRune(l.ch)
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

			t.Type = lexer.TokenType(token.StringContent)
			t.Value = content.String()
			return
		}
	}

	// If the type is still illegal, check these cases.
	if t.Type == lexer.TokenType(token.Illegal) {
		switch l.ch {
		case '{':
			t.Type = lexer.TokenType(token.LBrace)
		case '}':
			t.Type = lexer.TokenType(token.RBrace)
		case '(':
			t.Type = lexer.TokenType(token.LParen)
		case ')':
			t.Type = lexer.TokenType(token.RParen)
		case '[':
			t.Type = lexer.TokenType(token.LBracket)
		case ']':
			t.Type = lexer.TokenType(token.RBracket)
		case ',':
			t.Type = lexer.TokenType(token.Comma)
		case ';':
			t.Type = lexer.TokenType(token.Semicolon)
		case '+':
			t.Type = lexer.TokenType(token.Plus)
		case ':':
			t.Type = lexer.TokenType(token.Colon)
		case '*':
			t.Type = lexer.TokenType(token.Times)
		case '@':
			t.Type = lexer.TokenType(token.ErrorSuppress)
		case '|':
			if l.peekIs('|') {
				l.read()
				t.Value = "||"
				t.Type = lexer.TokenType(token.Or)
			} else {
				t.Type = lexer.TokenType(token.BinaryOr)
			}
		case '=':
			switch {
			case l.peekSeq('=', '='):
				l.readN(2)
				t.Type = lexer.TokenType(token.StrictEquals)
				t.Value = "==="
			case l.peekIs('='):
				l.read()
				t.Type = lexer.TokenType(token.Equals)
				t.Value = "=="
			case l.peekIs('>'):
				l.read()
				t.Type = lexer.TokenType(token.Arrow)
				t.Value = "=>"
			default:
				t.Type = lexer.TokenType(token.Assign)
			}
		case '&':
			if l.peekIs('&') {
				l.read()
				t.Value = "&&"
				t.Type = lexer.TokenType(token.And)
			} else {
				t.Type = lexer.TokenType(token.Reference)
			}
		case '!':
			switch {
			case l.peekSeq('=', '='):
				l.readN(2)
				t.Type = lexer.TokenType(token.StrictNotEquals)
				t.Value = "!=="
			case l.peekIs('='):
				l.read()
				t.Type = lexer.TokenType(token.NotEquals)
				t.Value = "!="
			default:
				t.Type = lexer.TokenType(token.Not)
			}
		case '/':
			switch {
			case l.peekIs('*'):
				t.Type = lexer.TokenType(token.BlockComment)
				t.Value = l.readBlockComment()
				return
			case l.peekIs('/'):
				t.Type = lexer.TokenType(token.LineComment)
				t.Value = l.readLineComment()
				return
			default:
				t.Type = lexer.TokenType(token.Divide)
			}
		case '\'':
			str, ok := l.readUntilIncl('\'')
			t.Value = str
			if !ok {
				return
			}

			t.Type = lexer.TokenType(token.SimpleString)
			return
		case '"':
			t.Type = lexer.TokenType(token.StringStart)
			l.inStr = true
		case '-':
			if l.peekIs('>') {
				l.read()
				t.Type = lexer.TokenType(token.ClassAccess)
				t.Value = "->"
			} else {
				t.Type = lexer.TokenType(token.Minus)
				t.Value = "-"
			}
		case '$':
			l.read() // $
			ident := l.readIdent()
			if ident == "" {
				return
			}

			t.Type = lexer.TokenType(token.Var)
			t.Value = "$" + ident
			return
		case '?':
			if l.peekIs('>') {
				l.read()
				l.inPHP = false
				t.Type = lexer.TokenType(token.PHPEnd)
				t.Value = "?>"
			} else {
				t.Type = lexer.TokenType(token.QuestionMark)
			}
		case '.':
			switch {
			case l.peekSeq(',', '.'):
				l.readN(2)
				t.Type = lexer.TokenType(token.Variadic)
				t.Value = "..."
			case l.peekIs('='):
				l.read()
				t.Value = ".="
				t.Type = lexer.TokenType(token.ConcatAssign)
			default:
				t.Type = lexer.TokenType(token.Concat)
			}
		default:
			if l.isNumberStart() {
				t.Value = l.readNumber()
				t.Type = lexer.TokenType(token.Number)
				return
			}

			if isIdent(l.ch) {
				t.Value = l.readIdent()
				t.Type = lexer.TokenType(token.LookupIdent(t.Value))
				return
			}
		}
	}

	l.read()
	return t, err
}

func (l *Lexer) pos() lexer.Position {
	return lexer.Position{
		Filename: l.name,
		Offset:   l.cursor,
		Line:     l.line,
		Column:   l.cursor - l.bol,
	}
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.read()
	}
}

// readNonPHP rads everything until either `<?=`, `<?php` or EOF.
func (l *Lexer) readNonPHP() string {
	res := strings.Builder{}
	for l.ch != 0 {
		_, _ = res.WriteRune(l.ch)
		l.read()

		if l.ch == '<' {
			if l.peekSeqWithSpace('?', '=') || l.peekSeqWithSpace('?', 'p', 'h', 'p') {
				return res.String()
			}
		}
	}

	return res.String()
}

func (l *Lexer) readBlockComment() string {
	res := strings.Builder{}
	_, _ = res.WriteRune(l.ch) // /
	l.read()
	_, _ = res.WriteRune(l.ch) // *
	l.read()

	for l.ch != 0 {
		_, _ = res.WriteRune(l.ch)
		l.read()

		if l.ch == '*' && l.peekIs('/') {
			_, _ = res.WriteRune(l.ch) // *
			l.read()
			_, _ = res.WriteRune(l.ch) // /
			l.read()
			break
		}
	}

	return res.String()
}

func (l *Lexer) readLineComment() string {
	res := strings.Builder{}
	_, _ = res.WriteRune(l.ch) // /
	l.read()
	_, _ = res.WriteRune(l.ch) // /
	l.read()

	for l.ch != 0 {
		_, _ = res.WriteRune(l.ch)
		l.read()

		if l.ch == '\n' {
			break
		} else if l.ch == '?' && l.peekIs('>') {
			break
		}
	}

	return res.String()
}
