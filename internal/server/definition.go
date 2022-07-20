package server

import (
	"context"
	"strings"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
)

func (s *server) Definition(
	ctx context.Context,
	params *protocol.DefinitionParams,
) (protocol.Definition, error) {
	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")

	pos, err := s.project.Definition(
		path,
		&project.Position{
			Row: uint(params.Position.Line + 1),
			Col: uint(params.Position.Character + 1),
		},
	)
	if err != nil {
		// TODO: should not return general errors, log and return an lsp error
		return nil, err
	}

	uri := params.TextDocument.URI
	if pos.Path != "" {
		uri = protocol.DocumentURI("file://" + pos.Path)
	}

	// TODO: Create helpers for creating this from a position.
	return []protocol.Location{{
		URI: uri,
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(pos.Row) - 1,
				Character: uint32(pos.Col) - 1,
			},
			End: protocol.Position{
				Line:      uint32(pos.Row) - 1,
				Character: uint32(pos.Col) - 1,
			},
		},
	}}, nil
}
