package lexer

import "strings"

// read reads the next rune into l.ch.
// if the peeked buffer is non-empty it consumes from there first.
func (l *Lexer) read() {
	if len(l.peeked) > 0 {
		l.ch = l.peeked[0]
		l.peeked = l.peeked[1:]
		l.cursor++
		return
	}

	ch, _, _ := l.input.ReadRune()
	l.ch = ch
	l.cursor++
}

func (l *Lexer) readN(n int) {
    for i := 0; i < n; i++ {
        l.read()
    }
}

// readUntil reads until any of the check runes is found (excluding this rune).
func (l *Lexer) readUntil(check ...rune) string {
	res := strings.Builder{}
	res.WriteRune(l.ch)
	l.read()

	escaping := false
	for l.ch != 0 {
		if !escaping {
			for _, c := range check {
				if l.ch == c {
					return res.String()
				}
			}
		}

		escaping = !escaping && l.ch == '\\'
		res.WriteRune(l.ch)
		l.read()
	}

	return res.String()
}

// readUntilIncl does readUntil including the check rune.
func (l *Lexer) readUntilIncl(check rune) (string, bool) {
	res := l.readUntil(check)
	if l.ch != check {
		return res, false
	}

	res += string(l.ch)
	l.read()
	return res, true
}
