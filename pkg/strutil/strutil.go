package strutil

import (
	"strings"
	"unicode"
)

// Fast way of removing all whitespace from a string, credit: https://stackoverflow.com/a/32081891.
func RemoveWhitespace(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, ch := range text {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}

	return b.String()
}
