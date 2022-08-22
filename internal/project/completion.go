package project

import (
	"errors"
	"strings"

	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symboltrie"
)

const maxCompletionResults = 25

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(pos *position.Position) ([]string, error) {
	line := p.GetLine(pos)
	fields := strings.Fields(line)
	field := fields[len(fields)-1]

	results := p.symbolTrie.SearchPrefix(field, maxCompletionResults)
	return symboltrie.SearchResultKeys(results), nil
}
