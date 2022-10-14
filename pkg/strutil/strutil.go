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
			_, _ = b.WriteRune(ch) // This always returns a nil error.
		}
	}

	return b.String()
}

func Lines(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}
