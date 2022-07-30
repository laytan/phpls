package project

import (
	"errors"
	"fmt"

	"github.com/laytan/elephp/internal/traversers"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/symboltrie"
)

const maxCompletionResults = 10

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(pos *position.Position) ([]string, error) {
	file := p.GetFile(pos.Path)
	if file == nil {
		return nil, fmt.Errorf("Error retrieving file content for %s", pos.Path)
	}

	ast, err := file.parse(p.parserConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s for completion: %w", pos.Path, err)
	}

	apos := position.LocToPos(file.content, pos.Row, pos.Col)
	nap := traversers.NewNodeAtPos(apos)
	ast.Walk(nap)

	if len(nap.Nodes) == 0 {
		return nil, ErrNoCompletionResults
	}

	query := symbol.GetIdentifier(nap.Nodes[len(nap.Nodes)-1])
	results := p.symbolTrie.SearchPrefix(query, maxCompletionResults)
	return symboltrie.SearchResultKeys(results), nil
}
