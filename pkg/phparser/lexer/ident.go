package lexer

import "strings"

func isIdent(r rune) bool {
	return r >= '0' && r <= '9' || isIdentStart(r)
}

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 0x7f && r <= 0xff) ||
		r == '_' ||
		r == '\\'
}

// readIdent keeps reading until a non-identifier character is found.
func (l *Lexer) readIdent() string {
	res := strings.Builder{}
	for {
		_, _ = res.WriteRune(l.ch)
		l.read()

		if !isIdent(l.ch) {
			// Special case for the 'yield from' keyword, which is the only keyword in PHP with a space.
			if l.ch == ' ' && res.String() == "yield" {
				if l.peekSeqWithSpace('f', 'r', 'o', 'm') {
					continue
				}
			}

			break
		}
	}

	return res.String()
}
