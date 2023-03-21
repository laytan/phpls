package lexer

import "unicode"

// peek returns the next rune without consuming it or increasing the cursor.
func (l *Lexer) peek() rune {
	if len(l.peeked) > 0 {
		return l.peeked[0]
	}

	ch, _, _ := l.input.ReadRune()
	l.peeked = append(l.peeked, ch)
	return ch
}

// peekIs checks if the next rune is the given rune.
func (l *Lexer) peekIs(check rune) bool {
	return l.peek() == check
}

// peekUntil keeps peeking until the rune matches any of the given runes,
// returns the matched rune.
func (l *Lexer) peekUntil(check ...rune) rune {
	for _, p := range l.peeked {
		for _, c := range check {
			if p == c {
				return p
			}
		}
	}

	ch := l.ch
	escaping := true
	for ch != 0 {
		if !escaping {
			for _, c := range check {
				if ch == c {
					return ch
				}
			}
		}

		escaping = !escaping && ch == '\\'
		ch, _, _ = l.input.ReadRune()
		l.peeked = append(l.peeked, ch)
	}

	return 0
}

// peekSeq peeks len(seq) and checks if the sequence matches what was peeked.
func (l *Lexer) peekSeq(seq ...rune) bool {
	// Fill l.peeked to at least the length of seq.
	for len(seq) > len(l.peeked) {
		ch, _, _ := l.input.ReadRune()
		if ch == 0 {
			return false
		}

		l.peeked = append(l.peeked, ch)
	}

	i := 0
	for ; i < len(seq); i++ {
		if l.peeked[i] != seq[i] {
			return false
		}
	}

	return i == len(seq)
}

// peekSeqWithSpace does peekSeq, but checks that the next rune is whitespace.
func (l *Lexer) peekSeqWithSpace(seq ...rune) bool {
	if !l.peekSeq(seq...) {
		return false
	}

	// If we don't have the next token, peek it.
	if len(l.peeked) <= len(seq) {
		ch, _, _ := l.input.ReadRune()
		l.peeked = append(l.peeked, ch)
	}

	return unicode.IsSpace(l.peeked[len(seq)])
}
