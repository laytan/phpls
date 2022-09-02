package server

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
)

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	start := time.Now()
	defer func() {
		log.Printf("Retrieving hover took %s\n", time.Since(start))
	}()

	pos := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	content, err := s.project.Hover(pos)
	if err != nil {
		log.Println(err)

		if !errors.Is(err, project.ErrNoDefinitionFound) {
			return nil, lsperrors.ErrRequestFailed(err.Error())
		}
	}

	if content == "" {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
	}, nil
}
