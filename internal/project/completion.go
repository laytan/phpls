package project

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
)

const maxCompletionResults = 20

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(
	pos *position.Position,
) []*traversers.TrieNode {
	query := p.getCompletionQuery(pos)
	if query == "" {
		return nil
	}

	return Index().FindPrefix(query, maxCompletionResults)
}

// Gets the current word ([a-zA-Z0-9]*) that the position is at.
func (p *Project) getCompletionQuery(pos *position.Position) string {
	content, err := Wrkspc().ContentOf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf("Getting file content for completion query: %w", err).Error(),
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

// Returns the position for the namespace statement that matches the given position.
func (p *Project) Namespace(pos *position.Position) *position.Position {
	content, root, err := Wrkspc().AllOf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf(
				"Getting namespace, could not get content/nodes of %s: %w",
				pos.Path,
				err,
			),
		)
	}

	traverser := traversers.NewNamespace(pos.Row)
	root.Walk(traverser)

	if traverser.Result == nil {
		log.Println("Did not find namespace")
		return nil
	}

	row, col := position.PosToLoc(content, uint(traverser.Result.Position.StartPos))

	return &position.Position{
		Row:  row,
		Col:  col,
		Path: pos.Path,
	}
}

// Returns whether the file at given pos needs a use statement for the given fqn.
func (p *Project) NeedsUseStmtFor(pos *position.Position, fqn string) bool {
	root, err := Wrkspc().IROf(pos.Path)
	if err != nil {
		log.Println(
			fmt.Errorf(
				"Checking use statement needed, could not get nodes of %s: %w",
				pos.Path,
				err,
			),
		)
	}

	parts := strings.Split(fqn, `\`)
	className := parts[len(parts)-1]

	// Get how it would be resolved in the current file state.
	actFQN := common.FullyQualify(root, className)

	// If the resolvement in current state equals the wanted fqn, no use stmt is needed.
	return actFQN.String() != fqn
}
