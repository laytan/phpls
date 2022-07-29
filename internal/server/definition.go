package server

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	log "github.com/sirupsen/logrus"
)

func (s *Server) Definition(
	ctx context.Context,
	params *protocol.DefinitionParams,
) (protocol.Definition, error) {
	start := time.Now()
	defer func() { log.Infof("Retrieving definition took %s\n", time.Since(start)) }()

	path := strings.TrimPrefix(string(params.TextDocument.URI), "file://")

	pos, err := s.project.Definition(
		path,
		&project.Position{
			Row: uint(params.Position.Line + 1),
			Col: uint(params.Position.Character + 1),
		},
	)
	if err != nil {
		if errors.Is(err, project.ErrNoDefinitionFound) {
			log.Warn(err)
			return nil, nil
		}

		log.Error(err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
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
