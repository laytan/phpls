package project

import (
	"bufio"
	"errors"
	"strings"
	"unicode"

	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
)

const maxCompletionResults = 20

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(
	pos *position.Position,
) []*index.INode {
	query := p.getCompletionQuery(pos)
	if query == "" {
		return nil
	}

	return index.Current.FindPrefix(query, maxCompletionResults)
}

// Gets the current word ([a-zA-Z0-9]*) that the position is at.
func (p *Project) getCompletionQuery(pos *position.Position) string {
	content := wrkspc.Current.FContentOf(pos.Path)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for i := 0; scanner.Scan(); i++ {
		// The target line:
		if uint(i) != pos.Row-1 {
			continue
		}

		content := scanner.Text()
		rContent := []rune(content)
		if len(rContent) == 0 {
			break
		}

		start := uint(0)
		end := uint(len(rContent))
		startI := pos.Col - 2
		if startI >= end {
			startI = end - 1
		}

		for i := startI; i > 0; i-- {
			ch := rContent[i]

			if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
				continue
			}

			start = i + 1
			break
		}

		for i := startI; int(i) < len(rContent); i++ {
			ch := rContent[i]

			if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
				continue
			}

			end = i
			break
		}

		if start >= end {
			return ""
		}

		return strings.TrimSpace(string(rContent[start:end]))
	}

	return ""
}
