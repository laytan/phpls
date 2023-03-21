package lexer

import "strings"

func (l *Lexer) isNumberStart() bool {
	c := l.ch
	if c == '-' || c == '+' {
		c = l.peek()
	}

	return c >= '0' && c <= '9'
}

// isNumber checks if the character can be INSIDE a number, the start of a number has different (stricter) rules, see Lexer.isNumberStart().
func isNumber(r rune) bool {
	return r == '_' || r == 'x' || r == 'X' || r == 'o' || r == 'O' || r == 'b' || r == 'B' ||
		(r >= 'A' && r <= 'F') ||
		(r >= 'a' && r <= 'f') ||
		(r >= '0' && r <= '9')
}

// readNumber keeps reading until a non-number character is found.
// This does not mean the number is valid, this would return oxZ for example which is not valid hex (but is valid format of hex).
func (l *Lexer) readNumber() string {
	res := strings.Builder{}
	// This is the first character, checked before calling this function.
	res.WriteRune(l.ch)
	l.read()

	for isNumber(l.ch) {
		res.WriteRune(l.ch)
		l.read()
	}

	return res.String()
}
