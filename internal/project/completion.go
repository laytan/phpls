package project

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
)

const maxCompletionResults = 20

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(
	pos *position.Position,
) []*traversers.TrieNode {
	query := p.getCompletionQuery(pos)
	if len(query) == 0 {
		return nil
	}

	return p.index.FindPrefix(query, maxCompletionResults)
}

// Gets the current word ([a-zA-Z0-9]*) that the position is at.
func (p *Project) getCompletionQuery(pos *position.Position) string {
	content, err := p.wrksp.ContentOf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf("[ERROR] getting file content for completion query: %w", err).Error(),
		)
		return ""
	}

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
