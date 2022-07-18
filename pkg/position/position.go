package position

import (
	"bufio"
	"strings"
)

func ToLocation(content string, pos uint) (row uint, col uint) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var line uint
	var curr uint
	for scanner.Scan() {
		line++

		lineLen := uint(len(scanner.Text())) + 1
		start := curr
		curr += lineLen

		if curr <= pos {
			continue
		}

		return line, (pos - start + 1)
	}

	return 0, 0
}

func FromLocation(content string, row uint, col uint) uint {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var line uint
	var curr uint
	for scanner.Scan() {
		line++

		if line < row {
			lineLen := uint(len(scanner.Text())) + 1
			curr += lineLen
			continue
		}

		return curr + col - 1
	}

	return 0
}
